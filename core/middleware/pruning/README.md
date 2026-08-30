# Pruning Middleware

剪枝中间件用于管理和优化工具消息，通过多种策略减少上下文长度。

## 功能特性

1. **Token阈值剪枝**：当总token数超过阈值时，从旧到新删除工具消息对
2. **保留最近轮次**：只保留最近N轮工具消息
3. **智能长度截断**：
   - 最近一轮工具消息原样保留
   - 最近时间窗口内且长度合适的消息原样保留
   - 超过指定长度的消息截断头尾部分

## 使用方法

```go
import (
    "context"
    "time"
    "goraven/core/middleware/pruning"
    "github.com/cloudwego/eino/adk"
)

func main() {
    ctx := context.Background()
    
    config := &pruning.Config{
        TokenThreshold:       160000,        // Token阈值，默认160000
        KeepRecentToolRounds: 3,             // 保留最近3轮工具消息
        RecentTimeWindow:     5 * time.Minute, // 最近5分钟的时间窗口
        MaxToolResultLength:  2000,          // 最大工具结果长度
        HeadLength:           1000,          // 截断时保留的头部长度
        TailLength:           1000,          // 截断时保留的尾部长度
        // TokenCounter: 自定义token计数器，默认使用字符数/4
    }
    
    mw, err := pruning.New(ctx, config)
    if err != nil {
        panic(err)
    }
    
    // 使用中间件
    agent := adk.NewChatModelAgent(
        adk.WithModel(model),
        adk.WithMiddleware(mw),
    )
}
```

## 配置说明

### TokenThreshold
- 类型：`int64`
- 默认值：`160000`
- 说明：当总token数超过此阈值时触发剪枝

### KeepRecentToolRounds
- 类型：`int`
- 默认值：`3`
- 说明：保留最近N轮工具消息

### RecentTimeWindow
- 类型：`time.Duration`
- 默认值：`5 * time.Minute`
- 说明：最近时间窗口，在此窗口内且长度合适的消息会被保留

### MaxToolResultLength
- 类型：`int`
- 默认值：`2000`
- 说明：工具结果的最大长度，超过此长度会被截断

### HeadLength
- 类型：`int`
- 默认值：`1000`
- 说明：截断时保留的头部字符数

### TailLength
- 类型：`int`
- 默认值：`1000`
- 说明：截断时保留的尾部字符数

### TokenCounter
- 类型：`func(ctx context.Context, msgs []adk.Message, tools []*schema.ToolInfo) (int64, error)`
- 默认值：字符数除以4的简单估算
- 说明：自定义token计数函数

## 剪枝策略

中间件按以下顺序应用三种策略：

1. **Token阈值策略**：当总token数超过阈值时，从旧到新删除工具消息对，直到满足阈值要求

2. **保留最近轮次策略**：只保留最近N轮工具消息对，删除所有更旧的工具消息

3. **智能长度截断策略**：
   - 最近一轮的工具消息原样保留
   - 在最近时间窗口内且长度不超过阈值的工具消息原样保留
   - 超过长度阈值的工具消息截断中间部分，保留头尾

## 工具消息对

中间件确保工具消息的成对性：
- 工具调用（assistant消息，包含tool_calls）
- 工具响应（tool消息）

删除或截断时，总是以消息对为单位进行操作，保证LLM不会因为消息格式错误而报错。

## 示例

```go
// 基本使用
config := &pruning.Config{
    TokenThreshold:       100000,
    KeepRecentToolRounds: 5,
    RecentTimeWindow:     10 * time.Minute,
    MaxToolResultLength:  3000,
    HeadLength:           1500,
    TailLength:           1500,
}

mw, _ := pruning.New(ctx, config)

// 使用自定义token计数器
config.TokenCounter = func(ctx context.Context, msgs []adk.Message, tools []*schema.ToolInfo) (int64, error) {
    // 使用更精确的tokenizer
    return preciseTokenCount(msgs), nil
}
```

## 注意事项

1. 只处理工具类型的消息，普通对话消息不受影响
2. 工具消息的时间戳从消息的Extra字段中读取（key为"timestamp"）
3. 截断长消息时会在中间插入提示信息
4. 保留的消息顺序与原始顺序一致
