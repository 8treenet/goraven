package repository_test

import (
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/util"
	unit_test "goraven/util/unit"
	"testing"
	"time"
)

func TestMsgSessionRepository_SaveChatMessage(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	msg := &po.Message{
		SessionId:             "0011",
		Timestamp:             util.Millisecond(),
		ContextState:          0,
		Content:               "大家都好",
		RoleType:              po.RoleTypeUser,
		PromptTokensCount:     11,
		CompletionTokensCount: 22,
		Duration:              3,
		AsstError:             "",
		Created:               time.Now(),
	}
	t.Log(repo.SaveChatMessage(msg.SessionId, msg))
}

func TestMsgSessionRepository_GetChatMessages(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)
	data, _ := repo.GetChatMessages("5577")
	t.Log(unit_test.JsonLog(data))
}

func TestMsgSessionRepository_SaveSession(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "0011",
		UserId:    "999",
		Title:     "你好",
	}
	t.Log(repo.SaveSession(session))
}

func TestMsgSessionRepository_UpdateSessionStatus(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	t.Log(repo.UpdateSessionStatus("0011", 1))
}

func TestMsgSessionRepository_AddSessionCount(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	t.Log(repo.AddSessionTokens("0011", 100, 200, 10))
}

func TestMsgSessionRepository_GetSession(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)
	data, _ := repo.GetSession("5577")
	t.Log(unit_test.JsonLog(data))
}
