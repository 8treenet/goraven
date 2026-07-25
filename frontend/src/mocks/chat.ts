/* ============================================
   Chat / Session mock data
   ============================================ */

import type {
  SessionListItem,
  SessionDetail,
  MessageItem,
  ReasoningItem,
  ChatRequest,
  ChatResponse,
  ShareInfo,
  ShareLinkRequest,
  AvailableModel,
  McpEndpoint,
  SimpleSkill,
  PersonaListItem,
  PaginatedResponse,
} from './types'
import { listDelay, itemDelay, mutationDelay, heavyDelay } from './delay'

/* ============================================
   Local rich types (preserving tool calls)
   ============================================ */

export interface ToolCall {
  toolName: string
  displayName: string
  icon: string
  action: string
  duration?: number
  success?: boolean
}

export interface RichMessage {
  id: number
  role: 'user' | 'assistant' | 'summary'
  content: string
  reasoningContent?: string
  toolCalls?: ToolCall[]
  timestamp: string
  roundId: number
}

export interface RichSession {
  id: number
  title: string
  modelId: number
  personaId: number | null
  mcpIds: number[]
  skillIds: number[]
  lastChatTime: string
  status: number
  messages: RichMessage[]
}

/* ============================================
   Constants
   ============================================ */

export const MOCK_MODELS: AvailableModel[] = [
  { aiModelId: 1, providerDisplayName: 'DeepSeek', displayName: 'DeepSeek V4 Pro', modelName: 'DeepSeek V4 Pro', icon: 'brain', contextLen: 131072, isDefault: true },
  { aiModelId: 2, providerDisplayName: 'Alibaba', displayName: 'Qwen Max', modelName: 'Qwen Max', icon: 'cpu', contextLen: 32768, isDefault: false },
  { aiModelId: 3, providerDisplayName: 'Zhipu', displayName: 'GLM-4', modelName: 'GLM-4', icon: 'cpu', contextLen: 131072, isDefault: false },
  { aiModelId: 4, providerDisplayName: 'OpenAI', displayName: 'GPT-4o', modelName: 'GPT-4o', icon: 'cpu', contextLen: 131072, isDefault: false },
  { aiModelId: 5, providerDisplayName: 'Anthropic', displayName: 'Claude 3.5 Sonnet', modelName: 'Claude 3.5 Sonnet', icon: 'cpu', contextLen: 200000, isDefault: false },
  { aiModelId: 6, providerDisplayName: 'Google', displayName: 'Gemini 2.0 Flash', modelName: 'Gemini 2.0 Flash', icon: 'cpu', contextLen: 1048576, isDefault: false },
]

export const MOCK_MCP_ENDPOINTS: McpEndpoint[] = [
  { mcpId: 1, name: 'filesystem', displayName: '文件系统', icon: 'file-text', description: '文件系统操作' },
  { mcpId: 2, name: 'brave-search', displayName: 'Brave 搜索', icon: 'search', description: 'Brave 网页搜索' },
  { mcpId: 3, name: 'github', displayName: 'GitHub', icon: 'code', description: 'GitHub API 交互' },
  { mcpId: 4, name: 'postgres', displayName: 'PostgreSQL', icon: 'terminal', description: 'PostgreSQL 数据库查询' },
  { mcpId: 5, name: 'redis', displayName: 'Redis', icon: 'terminal', description: 'Redis 缓存操作' },
  { mcpId: 6, name: 'docker', displayName: 'Docker', icon: 'terminal', description: 'Docker 容器管理' },
  { mcpId: 7, name: 'slack', displayName: 'Slack', icon: 'bot', description: 'Slack 消息推送' },
]

