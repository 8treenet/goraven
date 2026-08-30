package core_test

import (
	"context"
	"fmt"
	"math/rand"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/agent"
	unit_test "goraven/util/unit"
	"testing"
)

func TestSubAgentContextLoss(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: fmt.Sprint(rand.Intn(100000)),
		UserId:    "999",
		Title:     "我在太原现在的天气如何",
	}
	t.Log(repo.SaveSession(session))

	model, summaryModel := getTestDeepSeekModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(),
		`我是开发者，正在设计这个agent，我想测试下sub_agent调用工具的能力，你现在使用sub_agent调用获取天气的工具， 位置传太原即可。
	`)
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		fmt.Println(v)
	}
}
