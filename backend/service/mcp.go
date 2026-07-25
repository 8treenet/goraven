package service

import (
	"context"
	"encoding/json"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/core/sandbox"
	"goraven/core/tools"
	"goraven/util"
	"math/rand"
	"regexp"
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *McpService {
			return &McpService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *McpService) {
			initiator.FetchService(ctx, &service)
			return
		})
		initiator.BindBooting(func(bootManager freedom.BootManager) {
			freedom.ServiceLocator().Call(func(service *McpService) error {
				service.Worker.DeferRecycle()
				service.VerifierTimer()
				return nil
			})
		})
	})
}

type McpService struct {
	Worker      freedom.Worker
	UserRepo    *repository.UserRepository
	McpRepo     *repository.MCPRepository
	PersonaRepo *repository.PersonaRepository
}

func (service *McpService) VerifierTimer() {
	if !config.Get().System.Initialized {
		return
	}
	go func() {
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Second)
		for {
			service.verifyAllMCPs()
			time.Sleep(3 * time.Hour)
		}
	}()
}

func (service *McpService) verifyAllMCPs() {
	ctx := context.Background()
	superAdmin, err := service.UserRepo.FindSuperAdmin()
	if err != nil {
		return
	}

	endpoints, err := service.McpRepo.FindEnabledMCPEndpoints()
	if err != nil {
		return
	}

	box, boxerr := sandbox.NewSandbox(superAdmin.Username)
	if boxerr != nil {
		return
	}
	now := time.Now()

	for _, endpoint := range endpoints {
		mcpObj := service.BuildMCPObjectFromEndpoint(&endpoint)

		start := time.Now()
		_, validateErr := tools.ValidateMCP(ctx, mcpObj, box)
		latencyMs := int(time.Since(start).Milliseconds())

		if validateErr != nil {
			_ = service.McpRepo.UpdateMCPEndpointStatus(endpoint.McpId, po.MCPEndpointStatusDisabled)
			_ = service.PersonaRepo.DeletePersonaToolByMcpId(endpoint.McpId)
		}
		_ = service.McpRepo.UpdateMCPEndpointHealth(endpoint.McpId, latencyMs, now)
	}
}

func (service *McpService) CheckAllMCPHealth() {
	go service.verifyAllMCPs()
}

func (service *McpService) ListEnabledMCPs() ([]vo.UserMCPItem, error) {
	endpoints, err := service.McpRepo.FindUserSelectableMCPEndpoints()
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserMCPItem, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, vo.UserMCPItem{
			McpId:       ep.McpId,
			Name:        ep.Name,
			DisplayName: ep.DisplayName,
			Icon:        ep.Icon,
			Description: ep.Description,
		})
	}
	return items, nil
}

func (service *McpService) ListEnabledMCPsByIDs(mcpIds []int) ([]vo.UserMCPItem, error) {
	endpoints, err := service.McpRepo.FindEnabledMCPEndpointsByIDs(mcpIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.UserMCPItem, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, vo.UserMCPItem{
			McpId:       ep.McpId,
			Name:        ep.Name,
			DisplayName: ep.DisplayName,
			Icon:        ep.Icon,
			Description: ep.Description,
		})
	}
	return items, nil
}

