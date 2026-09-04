package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goraven/config"
	"goraven/core/iface"
	"goraven/core/sandbox"
	"goraven/util"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// VisualUnderstandRequest 多模态识别工具的请求参数
type VisualUnderstandRequest struct {
	FilePath string `json:"file_path" jsonschema:"description=The workspace-relative file path, e.g. /temp/screenshot.jpg or /images/photo.jpg or /temp/recording.mp3"`
	Question string `json:"question" jsonschema:"description=The question or instruction about the file content, e.g. Describe what you see in this image"`
}

// VisualUnderstandResponse 多模态识别工具的响应
type VisualUnderstandResponse struct {
	Result string `json:"result" jsonschema:"description=The visual/auditory understanding result from the model"`
}

// VisualUnderstand 多模态识别工具，使用视觉模型分析图片、视频、音频文件
type VisualUnderstand struct {
	Name           string
	Desc           string
	userID         string
	model          iface.BaseChatModel
	box            sandbox.Sandbox
	dailyStatsRepo iface.DailyStatsRepo
}

const (
	VisualUnderstandToolDesc = `Analyzes images, videos, and audio files using a visual understanding model. Provide the file path and a question about the file content. Supports common image formats (jpg, png, gif, webp, etc.), video formats (mp4, avi, mov, etc.), and audio formats (mp3, wav, etc.). If the tool fails with an error indicating the model does not support multimodal input, tell the user to set a multimodal-capable model on the model management page.`

	VisualUnderstandToolDescChinese = `使用视觉理解模型分析图片、视频、音频文件。提供文件路径和关于文件内容的问题。支持常见图片格式（jpg、png、gif、webp等）、视频格式（mp4、avi、mov等）和音频格式（mp3、wav等）。如果工具返回错误提示模型不支持多模态输入，请告知用户在模型管理页面设置支持多模态识别的模型。`
)

var (
	imageExts = map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true, ".tiff": true, ".ico": true,
	}
	videoExts = map[string]bool{
		".mp4": true, ".avi": true, ".mov": true, ".mkv": true,
		".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	}
	audioExts = map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aac": true,
		".ogg": true, ".wma": true, ".m4a": true, ".opus": true,
	}
)

// mimeTypes 常见扩展名到MIME类型的映射
var mimeTypes = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".bmp": "image/bmp", ".webp": "image/webp",
	".svg": "image/svg+xml", ".tiff": "image/tiff", ".ico": "image/x-icon",
	".mp4": "video/mp4", ".avi": "video/x-msvideo", ".mov": "video/quicktime",
	".mkv": "video/x-matroska", ".wmv": "video/x-ms-wmv", ".flv": "video/x-flv",
	".webm": "video/webm", ".m4v": "video/x-m4v",
	".mp3": "audio/mpeg", ".wav": "audio/wav", ".flac": "audio/flac",
	".aac": "audio/aac", ".ogg": "audio/ogg", ".wma": "audio/x-ms-wma",
	".m4a": "audio/mp4", ".opus": "audio/opus",
}

// supportsNativeVideo 判断模型格式是否原生支持视频输入
func supportsNativeVideo(format iface.APIFormat) bool {
	return format == iface.APIFormatOpenAI || format == iface.APIFormatGemini
}

// supportsNativeAudio 判断模型格式是否原生支持音频输入
func supportsNativeAudio(format iface.APIFormat) bool {
	return format == iface.APIFormatOpenAI || format == iface.APIFormatGemini
}

// fileType 文件类型枚举
type fileType int

const (
	fileTypeImage fileType = iota
	fileTypeVideo
	fileTypeAudio
	fileTypeUnknown
)

// detectFileType 根据扩展名判断文件类型
func detectFileType(ext string) fileType {
	ext = strings.ToLower(ext)
	if imageExts[ext] {
		return fileTypeImage
	}
	if videoExts[ext] {
		return fileTypeVideo
	}
	if audioExts[ext] {
		return fileTypeAudio
	}
	return fileTypeUnknown
}

// getMimeType 根据扩展名获取MIME类型
func getMimeType(ext string) string {
	if mt, ok := mimeTypes[strings.ToLower(ext)]; ok {
		return mt
	}
	return "application/octet-stream"
}

// formatUnsupportedError 生成不支持的多媒体类型错误提示
func formatUnsupportedError(mediaType string, format iface.APIFormat) error {
	lang := config.Get().GetLanguage()
	if lang == "zh" {
		return fmt.Errorf("当前视觉模型的API格式(%s)不支持%s输入，请切换为支持%s的模型", format, mediaType, mediaType)
	}
	return fmt.Errorf("current visual model's API format (%s) does not support %s input, please switch to a model that supports %s", format, mediaType, mediaType)
}