export const MOCK_SKILLS: SimpleSkill[] = [
  { userSkillId: 1, skillName: 'code-review', description: '代码审查', icon: 'code', source: 'builtin', categoryId: 1, categoryName: '开发工具' },
  { userSkillId: 2, skillName: 'pdf-reader', description: 'PDF 读取', icon: 'file-text', source: 'builtin', categoryId: 2, categoryName: '文档处理' },
  { userSkillId: 3, skillName: 'excel-analyzer', description: 'Excel 数据分析', icon: 'search', source: 'builtin', categoryId: 3, categoryName: '数据处理' },
  { userSkillId: 4, skillName: 'api-doc-gen', description: 'API 文档生成', icon: 'file-text', source: 'builtin', categoryId: 2, categoryName: '文档处理' },
  { userSkillId: 5, skillName: 'db-migration', description: '数据库迁移脚本', icon: 'terminal', source: 'builtin', categoryId: 4, categoryName: '数据库' },
  { userSkillId: 6, skillName: 'unit-test-gen', description: '单元测试生成', icon: 'wrench', source: 'builtin', categoryId: 1, categoryName: '开发工具' },
]

export const MOCK_PERSONAS: PersonaListItem[] = [
  {
    personaId: 1,
    name: 'Go 后端专家',
    icon: '💻',
    categoryName: '后端开发',
    modelName: 'DeepSeek V4 Pro',
    roleInfo: '精通 Go 语言和分布式系统',
    mcpNames: ['filesystem', 'github'],
    skillNames: ['code-review'],
  },
  {
    personaId: 2,
    name: '前端架构师',
    icon: '🎨',
    categoryName: '前端开发',
    modelName: 'Qwen Max',
    roleInfo: '专注 React 和 TypeScript',
    mcpNames: ['filesystem'],
    skillNames: [],
  },
  {
    personaId: 3,
    name: '数据分析师',
    icon: '📊',
    categoryName: '数据分析',
    modelName: 'GPT-4o',
    roleInfo: 'Python 数据处理与可视化',
    mcpNames: ['postgres', 'redis'],
    skillNames: ['excel-analyzer'],
  },
  {
    personaId: 4,
    name: 'DevOps 工程师',
    icon: '🚀',
    categoryName: '运维',
    modelName: 'Claude 3.5 Sonnet',
    roleInfo: 'CI/CD、容器化和云基础设施',
    mcpNames: ['filesystem', 'docker'],
    skillNames: ['api-doc-gen', 'db-migration'],
  },
  {
    personaId: 5,
    name: '全栈开发者',
    icon: '🔧',
    categoryName: '全栈开发',
    modelName: 'Gemini 2.0 Flash',
    roleInfo: '前后端全链路开发',
    mcpNames: ['filesystem', 'brave-search', 'github'],
    skillNames: ['code-review', 'unit-test-gen'],
  },
]

export const reasoningSegments = [
  '用户想了解代码审查的问题，我需要先读取相关文件，分析代码结构。让我先看看项目的主要文件...',
  '已经读取了 main.go，发现了并发模式的使用和一些潜在问题。接下来需要检查 handler 层的实现...',
  'handler 中也存在并发安全问题。让我搜索一下 Go 并发最佳实践来确认修复方案，确保建议的准确性...',
]

export const streamingToolCall1: ToolCall = {
  toolName: 'read_file',
  displayName: '文件系统',
  icon: '📄',
  action: '正在读取文件 main.go',
}

export const streamingToolCall2: ToolCall = {
  toolName: 'read_file',
  displayName: '文件系统',
  icon: '📄',
  action: '正在读取文件 handler/api.go',
}

export const streamingToolCall3: ToolCall = {
  toolName: 'brave_search',
  displayName: '网页搜索',
  icon: '🔍',
  action: '搜索 Go 并发最佳实践',
}

