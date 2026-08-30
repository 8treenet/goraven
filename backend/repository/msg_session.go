package repository

import (
	"context"
	"goraven/backend/po"
	"goraven/backend/vo"
	"goraven/core/iface"
	"goraven/util"
	"time"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *MsgSessionRepository {
			return &MsgSessionRepository{}
		})
	})
}

var _ iface.MessageRepo = new(MsgSessionRepository)

// MsgSessionRepository .
type MsgSessionRepository struct {
	freedom.Repository
}

func (repo *MsgSessionRepository) GetSession(sessionId string) (session *po.Session, err error) {
	err = repo.db().First(&session, "session_id = ?", sessionId).Error
	return
}

// SaveSession 存储会话
func (repo *MsgSessionRepository) SaveSession(session *po.Session) error {
	if session.SessionId == "" {
		session.SessionId = util.UUID()
		return repo.db().Create(session).Error
	}
	return repo.db().Save(session).Error
}

// AddSessionTokens 累加session的token统计
func (repo *MsgSessionRepository) AddSessionTokens(sessionId string, promptTokens, completionTokens, promptCachedTokens int) error {
	return repo.db().Model(&po.Session{}).
		Where("session_id = ?", sessionId).
		Updates(map[string]interface{}{
			"prompt_tokens_count":     gorm.Expr("prompt_tokens_count + ?", promptTokens),
			"completion_tokens_count": gorm.Expr("completion_tokens_count + ?", completionTokens),
			"prompt_cached_tokens":    gorm.Expr("prompt_cached_tokens + ?", promptCachedTokens),
			"updated":                 time.Now(),
		}).Error
}

// SetContextTokens 设置当前上下文长度
func (repo *MsgSessionRepository) SetContextTokens(sessionId string, tokens int) error {
	return repo.db().Model(&po.Session{}).
		Where("session_id = ?", sessionId).
		Updates(map[string]interface{}{
			"context_tokens": tokens,
			"updated":        time.Now(),
		}).Error
}

// UpdateSessionStatus 更新session的状态
func (repo *MsgSessionRepository) UpdateSessionStatus(sessionId string, status int) error {
	updates := map[string]interface{}{
		"status":  status,
		"updated": time.Now(),
	}
	if status == 1 {
		updates["last_chat_time"] = time.Now()
	}
	return repo.db().Model(&po.Session{}).Where("session_id = ?", sessionId).Updates(updates).Error
}

// SaveChatMessage 存储消息
func (repo *MsgSessionRepository) SaveChatMessage(sessionId string, msg *po.Message) error {
	msg.SessionId = sessionId
	if msg.MsgId == "" {
		msg.MsgId = util.UUID()
		return repo.db().Create(msg).Error
	}
	return repo.db().Save(msg).Error
}

// GetChatMessages 获取正序的消息记录
func (repo *MsgSessionRepository) GetChatMessages(sessionId string) ([]*po.Message, error) {
	var messages []*po.Message
	err := repo.db().Where("session_id = ? AND context_state = 0 AND asst_error = ''", sessionId).Order("timestamp asc").Find(&messages).Error
	return messages, err
}

// MarkSessionCompressed 标记session中指定roundIds的消息为已压缩
func (repo *MsgSessionRepository) MarkSessionCompressed(sessionId string, roundIds []string) error {
	if len(roundIds) == 0 {
		return nil
	}
	return repo.db().Model(&po.Message{}).
		Where("session_id = ? AND round_id IN ?", sessionId, roundIds).
		Update("context_state", 1).Error
}

// ════════════════════════════════════════════════════════════════════════════
// 用户端会话 API
// ════════════════════════════════════════════════════════════════════════════

