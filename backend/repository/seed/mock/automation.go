package mock

import (
	"fmt"
	"time"

	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/config"
)

// 自动化任务演示数据（PreviewUser 在线展示用）
// 任务/需求/回复文案迁移自前端 mocks/automation.ts；
// 模型/角色/MCP/技能的 ID 与名称对齐 preview 种子库（02/03 SQL），
// 与站点其它演示页面（仪表盘排行等）保持一致。
// 文案区分中英文：按 config system.language 取值（bilingual），
// 模型/技能名为英文标识，语言无关。
// 日期基于当前时间动态生成，数值固定不变，刷新不跳动。

// bilingual 中英文文案对，按系统语言（config system.language）取值
type bilingual struct{ zh, en string }

func (b bilingual) s() string {
	if config.Get().GetLanguage() == "en" {
		return b.en
	}
	return b.zh
}

// bilingualList 双语名称列表转当前语言字符串列表，保证非 nil
func bilingualList(items []bilingual) []string {
	list := make([]string, 0, len(items))
	for _, b := range items {
		list = append(list, b.s())
	}
	return list
}

// automationTaskMock 单个演示任务的静态配置
type automationTaskMock struct {
	id              int
	title           bilingual
	requirement     bilingual
	execType        uint8
	runAt           *time.Time // 单次执行时间（ExecType=1）
	intervalMinutes int        // 执行间隔分钟数（ExecType=2）
	fixedTime       string     // 固定时间 HH:MM（ExecType=3/4）
	weekday         uint8      // 每周执行日（ExecType=4）
	nextRunAt       time.Time
	createdAt       time.Time
	updatedAt       time.Time
	mcpIds          string
	mcpNames        []bilingual
	skillIds        string
	skillNames      []string // 技能名为英文标识，语言无关
	aiModelId       int
	aiModelName     string // 模型名对齐种子库 displayName，语言无关
	personaId       int
	personaName     bilingual
	status          uint8
	// 执行记录生成参数：共 count 条，每天/每周 intervalDays 天一条，固定在 hour 点
	execCount        int
	execHour         int
	execIntervalDays int
}

// automationTaskSpecs 构建演示任务列表（时间字段随当前时间生成，按创建时间倒序与真实服务一致）
func automationTaskSpecs() []automationTaskMock {
	task4RunAt := automationDaysAgoAt(2, 6)
	return []automationTaskMock{
		{
			id: 1,
			title: bilingual{
				zh: "每日站会总结",
				en: "Daily Standup Summary",
			},
			requirement: bilingual{
				zh: "每天早上 9 点汇总昨日项目进展、今日计划与风险项，整理成结构化站会纪要，并通过邮件发送给团队成员。若存在未关闭的风险项，需要单独列出并给出建议处理方案。",
				en: "Every day at 9 AM, summarize yesterday's project progress, today's plan, and risk items into structured standup minutes, and email them to team members. Open risk items must be listed separately with suggested resolutions.",
			},
			execType: po.AutomationExecTypeDaily, fixedTime: "09:00",
			nextRunAt: automationDaysAgoAt(-1, 9),
			createdAt: automationDaysAgoAt(18, 2).Add(30 * time.Minute),
			updatedAt: automationDaysAgoAt(0, 1),
			mcpIds:    "[1,12]",
			mcpNames: []bilingual{
				{zh: "文件系统", en: "Filesystem"},
				{zh: "邮件助手", en: "Email Assistant"},
			},
			skillIds: "[5,9,10]", skillNames: []string{"essay-writer", "data-viz-assistant", "presentation-builder"},
			aiModelId: 1, aiModelName: "deepseek-v4-pro",
			status:    po.AutomationStatusEnabled,
			execCount: 28, execHour: 9, execIntervalDays: 1,
		},
		{
			id: 4,
			title: bilingual{
				zh: "季度经营报告生成",
				en: "Quarterly Business Report",
			},
			requirement: bilingual{
				zh: "汇总本季度经营数据（收入、成本、活跃度），结合目标完成度生成经营分析报告，输出 PDF 并发送给管理层。",
				en: "Aggregate this quarter's business data (revenue, costs, activity), generate a business analysis report against target completion, export PDF and send to management.",
			},
			execType: po.AutomationExecTypeOnce, runAt: &task4RunAt, nextRunAt: task4RunAt,
			createdAt: automationDaysAgoAt(23, 3),
			updatedAt: task4RunAt.Add(5 * time.Minute),
			skillIds:  "[3,4,9,10]", skillNames: []string{"financial-report-analyzer", "tax-helper", "data-viz-assistant", "presentation-builder"},
			personaId: 3, personaName: bilingual{zh: "财务分析助手", en: "Financial Analyst"},
			status:    po.AutomationStatusDone,
			execCount: 1, execHour: 14, execIntervalDays: 1,
		},
		{
			id: 3,
			title: bilingual{
				zh: "竞品动态监控",
				en: "Competitor Monitoring",
			},
			requirement: bilingual{
				zh: "每 30 分钟抓取竞品官网动态与行业新闻，汇总变更点，若出现价格调整或重大版本发布则立即邮件提醒。",
				en: "Every 30 minutes, crawl competitor websites and industry news, summarize changes, and email alerts immediately on price adjustments or major version releases.",
			},
			execType: po.AutomationExecTypeInterval, intervalMinutes: 30,
			createdAt: automationDaysAgoAt(27, 8),
			updatedAt: automationDaysAgoAt(8, 9).Add(30 * time.Minute),
			mcpIds:    "[4,7,1]",
			mcpNames: []bilingual{
				{zh: "Brave 搜索", en: "Brave Search"},
				{zh: "文档编辑器", en: "Document Editor"},
				{zh: "文件系统", en: "Filesystem"},
			},
			skillIds: "[9]", skillNames: []string{"data-viz-assistant"},
			aiModelId: 8, aiModelName: "qwen-3.7max",
			status:    po.AutomationStatusDisabled,
			execCount: 5, execHour: 14, execIntervalDays: 1,
		},
		{
			id: 2,
			title: bilingual{
				zh: "每周代码审查报告",
				en: "Weekly Code Review Report",
			},
			requirement: bilingual{
				zh: "每周一上午 10 点拉取本周合并的代码变更，执行代码审查，输出风险点清单与改进建议，按仓库归档保存。",
				en: "Every Monday at 10 AM, pull this week's merged code changes, run code review, produce a risk list with improvement suggestions, and archive them by repository.",
			},
			execType: po.AutomationExecTypeWeekly, fixedTime: "10:00", weekday: 1,
			nextRunAt: automationNextWeekdayAt(time.Monday, 10),
			createdAt: automationDaysAgoAt(39, 5),
			updatedAt: automationDaysAgoAt(3, 2),
			mcpIds:    "[3]",
			mcpNames: []bilingual{
				{zh: "GitHub", en: "GitHub"},
			},
			skillIds: "[1]", skillNames: []string{"go-project-scaffold"},
			personaId: 2, personaName: bilingual{zh: "代码审查员", en: "Code Reviewer"},
			status:    po.AutomationStatusEnabled,
			execCount: 12, execHour: 10, execIntervalDays: 7,
		},
	}
}