export const MOCK_STREAM_REPLY = `## 代码审查结果

我分析了你的代码，发现以下 **3 个关键问题**：

### 1. 错误处理缺失

第 23 行的 \`err\` 没有检查，建议添加错误处理：

\`\`\`go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
\`\`\`

### 2. 并发安全问题

第 45 行的 \`map\` 在 goroutine 中未加锁，可能导致 **data race**：

\`\`\`go
var mu sync.RWMutex

mu.Lock()
cache[key] = value
mu.Unlock()
\`\`\`

### 3. 资源泄漏

第 67 行的 \`http.Response.Body\` 未关闭：

\`\`\`go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
\`\`\`

---

| 问题 | 严重程度 | 建议 |
|------|----------|------|
| 错误处理 | 高 | 添加 \`if err != nil\` |
| 并发安全 | 高 | 使用 \`sync.RWMutex\` |
| 资源泄漏 | 中 | 添加 \`defer Close()\` |

> **提示**：可以使用 \`go vet ./...\` 和 \`golangci-lint\` 自动检测这些问题。

我已经生成了详细的审查报告和修复后的代码：

<goraven-file kind="doc" path="/documents/code-review-report.pdf" name="代码审查报告" description="包含3个问题的详细分析和修复建议" />
<goraven-file kind="image" path="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png" name="Freedom 框架架构图" />
<goraven-file kind="video" path="https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4" name="审查要点讲解" description="15分钟逐行分析关键修复点" />

### 代码质量趋势

最近 5 次提交的代码质量评分变化：

<goraven-chart type="line" title="代码质量评分趋势" x="['v1.0','v1.1','v1.2','v2.0','v2.1']" y1="[72,78,85,82,91]" y1name="质量分" height="260" />

各类型问题的分布情况：

<goraven-chart type="pie" title="问题类型分布" labels="['错误处理','并发安全','资源泄漏','性能','其他']" y1="[12,8,5,3,2]" height="240" />

需要我详细展开哪一点？`

/* ============================================
   Rich session data (preserves tool calls)
   ============================================ */