// NewVisualUnderstand 创建多模态识别工具
func NewVisualUnderstand(userID string, model iface.BaseChatModel, box sandbox.Sandbox, dailyStatsRepo iface.DailyStatsRepo) (tool.InvokableTool, error) {
	desc := VisualUnderstandToolDesc
	if config.Get().GetLanguage() == "zh" {
		desc = VisualUnderstandToolDescChinese
	}

	t := &VisualUnderstand{
		Name:           "goraven_visual_understand",
		Desc:           desc,
		userID:         userID,
		model:          model,
		box:            box,
		dailyStatsRepo: dailyStatsRepo,
	}

	invokable, err := utils.InferTool(t.Name, t.Desc, t.Invoke)
	if err != nil {
		return nil, fmt.Errorf("failed to infer tool: %w", err)
	}
	return invokable, nil
}

// Invoke 执行多模态识别
func (v *VisualUnderstand) Invoke(ctx context.Context, req *VisualUnderstandRequest) (*VisualUnderstandResponse, error) {
	ext := strings.ToLower(filepath.Ext(req.FilePath))
	ft := detectFileType(ext)
	if ft == fileTypeUnknown {
		return nil, fmt.Errorf("unsupported file type: %s (supported: image, video, audio)", ext)
	}

	// 通过沙盒下载文件到本地路径（本地沙盒直接返回原路径，远程沙盒会先下载）
	absPath := filepath.Join(v.box.GetWorkspace(), strings.TrimPrefix(req.FilePath, "/"))
	localPath, err := v.box.Download(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from sandbox: %w", err)
	}

	switch ft {
	case fileTypeImage:
		return v.invokeImage(ctx, localPath, ext, req.Question)
	case fileTypeVideo:
		return v.invokeVideo(ctx, localPath, ext, req.Question)
	case fileTypeAudio:
		return v.invokeAudio(ctx, localPath, ext, req.Question)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// invokeImage 处理图片：base64 编码后通过 UserInputMultiContent 传递
// 所有提供商都支持 base64 图片
func (v *VisualUnderstand) invokeImage(ctx context.Context, absPath, ext, question string) (*VisualUnderstandResponse, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	mimeType := getMimeType(ext)

	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: question,
			},
			{
				Type: schema.ChatMessagePartTypeImageURL,
				Image: &schema.MessageInputImage{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: &b64,
						MIMEType:   mimeType,
					},
				},
			},
		},
	}

	result, err := v.callModel(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &VisualUnderstandResponse{Result: result}, nil
}

// invokeVideo 处理视频：base64 编码后通过 UserInputMultiContent 传递
// 仅 OpenAI 和 Gemini 格式支持视频
func (v *VisualUnderstand) invokeVideo(ctx context.Context, absPath, ext, question string) (*VisualUnderstandResponse, error) {
	if !supportsNativeVideo(v.model.Format()) {
		return nil, formatUnsupportedError("video", v.model.Format())
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	mimeType := getMimeType(ext)

	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: question,
			},
			{
				Type: schema.ChatMessagePartTypeVideoURL,
				Video: &schema.MessageInputVideo{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: &b64,
						MIMEType:   mimeType,
					},
				},
			},
		},
	}

	result, err := v.callModel(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &VisualUnderstandResponse{Result: result}, nil
}

// invokeAudio 处理音频：base64 编码后通过 UserInputMultiContent 传递
// 仅 OpenAI 和 Gemini 格式支持音频
func (v *VisualUnderstand) invokeAudio(ctx context.Context, absPath, ext, question string) (*VisualUnderstandResponse, error) {
	if !supportsNativeAudio(v.model.Format()) {
		return nil, formatUnsupportedError("audio", v.model.Format())
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	mimeType := getMimeType(ext)

	msg := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: question,
			},
			{
				Type: schema.ChatMessagePartTypeAudioURL,
				Audio: &schema.MessageInputAudio{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: &b64,
						MIMEType:   mimeType,
					},
				},
			},
		},
	}

	result, err := v.callModel(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &VisualUnderstandResponse{Result: result}, nil
}

// callModel 调用视觉模型并返回文本结果
func (v *VisualUnderstand) callModel(ctx context.Context, msg *schema.Message) (string, error) {
	// 一次性识别运行：使用视觉模型配置的 header 名与新的运行 ID
	ctx = util.WithConversationHeader(ctx, v.model.GetConversationHeaderKey(), util.UUID())
	resp, err := v.model.Generate(ctx, []*schema.Message{msg})
	if err != nil {
		return "", fmt.Errorf("visual model call failed: %w", err)
	}
	if v.dailyStatsRepo != nil && resp != nil && resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		usage := resp.ResponseMeta.Usage
		v.dailyStatsRepo.AddDailyStats(v.userID, usage.PromptTokens, usage.CompletionTokens, usage.PromptTokenDetails.CachedTokens)
	}
	return resp.Content, nil
}
