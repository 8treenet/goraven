package service_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/service"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util"
	unit_test "goraven/util/unit"
)

func TestHFSService_GenerateURL(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *service.HFSService
	unitTest.FetchService(&service)

	t.Log(service.GenerateURL("999", "/documents/go_concurrency_plan_20260425_82731.md"))
}

func TestWriteConf(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	t.Log(config.Get().ModifyConfig("system", "initialized", "false"))
}

// TestResolveAkDownloadMyProjectPath 复现个人项目 HTML 预览流程：
// 前端 /api/myProject/:id/access 申请 dir 凭证，随后用 /api/hfs/ak/{ak}/projects/<名>/<文件> 打开预览。
// 凭证 Path 由 MyProjectService.CreateTempAccess 写入，请求路径由前端 buildAkPath + controller 归一化得到，
// 二者必须能正确匹配。
func TestResolveAkDownloadMyProjectPath(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var hfsService *service.HFSService
	unitTest.FetchService(&hfsService)
	var mpService *service.MyProjectService
	unitTest.FetchService(&mpService)

	// 用户名/userId 需全局唯一（user 表主键与 username 唯一索引），用纳秒时间戳避免与历史运行残留冲突
	userId := fmt.Sprintf("aktestuserid%d", time.Now().UnixNano())
	userName := fmt.Sprintf("aktestuser%d", time.Now().UnixNano())
	user := &po.User{UserId: userId, Username: userName, Password: "x", Status: po.UserStatusEnabled}
	if err := mpService.UserRepo.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	project := &po.UserProject{UserId: userId, ProjectName: "demo"}
	if err := mpService.Repo.Create(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mpService.Repo.Delete(project.Id) })

	workspace := config.Get().GetUserSpace(userName)
	fileAbs := filepath.Join(workspace, "projects", "demo", "docs", "index.html")
	if err := os.MkdirAll(filepath.Dir(fileAbs), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileAbs, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(workspace) })

	// dir 凭证：前端 html 预览申请当前目录（项目相对路径，如 /docs）
	rsp, err := mpService.CreateTempAccess(userId, userName, project.Id, &vo.TempAccessReq{Path: "/docs", Type: "dir"})
	if err != nil {
		t.Fatalf("CreateTempAccess(dir) error = %v", err)
	}
	// 前端 buildAkPath(file) = projects/demo/docs/index.html，controller 归一化后带前导 /
	got, err := hfsService.ResolveAkDownload(rsp.Ak, "/projects/demo/docs/index.html")
	if err != nil {
		t.Fatalf("ResolveAkDownload() error = %v, want resolve to file", err)
	}
	if got != fileAbs {
		t.Errorf("ResolveAkDownload() = %q, want %q", got, fileAbs)
	}
	if _, err := hfsService.ResolveAkDownload(rsp.Ak, "/projects/other/secret.html"); !errors.Is(err, errs.ErrTempAccessPathNotAllowed) {
		t.Errorf("out-of-scope path: want ErrTempAccessPathNotAllowed, got %v", err)
	}

	// file 凭证：office 预览申请单文件
	rspFile, err := mpService.CreateTempAccess(userId, userName, project.Id, &vo.TempAccessReq{Path: "/docs/index.html", Type: "file"})
	if err != nil {
		t.Fatalf("CreateTempAccess(file) error = %v", err)
	}
	got, err = hfsService.ResolveAkDownload(rspFile.Ak, "/projects/demo/docs/index.html")
	if err != nil {
		t.Fatalf("ResolveAkDownload() file-type error = %v, want resolve to file", err)
	}
	if got != fileAbs {
		t.Errorf("ResolveAkDownload() file-type = %q, want %q", got, fileAbs)
	}
}

