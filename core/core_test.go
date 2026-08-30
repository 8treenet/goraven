package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/core/agent"
	"goraven/core/iface"
	"goraven/core/provider"
	"goraven/plugins"
	unit_test "goraven/util/unit"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

func TestMsgSessionRepository_SaveSession(t *testing.T) {
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
		SessionId:    "newddd123",
		UserId:       "999",
		Title:        "北京现在的天气如何",
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
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)
	run, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	err = run.Query(context.Background(), "太原现在的天气如何")
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

func getTestTools() tool.BaseTool {
	weatherTool, _ := utils.InferTool("GetWeather", "获取气温需要传入位置", GetWeatherFunc)
	return weatherTool
}

func getParallelTestTools() []tool.BaseTool {
	weatherTool, _ := utils.InferTool("GetWeather", "获取气温需要传入位置", GetWeatherFuncSlow)
	return []tool.BaseTool{weatherTool}
}

func GetWeatherFunc(_ context.Context, content WeatherParams) (string, error) {
	return "晴天，28度", nil
}

func GetWeatherFuncSlow(_ context.Context, content WeatherParams) (string, error) {
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("%s天气：晴天，28度", content.LOC), nil
}

func TestQueryParallelMessages(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	sessionId := "test_parallel_qwen"

	messages, err := repo.GetChatMessages(sessionId)
	if err != nil {
		t.Logf("Error: %v", err)
		return
	}

	t.Logf("Found %d messages in session %s", len(messages), sessionId)
	for i, msg := range messages {
		t.Logf("=== Message %d ===", i)
		t.Logf("  RoleType: %s", msg.RoleType)
		t.Logf("  Content: %s", msg.Content)
		t.Logf("  Ext: %s", msg.Ext)
		t.Logf("  Tool: %d", msg.Tool)
	}
}

type WeatherParams struct {
	LOC string `json:"loc" jsonschema_description:"位置"`
}

// go test -test.fullpath=true -timeout 250s -run ^TestSystemAgent$ goraven/core -v -count=1
func TestSystemAgent(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.DailyStatsRepository
	unitTest.FetchRepository(&repo)

	model, _ := getTestDeepSeekModel()
	sa, _ := agent.NewSystemAgent(agent.SystemAgentParam{
		ChatModel:      model,
		SysCfg:         repository.NewDefaultSystemConfig(),
		UserId:         "999",
		DailyStatsRepo: repo,
	})
	runner, err := sa.NewRunner(context.Background())
	if err != nil {
		panic(err)
	}
	reuslt, err := runner.Run(context.Background(), sa.BuildInstallSkillMessage("bloomberg-headlines"))
	if err != nil {
		panic(err)
	}
	t.Log(reuslt)
}

type SystemTask struct {
}

func (task *SystemTask) Complete(replyContent string, err error) {
	fmt.Println("Complete:", replyContent, err)
}

func (task *SystemTask) GetUserId() string {
	return "999"
}

func TestEnablePlanTask(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "ff123d",
		UserId:    "999",
		Title:     "测试计划任务模式",
	}
	t.Logf("Session ID: %s", session.SessionId)
	if err := repo.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	model, summaryModel := getTestDeepSeekModel()
	param := agent.AgentParam{Session: session, MsgRepo: repo, ChatModel: model, SysCfg: repository.NewDefaultSystemConfig()}
	mainAgent, _ := agent.NewMainAgent(param)
	mainAgent.AddTool(getTestTools())
	mainAgent.SetFlashModel(summaryModel)

	runner, err := mainAgent.NewRunner(context.Background())
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	query := "我在做task的工具测试，你要覆盖TaskCreate，TaskGet、TaskUpdate、TaskList这4个工具，请帮我制定一个学习 Go 语言并发编程的计划，包括学习 goroutine、channel、sync 包等内容，并跟踪进度，我要markdown格式，写到我的文档目录里"
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
}

func getTestDeepSeekModel() (model iface.BaseChatModel, summaryModel iface.BaseChatModel) {
	pv, err := provider.GetProviderByName(provider.DeepseekProviderName, provider.ProviderConfig{APIKey: "sk-"})
	if err != nil {
		panic(err)
	}
	model, err = pv.CreateModel("deepseek-v4-flash", true, 1000000)
	if err != nil {
		panic(err)
	}
	summaryModel, err = pv.CreateModel("deepseek-v4-flash", false, 1000000)
	if err != nil {
		panic(err)
	}

	return model, summaryModel
}

func getTestFreeModel() (model iface.BaseChatModel, summaryModel iface.BaseChatModel) {
	pv, err := provider.GetProviderByName(provider.OpenrouterProviderName, provider.ProviderConfig{APIKey: "sk-or-v1-"})
	if err != nil {
		panic(err)
	}

	model, err = pv.CreateModel("tencent/hy3-preview:free", true, 1000000)
	if err != nil {
		panic(err)
	}
	summaryModel, err = pv.CreateModel("tencent/hy3-preview:free", false, 1000000)
	if err != nil {
		panic(err)
	}
	return model, summaryModel
}

func getTestDoubaoModel() (model iface.BaseChatModel, summaryModel iface.BaseChatModel) {
	pv, err := provider.GetProviderByName(provider.OpenAICompatibleProviderName, provider.ProviderConfig{APIKey: "24bcc56b-8a36-4dfd-a026-bed9f256e47c", BaseURL: "https://ark.cn-beijing.volces.com/api/v3"})
	if err != nil {
		panic(err)
	}
	model, err = pv.CreateModel("doubao-seed-2-0-pro-260215", true, 1000000)
	if err != nil {
		panic(err)
	}
	summaryModel, err = pv.CreateModel("doubao-seed-2-0-pro-260215", false, 1000000)
	if err != nil {
		panic(err)
	}
	return
}

func TestCoder(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	var repo *repository.MsgSessionRepository
	unitTest.FetchRepository(&repo)

	session := &po.Session{
		SessionId: "dddoo9123",
		UserId:    "999",
		Title:     "code123",
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
	err = run.Query(context.Background(), "但是这个网站打开没有样式？我浏览器直接打开index.html。 是什么原因？")
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
