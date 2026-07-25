package test

import (
	"goraven/backend/vo"
	"strconv"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestTeamProjectList(t *testing.T) {
	var rsp struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.TeamProjectListRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/list").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("list success: %d items", len(rsp.Data.Items))
	for _, item := range rsp.Data.Items {
		t.Log(item)
	}
}

func TestTeamProjectShareAndUnshare(t *testing.T) {

	mkdirReq := vo.FileManagerMkdirReq{Path: "projects/test_team_project"}
	requests.NewHTTPRequest(domain+"/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	shareReq := vo.TeamProjectShareReq{
		ProjectName: "js-learning",
		Description: "测试js-learning",
	}
	var shareRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/share").
		Post().
		SetJSONBody(shareReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&shareRsp)
	if httpResp.Error != nil {
		t.Fatal("share error:", httpResp.Error)
	}
	if shareRsp.Code != 0 {
		t.Fatalf("share failed: code=%d msg=%s", shareRsp.Code, shareRsp.Msg)
	}
	t.Log("share success: test_team_project")
}

func TestTeamProjectUpdateDescription(t *testing.T) {
	sharedId := 1
	if sharedId == 0 {
		t.Skip("请设置 sharedId 变量（从 TestTeamProjectShareAndUnshare 获取）")
	}

	updateReq := vo.TeamProjectUpdateReq{Description: "更新后的简介1122345"}
	var rsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/"+strconv.Itoa(sharedId)).
		Put().
		SetJSONBody(updateReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("update error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("update failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Log("update description success")
}

func TestTeamProjectUnshare(t *testing.T) {
	sharedId := 2
	if sharedId == 0 {
		t.Skip("请设置 sharedId 变量（从 TestTeamProjectShareAndUnshare 获取）")
	}

	var rsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/"+strconv.Itoa(sharedId)).
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("unshare error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("unshare failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Log("unshare success")
}

func TestTeamProjectShareInvalidName(t *testing.T) {
	shareReq := vo.TeamProjectShareReq{
		ProjectName: "foo/bar",
		Description: "",
	}
	var rsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/share").
		Post().
		SetJSONBody(shareReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("request error:", httpResp.Error)
	}
	if rsp.Code == 0 {
		t.Fatal("expected error for invalid project name, got success")
	}
	t.Logf("invalid name correctly rejected: code=%d msg=%s", rsp.Code, rsp.Msg)
}

func TestTeamProjectShareNonexistentDir(t *testing.T) {
	shareReq := vo.TeamProjectShareReq{
		ProjectName: "nonexistent_project_12345",
		Description: "",
	}
	var rsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/share").
		Post().
		SetJSONBody(shareReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("request error:", httpResp.Error)
	}
	if rsp.Code == 0 {
		t.Fatal("expected error for nonexistent directory, got success")
	}
	t.Logf("nonexistent dir correctly rejected: code=%d msg=%s", rsp.Code, rsp.Msg)
}

func TestTeamProjectFileOps(t *testing.T) {

	idStr := "1"

	var listFilesRsp struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.FileManagerListRsp `json:"data,omitempty"`
	}
	requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr+"/list").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&listFilesRsp)
	if listFilesRsp.Code != 0 {
		t.Fatalf("project list failed: code=%d msg=%s", listFilesRsp.Code, listFilesRsp.Msg)
	}
	found := false
	for _, item := range listFilesRsp.Data.Items {
		t.Log(item)
	}
	if !found {
		t.Fatal("test_subdir not found in project listing")
	}
	t.Log("project list success: found test_subdir")
	return
}

func TestTeamProjectRenameAndCompress(t *testing.T) {

	mkdirReq := vo.FileManagerMkdirReq{Path: "projects/test_team_project"}
	requests.NewHTTPRequest(domain+"/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	shareReq := vo.TeamProjectShareReq{ProjectName: "test_team_project", Description: ""}
	requests.NewHTTPRequest(domain+"/api/teamProject/share").
		Post().
		SetJSONBody(shareReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	var listRsp struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data vo.TeamProjectListRsp `json:"data,omitempty"`
	}
	requests.NewHTTPRequest(domain+"/api/teamProject/list").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&listRsp)

	var sharedId int
	for _, item := range listRsp.Data.Items {
		if item.ProjectName == "test_team_project" && item.IsOwner {
			sharedId = item.Id
			break
		}
	}
	if sharedId == 0 {
		t.Skip("shared project not found, skipping compress test")
	}

	idStr := strconv.Itoa(sharedId)

	requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr+"/mkdir").
		Post().
		SetJSONBody(vo.FileManagerMkdirReq{Path: "compress_src"}).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	renameReq := vo.FileManagerRenameReq{OldPath: "compress_src", NewPath: "compress_renamed"}
	var renameRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr+"/rename").
		Put().
		SetJSONBody(renameReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&renameRsp)
	if httpResp.Error != nil {
		t.Fatal("rename error:", httpResp.Error)
	}
	if renameRsp.Code != 0 {
		t.Fatalf("rename failed: code=%d msg=%s", renameRsp.Code, renameRsp.Msg)
	}
	t.Log("rename success: compress_src → compress_renamed")

	compressReq := vo.FileManagerCompressReq{
		Paths:      []string{"compress_renamed"},
		OutputName: "test_compress.zip",
	}
	var compressRsp struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Data vo.FileManagerCompressRsp `json:"data,omitempty"`
	}
	httpResp = requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr+"/compress").
		Post().
		SetJSONBody(compressReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&compressRsp)
	if httpResp.Error != nil {
		t.Fatal("compress error:", httpResp.Error)
	}
	if compressRsp.Code != 0 {
		t.Fatalf("compress failed: code=%d msg=%s", compressRsp.Code, compressRsp.Msg)
	}
	t.Logf("compress success: zipPath=%s", compressRsp.Data.ZipPath)

	decompressReq := vo.FileManagerDecompressReq{
		Path:     compressRsp.Data.ZipPath,
		ToSubDir: true,
	}
	var decompressRsp apiRsp
	httpResp = requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr+"/decompress").
		Post().
		SetJSONBody(decompressReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&decompressRsp)
	if httpResp.Error != nil {
		t.Fatal("decompress error:", httpResp.Error)
	}
	if decompressRsp.Code != 0 {
		t.Fatalf("decompress failed: code=%d msg=%s", decompressRsp.Code, decompressRsp.Msg)
	}
	t.Log("decompress success")

	requests.NewHTTPRequest(domain+"/api/teamProject/"+idStr).
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
	requests.NewHTTPRequest(domain+"/api/fileManager/delete").
		Delete().
		SetJSONBody(vo.FileManagerDeleteReq{Paths: []string{"projects/test_team_project"}}).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
}