func TestHFSService_CreateUploadRejectsUnsafeFileNames(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var hfsService *service.HFSService
	unitTest.FetchService(&hfsService)

	tests := []struct {
		name     string
		fileName string
	}{
		{name: "parent traversal", fileName: "../escape.txt"},
		{name: "nested traversal", fileName: "../../escape.txt"},
		{name: "absolute path", fileName: "/tmp/escape.txt"},
		{name: "forward slash", fileName: "dir/escape.txt"},
		{name: "backslash", fileName: `dir\escape.txt`},
		{name: "dot", fileName: "."},
		{name: "dot dot", fileName: ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hfsService.CreateUpload("hfs-security-user", &vo.ChunkUploadCreateReq{
				FileName:    tt.fileName,
				FileSize:    4,
				ChunkSize:   4,
				TotalChunks: 1,
			})
			if err == nil {
				t.Fatalf("CreateUpload(%q) accepted an unsafe file name", tt.fileName)
			}
		})
	}
}

func TestHFSService_MergeUploadRejectsUnsafeStoredFileName(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var hfsService *service.HFSService
	unitTest.FetchService(&hfsService)
	var hfsRepo *repository.HFSRepository
	unitTest.FetchRepository(&hfsRepo)

	tempDir := t.TempDir()
	uploadID := "hfs-merge-security-" + util.UUID()
	fileName := "../hfs-merge-escape-" + util.UUID()
	escapedPath := filepath.Join(filepath.Dir(tempDir), filepath.Base(fileName))
	if err := os.WriteFile(filepath.Join(tempDir, "chunk_0"), []byte("data"), 0644); err != nil {
		t.Fatalf("write chunk: %v", err)
	}
	if err := hfsRepo.CreateUpload(&po.ChunkUpload{
		UploadId:    uploadID,
		UserId:      "hfs-security-user",
		FileName:    fileName,
		FileSize:    4,
		ChunkSize:   4,
		TotalChunks: 1,
		TempDir:     tempDir,
		Status:      po.UploadStatusPending,
	}); err != nil {
		t.Fatalf("create upload: %v", err)
	}

	if _, err := hfsService.MergeUpload("hfs-security-user", uploadID); err == nil {
		t.Fatal("MergeUpload accepted an unsafe stored file name")
	}
	if _, err := os.Stat(escapedPath); !os.IsNotExist(err) {
		t.Fatalf("merge created escaped file %q: %v", escapedPath, err)
	}
}

func TestHFSService_CommitAssetsRejectsUnsafeStoredFileName(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var hfsService *service.HFSService
	unitTest.FetchService(&hfsService)
	var hfsRepo *repository.HFSRepository
	unitTest.FetchRepository(&hfsRepo)

	tempDir := t.TempDir()
	uploadDir := t.TempDir()
	previousUploadDir := config.Get().Paths.UploadDir
	config.Get().Paths.UploadDir = uploadDir
	t.Cleanup(func() { config.Get().Paths.UploadDir = previousUploadDir })

	uploadID := "hfs-assets-security-" + util.UUID()
	fileName := "../hfs-assets-escape-" + util.UUID()
	escapedPath := filepath.Join(filepath.Dir(tempDir), filepath.Base(fileName))
	contents := []byte("keep me")
	if err := os.WriteFile(escapedPath, contents, 0644); err != nil {
		t.Fatalf("write escaped source: %v", err)
	}
	if err := hfsRepo.CreateUpload(&po.ChunkUpload{
		UploadId:    uploadID,
		UserId:      "hfs-security-user",
		FileName:    fileName,
		FileSize:    int64(len(contents)),
		ChunkSize:   len(contents),
		TotalChunks: 1,
		TempDir:     tempDir,
		Status:      po.UploadStatusCompleted,
	}); err != nil {
		t.Fatalf("create upload: %v", err)
	}

	if _, err := hfsService.CommitAssets("hfs-security-user", uploadID); err == nil {
		t.Fatal("CommitAssets accepted an unsafe stored file name")
	}
	got, err := os.ReadFile(escapedPath)
	if err != nil {
		t.Fatalf("read escaped source: %v", err)
	}
	if string(got) != string(contents) {
		t.Fatalf("escaped source changed to %q, want %q", got, contents)
	}
}
