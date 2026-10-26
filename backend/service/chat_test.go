package service

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/util"
	unit_test "goraven/util/unit"
)

func TestChatService_resolveSession(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)
	sid := "ddd123"
	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession("999", &vo.ChatReq{
		SessionId: &sid,
		AIModelId: 1,
		// PersonaId: new(int),
		McpIds:    []int{1, 2},
		SkillIds:  []int{9, 10},
		Reasoning: 1,
	})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(session))
	t.Log(unit_test.JsonLog(persona))
	t.Log(unit_test.JsonLog(mcpIds))
	t.Log(unit_test.JsonLog(skillIds))
	t.Log(unit_test.JsonLog(userRole))
}

func TestChatService_resolveSession2(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)

	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession("999", &vo.ChatReq{
		AIModelId: 1,
		Content:   "啊多久啊就是觉得啥的",
		// PersonaId: new(int),
		McpIds:    []int{1, 2},
		SkillIds:  []int{9, 10},
		Reasoning: 1,
	})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(session))
	t.Log(unit_test.JsonLog(persona))
	t.Log(unit_test.JsonLog(mcpIds))
	t.Log(unit_test.JsonLog(skillIds))
	t.Log(unit_test.JsonLog(userRole))
}

func TestChatService_resolveSession3(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)
	personaId := 3
	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession("999", &vo.ChatReq{
		//SessionId: &sid,
		AIModelId: 1,
		Content:   "啊多久啊就是觉得啥的",
		PersonaId: &personaId,
		Reasoning: 1,
	})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(session))
	t.Log(unit_test.JsonLog(persona))
	t.Log(unit_test.JsonLog(mcpIds))
	t.Log(unit_test.JsonLog(skillIds))
	t.Log(unit_test.JsonLog(userRole))
}

func TestChatService_resolveSession4(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)
	sid := "15770c0341814e0b984ab3d63d4c3ff1"
	session, persona, mcpIds, skillIds, userRole, err := service.resolveSession("999", &vo.ChatReq{
		SessionId: &sid,
		AIModelId: 1,
		// PersonaId: new(int),
		McpIds:    []int{1},
		SkillIds:  []int{9},
		Reasoning: 1,
	})
	if err != nil {
		panic(err)
	}
	t.Log(unit_test.JsonLog(session))
	t.Log(unit_test.JsonLog(persona))
	t.Log(unit_test.JsonLog(mcpIds))
	t.Log(unit_test.JsonLog(skillIds))
	t.Log(unit_test.JsonLog(userRole))
}

func TestChatService_processAttachments(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)
	t.Log(service.processAttachments("999", []string{
		"017f992a529344ac84f9574173d7a126",
		"3ecc09a8f5164af99afcb0c8ab230311",
		"d67dd4240c5a4a4ca2926400c89f5df7",
	}, false, nil))
}

// TestChatService_processOneAttachmentDirectImage 直接模式图片附件应返回工作区相对路径的 MediaItem。
// 回归：util.CompressImage 返回的是绝对路径，被直接赋给 relPath 后 GenerateURL 拼出双重前缀导致
// “文件不存在”，直接模式图片附件全部回退成标签。
func TestChatService_processOneAttachmentDirectImage(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var service *ChatService
	unitTest.FetchService(&service)
	var hfs *HFSService
	unitTest.FetchService(&hfs)
	var userRepo *repository.UserRepository
	unitTest.FetchRepository(&userRepo)
	var hfsRepo *repository.HFSRepository
	unitTest.FetchRepository(&hfsRepo)

	userId := "directimguserid" + util.UUID()
	userName := "directimguser" + util.UUID()
	if err := userRepo.CreateUser(&po.User{UserId: userId, Username: userName, Password: "x", Status: po.UserStatusEnabled}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	service.Worker.Store().Set(infra.UserNameStoreKey, userName)

	tempDir := t.TempDir()
	fileName := "big.jpg"
	f, err := os.Create(filepath.Join(tempDir, fileName))
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := jpeg.Encode(f, image.NewRGBA(image.Rect(0, 0, 3000, 3000)), &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode image: %v", err)
	}
	f.Close()

	uploadId := util.UUID()
	if err := hfsRepo.CreateUpload(&po.ChunkUpload{
		UploadId:    uploadId,
		UserId:      userId,
		FileName:    fileName,
		FileSize:    100,
		ChunkSize:   0,
		TotalChunks: 1,
		TempDir:     tempDir,
		Status:      po.UploadStatusCompleted,
	}); err != nil {
		t.Fatalf("create upload: %v", err)
	}

	tag, media, err := service.processOneAttachment(userId, uploadId, true, hfs)
	if err != nil {
		t.Fatalf("processOneAttachment: %v", err)
	}
	if media == nil {
		t.Fatalf("expected direct media item, got tag=%q", tag)
	}
	if media.URL == "" {
		t.Fatalf("expected non-empty URL, got media=%+v", media)
	}
	if !strings.HasPrefix(media.Path, "/temp/") || strings.Contains(media.Path, "/data/users/") {
		t.Fatalf("MediaItem.Path must be workspace-relative like /temp/xxx.jpg, got %q", media.Path)
	}
}