// UpdateSession 更新会话（标题、归档等，只更新传入字段）
func (repo *MsgSessionRepository) UpdateSession(sessionId string, userId string, updates map[string]interface{}) error {
	return repo.db().Model(&po.Session{}).Where("session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Updates(updates).Error
}

// ListSessions 获取用户未归档且未删除的会话列表（侧边栏"所有对话"），分页
// 支持按 personaId 筛选，按 status DESC 置顶进行中的会话，再按 lastChatTime DESC
// 过滤 automation_task_id = 0：自动化任务产生的会话不在侧边栏展示（任务详情页单独查看）
func (repo *MsgSessionRepository) ListSessions(userId string, req *vo.SessionListReq) ([]po.Session, *PageResult, error) {
	query := repo.db().Model(&po.Session{}).Where("user_id = ? AND deleted = 0 AND is_archived = 0 AND automation_task_id = 0", userId)
	if req.PersonaId != nil {
		query = query.Where("persona_id = ?", *req.PersonaId)
	}
	query = query.Order("status DESC, last_chat_time DESC")

	var sessions []po.Session
	pageResult, err := Paginate(query, &PageQuery{Page: req.Page, PageSize: req.PageSize}, &sessions)
	if err != nil {
		return nil, nil, err
	}
	return sessions, pageResult, nil
}

// GetUserSession 获取会话详情（校验 userId 防止越权）
func (repo *MsgSessionRepository) GetUserSession(sessionId string, userId string) (*po.Session, error) {
	var session po.Session
	err := repo.db().First(&session, "session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Error
	return &session, err
}

// GetAllMessages 获取会话消息（用户端展示，不包含压缩摘要）
// 返回 roleType=user/assistant 且 tool=0 的所有消息
func (repo *MsgSessionRepository) GetAllMessages(sessionId string) ([]po.Message, error) {
	var messages []po.Message
	err := repo.db().Where("session_id = ? AND role_type IN ? AND tool = 0", sessionId, []string{po.RoleTypeUser, po.RoleTypeAssistant}).Order("timestamp ASC").Find(&messages).Error
	return messages, err
}

// GetToolReasoningContent 获取同轮次中其他 tool=1 消息的思考内容（排除当前消息）
func (repo *MsgSessionRepository) GetToolReasoningContent(sessionId, roundId, excludeMsgId string) ([]po.Message, error) {
	var messages []po.Message
	err := repo.db().Where("session_id = ? AND round_id = ? AND msg_id != ? AND role_type = ? AND tool = 1 AND reasoning_content != ''",
		sessionId, roundId, excludeMsgId, po.RoleTypeAssistant).Order("timestamp DESC").Limit(30).Find(&messages).Error
	if err != nil {
		return messages, err
	}
	// var parts []string
	// for _, m := range messages {
	// 	if v := strings.TrimSpace(m.ReasoningContent); v != "" {
	// 		parts = append(parts, v)
	// 	}
	// }
	// return strings.Join(parts, "\n\n"), nil
	return messages, err
}

// SoftDeleteSession 软删除会话（deleted=1，title 追加时间戳后缀防冲突）
func (repo *MsgSessionRepository) SoftDeleteSession(sessionId string, userId string) error {
	var session po.Session
	if err := repo.db().First(&session, "session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Error; err != nil {
		return err
	}

	return repo.db().Model(&po.Session{}).Where("session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Updates(map[string]interface{}{
		"deleted": 1,
		"updated": time.Now(),
	}).Error
}

// compressTaskCachePrefix 压缩任务缓存 key 前缀
const compressTaskCachePrefix = "chat:compress:"

// compressTaskCacheTTL 压缩任务缓存过期时间
const compressTaskCacheTTL = 10 * time.Minute

// SetCompressTaskStatus 设置压缩任务缓存状态
func (repo *MsgSessionRepository) SetCompressTaskStatus(taskId string, status string) {
	repo.Redis().Set(context.Background(), compressTaskCachePrefix+taskId, status, compressTaskCacheTTL)
}

// GetCompressTaskStatus 获取压缩任务缓存状态，不存在返回空字符串
func (repo *MsgSessionRepository) GetCompressTaskStatus(taskId string) (string, error) {
	return repo.Redis().Get(context.Background(), compressTaskCachePrefix+taskId).Result()
}

// GetFirstUserMessage 获取会话的第一条用户消息
func (repo *MsgSessionRepository) GetFirstUserMessage(sessionId string) (*po.Message, error) {
	var msg po.Message
	err := repo.db().Where("session_id = ? AND role_type = ?", sessionId, po.RoleTypeUser).
		Order("timestamp ASC").First(&msg).Error
	return &msg, err
}

// db .
func (repo *MsgSessionRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
