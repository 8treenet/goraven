# Mock Data SQL Files

用于新数据库初始化的 mock 数据。所有 SQL 使用 **SQLite** 语法。

## 文件说明

| 文件 | 包含的表 | 数据量 |
|------|---------|--------|
| `01_users_and_auth.sql` | user, user_auth | 6 用户 + 6 auth tokens |
| `02_models_and_mcp.sql` | ai_model, mcp_endpoint, system_skill | 14 模型 + 12 MCP + 5 系统技能 |
| `03_personas_and_skills.sql` | persona_template, user_persona, persona_tool, skill_market, user_skill | 12 模板 + 7 角色 + 24 工具绑定 + 16 市场技能 + 13 用户技能 |
| `04_sessions_and_messages.sql` | session, message | 15 会话 + 32+ 消息 |
| `05_links_and_stats.sql` | share_link, file_link, chunk_upload, user_daily_stats, tool_daily_stats | 6 分享 + 11 文件 + 5 上传 + 17 用户日统计 + 60+ 工具日统计 |
| `06_dashboard.sql` | user_daily_stats, tool_daily_stats | 30天用户日统计 + 180+ 工具/技能/MCP日统计 |

## 使用方式

### 一键导入（推荐）

```bash
cd backend/repository/seed/mock
sqlite3 /path/to/goraven.db < 01_users_and_auth.sql
sqlite3 /path/to/goraven.db < 02_models_and_mcp.sql
sqlite3 /path/to/goraven.db < 03_personas_and_skills.sql
sqlite3 /path/to/goraven.db < 04_sessions_and_messages.sql
sqlite3 /path/to/goraven.db < 05_links_and_stats.sql
sqlite3 /path/to/goraven.db < 06_dashboard.sql
```

### 或合并执行

```bash
cat 0*.sql | sqlite3 /path/to/goraven.db
```

## 主用户 ID

所有 mock 数据（会话、角色、技能、统计等）均归属同一个用户：

```
userId: 90a431bee756432492c134f510bad949
```

该用户为超级管理员（admin），密码 `123456` 的 MD5。

## Mock 数据概览

### 用户（6个）

| 用户 | 角色 | 说明 |
|------|------|------|
| admin（张管理） | 超级管理员 | 主操作账号，拥有所有 mock 数据 |
| lina（李财务） | 普通用户 | 财务角色 |
| wangwei（王编辑） | 普通用户 | 写作/编辑角色 |
| zhaolaoshi（赵老师） | 普通用户 | 教育角色 |
| chen_dev（陈开发） | 管理员 | 开发角色 |
| liu_data（刘数据） | 普通用户 | 数据分析角色 |

### MCP 端点（12个）

| 类别 | 端点 | 用途 |
|------|------|------|
| 开发 | filesystem, postgres, github | 文件操作、数据库、代码仓库 |
| 搜索 | brave-search | 网页搜索 |
| 财务 | financial-data, tax-calculator | 金融数据、税务计算 |
| 写作 | document-editor, translation | 文档编辑、翻译 |
| 教育 | arxiv-search, notion | 学术论文、知识库 |
| 办公 | weather, email | 天气、邮件 |

### 技能（16个市场技能）

| 类别 | 技能 |
|------|------|
| 通用 | file-organizer, presentation-builder |
| 开发 | go-project-scaffold, react-component-gen |
| 运维 | docker-compose-gen, k8s-manifest-gen |
| 数据 | data-viz-assistant |
| 财务 | financial-report-analyzer, tax-helper, budget-planner |
| 写作 | essay-writer, creative-copywriter, business-email-composer |
| 教育 | lesson-planner, research-paper-helper, flashcard-gen |

### 会话场景（15个）

- Go 项目架构设计
- Docker Compose 问题排查
- React 组件开发
- 代码安全审查
- K8s 集群配置
- Q1 财务报告分析
- 年度预算编制
- 产品落地页文案撰写
- 技术博客写作
- 商务邮件撰写
- 教案设计（TCP 三次握手）
- 英语技术词汇学习
- 论文写作（LLM 综述）
- 销售数据可视化分析
- SQL 查询优化

## SQLite 语法注意事项

- 使用 `INSERT OR IGNORE` 替代 MySQL 的 `INSERT IGNORE`
- 日期函数：`datetime('now')`、`datetime('now', '-7 days')`、`date('now')`
- 不支持 MySQL 的 `SET @var` 变量语法，主用户 ID 直接硬编码
- 时间修饰符使用 `+/- N days/hours` 格式
- 表名和列名不使用反引号（虽然 SQLite 也兼容）

## 注意事项

- 所有 `INSERT OR IGNORE` 可重复执行不会报错
- 密码字段存储的是 MD5 哈希值（`e10adc3949ba59abbe56e057f20f883e` = `123456` 的 MD5）
- API Key 和 Token 均为 mock 值，不可用于真实 API 调用
- `user_daily_stats` 和 `tool_daily_stats` 的日期使用 `date('now')` 动态计算