export const MOCK_SESSIONS: RichSession[] = [
  {
    id: 1,
    title: 'Go 并发模式讨论',
    modelId: 1,
    personaId: 1,
    mcpIds: [1, 3],
    skillIds: [1],
    lastChatTime: '3分钟前',
    status: 0,
    messages: [
      { id: 1, role: 'user', content: '请帮我审查这段并发代码的安全性', toolCalls: [], timestamp: '14:20', roundId: 1 },
      {
        id: 2,
        role: 'assistant',
        content: `## 并发安全审查报告

我分析了代码，发现以下问题：

### 问题 1: Channel 未关闭

\`\`\`go
// 危险: 如果 channel 未关闭，循环将永久阻塞
for v := range ch {
    results = append(results, v)
}
\`\`\`

**修复方案**：使用 \`context\` 或 \`select\` 控制退出：

\`\`\`go
for {
    select {
    case v, ok := <-ch:
        if !ok {
            return results
        }
        results = append(results, v)
    case <-ctx.Done():
        return results
    }
}
\`\`\`

### 问题 2: Goroutine 泄漏

| 风险点 | 描述 | 修复 |
|--------|------|------|
| 无缓冲 channel | 发送端可能永久阻塞 | 使用 \`context.WithCancel\` |
| WaitGroup 缺失 | 无法等待所有 goroutine 完成 | 添加 \`sync.WaitGroup\` |
| 共享 map | 并发读写导致 data race | 使用 \`sync.Map\` 或加锁 |

> **建议**：运行 \`go race detector\` 进行全面检测：
> \`\`\`bash
> go run -race ./...
> \`\`\``,
        reasoningContent: '先检查 goroutine 的数量是否可控，确认是否有泄漏风险\n再看 channel 的关闭机制，发现 range 循环没有退出条件\n最后分析数据竞争问题，共享 map 没有加锁保护',
        toolCalls: [
          { toolName: 'read_file', displayName: '文件系统', icon: '📄', action: '正在读取文件 main.go' },
          { toolName: 'read_file', displayName: '文件系统', icon: '📄', action: '正在读取文件 handler/api.go' },
        ],
        timestamp: '14:20',
        roundId: 1,
      },
    ],
  },
  {
    id: 2,
    title: 'Docker 部署配置',
    modelId: 2,
    personaId: null,
    mcpIds: [1],
    skillIds: [],
    lastChatTime: '1小时前',
    status: 0,
    messages: [
      { id: 3, role: 'user', content: '帮我写一个多阶段构建的 Dockerfile', toolCalls: [], timestamp: '13:10', roundId: 2 },
      {
        id: 4,
        role: 'assistant',
        content: `## 多阶段构建 Dockerfile

以下是一个生产级的多阶段构建示例：

\`\`\`dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \\
    -ldflags="-s -w" \\
    -o /server .

# Stage 2: Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
\`\`\`

### 关键优化点

1. **依赖缓存** — 先复制 \`go.mod\` 再复制源码，利用 Docker 层缓存
2. **静态链接** — \`CGO_ENABLED=0\` 生成纯静态二进制
3. **Strip 符号** — \`-ldflags="-s -w"\` 减小镜像体积约 **30%**
4. **最小基础镜像** — Alpine 仅 ~5MB，最终镜像 < 20MB`,
        toolCalls: [],
        timestamp: '13:11',
        roundId: 2,
      },
      { id: 5, role: 'user', content: '加上健康检查', toolCalls: [], timestamp: '13:12', roundId: 3 },
      {
        id: 6,
        role: 'assistant',
        content: `好的，在 runtime stage 中添加健康检查：

\`\`\`dockerfile
HEALTHCHECK --interval=30s --timeout=3s \\
  --start-period=5s --retries=3 \\
  CMD wget -q -O- http://localhost:8080/health || exit 1
\`\`\`

### 参数说明

| 参数 | 值 | 含义 |
|------|-----|------|
| \`--interval\` | 30s | 检查间隔 |
| \`--timeout\` | 3s | 单次超时 |
| \`--start-period\` | 5s | 启动宽限期 |
| \`--retries\` | 3 | 连续失败次数 |

> \`--start-period\` 很重要，避免应用启动期间被误判为不健康。`,
        toolCalls: [],
        timestamp: '13:12',
        roundId: 3,
      },
    ],
  },
  {
    id: 3,
    title: '微服务架构设计',
    modelId: 1,
    personaId: 1,
    mcpIds: [1, 3],
    skillIds: [1],
    lastChatTime: '刚刚',
    status: 1,
    messages: [
      { id: 7, role: 'user', content: '帮我设计一个订单系统的微服务架构，画出组件交互和部署流程', toolCalls: [], timestamp: '15:30', roundId: 7 },
      {
        id: 8,
        role: 'assistant',
        content: `## 订单系统微服务架构

核心服务拆分如下：

- **Gateway** — API 网关，统一鉴权、限流、路由
- **Order** — 订单服务，管理订单生命周期
- **Inventory** — 库存服务，处理库存扣减与回补
- **Payment** — 支付服务，对接第三方支付渠道
- **Notification** — 通知服务，订单状态变更推送

### 下单流程时序图

\`\`\`mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant O as Order
    participant I as Inventory
    participant P as Payment
    participant N as Notification

    C->>G: POST /orders
    G->>G: 鉴权 & 限流
    G->>O: 创建订单
    O->>I: 预扣库存
    I-->>O: 扣减成功
    O->>P: 发起支付
    P-->>O: 支付回调
    O->>I: 确认扣减
    O->>N: 发送通知
    N-->>C: 订单已支付
\`\`\`

### 部署架构

\`\`\`mermaid
graph TB
    subgraph K8s Cluster
        LB[Load Balancer]
        subgraph Services
            GW[Gateway x3]
            OR[Order x5]
            IV[Inventory x3]
            PY[Payment x3]
            NT[Notification x2]
        end
        subgraph Data
            PG[(PostgreSQL)]
            RD[(Redis)]
            MQ[RabbitMQ]
        end
    end

    LB --> GW
    GW --> OR
    OR --> IV
    OR --> PY
    OR --> MQ
    MQ --> NT
    OR --> PG
    IV --> RD
    PY --> PG
\`\`\`

### 核心数据关系

\`\`\`mermaid
erDiagram
    ORDER ||--o{ ORDER_ITEM : contains
    ORDER ||--|| PAYMENT : has
    ORDER }o--|| USER : belongs_to
    ORDER_ITEM }o--|| PRODUCT : references
    PAYMENT ||--o{ PAYMENT_LOG : records

    ORDER {
        bigint id PK
        string order_no UK
        bigint user_id FK
        decimal total_amount
        string status
        timestamp created_at
    }

    ORDER_ITEM {
        bigint id PK
        bigint order_id FK
        bigint product_id FK
        int quantity
        decimal unit_price
    }

    PAYMENT {
        bigint id PK
        bigint order_id FK
        string transaction_id
        decimal amount
        string status
    }
\`\`\`

> 以上架构使用 PostgreSQL 做主库，Redis 做库存缓存和分布式锁，RabbitMQ 解耦异步通知，各服务通过 gRPC 内部通信。`,
        toolCalls: [
          { toolName: 'read_file', displayName: '文件系统', icon: '📄', action: '读取 go.mod' },
          { toolName: 'search', displayName: '搜索', icon: '🔍', action: '搜索微服务最佳实践' },
        ],
        timestamp: '15:31',
        roundId: 7,
      },
    ],
  },
  {
    id: 4,
    title: 'Q2营收数据分析',
    modelId: 4,
    personaId: 3,
    mcpIds: [7],
    skillIds: [3],
    lastChatTime: '刚刚',
    status: 1,
    messages: [
      { id: 9, role: 'user', content: '帮我分析下Q2的营收数据，做个趋势和对比', toolCalls: [], timestamp: '16:00', roundId: 9 },
      {
        id: 10,
        role: 'assistant',
        content: `## Q2 营收数据分析报告

### 月度营收趋势

Q2 整体营收呈现稳步增长态势，6 月份受促销活动拉动，环比增长显著。

<goraven-chart type="bar" title="Q2 月度营收" x="['4月','5月','6月']" y1="[286,342,518]" y1name="营收(万元)" height="260" />

### 各产品线收入占比

旗舰产品和 SaaS 服务贡献了超过 70% 的收入，定制开发项目占比相对稳定。

<goraven-chart type="pie" title="产品线收入分布" labels="['旗舰产品','SaaS服务','定制开发','技术支持','其他']" y1="[156,138,82,45,28]" height="240" />

### 同比与环比增长

与去年同期相比，Q2 整体增长 36.8%。环比 Q1 增长 18.2%，其中 SaaS 服务增速最快。

<goraven-chart type="line" title="营收同比增长趋势" x="['1月','2月','3月','4月','5月','6月']" y1="[220,185,260,286,342,518]" y1name="2025" y2="[180,160,230,245,280,370]" y2name="2024" height="260" />

### 客户活跃度变化

日均活跃客户数在 Q2 稳步上升，周末活跃度有所回落但工作日恢复迅速。

<goraven-chart type="area" title="日均活跃客户数" x="['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" y1="[3400,3800,4200,4500,4100,2800,3200]" y1name="活跃客户" height="260" />

### 总结

- Q2 总营收 **1,146 万元**，同比增长 36.8%
- SaaS 服务增速最快，环比 +42%
- 6 月促销贡献增量 176 万元
- 建议 Q3 加大 SaaS 推广力度，优化定制开发利润率`,
        toolCalls: [
          { toolName: 'read_file', displayName: '文件系统', icon: '📄', action: '读取 Q2 销售数据' },
          { toolName: 'search', displayName: '搜索', icon: '🔍', action: '搜索行业营收基准' },
        ],
        timestamp: '16:01',
        roundId: 9,
      },
    ],
  },
]

