package service

import (
	"testing"

	"goraven/backend/po"
	"goraven/backend/vo"
)

func TestExtractAnswer(t *testing.T) {
	cases := []struct {
		name       string
		messages   []po.Message
		wantAnswer string
	}{
		{
			name: "常规问答对",
			messages: []po.Message{
				{RoleType: po.RoleTypeUser, Content: "每天早上整理昨日新闻"},
				{RoleType: po.RoleTypeAssistant, Content: "已为你整理完成。"},
			},
			wantAnswer: "已为你整理完成。",
		},
		{
			name: "多条助手消息取最后一条",
			messages: []po.Message{
				{RoleType: po.RoleTypeUser, Content: "问题"},
				{RoleType: po.RoleTypeAssistant, Content: "中间过程"},
				{RoleType: po.RoleTypeAssistant, Content: "最终回复"},
			},
			wantAnswer: "最终回复",
		},
		{
			name: "忽略 summary 等其他角色",
			messages: []po.Message{
				{RoleType: po.RoleTypeSummary, Content: "历史摘要"},
				{RoleType: po.RoleTypeUser, Content: "问题"},
				{RoleType: po.RoleTypeAssistant, Content: "回复"},
			},
			wantAnswer: "回复",
		},
		{
			name:       "空消息列表",
			messages:   []po.Message{},
			wantAnswer: "",
		},
		{
			name: "缺少助手回复",
			messages: []po.Message{
				{RoleType: po.RoleTypeUser, Content: "问题"},
			},
			wantAnswer: "",
		},
		{
			name: "缺少用户问题",
			messages: []po.Message{
				{RoleType: po.RoleTypeAssistant, Content: "回复"},
			},
			wantAnswer: "回复",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			answer := extractAnswer(c.messages)
			if answer != c.wantAnswer {
				t.Errorf("answer = %q, want %q", answer, c.wantAnswer)
			}
		})
	}
}

func TestFillDisplayFields(t *testing.T) {
	lookups := displayLookups{
		modelNames:   map[int]string{1: "DeepSeek - DeepSeek V3"},
		personaNames: map[int]string{2: "代码审查员"},
		projectNames: map[int]string{3: "raven-team"},
		mcpNames:     map[int]string{1: "网页抓取", 2: "数据库查询"},
		skillNames:   map[int]string{4: "数据分析", 5: "报告模板"},
	}

	t.Run("正常填充", func(t *testing.T) {
		item := vo.AutomationTaskItem{
			AIModelId:         1,
			PersonaId:         2,
			SharedProjectId:   3,
			McpIds:            "[1,2]",
			SkillIds:          "[5,4]",
			SharedProjectName: "",
		}
		fillDisplayFields(&item, lookups)
		if item.AIModelName != "DeepSeek - DeepSeek V3" {
			t.Errorf("AIModelName = %q", item.AIModelName)
		}
		if item.PersonaName != "代码审查员" {
			t.Errorf("PersonaName = %q", item.PersonaName)
		}
		if item.SharedProjectName != "raven-team" {
			t.Errorf("SharedProjectName = %q", item.SharedProjectName)
		}
		if len(item.McpNames) != 2 || item.McpNames[0] != "网页抓取" || item.McpNames[1] != "数据库查询" {
			t.Errorf("McpNames = %v", item.McpNames)
		}
		// 名称顺序跟随 skillIds 的 JSON 顺序，而非 ID 升序
		if len(item.SkillNames) != 2 || item.SkillNames[0] != "报告模板" || item.SkillNames[1] != "数据分析" {
			t.Errorf("SkillNames = %v", item.SkillNames)
		}
	})

	t.Run("ID为零或引用缺失留空", func(t *testing.T) {
		item := vo.AutomationTaskItem{
			AIModelId: 0,
			PersonaId: 99, // 映射中不存在
			McpIds:    "[1,88]",
			SkillIds:  "[]",
		}
		fillDisplayFields(&item, lookups)
		if item.AIModelName != "" || item.PersonaName != "" || item.SharedProjectName != "" {
			t.Errorf("未引用的名称应为空，得到 model=%q persona=%q project=%q", item.AIModelName, item.PersonaName, item.SharedProjectName)
		}
		if len(item.McpNames) != 1 || item.McpNames[0] != "网页抓取" {
			t.Errorf("缺失的 MCP 应跳过，McpNames = %v", item.McpNames)
		}
		if item.SkillNames == nil || len(item.SkillNames) != 0 {
			t.Errorf("空 SkillIds 应得到空切片，得到 %v", item.SkillNames)
		}
	})

	t.Run("非法JSON容错", func(t *testing.T) {
		item := vo.AutomationTaskItem{McpIds: "not-json", SkillIds: ""}
		fillDisplayFields(&item, lookups)
		if len(item.McpNames) != 0 || len(item.SkillNames) != 0 {
			t.Errorf("非法 JSON 应得到空列表，得到 %v / %v", item.McpNames, item.SkillNames)
		}
	})
}
