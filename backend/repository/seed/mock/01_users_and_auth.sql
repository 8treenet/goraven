-- ============================================================
-- Mock Data: Users & Auth
-- 表: user, user_auth
-- 数据库: SQLite
-- 使用方式:
--   sqlite3 /path/to/raven.db < 01_users_and_auth.sql
-- ============================================================

-- 定义主用户变量（该用户将拥有所有会话、角色、技能数据）
-- 注意：SQLite 不支持 SET 变量，此处仅作标记，实际值已硬编码到各条 INSERT 中
-- @user1 = '90a431bee756432492c134f510bad949'

-- ============================================================
-- user 表
-- 6个用户: 超级管理员、财务、编辑、教师、开发、数据分析
-- 密码均为 123456 的 MD5: e10adc3949ba59abbe56e057f20f883e
-- ============================================================

INSERT OR IGNORE INTO user (userId, username, password, email, role, superAdmin, status, nickname, avatar, deleted, created, updated) VALUES

-- 2. 财务人员
('u_f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c1', 'lina', 'e10adc3949ba59abbe56e057f20f883e', 'lina@example.com', 0, 0, 1, '李财务', '', 0, datetime('now', '-30 days'), datetime('now')),

-- 3. 内容编辑
('u_a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d2', 'wangwei', 'e10adc3949ba59abbe56e057f20f883e', 'wangwei@example.com', 0, 0, 1, '王编辑', '', 0, datetime('now', '-20 days'), datetime('now')),

-- 4. 教师
('u_b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e3', 'zhaolaoshi', 'e10adc3949ba59abbe56e057f20f883e', 'zhao@example.com', 0, 0, 1, '赵老师', '', 0, datetime('now', '-15 days'), datetime('now')),

-- 5. 开发工程师
('u_c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f4', 'chen_dev', 'e10adc3949ba59abbe56e057f20f883e', 'chen@example.com', 1, 0, 1, '陈开发', '', 0, datetime('now', '-10 days'), datetime('now')),

-- 6. 数据分析师
('u_d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a5', 'liu_data', 'e10adc3949ba59abbe56e057f20f883e', 'liu@example.com', 0, 0, 1, '刘数据', '', 0, datetime('now', '-5 days'), datetime('now'));

-- ============================================================
-- user_auth 表
-- 所有 token 归属 @user1 (admin)，模拟多端登录和历史 token
-- ============================================================

INSERT OR IGNORE INTO user_auth (userId, accessToken, expiresAt, clientIP, clientUA, created, updated) VALUES

-- 当前有效的 token (3个: Mac + Windows + Mobile)
('90a431bee756432492c134f510bad949', 'rvn_admin_current_00000000000000000000000000000001', datetime('now', '+30 days'), '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/130.0.0.0', datetime('now'), datetime('now')),
('90a431bee756432492c134f510bad949', 'rvn_admin_windows_00000000000000000000000000000002', datetime('now', '+30 days'), '172.16.0.20', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edge/130.0', datetime('now', '-2 days'), datetime('now')),
('90a431bee756432492c134f510bad949', 'rvn_admin_mobile_0000000000000000000000000000000003', datetime('now', '+30 days'), '10.0.0.55', 'RavenApp/1.0.0 (iPhone; iOS 18.0)', datetime('now', '-1 day'), datetime('now')),

-- 历史 token (3个: 已过期或旧设备)
('90a431bee756432492c134f510bad949', 'rvn_admin_old_mac_00000000000000000000000000000000004', datetime('now', '-5 days'), '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/129.0.0.0', datetime('now', '-30 days'), datetime('now')),
('90a431bee756432492c134f510bad949', 'rvn_admin_old_linux_0000000000000000000000000000000005', datetime('now', '-10 days'), '10.10.10.88', 'Mozilla/5.0 (X11; Linux x86_64) Firefox/131.0', datetime('now', '-15 days'), datetime('now')),
('90a431bee756432492c134f510bad949', 'rvn_admin_ipad_0000000000000000000000000000000000006', datetime('now', '+7 days'), '192.168.1.200', 'Mozilla/5.0 (iPad; iPadOS 18.0) Safari/605.1', datetime('now', '-3 days'), datetime('now'));