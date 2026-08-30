package repository

import (
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository/seed"
	"goraven/backend/vo"
	"goraven/util"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *MCPRepository {
			return &MCPRepository{}
		})
	})
}

type MCPRepository struct {
	freedom.Repository
}

// FindEnabledMCPEndpoints 查询所有启用且未删除的 MCP 端点
func (repo *MCPRepository) FindEnabledMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ?", 1, 0).Find(&endpoints).Error
	return endpoints, err
}

// FindUserSelectableMCPEndpoints 查询用户可选的 MCP 端点（启用且未删除且非始终启用）
// 始终启用的 MCP 由管理员配置，用户无需也无需选择，故从可选列表排除
func (repo *MCPRepository) FindUserSelectableMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ? AND always_on = ?", 1, 0, 0).Find(&endpoints).Error
	return endpoints, err
}

// FindEnabledMCPEndpointsByIDs 按 mcpId 列表查询启用且未删除的 MCP 端点
func (repo *MCPRepository) FindEnabledMCPEndpointsByIDs(mcpIds []int) ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ? AND mcp_id IN ?", 1, 0, mcpIds).Find(&endpoints).Error
	return endpoints, err
}

// FindAlwaysOnMcpIds 获取始终启用的 MCP 端点 ID 列表（启用且未删除且 always_on=1）
func (repo *MCPRepository) FindAlwaysOnMcpIds() ([]int, error) {
	var ids []int
	err := repo.db().Model(&po.MCPEndpoint{}).
		Where("status = ? AND deleted = ? AND always_on = ?", 1, 0, 1).
		Pluck("mcp_id", &ids).Error
	return ids, err
}

// UpdateMCPEndpointStatus 更新 MCP 端点状态
func (repo *MCPRepository) UpdateMCPEndpointStatus(mcpId int, status uint8) error {
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ?", mcpId).Update("status", status).Error
}

// PaginateMCPEndpoints 分页查询 MCP 端点列表
func (repo *MCPRepository) PaginateMCPEndpoints(req *vo.AdminMCPListReq) ([]vo.AdminMCPItem, *PageResult, error) {
	query := repo.db().Model(&po.MCPEndpoint{}).Where("deleted = 0")
	if req.Search != "" {
		query = query.Where("name LIKE ? OR display_name LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Transport != "" {
		query = query.Where("transport = ?", req.Transport)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var endpoints []po.MCPEndpoint
	pr, err := Paginate(query.Order("status DESC, created DESC"), &PageQuery{Page: req.Page, PageSize: req.PageSize}, &endpoints)
	if err != nil {
		return nil, nil, err
	}

	items := make([]vo.AdminMCPItem, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, vo.AdminMCPItem{
			McpId:           ep.McpId,
			Name:            ep.Name,
			DisplayName:     ep.DisplayName,
			Icon:            ep.Icon,
			Description:     ep.Description,
			Transport:       ep.Transport,
			HttpUrl:         ep.HttpUrl,
			HttpHeader:      util.MaskJSONValues(ep.HttpHeader),
			HttpProxyUrl:    ep.HttpProxyURL,
			StdioType:       ep.StdioType,
			StdioEnv:        util.MaskJSONValues(ep.StdioEnv),
			StdioArgs:       ep.StdioArgs,
			Status:          ep.Status,
			AlwaysOn:        ep.AlwaysOn,
			HealthLatency:   ep.HealthLatency,
			HealthCheckedAt: ep.HealthCheckedAt,
			Remark:          ep.Remark,
			Created:         ep.Created,
			Updated:         ep.Updated,
		})
	}

	return items, pr, nil
}

// GetMCPEndpointByID 根据 ID 查询 MCP 端点详情
func (repo *MCPRepository) GetMCPEndpointByID(mcpId int) (*po.MCPEndpoint, error) {
	var endpoint po.MCPEndpoint
	err := repo.db().First(&endpoint, "mcp_id = ? AND deleted = 0", mcpId).Error
	return &endpoint, err
}

// CreateMCPEndpoint 创建 MCP 端点
func (repo *MCPRepository) CreateMCPEndpoint(endpoint *po.MCPEndpoint) error {
	return repo.db().Create(endpoint).Error
}

// UpdateMCPEndpoint 更新 MCP 端点
func (repo *MCPRepository) UpdateMCPEndpoint(mcpId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ? AND deleted = 0", mcpId).Updates(updates).Error
}

// SoftDeleteMCPEndpoint 软删除 MCP 端点，name 追加时间戳后缀以允许复用
func (repo *MCPRepository) SoftDeleteMCPEndpoint(mcpId int) error {
	var endpoint po.MCPEndpoint
	if err := repo.db().First(&endpoint, "mcp_id = ? AND deleted = 0", mcpId).Error; err != nil {
		return err
	}

	suffixedName := fmt.Sprintf("%s-%d", endpoint.Name, time.Now().Unix())
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ? AND deleted = 0", mcpId).Updates(map[string]interface{}{
		"deleted": 1,
		"name":    suffixedName,
		"updated": time.Now(),
	}).Error
}

// UpdateMCPEndpointHealth 更新 MCP 端点延迟和检测时间
func (repo *MCPRepository) UpdateMCPEndpointHealth(mcpId int, latencyMs int, checkedAt time.Time) error {
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ? AND deleted = 0", mcpId).Updates(map[string]interface{}{
		"health_latency":   latencyMs,
		"health_checked_at": checkedAt,
		"updated":         time.Now(),
	}).Error
}

// FindMCPEndpointByName 根据名称查询未删除的 MCP 端点
func (repo *MCPRepository) FindMCPEndpointByName(name string) (*po.MCPEndpoint, error) {
	var endpoint po.MCPEndpoint
	err := repo.db().First(&endpoint, "name = ? AND deleted = 0", name).Error
	return &endpoint, err
}

// FindAllActiveMCPEndpoints 查询所有启用且未删除的 MCP 端点（用于推荐列表匹配安装状态）
func (repo *MCPRepository) FindAllActiveMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = 1 AND deleted = 0").Find(&endpoints).Error
	return endpoints, err
}

// GetRecommendMCPEndpoints 推荐列表
func (repo *MCPRepository) GetRecommendMCPEndpoints() ([]po.MCPEndpoint, error) {
	return seed.RecommendMCPEndpoints, nil
}

// GetMCPEndpointsByIDs 根据 ID 列表批量查询启用且未删除的 MCP 端点
func (repo *MCPRepository) GetMCPEndpointsByIDs(mcpIds []int) ([]po.MCPEndpoint, error) {
	if len(mcpIds) == 0 {
		return nil, nil
	}
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("mcp_id IN ? AND status = 1 AND deleted = 0", mcpIds).Find(&endpoints).Error
	return endpoints, err
}

func (repo *MCPRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
