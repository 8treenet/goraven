-- ============================================================
-- Mock Data: Share Links, File Links, Chunk Uploads, Daily Stats
-- 表: share_link, file_link, chunk_upload, user_daily_stats, tool_daily_stats
-- 数据库: SQLite
-- 依赖: 01_users_and_auth.sql, 04_sessions_and_messages.sql
-- 所有数据归属 @user1 = '90a431bee756432492c134f510bad949'
-- ============================================================

-- ============================================================
-- share_link 表 (对话分享链接)
-- 全部归属 @user1
-- ============================================================

INSERT OR IGNORE INTO share_link (shareId, userId, sessionId, title, expiresAt, viewCount, deleted, created, updated) VALUES

('shr_a1b2c3d4', '90a431bee756432492c134f510bad949', 'sess_001', 'Go 项目架构设计讨论 - 完整对话', datetime('now', '+30 days'), 128, 0, datetime('now', '-5 days'), datetime('now')),
('shr_e5f6g7h8', '90a431bee756432492c134f510bad949', 'sess_006', 'Q1 财务报告分析 - 干货分享', datetime('now', '+7 days'), 256, 0, datetime('now', '-2 days'), datetime('now')),
('shr_i9j0k1l2', '90a431bee756432492c134f510bad949', 'sess_004', '代码审查：认证模块安全检查', datetime('now', '-1 day'), 32, 0, datetime('now', '-18 days'), datetime('now')),
('shr_m3n4o5p6', '90a431bee756432492c134f510bad949', 'sess_008', '产品落地页文案撰写 - 参考', datetime('now', '+90 days'), 89, 0, datetime('now', '-1 day'), datetime('now')),
('shr_q7r8s9t0', '90a431bee756432492c134f510bad949', 'sess_011', '教案设计：TCP 三次握手 - 教学参考', datetime('now', '+30 days'), 198, 0, datetime('now', '-2 days'), datetime('now')),
('shr_y5z6a7b8', '90a431bee756432492c134f510bad949', 'sess_013', '论文写作：LLM在软件工程中的应用', datetime('now', '+60 days'), 67, 0, datetime('now', '-3 days'), datetime('now'));

-- ============================================================
-- file_link 表 (文件外链)
-- 全部归属 @user1，涵盖各类文件类型
-- ============================================================

INSERT OR IGNORE INTO file_link (linkId, userId, filePath, fileName, expiresAt, deleted, created, updated) VALUES

-- 财务类文件
('fl_a001', '90a431bee756432492c134f510bad949', '/documents/q1_2026_financial_report.pdf', '2026年Q1财务分析报告.pdf', datetime('now', '+72 hours'), 0, datetime('now', '-5 days'), datetime('now')),
('fl_a002', '90a431bee756432492c134f510bad949', '/documents/budget_2026_draft.xlsx', '2026年度预算初稿.xlsx', datetime('now', '+168 hours'), 0, datetime('now', '-3 days'), datetime('now')),

-- 写作/文档
('fl_a003', '90a431bee756432492c134f510bad949', '/documents/eino_framework_blog.md', 'Eino框架入门博客草稿.md', datetime('now', '+72 hours'), 0, datetime('now', '-2 days'), datetime('now')),
('fl_a004', '90a431bee756432492c134f510bad949', '/documents/client_email_draft.txt', '海外客户合作邮件草稿.txt', datetime('now', '+24 hours'), 0, datetime('now', '-1 day'), datetime('now')),

-- 教育/教案
('fl_a005', '90a431bee756432492c134f510bad949', '/documents/tcp_handshake_lesson_plan.md', 'TCP三次握手教案.md', datetime('now', '+168 hours'), 0, datetime('now', '-3 days'), datetime('now')),
('fl_a006', '90a431bee756432492c134f510bad949', '/documents/english_tech_vocabulary.md', '英语技术词汇学习笔记.md', datetime('now', '+72 hours'), 0, datetime('now', '-6 days'), datetime('now')),

-- 数据/分析
('fl_a007', '90a431bee756432492c134f510bad949', '/documents/sales_h1_2026.csv', '2026上半年销售数据.csv', datetime('now', '+72 hours'), 0, datetime('now', '-4 days'), datetime('now')),
('fl_a008', '90a431bee756432492c134f510bad949', '/projects/sales_report_h1_2026.html', '上半年销售数据可视化报告.html', datetime('now', '+48 hours'), 0, datetime('now', '-1 day'), datetime('now')),