func (service *McpService) ListMCPEndpoints(req *vo.AdminMCPListReq) (*infra.PageResponse, error) {
	items, pr, err := service.McpRepo.PaginateMCPEndpoints(req)
	if err != nil {
		return nil, err
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

func (service *McpService) GetMCPEndpointDetail(mcpId int) (*vo.AdminMCPDetailRsp, error) {
	endpoint, err := service.McpRepo.GetMCPEndpointByID(mcpId)
	if err != nil {
		return nil, errs.ErrMCPNotFound
	}
	return &vo.AdminMCPDetailRsp{
		McpId:           endpoint.McpId,
		Name:            endpoint.Name,
		DisplayName:     endpoint.DisplayName,
		Icon:            endpoint.Icon,
		Description:     endpoint.Description,
		Transport:       endpoint.Transport,
		HttpUrl:         endpoint.HttpUrl,
		HttpHeader:      endpoint.HttpHeader,
		HttpProxyUrl:    endpoint.HttpProxyURL,
		StdioType:       endpoint.StdioType,
		StdioEnv:        endpoint.StdioEnv,
		StdioArgs:       endpoint.StdioArgs,
		Status:          endpoint.Status,
		AlwaysOn:        endpoint.AlwaysOn,
		HealthLatency:   endpoint.HealthLatency,
		HealthCheckedAt: endpoint.HealthCheckedAt,
		Remark:          endpoint.Remark,
		Created:         endpoint.Created,
		Updated:         endpoint.Updated,
	}, nil
}

func (service *McpService) CreateMCPEndpoint(req *vo.AdminCreateMCPReq) error {
	if err := service.validateCreateReq(req); err != nil {
		return err
	}

	if _, err := service.McpRepo.FindMCPEndpointByName(req.Name); err == nil {
		return errs.ErrMCPNameAlreadyExists
	}

	httpHeaderJSON, err := util.MapToJSON(req.HttpHeader)
	if err != nil {
		return errs.NewFormatError("httpHeader JSON error: %v", "httpHeader JSON 格式错误: %v", err)
	}
	stdioEnvJSON, err := util.MapToEnvJSON(req.StdioEnv)
	if err != nil {
		return errs.NewFormatError("stdioEnv JSON error: %v", "stdioEnv JSON 格式错误: %v", err)
	}
	stdioArgsJSON, err := util.SliceToJSON(req.StdioArgs)
	if err != nil {
		return errs.NewFormatError("stdioArgs JSON error: %v", "stdioArgs JSON 格式错误: %v", err)
	}

	mcpObj := service.buildMCPObject(req.Transport, req.HttpUrl, httpHeaderJSON, req.HttpProxyUrl, req.StdioType, stdioEnvJSON, stdioArgsJSON)

	startCheckTime := time.Now()
	if err := service.testMCPConnection(mcpObj); err != nil {
		return err
	}
	latencyMs := int(time.Since(startCheckTime).Milliseconds())

	endpoint := &po.MCPEndpoint{
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Icon:            req.Icon,
		Description:     req.Description,
		Transport:       req.Transport,
		HttpUrl:         req.HttpUrl,
		HttpHeader:      httpHeaderJSON,
		HttpProxyURL:    req.HttpProxyUrl,
		StdioType:       req.StdioType,
		StdioEnv:        stdioEnvJSON,
		StdioArgs:       stdioArgsJSON,
		Status:          po.MCPEndpointStatusEnabled,
		Remark:          req.Remark,
		HealthCheckedAt: &startCheckTime,
		HealthLatency:   latencyMs,
	}

	return service.McpRepo.CreateMCPEndpoint(endpoint)
}

func (service *McpService) UpdateMCPEndpoint(mcpId int, req *vo.AdminUpdateMCPReq) error {
	existing, err := service.McpRepo.GetMCPEndpointByID(mcpId)
	if err != nil {
		return errs.ErrMCPNotFound
	}

	updates := map[string]interface{}{}
	needTest := false

	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.HttpUrl != "" {
		updates["http_url"] = req.HttpUrl
		if req.HttpUrl != existing.HttpUrl {
			needTest = true
		}
	}
	if req.HttpHeader != nil {
		httpHeaderJSON, err := util.MapToJSON(req.HttpHeader)
		if err != nil {
			return errs.NewFormatError("httpHeader JSON error: %v", "httpHeader JSON 格式错误: %v", err)
		}
		updates["http_header"] = httpHeaderJSON
		needTest = true
	}
	if req.HttpProxyUrl != "" {
		updates["proxy_url"] = req.HttpProxyUrl
		if req.HttpProxyUrl != existing.HttpProxyURL {
			needTest = true
		}
	}
	if req.StdioType != "" {
		updates["stdio_type"] = req.StdioType
		if req.StdioType != existing.StdioType {
			needTest = true
		}
	}
	if req.StdioEnv != nil {
		stdioEnvJSON, err := util.MapToEnvJSON(req.StdioEnv)
		if err != nil {
			return errs.NewFormatError("stdioEnv JSON error: %v", "stdioEnv JSON 格式错误: %v", err)
		}
		updates["stdio_env"] = stdioEnvJSON
		needTest = true
	}
	if req.StdioArgs != nil {
		stdioArgsJSON, err := util.SliceToJSON(req.StdioArgs)
		if err != nil {
			return errs.NewFormatError("stdioArgs JSON error: %v", "stdioArgs JSON 格式错误: %v", err)
		}
		updates["stdio_args"] = stdioArgsJSON
		if stdioArgsJSON != existing.StdioArgs {
			needTest = true
		}
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Status != nil {
		updates["status"] = int(*req.Status)
		if *req.Status == po.MCPEndpointStatusEnabled && existing.Status == po.MCPEndpointStatusDisabled {
			needTest = true
		}
	}

	if len(updates) == 0 {
		return nil
	}

	if needTest {
		httpUrl := existing.HttpUrl
		if v, ok := updates["http_url"]; ok {
			httpUrl = v.(string)
		}
		httpHeader := existing.HttpHeader
		if v, ok := updates["http_header"]; ok {
			httpHeader = v.(string)
		}
		proxyUrl := existing.HttpProxyURL
		if v, ok := updates["proxy_url"]; ok {
			proxyUrl = v.(string)
		}
		stdioType := existing.StdioType
		if v, ok := updates["stdio_type"]; ok {
			stdioType = v.(string)
		}
		stdioEnv := existing.StdioEnv
		if v, ok := updates["stdio_env"]; ok {
			stdioEnv = v.(string)
		}
		stdioArgs := existing.StdioArgs
		if v, ok := updates["stdio_args"]; ok {
			stdioArgs = v.(string)
		}

		mcpObj := service.buildMCPObject(existing.Transport, httpUrl, httpHeader, proxyUrl, stdioType, stdioEnv, stdioArgs)
		startCheckTime := time.Now()
		if err := service.testMCPConnection(mcpObj); err != nil {
			return err
		}
		latencyMs := int(time.Since(startCheckTime).Milliseconds())
		updates["health_latency"] = latencyMs
		updates["health_checked_at"] = startCheckTime
	}

	return service.McpRepo.UpdateMCPEndpoint(mcpId, updates)
}

func (service *McpService) DeleteMCPEndpoint(mcpId int) error {
	if _, err := service.McpRepo.GetMCPEndpointByID(mcpId); err != nil {
		return errs.ErrMCPNotFound
	}

	_ = service.PersonaRepo.DeletePersonaToolByMcpId(mcpId)

	return service.McpRepo.SoftDeleteMCPEndpoint(mcpId)
}

func (service *McpService) UpdateMCPEndpointStatus(mcpId int, status uint8) error {
	endpoint, err := service.McpRepo.GetMCPEndpointByID(mcpId)
	if err != nil {
		return errs.ErrMCPNotFound
	}

	if status == po.MCPEndpointStatusEnabled {
		mcpObj := service.BuildMCPObjectFromEndpoint(endpoint)
		if err := service.testMCPConnection(mcpObj); err != nil {
			return err
		}
	}

	if err := service.McpRepo.UpdateMCPEndpoint(mcpId, map[string]interface{}{"status": status}); err != nil {
		return err
	}

	if status == po.MCPEndpointStatusDisabled {
		_ = service.PersonaRepo.DeletePersonaToolByMcpId(mcpId)
	}
	return nil
}

func (service *McpService) ToggleMCPAlwaysOn(mcpId int, req *vo.AdminMCPToggleAlwaysOnReq) error {
	if _, err := service.McpRepo.GetMCPEndpointByID(mcpId); err != nil {
		return errs.ErrMCPNotFound
	}

	return service.McpRepo.UpdateMCPEndpoint(mcpId, map[string]interface{}{
		"always_on": req.AlwaysOn,
	})
}

func (service *McpService) GetRecommendMCPs() []vo.MCPRecommendItem {
	activeEndpoints, _ := service.McpRepo.FindAllActiveMCPEndpoints()
	installedMap := make(map[string]*po.MCPEndpoint)
	for i := range activeEndpoints {
		installedMap[activeEndpoints[i].Name] = &activeEndpoints[i]
	}

	items, _ := service.McpRepo.GetRecommendMCPEndpoints()
	reuslt := []vo.MCPRecommendItem{}
	for _, item := range items {
		item := vo.MCPRecommendItem{
			Name:        item.Name,
			DisplayName: item.DisplayName,
			Icon:        item.Icon,
			Description: item.Description,
			Transport:   item.Transport,
			HttpUrl:     item.HttpUrl,
			HttpHeader:  item.HttpHeader,
			StdioType:   item.StdioType,
			StdioArgs:   item.StdioArgs,
			StdioEnv:    item.StdioEnv,
		}
		if ep, ok := installedMap[item.Name]; ok {
			item.Installed = true
			item.McpId = ep.McpId
			item.McpStatus = ep.Status
		}
		reuslt = append(reuslt, item)
	}
	return reuslt
}

func (service *McpService) validateCreateReq(req *vo.AdminCreateMCPReq) error {
	if req.Name == "" {
		return errs.ErrMCPNameRequired
	}

	var mcpNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]{1,63}$`)
	if !mcpNamePattern.MatchString(req.Name) {
		return errs.ErrMCPNameInvalid
	}
	if req.DisplayName == "" {
		return errs.ErrMCPDisplayNameRequired
	}
	if req.Transport == "" {
		return errs.ErrMCPTransportRequired
	}
	if req.Transport != tools.MCPTransportStdio && req.Transport != tools.MCPTransportSSE && req.Transport != tools.MCPTransportHTTP {
		return errs.ErrMCPTransportRequired
	}
	if req.Transport == tools.MCPTransportSSE || req.Transport == tools.MCPTransportHTTP {
		if req.HttpUrl == "" {
			return errs.ErrMCPHttpURLRequired
		}
	}
	if req.Transport == tools.MCPTransportStdio {
		if len(req.StdioArgs) == 0 {
			return errs.ErrMCPStdioArgsRequired
		}
	}
	return nil
}

func (service *McpService) testMCPConnection(mcpObj tools.MCP) error {
	ctx := context.Background()
	superAdmin, err := service.UserRepo.FindSuperAdmin()
	if err != nil {
		return errs.NewFormatError("failed to get super admin: %v", "获取超级管理员失败: %v", err)
	}
	box, boxerr := sandbox.NewSandbox(superAdmin.Username)
	if boxerr != nil {
		return boxerr
	}

	_, validateErr := tools.ValidateMCP(ctx, mcpObj, box)
	if validateErr != nil {
		return errs.NewFormatError("MCP connection test failed: %v", "MCP 连接测试失败: %v", validateErr)
	}
	return nil
}

func (service *McpService) CheckMCPToolNameConflicts(mcpIds []int) error {
	if len(mcpIds) <= 1 {
		return nil
	}

	endpoints, err := service.McpRepo.GetMCPEndpointsByIDs(mcpIds)
	if err != nil {
		return err
	}

	superAdmin, err := service.UserRepo.FindSuperAdmin()
	if err != nil {
		return errs.NewFormatError("failed to get super admin: %v", "获取超级管理员失败: %v", err)
	}
	box, boxerr := sandbox.NewSandbox(superAdmin.Username)
	if boxerr != nil {
		return boxerr
	}

	toolOwnerMap := make(map[string]string)
	ctx := context.Background()

	for _, ep := range endpoints {
		mcpObj := service.BuildMCPObjectFromEndpoint(&ep)
		toolNames, validateErr := tools.ValidateMCP(ctx, mcpObj, box)
		if validateErr != nil {

			continue
		}
		for _, name := range toolNames {
			if prev, exists := toolOwnerMap[name]; exists {
				return errs.NewFormatError(
					"MCP tool name conflict: '%s' is exposed by both '%s' and '%s'",
					"MCP 工具名称冲突：'%s' 同时存在于 '%s' 和 '%s'",
					name, prev, ep.DisplayName,
				)
			}
			toolOwnerMap[name] = ep.DisplayName
		}
	}
	return nil
}

func (service *McpService) BuildMCPObjectFromEndpoint(ep *po.MCPEndpoint) tools.MCP {
	mcpObj := tools.MCP{
		Transport:   ep.Transport,
		Name:        ep.Name,
		DisplayName: ep.DisplayName,
		HttpURL:     ep.HttpUrl,
		HttpHeader:  ep.HttpHeader,
		ProxyURL:    ep.HttpProxyURL,
		StdioType:   ep.StdioType,
	}

	if ep.StdioEnv != "" {
		_ = json.Unmarshal([]byte(ep.StdioEnv), &mcpObj.StdioENV)
	}
	if ep.StdioArgs != "" {
		_ = json.Unmarshal([]byte(ep.StdioArgs), &mcpObj.StdioArgs)
	}

	return mcpObj
}

func (service *McpService) buildMCPObject(transport, httpUrl, httpHeader, proxyUrl, stdioType, stdioEnv, stdioArgs string) tools.MCP {
	mcpObj := tools.MCP{
		Transport:  transport,
		HttpURL:    httpUrl,
		HttpHeader: httpHeader,
		ProxyURL:   proxyUrl,
		StdioType:  stdioType,
	}

	if stdioEnv != "" {
		_ = json.Unmarshal([]byte(stdioEnv), &mcpObj.StdioENV)
	}
	if stdioArgs != "" {
		_ = json.Unmarshal([]byte(stdioArgs), &mcpObj.StdioArgs)
	}

	return mcpObj
}
