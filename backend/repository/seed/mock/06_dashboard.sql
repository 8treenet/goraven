-- ============================================================
-- Mock Data: Dashboard (仪表盘数据)
-- 表: user_daily_stats, tool_daily_stats
-- 数据库: SQLite
-- 依赖: 01_users_and_auth.sql (user 表), 04_sessions_and_messages.sql (session 表)
-- 所有数据归属 @user1 = '90a431bee756432492c134f510bad949'
--
-- 覆盖接口:
--   GET /api/dashboard        → Dashboard()        用户仪表盘聚合数据
--   GET /api/dashboard/tokenTrend → GetTokenTrend() 用户 Token 消耗趋势
-- ============================================================

-- ============================================================
-- user_daily_stats 表 (用户日统计)
-- 最近 30 天完整覆盖，用于仪表盘 Token 趋势、Sparkline 迷你图
-- ============================================================

INSERT OR IGNORE INTO user_daily_stats (userId, statDate, promptTokens, promptCachedTokens, completionTokens, messageCount, roundCount, created, updated) VALUES

-- 今天 (高活跃度)
('90a431bee756432492c134f510bad949', date('now'), 52000, 14000, 22000, 95, 28, datetime('now'), datetime('now')),

-- 最近 1-7 天 (活跃期)
('90a431bee756432492c134f510bad949', date('now', '-1 day'), 45000, 12000, 18000, 85, 24, datetime('now', '-1 day'), datetime('now', '-1 day')),
('90a431bee756432492c134f510bad949', date('now', '-2 days'), 38000, 10000, 15000, 72, 20, datetime('now', '-2 days'), datetime('now', '-2 days')),
('90a431bee756432492c134f510bad949', date('now', '-3 days'), 42000, 11000, 17000, 68, 18, datetime('now', '-3 days'), datetime('now', '-3 days')),
('90a431bee756432492c134f510bad949', date('now', '-4 days'), 48000, 13000, 20000, 80, 22, datetime('now', '-4 days'), datetime('now', '-4 days')),
('90a431bee756432492c134f510bad949', date('now', '-5 days'), 35000, 9000, 14000, 55, 15, datetime('now', '-5 days'), datetime('now', '-5 days')),
('90a431bee756432492c134f510bad949', date('now', '-6 days'), 28000, 7000, 11000, 45, 12, datetime('now', '-6 days'), datetime('now', '-6 days')),

-- 8-14 天前 (中等活跃)
('90a431bee756432492c134f510bad949', date('now', '-7 days'), 32000, 8000, 13000, 50, 14, datetime('now', '-7 days'), datetime('now', '-7 days')),
('90a431bee756432492c134f510bad949', date('now', '-8 days'), 25000, 6000, 10000, 40, 10, datetime('now', '-8 days'), datetime('now', '-8 days')),
('90a431bee756432492c134f510bad949', date('now', '-9 days'), 20000, 5000, 8000, 32, 8, datetime('now', '-9 days'), datetime('now', '-9 days')),
('90a431bee756432492c134f510bad949', date('now', '-10 days'), 22000, 5000, 9000, 35, 9, datetime('now', '-10 days'), datetime('now', '-10 days')),
('90a431bee756432492c134f510bad949', date('now', '-11 days'), 18000, 4000, 7000, 28, 7, datetime('now', '-11 days'), datetime('now', '-11 days')),
('90a431bee756432492c134f510bad949', date('now', '-12 days'), 28000, 7000, 11000, 42, 11, datetime('now', '-12 days'), datetime('now', '-12 days')),
('90a431bee756432492c134f510bad949', date('now', '-13 days'), 15000, 3500, 6000, 25, 6, datetime('now', '-13 days'), datetime('now', '-13 days')),
('90a431bee756432492c134f510bad949', date('now', '-14 days'), 40000, 10000, 16000, 60, 18, datetime('now', '-14 days'), datetime('now', '-14 days')),

