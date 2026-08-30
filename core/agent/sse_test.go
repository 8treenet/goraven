package agent

import (
	"testing"
)

func TestExtractArgumentValue(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		key      string
		expected string
	}{
		{
			name:     "string value",
			json:     `{"path": "/tmp/test.go"}`,
			key:      "path",
			expected: "/tmp/test.go",
		},
		{
			name:     "number value",
			json:     `{"count": 42}`,
			key:      "count",
			expected: "42",
		},
		{
			name:     "missing key",
			json:     `{"path": "/tmp/test.go"}`,
			key:      "missing",
			expected: "",
		},
		{
			name:     "invalid json",
			json:     `{invalid}`,
			key:      "path",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractArgumentValue(tt.json, tt.key)
			if got != tt.expected {
				t.Errorf("extractArgumentValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// newMainAgentWithTasks 构造仅含 planTaskBackend 的 MainAgent 测试桩，
// 用于 resolveTaskUpdateDisplay 反查 subject。其余字段留空，避免测试关心无关依赖。
func newMainAgentWithTasks(tasks map[string]string) *MainAgent {
	backend := newInMemoryBackend()
	baseDir := planTaskBaseDir
	for id, subject := range tasks {
		backend.files[baseDir+"/"+id+".json"] = `{"id":"` + id + `","subject":"` + subject + `"}`
	}
	return &MainAgent{planTaskBackend: backend}
}

func TestResolveTaskUpdateDisplay(t *testing.T) {
	base := toolRegistry["TaskUpdate"]
	tests := []struct {
		name      string
		arguments string
		main      *MainAgent
		wantZh    string
		wantEn    string
	}{
		{
			"in_progress with activeForm uses activeForm",
			`{"taskId":"3","status":"in_progress","activeForm":"正在安装依赖"}`,
			newMainAgentWithTasks(map[string]string{"3": "实现登录"}),
			"正在安装依赖",
			"正在安装依赖",
		},
		{
			"completed with taskId and subject resolved",
			`{"taskId":"7","status":"completed"}`,
			newMainAgentWithTasks(map[string]string{"7": "实现登录"}),
			"完成任务 #7：实现登录",
			"Complete task #7: 实现登录",
		},
		{
			"completed with taskId but subject missing falls back to #id form",
			`{"taskId":"7","status":"completed"}`,
			newMainAgentWithTasks(map[string]string{}),
			"完成任务 #7",
			"Complete task #7",
		},
		{
			"completed with taskId but nil main falls back to #id form",
			`{"taskId":"7","status":"completed"}`,
			nil,
			"完成任务 #7",
			"Complete task #7",
		},
		{
			"deleted with taskId and subject resolved",
			`{"taskId":"2","status":"deleted"}`,
			newMainAgentWithTasks(map[string]string{"2": "废弃方案"}),
			"删除任务 #2：废弃方案",
			"Delete task #2: 废弃方案",
		},
		{
			"in_progress without activeForm uses subject when available",
			`{"taskId":"4","status":"in_progress"}`,
			newMainAgentWithTasks(map[string]string{"4": "编写测试"}),
			"开始任务 #4：编写测试",
			"Start task #4: 编写测试",
		},
		{
			"no taskId falls back",
			`{"status":"completed"}`,
			newMainAgentWithTasks(map[string]string{}),
			base.ActionZh,
			base.ActionEn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveTaskUpdateDisplay(tt.arguments, tt.main)
			if !ok {
				t.Fatalf("resolveTaskUpdateDisplay() ok=false, want true")
			}
			if got.ActionZh != tt.wantZh {
				t.Errorf("ActionZh = %q, want %q", got.ActionZh, tt.wantZh)
			}
			if got.ActionEn != tt.wantEn {
				t.Errorf("ActionEn = %q, want %q", got.ActionEn, tt.wantEn)
			}
		})
	}
}

func TestExecuteDisplayResolverRegistered(t *testing.T) {
	resolver, ok := toolDisplayResolvers["execute"]
	if !ok {
		t.Fatal("execute resolver not registered")
	}
	got, matched := resolver(`{"command":"git status"}`, nil)
	if !matched {
		t.Fatalf("execute resolver matched=false, want true")
	}
	if got.ActionZh != "正在执行 Git 命令" {
		t.Errorf("ActionZh = %q, want 正在执行 Git 命令", got.ActionZh)
	}
}

func TestDeferredToolEventRegistration(t *testing.T) {
	for _, name := range []string{"execute", "TaskUpdate"} {
		if !isDeferredToolEvent(name) {
			t.Errorf("isDeferredToolEvent(%q) = false, want true", name)
		}
	}
	if isDeferredToolEvent("ls") {
		t.Errorf("isDeferredToolEvent(\"ls\") = true, want false")
	}
}

func TestInMemoryBackendLookupTaskSubject(t *testing.T) {
	backend := newInMemoryBackend()
	// 写入两份任务文件，模拟 plantask.New 持久化后的内存布局。
	backend.files["/tmp/tasks/1.json"] = `{"id":"1","subject":"实现登录","status":"in_progress"}`
	backend.files["/tmp/tasks/2.json"] = `{"id":"2","subject":"","status":"pending"}`

	cases := []struct {
		name   string
		taskID string
		want   string
		ok     bool
	}{
		{"existing with subject", "1", "实现登录", true},
		{"existing but empty subject", "2", "", false},
		{"missing task", "999", "", false},
		{"invalid json falls back", "3", "", false},
	}
	// 准备一份坏 JSON 测试 missing 分支以外，再加坏 JSON 文件。
	backend.files["/tmp/tasks/3.json"] = `{not-json}`

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := backend.lookupTaskSubject("/tmp/tasks", c.taskID)
			if got != c.want {
				t.Errorf("lookupTaskSubject(%q) got = %q, want %q", c.taskID, got, c.want)
			}
			if ok != c.ok {
				t.Errorf("lookupTaskSubject(%q) ok = %v, want %v", c.taskID, ok, c.ok)
			}
		})
	}
}
