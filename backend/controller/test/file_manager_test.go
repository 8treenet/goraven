package test

import (
	"encoding/json"
	"goraven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

// apiRsp 通用 API 响应结构
type apiRsp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

// TestFileManagerListRoot 列出根目录 GET /api/fileManager/list
func TestFileManagerListRoot(t *testing.T) {
	var rsp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data vo.FileManagerListRsp  `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/list").
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

// TestFileManagerListDocuments 列出 documents 目录 GET /api/fileManager/list?dir=documents
func TestFileManagerListDocuments(t *testing.T) {
	var rsp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data vo.FileManagerListRsp  `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/list?dir=documents").
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

// TestFileManagerListSortBySize 按大小排序 GET /api/fileManager/list?dir=documents&sort=size&order=desc
func TestFileManagerListSortBySize(t *testing.T) {
	var rsp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data vo.FileManagerListRsp  `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/list?dir=documents&sort=size&order=desc").
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

// TestFileManagerMkdirAndDelete 创建目录后删除 POST /api/fileManager/mkdir → DELETE /api/fileManager/delete
func TestFileManagerMkdirAndDelete(t *testing.T) {
	// 创建目录
	mkdirReq := vo.FileManagerMkdirReq{
		Path: "test_mkdir_dir",
	}
	var mkdirRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/mkdir").
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

	// 验证目录存在
	var listRsp struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data vo.FileManagerListRsp  `json:"data,omitempty"`
	}
	requests.NewHTTPRequest(domain + "/api/fileManager/list").
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

	// 删除目录
	deleteReq := vo.FileManagerDeleteReq{
		Paths: []string{"test_mkdir_dir"},
	}
	var delRsp apiRsp
	httpResp = requests.NewHTTPRequest(domain + "/api/fileManager/delete").
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

// TestFileManagerRename 重命名 PUT /api/fileManager/rename
func TestFileManagerRename(t *testing.T) {
	// 先创建测试目录
	mkdirReq := vo.FileManagerMkdirReq{Path: "test_rename_before"}
	var mkdirRsp apiRsp
	requests.NewHTTPRequest(domain + "/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&mkdirRsp)
	if mkdirRsp.Code != 0 {
		t.Fatalf("mkdir for rename failed: code=%d msg=%s", mkdirRsp.Code, mkdirRsp.Msg)
	}

	// 重命名
	renameReq := vo.FileManagerRenameReq{
		OldPath: "test_rename_before",
		NewPath: "test_rename_after",
	}
	var renameRsp apiRsp
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/rename").
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

	// 清理
	deleteReq := vo.FileManagerDeleteReq{Paths: []string{"test_rename_after"}}
	requests.NewHTTPRequest(domain + "/api/fileManager/delete").
		Delete().
		SetJSONBody(deleteReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
}

// TestFileManagerCompressAndDecompress 压缩和解压 POST /api/fileManager/compress → POST /api/fileManager/decompress
func TestFileManagerCompressAndDecompress(t *testing.T) {
	// 创建测试文件
	mkdirReq := vo.FileManagerMkdirReq{Path: "test_compress_src"}
	requests.NewHTTPRequest(domain + "/api/fileManager/mkdir").
		Post().
		SetJSONBody(mkdirReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})

	// 压缩 documents 目录
	compressReq := vo.FileManagerCompressReq{
		Paths:      []string{"test_compress_src"},
		OutputName: "test_compress_src.zip",
	}
	var compressRsp struct {
		Code int                        `json:"code"`
		Msg  string                     `json:"msg"`
		Data vo.FileManagerCompressRsp  `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/compress").
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

	// 解压到同名子目录
	decompressReq := vo.FileManagerDecompressReq{
		Path:     compressRsp.Data.ZipPath,
		ToSubDir: true,
	}
	var decompressRsp apiRsp
	httpResp = requests.NewHTTPRequest(domain + "/api/fileManager/decompress").
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

	// 清理
	deleteReq := vo.FileManagerDeleteReq{
		Paths: []string{"test_compress_src", "test_compress_src.zip", "test_compress_src"},
	}
	requests.NewHTTPRequest(domain + "/api/fileManager/delete").
		Delete().
		SetJSONBody(deleteReq).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&apiRsp{})
}

// TestFileManagerUsage 磁盘使用统计 GET /api/fileManager/usage
func TestFileManagerUsage(t *testing.T) {
	var rsp struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data vo.FileManagerUsageRsp  `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/usage").
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

// TestFileManagerUpload 上传文件到用户空间 POST /api/fileManager/upload
// 前置条件：先完成 HFS 分片上传（create → chunk → merge），拿到 uploadId
func TestFileManagerUpload(t *testing.T) {
	// 设置 uploadId（需要先手动执行 HFS 上传流程获取，或通过 TestHfsCreateUpload/TestHfsUploadChunk/TestHfsMergeUpload 自动化获取）
	uploadId := ""
	if uploadId == "" {
		t.Skip("请设置 uploadId 变量（先完成 HFS 分片上传流程获取）")
	}

	uploadReq := vo.FileManagerUploadReq{
		UploadId: uploadId,
		Dir:      "documents",
	}
	var rsp struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Data vo.FileManagerUploadRsp   `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain + "/api/fileManager/upload").
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