/* ============================================
   Helper maps
   ============================================ */

const modelMap = new Map(MOCK_MODELS.map((m) => [m.aiModelId, m]))
const personaMap = new Map(MOCK_PERSONAS.map((p) => [p.personaId, p]))

function sessionToListItem(s: RichSession): SessionListItem {
  return {
    sessionId: String(s.id),
    title: s.title,
    status: s.status,
    personaId: s.personaId ?? 0,
    lastChatTime: s.lastChatTime,
    created: '2025-06-01T10:00:00Z',
  }
}

function sessionToDetail(s: RichSession): SessionDetail {
  const model = modelMap.get(s.modelId)
  const persona = s.personaId ? personaMap.get(s.personaId) : null
  return {
    sessionId: String(s.id),
    title: s.title,
    status: s.status,
    personaId: s.personaId ?? 0,
    aiModelId: s.modelId,
    contextTokens: 0,
    promptTokensCount: 0,
    completionTokensCount: 0,
    mcpIds: s.mcpIds,
    skillIds: s.skillIds,
    lastChatTime: s.lastChatTime,
    created: '2025-06-01T10:00:00Z',
    modelName: model?.modelName ?? '',
    personaName: persona?.name ?? '',
    personaIcon: persona?.icon ?? '',
    contextLimit: model?.contextLen ?? 131072,
  }
}