-- 15-21 天前 (波动期)
('90a431bee756432492c134f510bad949', date('now', '-15 days'), 12000, 3000, 5000, 20, 5, datetime('now', '-15 days'), datetime('now', '-15 days')),
('90a431bee756432492c134f510bad949', date('now', '-16 days'), 18000, 4500, 7500, 30, 8, datetime('now', '-16 days'), datetime('now', '-16 days')),
('90a431bee756432492c134f510bad949', date('now', '-17 days'), 16000, 4000, 6500, 26, 7, datetime('now', '-17 days'), datetime('now', '-17 days')),
('90a431bee756432492c134f510bad949', date('now', '-18 days'), 35000, 9000, 14000, 55, 15, datetime('now', '-18 days'), datetime('now', '-18 days')),
('90a431bee756432492c134f510bad949', date('now', '-19 days'), 10000, 2500, 4000, 18, 4, datetime('now', '-19 days'), datetime('now', '-19 days')),
('90a431bee756432492c134f510bad949', date('now', '-20 days'), 14000, 3500, 5500, 22, 6, datetime('now', '-20 days'), datetime('now', '-20 days')),
('90a431bee756432492c134f510bad949', date('now', '-21 days'), 30000, 8000, 12000, 48, 13, datetime('now', '-21 days'), datetime('now', '-21 days')),

-- 22-30 天前 (早期)
('90a431bee756432492c134f510bad949', date('now', '-22 days'), 11000, 2500, 4500, 18, 5, datetime('now', '-22 days'), datetime('now', '-22 days')),
('90a431bee756432492c134f510bad949', date('now', '-23 days'), 9000, 2000, 3500, 15, 4, datetime('now', '-23 days'), datetime('now', '-23 days')),
('90a431bee756432492c134f510bad949', date('now', '-24 days'), 13000, 3000, 5000, 20, 5, datetime('now', '-24 days'), datetime('now', '-24 days')),
('90a431bee756432492c134f510bad949', date('now', '-25 days'), 22000, 5000, 9000, 35, 9, datetime('now', '-25 days'), datetime('now', '-25 days')),
('90a431bee756432492c134f510bad949', date('now', '-26 days'), 8000, 2000, 3000, 12, 3, datetime('now', '-26 days'), datetime('now', '-26 days')),
('90a431bee756432492c134f510bad949', date('now', '-27 days'), 17000, 4000, 7000, 28, 8, datetime('now', '-27 days'), datetime('now', '-27 days')),
('90a431bee756432492c134f510bad949', date('now', '-28 days'), 12000, 3000, 5000, 20, 5, datetime('now', '-28 days'), datetime('now', '-28 days')),
('90a431bee756432492c134f510bad949', date('now', '-29 days'), 15000, 3500, 6000, 24, 6, datetime('now', '-29 days'), datetime('now', '-29 days')),
('90a431bee756432492c134f510bad949', date('now', '-30 days'), 18000, 4500, 7000, 28, 7, datetime('now', '-30 days'), datetime('now', '-30 days'));

-- ============================================================
-- tool_daily_stats 表 (工具/技能/MCP 日统计)
-- 覆盖近 7 天的技能(skill)、MCP(mcp)、工具(tool) 使用数据
-- 支撑仪表盘技能/MCP/工具使用排行
-- ============================================================

-- ── MCP 工具使用 ──

-- filesystem (最常用)
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now'), 52, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-1 day'), 48, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-2 days'), 38, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-3 days'), 42, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-4 days'), 35, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-5 days'), 28, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'filesystem', date('now', '-6 days'), 20, datetime('now'), datetime('now'));

-- postgres
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now'), 25, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-1 day'), 20, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-2 days'), 28, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-3 days'), 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-4 days'), 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-5 days'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'postgres', date('now', '-6 days'), 10, datetime('now'), datetime('now'));

-- financial-data
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now'), 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-1 day'), 14, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-2 days'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-3 days'), 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'financial-data', date('now', '-4 days'), 22, datetime('now'), datetime('now'));

-- document-editor
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now', '-1 day'), 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now', '-2 days'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'document-editor', date('now', '-3 days'), 10, datetime('now'), datetime('now'));

-- github
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now', '-1 day'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now', '-2 days'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'github', date('now', '-4 days'), 5, datetime('now'), datetime('now'));

-- arxiv-search
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now', '-1 day'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'arxiv-search', date('now', '-3 days'), 6, datetime('now'), datetime('now'));

-- brave-search
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'brave-search', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'brave-search', date('now', '-1 day'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'brave-search', date('now', '-3 days'), 3, datetime('now'), datetime('now'));

-- translation
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'translation', date('now'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'translation', date('now', '-1 day'), 4, datetime('now'), datetime('now'));

-- weather
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'weather', date('now'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'weather', date('now', '-3 days'), 2, datetime('now'), datetime('now'));