-- 开发/运维
('fl_a009', '90a431bee756432492c134f510bad949', '/projects/raven_deploy.yaml', 'Raven K8s部署配置.yaml', datetime('now', '+72 hours'), 0, datetime('now', '-1 day'), datetime('now')),
('fl_a010', '90a431bee756432492c134f510bad949', '/temp/docker_compose_debug.yaml', 'Docker Compose调试配置.yaml', datetime('now', '+24 hours'), 0, datetime('now', '-5 days'), datetime('now')),

-- 已过期
('fl_a011', '90a431bee756432492c134f510bad949', '/temp/old_screenshot.png', '调试截图_old.png', datetime('now', '-1 hour'), 0, datetime('now', '-7 days'), datetime('now'));

-- ============================================================
-- chunk_upload 表 (分片上传任务)
-- 全部归属 @user1
-- ============================================================

INSERT OR IGNORE INTO chunk_upload (uploadId, userId, fileName, fileSize, chunkSize, totalChunks, tempDir, status, deleted, created, updated) VALUES

-- 大文件上传
('up_a001', '90a431bee756432492c134f510bad949', 'q1_financial_data_pack.zip', 52428800, 5242880, 10, '/raven/user_space/temp/up_a001', 1, 0, datetime('now', '-3 days'), datetime('now')),
('up_a002', '90a431bee756432492c134f510bad949', 'teaching_materials_bundle.zip', 157286400, 5242880, 30, '/raven/user_space/temp/up_a002', 0, 0, datetime('now', '-1 hour'), datetime('now')),
('up_a003', '90a431bee756432492c134f510bad949', 'legacy_sales_data_2024.sql', 104857600, 5242880, 20, '/raven/user_space/temp/up_a003', 2, 0, datetime('now', '-10 days'), datetime('now')),
('up_a004', '90a431bee756432492c134f510bad949', 'paper_references.pdf', 2097152, 1048576, 2, '/raven/user_space/temp/up_a004', 1, 0, datetime('now', '-2 days'), datetime('now')),
('up_a005', '90a431bee756432492c134f510bad949', 'product_demo_video.mp4', 209715200, 5242880, 40, '/raven/user_space/temp/up_a005', 3, 0, datetime('now', '-5 days'), datetime('now'));

-- ============================================================
-- user_daily_stats 表 (用户日统计)
-- 最近30天数据，全部归属 @user1
-- ============================================================

INSERT OR IGNORE INTO user_daily_stats (userId, statDate, promptTokens, promptCachedTokens, completionTokens, messageCount, roundCount, created, updated) VALUES

-- 最近7天 (活动活跃期：多场景使用)
('90a431bee756432492c134f510bad949', date('now'), 52000, 14000, 22000, 95, 28, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-1 day'), 45000, 12000, 18000, 85, 24, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-2 day'), 38000, 10000, 15000, 72, 20, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-3 day'), 42000, 11000, 17000, 68, 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-4 day'), 48000, 13000, 20000, 80, 22, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-5 day'), 35000, 9000, 14000, 55, 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-6 day'), 28000, 7000, 11000, 45, 12, datetime('now'), datetime('now')),

-- 8-30天前 (部分日期活跃)
('90a431bee756432492c134f510bad949', date('now', '-7 day'), 32000, 8000, 13000, 50, 14, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-8 day'), 25000, 6000, 10000, 40, 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-10 day'), 22000, 5000, 9000, 35, 9, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-12 day'), 28000, 7000, 11000, 42, 11, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-14 day'), 40000, 10000, 16000, 60, 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-16 day'), 18000, 4500, 7500, 30, 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-18 day'), 35000, 9000, 14000, 55, 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-21 day'), 30000, 8000, 12000, 48, 13, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-25 day'), 22000, 5000, 9000, 35, 9, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', date('now', '-30 day'), 18000, 4500, 7000, 28, 7, datetime('now'), datetime('now'));

-- ============================================================
-- tool_daily_stats 表 (工具/技能日统计)
-- 全部归属 @user1
-- ============================================================

INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES

-- MCP 工具使用: 文件系统 (最常用)
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now'), 52, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-1 day'), 48, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-2 day'), 38, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-3 day'), 42, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-4 day'), 35, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-5 day'), 28, datetime('now'), datetime('now')),

