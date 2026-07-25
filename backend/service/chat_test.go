package service

import (
	"goraven/backend/vo"
	unit_test "goraven/util/unit"
	"testing"
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
	}))
}
