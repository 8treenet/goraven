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

func (repo *MCPRepository) FindEnabledMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ?", 1, 0).Find(&endpoints).Error
	return endpoints, err
}

func (repo *MCPRepository) FindUserSelectableMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ? AND always_on = ?", 1, 0, 0).Find(&endpoints).Error
	return endpoints, err
}

func (repo *MCPRepository) FindEnabledMCPEndpointsByIDs(mcpIds []int) ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = ? AND deleted = ? AND mcp_id IN ?", 1, 0, mcpIds).Find(&endpoints).Error
	return endpoints, err
}

func (repo *MCPRepository) FindAlwaysOnMcpIds() ([]int, error) {
	var ids []int
	err := repo.db().Model(&po.MCPEndpoint{}).
		Where("status = ? AND deleted = ? AND always_on = ?", 1, 0, 1).
		Pluck("mcp_id", &ids).Error
	return ids, err
}

func (repo *MCPRepository) UpdateMCPEndpointStatus(mcpId int, status uint8) error {
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ?", mcpId).Update("status", status).Error
}

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

func (repo *MCPRepository) GetMCPEndpointByID(mcpId int) (*po.MCPEndpoint, error) {
	var endpoint po.MCPEndpoint
	err := repo.db().First(&endpoint, "mcp_id = ? AND deleted = 0", mcpId).Error
	return &endpoint, err
}

func (repo *MCPRepository) CreateMCPEndpoint(endpoint *po.MCPEndpoint) error {
	return repo.db().Create(endpoint).Error
}

func (repo *MCPRepository) UpdateMCPEndpoint(mcpId int, updates map[string]interface{}) error {
	updates["updated"] = time.Now()
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ? AND deleted = 0", mcpId).Updates(updates).Error
}

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

func (repo *MCPRepository) UpdateMCPEndpointHealth(mcpId int, latencyMs int, checkedAt time.Time) error {
	return repo.db().Model(&po.MCPEndpoint{}).Where("mcp_id = ? AND deleted = 0", mcpId).Updates(map[string]interface{}{
		"health_latency":    latencyMs,
		"health_checked_at": checkedAt,
		"updated":           time.Now(),
	}).Error
}

func (repo *MCPRepository) FindMCPEndpointByName(name string) (*po.MCPEndpoint, error) {
	var endpoint po.MCPEndpoint
	err := repo.db().First(&endpoint, "name = ? AND deleted = 0", name).Error
	return &endpoint, err
}

func (repo *MCPRepository) FindAllActiveMCPEndpoints() ([]po.MCPEndpoint, error) {
	var endpoints []po.MCPEndpoint
	err := repo.db().Where("status = 1 AND deleted = 0").Find(&endpoints).Error
	return endpoints, err
}

func (repo *MCPRepository) GetRecommendMCPEndpoints() ([]po.MCPEndpoint, error) {
	return seed.RecommendMCPEndpoints, nil
}

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
