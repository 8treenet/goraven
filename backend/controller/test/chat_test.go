package test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"goraven/backend/vo"
	"strings"
	"testing"
	"time"

	"github.com/8treenet/freedom/infra/requests"
)

func init() {
	requests.SetHTTPClient(requests.NewHTTPClient(1000*time.Second, 5*time.Second))
}

// TestChat 发起对话 POST /api/chat
// go test -test.fullpath=true -timeout 1150s -run ^TestChat$ goraven/backend/controller/test -v -count=1
func TestChat(t *testing.T) {
	content := "我在学习golang语言,给我一出一份学习计划"
	modelId := 1
	req := vo.ChatReq{
		Content:   content,
		AIModelId: modelId,
	}

	var rsp struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data vo.ChatRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/chat").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("chat error:", httpResp.Error)
	}
	t.Logf("chat response: code=%d msg=%s sessionId=%s", rsp.Code, rsp.Msg, rsp.Data.SessionId)

	chatsteam(t, rsp.Data.SessionId)
}

// TestChat 发起对话 POST /api/chat
// go test -test.fullpath=true -timeout 150s -run ^TestPersonaChat$ goraven/backend/controller/test -v -count=1
func TestPersonaChat(t *testing.T) {
	content := "分析下这个项目"
	modelId := 2
	//personaId := 2
	req := vo.ChatReq{
		Content:   content,
		AIModelId: modelId,
		Project:   "golang-learning",
	}

	var rsp struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data vo.ChatRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/chat").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("chat error:", httpResp.Error)
	}
	t.Logf("chat response: code=%d msg=%s sessionId=%s", rsp.Code, rsp.Msg, rsp.Data.SessionId)

	chatsteam(t, rsp.Data.SessionId)
}

// TestChat 发起对话 POST /api/chat
// go test -test.fullpath=true -timeout 1150s -run ^TestPersonaSkillChat$ goraven/backend/controller/test -v -count=1
func TestPersonaSkillChat(t *testing.T) {
	content := "给我查下最新的彭博社d头条"
	modelId := 1
	personaId := 3
	req := vo.ChatReq{
		Content:   content,
		AIModelId: modelId,
		PersonaId: &personaId,
	}

	var rsp struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data vo.ChatRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/chat").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("chat error:", httpResp.Error)
	}
	t.Logf("chat response: code=%d msg=%s sessionId=%s", rsp.Code, rsp.Msg, rsp.Data.SessionId)

	chatsteam(t, rsp.Data.SessionId)
}

// go test -test.fullpath=true -timeout 350s -run ^TestMCPChat$ goraven/backend/controller/test -v -count=1
func TestMCPChat(t *testing.T) {
	content := "我现在加载了github的mcp，我想测试下是否正常。你调用一下试试"
	modelId := 1
	req := vo.ChatReq{
		Content:   content,
		AIModelId: modelId,
		McpIds:    []int{3},
	}

	var rsp struct {
		Code int        `json:"code"`
		Msg  string     `json:"msg"`
		Data vo.ChatRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/chat").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("chat error:", httpResp.Error)
	}
	t.Logf("chat response: code=%d msg=%s sessionId=%s", rsp.Code, rsp.Msg, rsp.Data.SessionId)

	chatsteam(t, rsp.Data.SessionId)
}

// TestChatStop 停止生成 POST /api/chat/stop
func TestChatStop(t *testing.T) {
	stopReq := vo.ChatStopReq{
		SessionId: "2e6b6788ad124f29b3146438de169bb4",
	}
	var stopRsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	stopResp := requests.NewHTTPRequest(domain+"/api/chat/stop").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(stopReq).
		ToJSON(&stopRsp)
	if stopResp.Error != nil {
		t.Fatal("stop error:", stopResp.Error)
	}
	t.Logf("stop response: code=%d msg=%s", stopRsp.Code, stopRsp.Msg)
}

