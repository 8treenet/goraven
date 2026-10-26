package agent

import (
	"goraven/backend/po"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestBuildUserMessageWithMedia(t *testing.T) {
	msg := BuildUserMessage("看下这张图", []MediaItem{
		{Type: "image", Path: "/temp/a.jpg", URL: "https://x/api/hfs/public/a.jpg"},
		{Type: "audio", Path: "/temp/b.mp3", URL: "https://x/api/hfs/public/b.mp3"},
	})
	if msg.Role != schema.User || len(msg.UserInputMultiContent) != 3 {
		t.Fatalf("unexpected message: role=%v parts=%d", msg.Role, len(msg.UserInputMultiContent))
	}
	if msg.UserInputMultiContent[0].Type != schema.ChatMessagePartTypeText || msg.UserInputMultiContent[0].Text != "看下这张图" {
		t.Fatalf("first part should be text, got %+v", msg.UserInputMultiContent[0])
	}
	img := msg.UserInputMultiContent[1]
	if img.Type != schema.ChatMessagePartTypeImageURL || img.Image == nil || img.Image.URL == nil || *img.Image.URL != "https://x/api/hfs/public/a.jpg" {
		t.Fatalf("image part mismatch: %+v", img)
	}
	audio := msg.UserInputMultiContent[2]
	if audio.Type != schema.ChatMessagePartTypeAudioURL || audio.Audio == nil || audio.Audio.URL == nil || *audio.Audio.URL != "https://x/api/hfs/public/b.mp3" {
		t.Fatalf("audio part mismatch: %+v", audio)
	}
}

func TestBuildUserMessagePlain(t *testing.T) {
	msg := BuildUserMessage("hello", nil)
	if msg.Role != schema.User || msg.Content != "hello" || len(msg.UserInputMultiContent) != 0 {
		t.Fatalf("plain message mismatch: %+v", msg)
	}
}

func TestBuildUserMessageVideo(t *testing.T) {
	msg := BuildUserMessage("看视频", []MediaItem{
		{Type: "video", Path: "/temp/v.mp4", URL: "https://x/api/hfs/public/v.mp4"},
	})
	part := msg.UserInputMultiContent[1]
	if part.Type != schema.ChatMessagePartTypeVideoURL || part.Video == nil || part.Video.URL == nil || *part.Video.URL != "https://x/api/hfs/public/v.mp4" {
		t.Fatalf("video part mismatch: %+v", part)
	}
}

func TestBuildHistoryFromMessagesWithMedia(t *testing.T) {
	list := []*po.Message{
		{MsgId: "m1", SessionId: "s1", RoundId: "r1", Timestamp: 1, RoleType: po.RoleTypeUser, Content: "看图",
			Media: `[{"type":"image","path":"/temp/a.jpg","url":"https://x/api/hfs/public/a.jpg"}]`},
		{MsgId: "m2", SessionId: "s1", RoundId: "r1", Timestamp: 2, RoleType: po.RoleTypeAssistant, Content: "好的"},
	}
	msgs := BuildHistoryFromMessages(list)
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	first := msgs[0]
	if first.Role != schema.User || len(first.UserInputMultiContent) != 2 {
		t.Fatalf("want 2 parts, got %d (role=%v)", len(first.UserInputMultiContent), first.Role)
	}
	if first.UserInputMultiContent[0].Type != schema.ChatMessagePartTypeText || first.UserInputMultiContent[0].Text != "看图" {
		t.Fatalf("first part should be text: %+v", first.UserInputMultiContent[0])
	}
	img := first.UserInputMultiContent[1]
	if img.Type != schema.ChatMessagePartTypeImageURL || img.Image == nil || img.Image.URL == nil || *img.Image.URL != "https://x/api/hfs/public/a.jpg" {
		t.Fatalf("image part mismatch: %+v", img)
	}
	if msgs[1].Content != "好的" {
		t.Fatalf("assistant message mismatch: %+v", msgs[1])
	}
}

func TestBuildHistoryFromMessagesMediaBrokenJSON(t *testing.T) {
	list := []*po.Message{
		{MsgId: "m1", SessionId: "s1", RoundId: "r1", Timestamp: 1, RoleType: po.RoleTypeUser, Content: "看图", Media: "{bad json"},
	}
	msgs := BuildHistoryFromMessages(list)
	if len(msgs) != 1 || msgs[0].Content != "看图" || len(msgs[0].UserInputMultiContent) != 0 {
		t.Fatalf("broken media json should degrade to plain text: %+v", msgs)
	}
}