-- email
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'mcp', 'email', date('now'), 4, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'mcp', 'email', date('now', '-2 days'), 6, datetime('now'), datetime('now'));

-- ── 技能(skill) 使用 ──

-- financial-report-analyzer (最常用技能)
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now', '-1 day'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now', '-2 days'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now', '-3 days'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'financial-report-analyzer', date('now', '-4 days'), 7, datetime('now'), datetime('now'));

-- data-viz-assistant
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now', '-1 day'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now', '-2 days'), 9, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'data-viz-assistant', date('now', '-4 days'), 4, datetime('now'), datetime('now'));

-- lesson-planner
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now'), 7, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now', '-1 day'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now', '-3 days'), 9, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'lesson-planner', date('now', '-4 days'), 4, datetime('now'), datetime('now'));

-- essay-writer
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'essay-writer', date('now'), 6, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'essay-writer', date('now', '-1 day'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'essay-writer', date('now', '-2 days'), 4, datetime('now'), datetime('now'));

-- research-paper-helper
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'research-paper-helper', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'research-paper-helper', date('now', '-1 day'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'research-paper-helper', date('now', '-3 days'), 7, datetime('now'), datetime('now'));

-- tax-helper
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'tax-helper', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'tax-helper', date('now', '-2 days'), 4, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'tax-helper', date('now', '-4 days'), 3, datetime('now'), datetime('now'));

-- creative-copywriter
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'creative-copywriter', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'creative-copywriter', date('now', '-1 day'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'creative-copywriter', date('now', '-2 days'), 6, datetime('now'), datetime('now'));

-- go-project-scaffold
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'go-project-scaffold', date('now'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'go-project-scaffold', date('now', '-2 days'), 5, datetime('now'), datetime('now'));

-- react-component-gen
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'react-component-gen', date('now'), 2, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'react-component-gen', date('now', '-4 days'), 4, datetime('now'), datetime('now'));

-- docker-compose-gen
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'docker-compose-gen', date('now'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'docker-compose-gen', date('now', '-4 days'), 5, datetime('now'), datetime('now'));

-- budget-planner
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'budget-planner', date('now'), 4, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'budget-planner', date('now', '-2 days'), 3, datetime('now'), datetime('now'));

-- business-email-composer
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'business-email-composer', date('now'), 2, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'business-email-composer', date('now', '-1 day'), 3, datetime('now'), datetime('now'));

-- presentation-builder
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'presentation-builder', date('now'), 4, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'skill', 'presentation-builder', date('now', '-3 days'), 2, datetime('now'), datetime('now'));

-- file-organizer
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'file-organizer', date('now'), 3, datetime('now'), datetime('now'));

-- k8s-manifest-gen
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'k8s-manifest-gen', date('now', '-5 days'), 3, datetime('now'), datetime('now'));

-- flashcard-gen
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'skill', 'flashcard-gen', date('now', '-6 days'), 2, datetime('now'), datetime('now'));

-- ── 内置工具(tool) 使用 ──

-- web_fetch
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now'), 5, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now', '-1 day'), 3, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now', '-2 days'), 7, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'web_fetch', date('now', '-4 days'), 4, datetime('now'), datetime('now'));

-- read_file
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'read_file', date('now'), 18, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'read_file', date('now', '-1 day'), 15, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'read_file', date('now', '-2 days'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'read_file', date('now', '-3 days'), 20, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'read_file', date('now', '-4 days'), 10, datetime('now'), datetime('now'));

-- write_file
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'write_file', date('now'), 12, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'write_file', date('now', '-1 day'), 10, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'write_file', date('now', '-2 days'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'write_file', date('now', '-3 days'), 14, datetime('now'), datetime('now'));

-- delete_file
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'delete_file', date('now'), 2, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'delete_file', date('now', '-2 days'), 3, datetime('now'), datetime('now'));

-- file_url
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'file_url', date('now'), 4, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'file_url', date('now', '-1 day'), 3, datetime('now'), datetime('now'));

-- unique_id
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'unique_id', date('now'), 8, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'unique_id', date('now', '-1 day'), 6, datetime('now'), datetime('now'));

-- copy_file
INSERT OR IGNORE INTO tool_daily_stats (userId, toolType, toolName, statDate, count, created, updated) VALUES
('90a431bee756432492c134f510bad949', 'tool', 'copy_file', date('now'), 1, datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'tool', 'copy_file', date('now', '-3 days'), 2, datetime('now'), datetime('now'));