// buildAutomationTaskItem 静态配置转列表条目 VO
// 名称列表统一输出空数组而非 null，避免前端对 null 取 length 报错
func buildAutomationTaskItem(t automationTaskMock) vo.AutomationTaskItem {
	skillNames := t.skillNames
	if skillNames == nil {
		skillNames = []string{}
	}
	return vo.AutomationTaskItem{
		Id:              t.id,
		Title:           t.title.s(),
		UserId:          config.Get().Behavior.PreviewUser,
		ExecType:        t.execType,
		RunAt:           t.runAt,
		IntervalMinutes: t.intervalMinutes,
		FixedTime:       t.fixedTime,
		Weekday:         t.weekday,
		McpIds:          t.mcpIds,
		SkillIds:        t.skillIds,
		AIModelId:       t.aiModelId,
		PersonaId:       t.personaId,
		AIModelName:     t.aiModelName,
		PersonaName:     t.personaName.s(),
		McpNames:        bilingualList(t.mcpNames),
		SkillNames:      skillNames,
		NextRunAt:       t.nextRunAt,
		Status:          t.status,
		Created:         t.createdAt,
		Updated:         t.updatedAt,
	}
}

// automationAnswerMock 精写的执行回复内容
type automationAnswerMock struct {
	answer bilingual
}

// automationAnswerFixed 固定回复表，key 为 "taskId_executionId"
var automationAnswerFixed = map[string]automationAnswerMock{
	"1_128": {
		answer: bilingual{
			zh: "【站会纪要】昨日：完成报表模块联调，接口通过率 100%；今日：启动数据看板开发，预计本周五提测；风险：MCP 数据库连接偶发超时，已反馈管理员跟进。",
			en: "[Standup Minutes] Yesterday: completed report module integration testing with 100% API pass rate; Today: started data dashboard development, QA submission expected Friday; Risks: occasional MCP database connection timeouts, reported to admin for follow-up.",
		},
	},
	"1_127": {
		answer: bilingual{
			zh: "【站会纪要】昨日：修复权限模块越权漏洞并发布；今日：推进报表导出功能，下午评审数据模型；风险：无未关闭风险项。",
			en: "[Standup Minutes] Yesterday: fixed the privilege escalation vulnerability in the permission module and released; Today: advancing report export feature, data model review this afternoon; Risks: no open risk items.",
		},
	},
	"2_212": {
		answer: bilingual{
			zh: "【审查报告】本周共 14 个 PR 合并。风险：订单服务存在一处 N+1 查询，建议加索引；工具库新增全局单例未做并发保护，已提单跟进。",
			en: "[Review Report] 14 PRs merged this week. Risks: an N+1 query in the order service, index recommended; a new global singleton in the utility library lacks concurrency protection, ticket filed for follow-up.",
		},
	},
	"4_401": {
		answer: bilingual{
			zh: "【经营报告】Q3 收入环比增长 12.4%，成本率下降 2.1 个百分点；活跃用户增长 8.7%。目标完成度 92%，详见报告附件。",
			en: "[Business Report] Q3 revenue grew 12.4% QoQ and cost ratio dropped 2.1 percentage points; active users grew 8.7%. Goal completion at 92%, see the report attachment for details.",
		},
	},
}