function richMessageToItem(m: RichMessage): MessageItem {
  const reasoningContent: ReasoningItem[] = []
  if (m.reasoningContent) {
    reasoningContent.push({ eventType: 'reasoning', content: m.reasoningContent })
  }
  if (m.toolCalls) {
    for (const tc of m.toolCalls) {
      reasoningContent.push({
        eventType: 'tool',
        tool: { name: tc.toolName, displayName: tc.displayName, icon: tc.icon, action: tc.action },
      })
    }
  }
  return {
    msgId: String(m.id),
    roundId: String(m.roundId),
    contextState: 0,
    content: m.content,
    reasoningContent,
    roleType: m.role,
    created: m.timestamp,
  }
}

/* ============================================
   In-memory state for create/stop/compress
   ============================================ */

let nextSessionId = 100
const activeGenerating = new Set<string>()

/* ============================================
   Exported async functions
   ============================================ */

/** List sessions, optionally filtered by persona */
export async function getSessions(personaId?: number): Promise<SessionListItem[]> {
  await listDelay()
  let filtered = MOCK_SESSIONS
  if (personaId !== undefined) {
    filtered = MOCK_SESSIONS.filter((s) => s.personaId === personaId)
  }
  return filtered.map(sessionToListItem)
}

/** Get single session detail */
export async function getSessionDetail(sessionId: string): Promise<SessionDetail> {
  await itemDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  return sessionToDetail(session)
}

/** Update session title or archive status */
export async function updateSession(
  sessionId: string,
  req: { title?: string; isArchived?: boolean },
): Promise<void> {
  await mutationDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  if (req.title !== undefined) session.title = req.title
  if (req.isArchived !== undefined) session.status = req.isArchived ? 1 : 0
}

/** Delete a session */
export async function deleteSession(sessionId: string): Promise<void> {
  await mutationDelay()
  const idx = MOCK_SESSIONS.findIndex((s) => String(s.id) === sessionId)
  if (idx === -1) throw new Error(`Session ${sessionId} not found`)
  MOCK_SESSIONS.splice(idx, 1)
}

/** Get messages for a session */
export async function getSessionMessages(sessionId: string): Promise<MessageItem[]> {
  await listDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  return session.messages.map(richMessageToItem)
}