// chatsteam SSE 流输出
// reasoning 和 content 实时追加输出，tool 和 retry 单行输出
func chatsteam(t *testing.T, sessionId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", domain+"/api/chat/"+sessionId+"/stream", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("stream error:", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("stream content-type: %s, status: %d\n", contentType, resp.StatusCode)
	if contentType != "text/event-stream" {
		t.Errorf("expected content-type text/event-stream, got %s", contentType)
	}

	scanner := bufio.NewScanner(resp.Body)
	var (
		eventType string
		dataLine  string
		lastBlock string
	)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if dataLine == "" {
				continue
			}
			dataStr := strings.TrimPrefix(dataLine, "data: ")
			dataStr = strings.TrimPrefix(dataStr, "data:")

			switch eventType {
			case "reasoning":
				if lastBlock != "reasoning" {
					if lastBlock != "" {
						fmt.Println()
					}
					fmt.Print("[思考] ")
					lastBlock = "reasoning"
				}
				var evt struct {
					Content string `json:"content"`
				}
				if json.Unmarshal([]byte(dataStr), &evt) == nil {
					fmt.Print(evt.Content)
				}

			case "content":
				if lastBlock != "content" {
					if lastBlock != "" {
						fmt.Print("\n\n")
					}
					fmt.Print("[回复] ")
					lastBlock = "content"
				}
				var evt struct {
					Content string `json:"content"`
				}
				if json.Unmarshal([]byte(dataStr), &evt) == nil {
					fmt.Print(evt.Content)
				}

			case "tool":
				if lastBlock != "" {
					fmt.Println()
				}
				var evt struct {
					Tool *struct {
						Icon        string `json:"icon"`
						DisplayName string `json:"displayName"`
						Action      string `json:"action"`
					} `json:"tool"`
				}
				if json.Unmarshal([]byte(dataStr), &evt) == nil && evt.Tool != nil {
					fmt.Printf("%s %s - %s\n", evt.Tool.Icon, evt.Tool.DisplayName, evt.Tool.Action)
				}
				lastBlock = ""

			case "retry":
				if lastBlock != "" {
					fmt.Println()
				}
				var evt struct {
					Retry *struct {
						Attempt    int    `json:"attempt"`
						MaxRetries int    `json:"maxRetries"`
						Error      string `json:"error"`
					} `json:"retry"`
				}
				if json.Unmarshal([]byte(dataStr), &evt) == nil && evt.Retry != nil {
					fmt.Printf("[retry] %d/%d: %s\n", evt.Retry.Attempt, evt.Retry.MaxRetries, evt.Retry.Error)
				}
				lastBlock = ""

			case "end":
				fmt.Println()
				fmt.Println("--- end ---")
			}
			eventType = ""
			dataLine = ""
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
		if strings.HasPrefix(line, "data:") {
			dataLine = line
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}
}

// TestChatCompress 手动压缩上下文 POST /api/chat/compress
func TestChatCompress(t *testing.T) {
	req := vo.ChatCompressReq{
		SessionId: "newddd123",
	}

	var rsp struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data vo.ChatCompressRsp `json:"data,omitempty"`
	}
	httpResp := requests.NewHTTPRequest(domain+"/api/chat/compress").
		Post().
		SetHeaderValue("Authorization", "Bearer "+token).
		SetJSONBody(req).
		ToJSON(&rsp)
	if httpResp.Error != nil {
		t.Fatal("chat compress error:", httpResp.Error)
	}
	t.Logf("compress response: code=%d msg=%s taskId=%s", rsp.Code, rsp.Msg, rsp.Data.TaskId)
	ChatPollCompress(rsp.Data.TaskId)
}

func ChatPollCompress(taskId string) {
	for i := 0; i < 100; i++ {
		time.Sleep(500 * time.Millisecond)
		var rsp struct {
			Code int                    `json:"code"`
			Msg  string                 `json:"msg"`
			Data vo.ChatCompressPollRsp `json:"data,omitempty"`
		}
		httpResp := requests.NewHTTPRequest(domain+"/api/chat/compress/"+taskId).
			Get().
			SetHeaderValue("Authorization", "Bearer "+token).
			ToJSON(&rsp)
		if httpResp.Error != nil {
			panic(httpResp.Error)
		}
		fmt.Printf("poll compress response: code=%d msg=%s status=%s message=%s \n",
			rsp.Code, rsp.Msg, rsp.Data.Status, rsp.Data.Message)
		if rsp.Data.Status == "done" {
			break
		}
	}
}
