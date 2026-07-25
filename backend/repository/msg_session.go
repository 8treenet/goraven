package repository

import (
	"context"
	"goraven/backend/po"
	"goraven/backend/vo"
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

type MsgSessionRepository struct {
	freedom.Repository
}

func (repo *MsgSessionRepository) GetSession(sessionId string) (session *po.Session, err error) {
	err = repo.db().First(&session, "session_id = ?", sessionId).Error
	return
}

func (repo *MsgSessionRepository) SaveSession(session *po.Session) error {
	if session.SessionId == "" {
		session.SessionId = util.UUID()
		return repo.db().Create(session).Error
	}
	return repo.db().Save(session).Error
}

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

func (repo *MsgSessionRepository) SetContextTokens(sessionId string, tokens int) error {
	return repo.db().Model(&po.Session{}).
		Where("session_id = ?", sessionId).
		Updates(map[string]interface{}{
			"context_tokens": tokens,
			"updated":        time.Now(),
		}).Error
}

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

func (repo *MsgSessionRepository) SaveChatMessage(sessionId string, msg *po.Message) error {
	msg.SessionId = sessionId
	if msg.MsgId == "" {
		msg.MsgId = util.UUID()
		return repo.db().Create(msg).Error
	}
	return repo.db().Save(msg).Error
}

func (repo *MsgSessionRepository) GetChatMessages(sessionId string) ([]*po.Message, error) {
	var messages []*po.Message
	err := repo.db().Where("session_id = ? AND context_state = 0 AND asst_error = ''", sessionId).Order("timestamp asc").Find(&messages).Error
	return messages, err
}

func (repo *MsgSessionRepository) MarkSessionCompressed(sessionId string, roundIds []string) error {
	if len(roundIds) == 0 {
		return nil
	}
	return repo.db().Model(&po.Message{}).
		Where("session_id = ? AND round_id IN ?", sessionId, roundIds).
		Update("context_state", 1).Error
}

func (repo *MsgSessionRepository) UpdateSession(sessionId string, userId string, updates map[string]interface{}) error {
	return repo.db().Model(&po.Session{}).Where("session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Updates(updates).Error
}

func (repo *MsgSessionRepository) ListSessions(userId string, req *vo.SessionListReq) ([]po.Session, *PageResult, error) {
	query := repo.db().Model(&po.Session{}).Where("user_id = ? AND deleted = 0 AND is_archived = 0", userId)
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

func (repo *MsgSessionRepository) GetUserSession(sessionId string, userId string) (*po.Session, error) {
	var session po.Session
	err := repo.db().First(&session, "session_id = ? AND user_id = ? AND deleted = 0", sessionId, userId).Error
	return &session, err
}

func (repo *MsgSessionRepository) GetAllMessages(sessionId string) ([]po.Message, error) {
	var messages []po.Message
	err := repo.db().Where("session_id = ? AND role_type IN ? AND tool = 0", sessionId, []string{po.RoleTypeUser, po.RoleTypeAssistant}).Order("timestamp ASC").Find(&messages).Error
	return messages, err
}

func (repo *MsgSessionRepository) GetToolReasoningContent(sessionId, roundId, excludeMsgId string) ([]po.Message, error) {
	var messages []po.Message
	err := repo.db().Where("session_id = ? AND round_id = ? AND msg_id != ? AND role_type = ? AND tool = 1 AND reasoning_content != ''",
		sessionId, roundId, excludeMsgId, po.RoleTypeAssistant).Order("timestamp DESC").Limit(30).Find(&messages).Error
	if err != nil {
		return messages, err
	}

	return messages, err
}

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

const compressTaskCachePrefix = "chat:compress:"

const compressTaskCacheTTL = 10 * time.Minute

func (repo *MsgSessionRepository) SetCompressTaskStatus(taskId string, status string) {
	repo.Redis().Set(context.Background(), compressTaskCachePrefix+taskId, status, compressTaskCacheTTL)
}

func (repo *MsgSessionRepository) GetCompressTaskStatus(taskId string) (string, error) {
	return repo.Redis().Get(context.Background(), compressTaskCachePrefix+taskId).Result()
}

func (repo *MsgSessionRepository) GetFirstUserMessage(sessionId string) (*po.Message, error) {
	var msg po.Message
	err := repo.db().Where("session_id = ? AND role_type = ?", sessionId, po.RoleTypeUser).
		Order("timestamp ASC").First(&msg).Error
	return &msg, err
}

func (repo *MsgSessionRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