/** Start a new chat. Returns the session ID to stream from. */
export async function createChat(req: ChatRequest): Promise<ChatResponse> {
  await itemDelay()
  const id = req.sessionId ? Number(req.sessionId) : nextSessionId++
  const sessionId = String(id)

  if (!req.sessionId) {
    // Create a new session in the mock list
    const title = req.content.slice(0, 30) + (req.content.length > 30 ? '...' : '')
    MOCK_SESSIONS.unshift({
      id,
      title,
      modelId: req.aiModelId,
      personaId: req.personaId ?? null,
      mcpIds: req.mcpIds,
      skillIds: req.skillIds,
      lastChatTime: '刚刚',
      status: 0,
      messages: [],
    })
  }

  activeGenerating.add(sessionId)
  return { sessionId }
}

/** Stop generation for a session */
export async function stopChat(sessionId: string): Promise<void> {
  await itemDelay()
  activeGenerating.delete(sessionId)
}

/** Trigger context compression */
export async function compressChat(sessionId: string): Promise<{ taskId: string }> {
  await mutationDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  const taskId = `compress-${Date.now()}`
  return { taskId }
}

/** Poll compression status */
export async function getCompressStatus(_taskId: string): Promise<{ status: string }> {
  await listDelay()
  return { status: 'done' }
}

/** Create a share link for a session */
export async function createShare(sessionId: string, req: ShareLinkRequest): Promise<ShareInfo> {
  await mutationDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  return {
    shareId: `share-${Date.now()}`,
    sessionId: String(session.id),
    title: req.title ?? session.title,
    expiresAt: '2026-06-09T10:00:00Z',
    viewCount: 0,
    isExpired: false,
    created: new Date().toISOString(),
  }
}

/** Get share info for a session */
export async function getShare(sessionId: string): Promise<ShareInfo> {
  await itemDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  return {
    shareId: `share-${session.id}`,
    sessionId: String(session.id),
    title: session.title,
    expiresAt: '2026-06-09T10:00:00Z',
    viewCount: 12,
    isExpired: false,
    created: '2025-06-01T10:00:00Z',
  }
}

/** Delete a share link */
export async function deleteShare(sessionId: string): Promise<void> {
  await mutationDelay()
  const session = MOCK_SESSIONS.find((s) => String(s.id) === sessionId)
  if (!session) throw new Error(`Session ${sessionId} not found`)
  // no-op in mock
}

/** List current user's shares */
export async function getMyShares(params?: {
  page?: number
  pageSize?: number
}): Promise<PaginatedResponse<ShareInfo>> {
  await listDelay()
  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 10
  const all: ShareInfo[] = MOCK_SESSIONS.map((s) => ({
    shareId: `share-${s.id}`,
    sessionId: String(s.id),
    title: s.title,
    expiresAt: '2026-06-09T10:00:00Z',
    viewCount: Math.floor(Math.random() * 50),
    isExpired: false,
    created: '2025-06-01T10:00:00Z',
  }))
  const totalCount = all.length
  const totalPage = Math.ceil(totalCount / pageSize)
  const start = (page - 1) * pageSize
  return {
    list: all.slice(start, start + pageSize),
    totalPage,
    totalCount,
    page,
    pageSize,
  }
}

/** List available models */
export async function getModels(): Promise<AvailableModel[]> {
  await listDelay()
  return MOCK_MODELS
}

/** List MCP endpoints */
export async function getMcpEndpoints(): Promise<McpEndpoint[]> {
  await listDelay()
  return MOCK_MCP_ENDPOINTS
}

/** List installed skills */
export async function getSkills(): Promise<SimpleSkill[]> {
  await listDelay()
  return MOCK_SKILLS
}

/** List personas */
export async function getPersonas(): Promise<PersonaListItem[]> {
  await listDelay()
  return MOCK_PERSONAS
}

/** Upload a file */
export async function uploadFile(file: File): Promise<{
  uploadId: string
  filePath: string
  fileName: string
}> {
  await heavyDelay()
  return {
    uploadId: `upload-${Date.now()}`,
    filePath: `/documents/${file.name}`,
    fileName: file.name,
  }
}
