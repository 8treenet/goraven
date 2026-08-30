package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/agent"
	"goraven/core/tools"
	"goraven/plugins"
	unit_test "goraven/util/unit"
	"testing"
	"time"
)

func TestVisualUnderstand(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "fffa11",
		UserId:    "999",
		Title:     "测试视觉理解",
	}
	if err := repo.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	model, summaryModel := getTestFreeModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.SetFlashModel(summaryModel)
	runner, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	query := "我是goraven的开发者，现在要测试下visual_understand的工具。你在我的目录里找个图片文件使用下，然后告诉我图片的内容"
	err = runner.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	c := runner.StartFetch()
	for v := range c {
		fmt.Println(v)
		if v.Type == agent.SSEEventTypeTool {
			jdata, _ := json.Marshal(v.Tool)
			fmt.Println(string(jdata))
		}
	}
	t.Log(runner.GetReplyContent())
}

func TestWebFetch(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "1113213",
		UserId:    "999",
		Title:     "测试webfetch",
	}
	t.Log(repo.SaveSession(session))

	model, summaryModel := getTestFreeModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "https://www.deepseek.com  查看下这个网站")
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		fmt.Println(v)
	}

	fmt.Println(run.GetReplyContent())
}

func TestMCPSearchFetch(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "mcptest",
		UserId:    "999",
		Title:     "测试mcptest",
	}

	mcpObj := tools.MCP{
		Transport:    tools.MCPTransportHTTP,
		Name:         "WebSearch",
		HttpURL:      "https://dashscope.aliyuncs.com/api/v1/mcps/WebSearch/mcp",
		HttpHeader:   `{"Authorization":"Bearer "}`,
		StdioCommand: "",
		StdioENV:     []string{},
		StdioArgs:    []string{},
	}

	model, summaryModel := getTestDeepSeekModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddMCP(mcpObj)
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "使用bailian_web_search 帮我查下美国和伊朗的最新战况")
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		fmt.Println(v)
	}

	fmt.Println(run.GetReplyContent())
}

func TestGetSkill(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()
	plugins.RegisterAll()

	var repo *repository.MsgSessionRepository
	var skillRepo *repository.SkillRepository
	unitTest.FetchRepository(&repo)
	unitTest.FetchRepository(&skillRepo)

	session := &po.Session{
		SessionId: "ddddasd",
		UserId:    "999",
		Title:     "查看某个技能",
	}
	t.Log(repo.SaveSession(session))

	model, summaryModel := getTestDeepSeekModel()

	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SystemSkillProvider: skillRepo}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "我现在是raven的作者，我正在测试系统，你现在能看到goraven-test-skill这个技能吗？调用下，帮我完成测试")
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		fmt.Println("sse:", v)
		if v.Type == agent.SSEEventTypeEnd {
			fmt.Println("test over")
		}
	}
	t.Log(run.GetReplyContent())
}

// go test -test.fullpath=true -timeout 350s -run ^TestCLI$ goraven/core -v -count=1
func TestCLI(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		UserId: "999",
		Title:  "测试cli",
	}
	if err := repo.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	model, summaryModel := getTestDeepSeekModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.SetFlashModel(summaryModel)
	runner, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	query := "我是goraven的开发者，现在要测试下拷贝文件、创建目录、删除文件、移动文件。 可以拿我/documents目录下的文件做测试"
	err = runner.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	c := runner.StartFetch()
	for v := range c {
		fmt.Println(v)
		if v.Type == agent.SSEEventTypeTool {
			jdata, _ := json.Marshal(v.Tool)
			fmt.Println(string(jdata))
		}
	}
	t.Log(runner.GetReplyContent())
}

func TestSkill(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()
	plugins.RegisterAll()

	var repo *repository.MsgSessionRepository
	var skillRepo *repository.SkillRepository
	var ssrepo *repository.SystemSettingRepository
	var dailyStatsRepo *repository.DailyStatsRepository
	unitTest.FetchRepository(&ssrepo)
	unitTest.FetchRepository(&repo)
	unitTest.FetchRepository(&skillRepo)
	unitTest.FetchRepository(&dailyStatsRepo)

	session := &po.Session{
		SessionId:    "11ds",
		UserId:       "999",
		Title:        "给我查下最新的彭博社的头条",
		LastChatTime: time.Now(),
		Created:      time.Now(),
	}
	//t.Log(repo.SaveSession(session))
	sconf, err := ssrepo.LoadConfig()
	if err != nil {
		panic(err)
	}

	model, summaryModel := getTestDeepSeekModel()

	param := agent.AgentParam{
		Session: session,
		MsgRepo: repo, ChatModel: model,
		SystemSkillProvider: skillRepo,
		SysCfg:              sconf,
		DailyStatsRepo:      dailyStatsRepo,
	}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "那么你觉得 中国为打击芬太尼滥用限制对美化学品出口是为了什么？")
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		if v.Tool != nil {
			fmt.Println("sse: ", *v.Tool)
		} else {
			fmt.Println("sse: ", v)
		}
		if v.Type == agent.SSEEventTypeEnd {
			fmt.Println("test over")
		}
	}
	t.Log(run.GetReplyContent())
}

func TestVisual(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	var skillRepo *repository.SkillRepository
	var ssrepo *repository.SystemSettingRepository
	var dailyStatsRepo *repository.DailyStatsRepository
	unitTest.FetchRepository(&ssrepo)
	unitTest.FetchRepository(&repo)
	unitTest.FetchRepository(&skillRepo)
	unitTest.FetchRepository(&dailyStatsRepo)

	session := &po.Session{
		SessionId:    "dudduddutu",
		UserId:       "999",
		Title:        "查看图片",
		LastChatTime: time.Now(),
		Created:      time.Now(),
	}
	//t.Log(repo.SaveSession(session))
	sconf, err := ssrepo.LoadConfig()
	if err != nil {
		panic(err)
	}

	model, summaryModel := getTestDeepSeekModel()
	dmodel, _ := getTestDoubaoModel()

	param := agent.AgentParam{
		Session: session,
		MsgRepo: repo, ChatModel: model,
		SystemSkillProvider: skillRepo,
		SysCfg:              sconf,
		DailyStatsRepo:      dailyStatsRepo,
	}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)
	mainAgent.SetVisualModel(dmodel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "帮我读取下这个图片的信息，位置在/images/zxg.jpg")
	if err != nil {
		panic(err)
	}
	c := run.StartFetch()
	for v := range c {
		fmt.Println("sse:", v)
		if v.Type == agent.SSEEventTypeEnd {
			fmt.Println("test over")
		}
	}
	t.Log(run.GetReplyContent())
}
