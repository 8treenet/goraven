package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"goraven/config"
	"goraven/util"
	"os"
	"path/filepath"
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

type ClawHubRepository struct {
	freedom.Repository
	ClawHubAPIURL string
	ClawHubToken  string
}

type ClawHubSearchResult struct {
	Slug        string  `json:"slug"`
	DisplayName string  `json:"displayName"`
	Summary     string  `json:"summary"`
	Version     string  `json:"version"`
	Score       float64 `json:"score"`
	UpdatedAt   int64   `json:"updatedAt"`
}

type ClawHubSearchResponse struct {
	Results []ClawHubSearchResult `json:"results"`
}

type ClawHubSkillListItem struct {
	Slug        string            `json:"slug"`
	DisplayName string            `json:"displayName"`
	Summary     string            `json:"summary"`
	Stats       ClawHubSkillStats `json:"stats"`
	CreatedAt   int64             `json:"createdAt"`
	UpdatedAt   int64             `json:"updatedAt"`
}

type ClawHubSkillStats struct {
	Comments        int `json:"comments"`
	Downloads       int `json:"downloads"`
	InstallsAllTime int `json:"installsAllTime"`
	InstallsCurrent int `json:"installsCurrent"`
	Stars           int `json:"stars"`
	Versions        int `json:"versions"`
}

type ClawHubExploreResponse struct {
	Items      []ClawHubSkillListItem `json:"items"`
	NextCursor string                 `json:"nextCursor"`
}

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

func (repo *ClawHubRepository) Explore(sort string) (*ClawHubExploreResponse, error) {
	apiSort := repo.ResolveExploreSort(sort)
	cacheKey := fmt.Sprintf("clawhub:explore:%d:%s", 100, apiSort)

	cached, err := repo.Redis().Get(context.Background(), cacheKey).Bytes()
	if err == nil {
		var result ClawHubExploreResponse
		if json.Unmarshal(cached, &result) == nil {
			return &result, nil
		}
	}

	url := repo.registryURL("/api/v1/skills")

	if apiSort != "" && apiSort != "createdAt" {
		url = fmt.Sprintf("%s?sort=%s", url, apiSort)
	}

	var result ClawHubExploreResponse
	resp := repo.newGetRequest(url).ToJSON(&result)
	if resp.Error != nil {
		return nil, fmt.Errorf("clawhub explore failed: %w", resp.Error)
	}

	if data, err := json.Marshal(&result); err == nil {
		repo.Redis().Set(context.Background(), cacheKey, data, 24*time.Hour)
	}

	return &result, nil
}

func (repo *ClawHubRepository) FetchFile(slug, filePath string) (string, error) {
	cacheDir := config.Get().GetClawHUBCacheDir()
	cachePath := filepath.Join(cacheDir, slug+".md")

	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < 10*24*time.Hour {
			data, err := os.ReadFile(cachePath)
			if err == nil {
				return string(data), nil
			}
		}
	}

	url := repo.registryURL(fmt.Sprintf("/api/v1/skills/%s/file", util.EncodePath(slug)))
	url = fmt.Sprintf("%s?path=%s", url, util.EncodeQuery(filePath))

	text, resp := repo.newGetRequest(url).ToString()
	if resp.Error != nil {
		return "", fmt.Errorf("clawhub fetch file failed: %w", resp.Error)
	}

	os.WriteFile(cachePath, []byte(text), 0644)

	return text, nil
}

func (repo *ClawHubRepository) Download(slug string) (string, error) {
	url := repo.registryURL("/api/v1/download")
	url = fmt.Sprintf("%s?slug=%s", url, util.EncodeQuery(slug))

	zipPath, err := repo.downloadZip(url)
	return zipPath, err
}

func (repo *ClawHubRepository) registryURL(path string) string {
	base := "https://clawhub.ai"
	if repo.ClawHubAPIURL != "" {
		base = repo.ClawHubAPIURL
	}
	base = strings.TrimRight(base, "/")
	return base + path
}

func (repo *ClawHubRepository) newGetRequest(url string) requests.Request {
	req := requests.NewHTTPRequest(url).Get()
	if repo.ClawHubToken != "" {
		req = req.SetHeaderValue("Authorization", "Bearer "+repo.ClawHubToken)
	}
	return req
}

func (repo *ClawHubRepository) downloadZip(url string) (string, error) {
	downloadDir := config.Get().GetDownloadTempDir()

	tmpPath := filepath.Join(downloadDir, util.UUID()+".zip")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %w", err)
	}

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