// automationFallbackAnswer 兜底回复模板（%s 为任务标题）
var automationFallbackAnswer = bilingual{
	zh: "已执行任务「%s」，执行结果已生成，可展开执行记录查看本次输出。",
	en: "Task \"%s\" has been executed and the results are ready. Expand the execution record to view this run's output.",
}

// BuildAutomationTasks 构建自动化任务分页列表 mock 数据，status 非 nil 时按状态过滤
func BuildAutomationTasks(page, pageSize int, status *uint8) *infra.PageResponse {
	specs := automationTaskSpecs()
	items := make([]vo.AutomationTaskItem, 0, len(specs))
	for _, s := range specs {
		if status != nil && s.status != *status {
			continue
		}
		items = append(items, buildAutomationTaskItem(s))
	}
	return paginateAutomation(items, page, pageSize)
}

// BuildAutomationTask 构建自动化任务详情 mock 数据（含需求描述），不存在返回 nil
func BuildAutomationTask(id int) *vo.AutomationTaskDetail {
	for _, s := range automationTaskSpecs() {
		if s.id == id {
			return &vo.AutomationTaskDetail{
				AutomationTaskItem: buildAutomationTaskItem(s),
				Requirement:        s.requirement.s(),
			}
		}
	}
	return nil
}

// BuildAutomationExecutions 构建任务执行记录分页 mock 数据（新→旧），任务不存在返回空页
func BuildAutomationExecutions(taskId, page, pageSize int) *infra.PageResponse {
	spec, ok := findAutomationTask(taskId)
	if !ok {
		return paginateAutomation([]vo.AutomationExecutionItem{}, page, pageSize)
	}
	items := make([]vo.AutomationExecutionItem, 0, spec.execCount)
	for i := 0; i < spec.execCount; i++ {
		started := automationDaysAgoAt(i*spec.execIntervalDays, spec.execHour)
		duration := time.Duration(3+(i*37)%4)*time.Minute + 40*time.Second
		items = append(items, vo.AutomationExecutionItem{
			Id:         taskId*100 + spec.execCount - i,
			StartedAt:  started,
			FinishedAt: started.Add(duration),
		})
	}
	return paginateAutomation(items, page, pageSize)
}

// BuildAutomationAnswer 构建执行回复 mock 数据：优先取精写内容，
// 其余执行记录按通用回复模板生成
func BuildAutomationAnswer(taskId, executionId int) *vo.AutomationAnswerRsp {
	if ans, ok := automationAnswerFixed[fmt.Sprintf("%d_%d", taskId, executionId)]; ok {
		return &vo.AutomationAnswerRsp{Answer: ans.answer.s()}
	}
	spec, ok := findAutomationTask(taskId)
	if !ok {
		return &vo.AutomationAnswerRsp{}
	}
	return &vo.AutomationAnswerRsp{
		Answer: fmt.Sprintf(automationFallbackAnswer.s(), spec.title.s()),
	}
}

// findAutomationTask 按 ID 查找演示任务配置
func findAutomationTask(id int) (automationTaskMock, bool) {
	for _, s := range automationTaskSpecs() {
		if s.id == id {
			return s, true
		}
	}
	return automationTaskMock{}, false
}

// paginateAutomation 演示数据分页（新→旧列表切片）
func paginateAutomation[T any](items []T, page, pageSize int) *infra.PageResponse {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPage := (len(items) + pageSize - 1) / pageSize
	if totalPage < 1 {
		totalPage = 1
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return &infra.PageResponse{
		List:       items[start:end],
		TotalPage:  totalPage,
		TotalCount: len(items),
		Page:       page,
		PageSize:   pageSize,
	}
}

// automationDaysAgoAt daysAgo 天前（负数为未来）的 hour 点整（本地时区）
func automationDaysAgoAt(daysAgo, hour int) time.Time {
	now := time.Now()
	d := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	return d.AddDate(0, 0, -daysAgo)
}

// automationNextWeekdayAt 下一个目标星期（今天为该星期且时刻已过则推到下周）的 hour 点整
func automationNextWeekdayAt(weekday time.Weekday, hour int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	diff := (int(weekday) - int(now.Weekday()) + 7) % 7
	if diff == 0 && !next.After(now) {
		diff = 7
	}
	return next.AddDate(0, 0, diff)
}
