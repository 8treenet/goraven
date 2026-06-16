package test

import (
	"encoding/json"
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

type apiRsp struct {
	Code	int		`json:"code"`
	Msg	string		`json:"msg"`
	Data	json.RawMessage	`json:"data,omitempty"`
}

func TestFileManagerListRoot(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerListRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/list").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list root error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list root failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("list root success: %d items", len(rsp.Data.Items))
	for _, item := range rsp.Data.Items {
		t.Logf("  %s isDir=%v size=%d modTime=%s", item.Name, item.IsDir, item.Size, item.ModTime)
	}
}

func TestFileManagerListDocuments(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerListRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/list?dir=documents").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list documents error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list documents failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("list documents success: %d items", len(rsp.Data.Items))
	for _, item := range rsp.Data.Items {
		t.Logf("  %s isDir=%v size=%d", item.Name, item.IsDir, item.Size)
	}
}

func TestFileManagerListSortBySize(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerListRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/list?dir=documents&sort=size&order=desc").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("list sort by size error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("list sort by size failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("list sort by size desc: %d items", len(rsp.Data.Items))
	for _, item := range rsp.Data.Items {
		t.Logf("  %s isDir=%v size=%d", item.Name, item.IsDir, item.Size)
	}
}

func TestFileManagerMkdirAndDelete(t *testing.T) {

	mkdirReq := vo.FileManagerMkdirReq{
		Path: "test_mkdir_dir",
	}
	var mkdirRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&mkdirRsp)
	if httpResp.Error != nil {
		t.Fatal("mkdir error:", httpResp.Error)
	}
	if mkdirRsp.Code != 0 {
		t.Fatalf("mkdir failed: code=%d msg=%s", mkdirRsp.Code, mkdirRsp.Msg)
	}
	t.Log("mkdir success: test_mkdir_dir")

	var listRsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerListRsp	`json:"data,omitempty"`
	}
	requests.NewHTTPRequest(domain+"/api/fileManager/list").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&listRsp)
	found := false
	for _, item := range listRsp.Data.Items {
		if item.Name == "test_mkdir_dir" && item.IsDir {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("mkdir verification failed: test_mkdir_dir not found in root listing")
	}

	deleteReq := vo.FileManagerDeleteReq{
		Paths: []string{"test_mkdir_dir"},
	}
	var delRsp apiRsp
	httpResp = requests.NewHTTPRequest(domain+"/api/fileManager/delete").
		Delete().
		SetJSONBody(deleteReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&delRsp)
	if httpResp.Error != nil {
		t.Fatal("delete error:", httpResp.Error)
	}
	if delRsp.Code != 0 {
		t.Fatalf("delete failed: code=%d msg=%s", delRsp.Code, delRsp.Msg)
	}
	t.Log("delete success: test_mkdir_dir")
}

func TestFileManagerRename(t *testing.T) {

	mkdirReq := vo.FileManagerMkdirReq{Path: "test_rename_before"}
	var mkdirRsp apiRsp
	requests.NewHTTPRequest(domain+"/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&mkdirRsp)
	if mkdirRsp.Code != 0 {
		t.Fatalf("mkdir for rename failed: code=%d msg=%s", mkdirRsp.Code, mkdirRsp.Msg)
	}

	renameReq := vo.FileManagerRenameReq{
		OldPath:	"test_rename_before",
		NewPath:	"test_rename_after",
	}
	var renameRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/rename").
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
	t.Log("rename success: test_rename_before → test_rename_after")

	deleteReq := vo.FileManagerDeleteReq{Paths: []string{"test_rename_after"}}
	requests.NewHTTPRequest(domain+"/api/fileManager/delete").
		Delete().
		SetJSONBody(deleteReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
}

func TestFileManagerCompressAndDecompress(t *testing.T) {

	mkdirReq := vo.FileManagerMkdirReq{Path: "test_compress_src"}
	requests.NewHTTPRequest(domain+"/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	compressReq := vo.FileManagerCompressReq{
		Paths:		[]string{"test_compress_src"},
		OutputName:	"test_compress_src.zip",
	}
	var compressRsp struct {
		Code	int				`json:"code"`
		Msg	string				`json:"msg"`
		Data	vo.FileManagerCompressRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/compress").
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
		Path:		compressRsp.Data.ZipPath,
		ToSubDir:	true,
	}
	var decompressRsp apiRsp
	httpResp = requests.NewHTTPRequest(domain+"/api/fileManager/decompress").
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

	deleteReq := vo.FileManagerDeleteReq{
		Paths: []string{"test_compress_src", "test_compress_src.zip", "test_compress_src"},
	}
	requests.NewHTTPRequest(domain+"/api/fileManager/delete").
		Delete().
		SetJSONBody(deleteReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
}

func TestFileManagerUsage(t *testing.T) {
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerUsageRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/usage").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("usage error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("usage failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("usage success: totalSize=%d usedSize=%d fileCount=%d",
		rsp.Data.TotalSize, rsp.Data.UsedSize, rsp.Data.FileCount)
}

func TestFileManagerUpload(t *testing.T) {

	uploadId := ""
	if uploadId == "" {
		t.Skip("请设置 uploadId 变量（先完成 HFS 分片上传流程获取）")
	}

	uploadReq := vo.FileManagerUploadReq{
		UploadId:	uploadId,
		Dir:		"documents",
	}
	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.FileManagerUploadRsp	`json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/fileManager/upload").
		Post().
		SetJSONBody(uploadReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("upload error:", httpResp.Error)
	}
	if rsp.Code != 0 {
		t.Fatalf("upload failed: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	t.Logf("upload success: path=%s", rsp.Data.Path)
}
