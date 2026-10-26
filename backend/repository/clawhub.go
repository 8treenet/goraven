package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"goraven/config"
	"goraven/util"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/infra/requests"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *ClawHubRepository {
			return &ClawHubRepository{}
		})
	})
}

// ClawHubRepository ClawHub HTTP API 仓库
// 通过 HTTP 调用 ClawHub Registry API（/api/v1/*）实现技能搜索、浏览、详情、下载等功能。
type ClawHubRepository struct {
	freedom.Repository
	ClawHubAPIURL string
	ClawHubToken  string
}

// ──────────────────────────── 响应结构体 ────────────────────────────

// ClawHubSearchResult 搜索结果条目
type ClawHubSearchResult struct {
	Slug        string  `json:"slug"`        // 唯一标识
	DisplayName string  `json:"displayName"` // 展示名
	Summary     string  `json:"summary"`     // 简短描述
	Version     string  `json:"version"`     // 匹配的版本号
	Score       float64 `json:"score"`       // 搜索相关度分数
	UpdatedAt   int64   `json:"updatedAt"`   // 更新时间（毫秒时间戳）
}

// ClawHubSearchResponse 搜索接口响应
type ClawHubSearchResponse struct {
	Results []ClawHubSearchResult `json:"results"` // 搜索结果列表
}

// ClawHubSkillListItem 技能列表条目（explore 接口返回）
type ClawHubSkillListItem struct {
	Slug        string            `json:"slug"`        // 唯一标识
	DisplayName string            `json:"displayName"` // 展示名
	Summary     string            `json:"summary"`     // 简短描述
	Stats       ClawHubSkillStats `json:"stats"`       // 统计数据
	CreatedAt   int64             `json:"createdAt"`   // 创建时间（毫秒时间戳）
	UpdatedAt   int64             `json:"updatedAt"`   // 更新时间（毫秒时间戳）
}

// ClawHubSkillStats 技能统计
type ClawHubSkillStats struct {
	Comments        int `json:"comments"`        // 评论数
	Downloads       int `json:"downloads"`       // 下载量
	InstallsAllTime int `json:"installsAllTime"` // 累计安装数
	InstallsCurrent int `json:"installsCurrent"` // 当前安装数
	Stars           int `json:"stars"`           // 星标数
	Versions        int `json:"versions"`        // 版本数
}

// ClawHubExploreResponse 技能列表（explore）接口响应
type ClawHubExploreResponse struct {
	Items      []ClawHubSkillListItem `json:"items"`      // 技能列表
	NextCursor string                 `json:"nextCursor"` // 分页游标
}

// ──────────────────────────── 公开方法 ────────────────────────────

// Search 向量搜索技能
// 调用 GET /api/v1/search?q={query}&limit={limit}
func (repo *ClawHubRepository) Search(query string, limit int) (*ClawHubSearchResponse, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}

	url := repo.registryURL("/api/v1/search")
	url = fmt.Sprintf("%s?q=%s&limit=%d", url, util.EncodeQuery(query), limit)

	var result ClawHubSearchResponse
	resp := repo.newGetRequest(url).ToJSON(&result)
	if resp.Error != nil {
		return nil, fmt.Errorf("clawhub search failed: %w", resp.Error)
	}
	return &result, nil
}

// Explore 浏览技能列表（按排序规则），带 Redis 缓存（24 小时）
// 调用 GET /api/v1/skills?limit={limit}&sort={sort}
// sort 可选值：createdAt（最新）、updated（最近更新）、downloads、stars、installsCurrent、installsAllTime、trending
func (repo *ClawHubRepository) Explore(sort string) (*ClawHubExploreResponse, error) {
	apiSort := repo.ResolveExploreSort(sort)
	cacheKey := fmt.Sprintf("clawhub:explore:%d:%s", 100, apiSort)

	// 尝试读取 Redis 缓存
	cached, err := repo.Redis().Get(context.Background(), cacheKey).Bytes()
	if err == nil {
		var result ClawHubExploreResponse
		if json.Unmarshal(cached, &result) == nil {
			return &result, nil
		}
	}

	// 缓存未命中，走远程拉取
	url := repo.registryURL("/api/v1/skills")
	// url = fmt.Sprintf("%s?limit=%d", url, limit)
	if apiSort != "" && apiSort != "createdAt" {
		url = fmt.Sprintf("%s?sort=%s", url, apiSort)
	}

	var result ClawHubExploreResponse
	resp := repo.newGetRequest(url).ToJSON(&result)
	if resp.Error != nil {
		return nil, fmt.Errorf("clawhub explore failed: %w", resp.Error)
	}

	// 写入 Redis 缓存，24 小时过期
	if data, err := json.Marshal(&result); err == nil {
		repo.Redis().Set(context.Background(), cacheKey, data, 24*time.Hour)
	}

	return &result, nil
}