-- MCP 工具使用: 财务数据
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now'), 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-2 day'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-4 day'), 15, datetime('now'), datetime('now')),

-- MCP 工具使用: PostgreSQL
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now'), 25, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-1 day'), 20, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-2 day'), 28, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-3 day'), 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-4 day'), 15, datetime('now'), datetime('now')),

-- MCP 工具使用: 文档编辑器
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now', '-1 day'), 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now', '-2 day'), 8, datetime('now'), datetime('now')),

-- MCP 工具使用: 学术论文
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now', '-1 day'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now', '-3 day'), 6, datetime('now'), datetime('now')),

-- MCP 工具使用: GitHub
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now', '-1 day'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now', '-2 day'), 12, datetime('now'), datetime('now')),

-- MCP 工具使用: Brave 搜索
('90a431bee756432492c134f510bad949', 'mcp', 'brave-search', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'brave-search', date('now', '-1 day'), 8, datetime('now'), datetime('now')),

-- MCP 工具使用: 翻译服务
('90a431bee756432492c134f510bad949', 'mcp', 'translation', date('now'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'translation', date('now', '-1 day'), 4, datetime('now'), datetime('now')),

-- 技能使用: 财务报表分析
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now', '-2 day'), 10, datetime('now'), datetime('now')),

-- 技能使用: 税务助手
('90a431bee756432492c134f510bad949', 'skill', 'tax-helper', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'tax-helper', date('now', '-2 day'), 4, datetime('now'), datetime('now')),

-- 技能使用: 文章写作
('90a431bee756432492c134f510bad949', 'skill', 'essay-writer', date('now'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'essay-writer', date('now', '-1 day'), 8, datetime('now'), datetime('now')),

-- 技能使用: 创意文案
('90a431bee756432492c134f510bad949', 'skill', 'creative-copywriter', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'creative-copywriter', date('now', '-1 day'), 3, datetime('now'), datetime('now')),

-- 技能使用: 教案设计
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now'), 7, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now', '-3 day'), 9, datetime('now'), datetime('now')),

-- 技能使用: 论文助手
('90a431bee756432492c134f510bad949', 'skill', 'research-paper-helper', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'research-paper-helper', date('now', '-3 day'), 7, datetime('now'), datetime('now')),

-- 技能使用: 数据可视化
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now', '-1 day'), 6, datetime('now'), datetime('now')),

-- 技能使用: Docker Compose
('90a431bee756432492c134f510bad949', 'skill', 'docker-compose-gen', date('now'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'docker-compose-gen', date('now', '-4 day'), 5, datetime('now'), datetime('now')),

-- 技能使用: 演示文稿
('90a431bee756432492c134f510bad949', 'skill', 'presentation-builder', date('now'), 4, datetime('now'), datetime('now')),

-- 技能使用: 商务邮件
('90a431bee756432492c134f510bad949', 'skill', 'business-email-composer', date('now', '-1 day'), 3, datetime('now'), datetime('now')),

-- 工具使用: web_fetch
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now', '-1 day'), 3, datetime('now'), datetime('now')),

-- 全局统计 (userId 为空字符串)
('', 'mcp', 'filesystem', date('now'), 52, datetime('now'), datetime('now')),
('', 'mcp', 'financial-data', date('now'), 18, datetime('now'), datetime('now')),
('', 'mcp', 'postgres', date('now'), 25, datetime('now'), datetime('now')),
('', 'mcp', 'github', date('now'), 10, datetime('now'), datetime('now')),
('', 'mcp', 'document-editor', date('now'), 12, datetime('now'), datetime('now')),
('', 'mcp', 'arxiv-search', date('now'), 8, datetime('now'), datetime('now')),
('', 'skill', 'financial-report-analyzer', date('now'), 8, datetime('now'), datetime('now')),
('', 'skill', 'essay-writer', date('now'), 6, datetime('now'), datetime('now')),
('', 'skill', 'lesson-planner', date('now'), 7, datetime('now'), datetime('now')),
('', 'skill', 'data-viz-assistant', date('now'), 8, datetime('now'), datetime('now')),
('', 'skill', 'tax-helper', date('now'), 5, datetime('now'), datetime('now')),
('', 'skill', 'creative-copywriter', date('now'), 5, datetime('now'), datetime('now'));