// FetchFile 获取技能内单个文件的文本内容（带本地文件缓存，10 天过期）
// 缓存路径：{GetClawHUBCacheDir}/{slug}.md，先读本地缓存，超过 10 天则远程拉取并更新缓存
func (repo *ClawHubRepository) FetchFile(slug, filePath string) (string, error) {
	cacheDir := config.Get().GetClawHUBCacheDir()
	cachePath := filepath.Join(cacheDir, slug+".md")

	// 尝试读取本地缓存
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < 10*24*time.Hour {
			data, err := os.ReadFile(cachePath)
			if err == nil {
				return string(data), nil
			}
		}
	}

	// 缓存不存在或超过 10 天，走远程拉取
	url := repo.registryURL(fmt.Sprintf("/api/v1/skills/%s/file", util.EncodePath(slug)))
	url = fmt.Sprintf("%s?path=%s", url, util.EncodeQuery(filePath))

	text, resp := repo.newGetRequest(url).ToString()
	if resp.Error != nil {
		return "", fmt.Errorf("clawhub fetch file failed: %w", resp.Error)
	}

	// 写入本地缓存
	os.WriteFile(cachePath, []byte(text), 0644)

	return text, nil
}

// Download 下载技能 zip 到本地目录并解压
// 调用 GET /api/v1/download?slug={slug}，返回 zip 字节流
// 解压到 config.Paths.SkillsHub/{slug}/ 目录
func (repo *ClawHubRepository) Download(slug string) (string, error) {
	url := repo.registryURL("/api/v1/download")
	url = fmt.Sprintf("%s?slug=%s", url, util.EncodeQuery(slug))

	// 下载 zip 到临时文件
	zipPath, err := repo.downloadZip(url)
	return zipPath, err
}

// ──────────────────────────── 内部方法 ────────────────────────────

// registryURL 拼接 registry 基础地址和 API 路径
func (repo *ClawHubRepository) registryURL(path string) string {
	base := "https://clawhub.ai"
	if repo.ClawHubAPIURL != "" {
		base = repo.ClawHubAPIURL
	}
	base = strings.TrimRight(base, "/")
	return base + path
}

// newGetRequest 创建带认证的 GET 请求
func (repo *ClawHubRepository) newGetRequest(url string) requests.Request {
	req := requests.NewHTTPRequest(url).Get()
	if repo.ClawHubToken != "" {
		req = req.SetHeaderValue("Authorization", "Bearer "+repo.ClawHubToken)
	}
	return req
}

// downloadZip 下载 zip 文件到临时目录
// 官方 download 接口返回二进制 zip 流，不走 registry 代理
func (repo *ClawHubRepository) downloadZip(url string) (string, error) {
	downloadDir := config.Get().GetDownloadTempDir()

	// 创建临时 zip 文件
	tmpPath := filepath.Join(downloadDir, util.UUID()+".zip")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %w", err)
	}

	// 发起 HTTP 请求获取 zip 内容
	req := requests.NewHTTPRequest(url).Get()
	if repo.ClawHubToken != "" {
		req = req.SetHeaderValue("Authorization", "Bearer "+repo.ClawHubToken)
	}

	data, resp := req.ToBytes()
	if resp.Error != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("download zip failed: %w", resp.Error)
	}

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("write zip file failed: %w", err)
	}
	tmpFile.Close()

	return tmpPath, nil
}

// ResolveExploreSort 将前端排序参数映射为 API sort 参数
// 支持的值：newest/updated/downloads/rating/installs/installsAllTime/trending
func (repo *ClawHubRepository) ResolveExploreSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "newest", "createdat", "":
		return "createdAt"
	case "updated":
		return "updated"
	case "downloads", "download":
		return "downloads"
	case "rating", "stars", "star":
		return "stars"
	case "installs", "install", "current":
		return "installsCurrent"
	case "installsalltime":
		return "installsAllTime"
	case "trending":
		return "trending"
	default:
		return "createdAt"
	}
}
