package po

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	SessionId             string    `gorm:"primaryKey;column:session_id;type:varchar(64)"`                                                //主键唯一id
	UserId                string    `gorm:"index;index:idx_user_del_arch,priority:1;column:user_id"`                                      //用户id
	Title                 string    `gorm:"column:title;type:varchar(255)"`                                                               //会话标题
	IsArchived            int       `gorm:"index:idx_user_del_arch,priority:3;column:is_archived"`                                        //是否归档 0否 1是
	PromptTokensCount     int       `gorm:"column:prompt_tokens_count"`                                                                   //累计promptTokensCount：会话级总量，由AddSessionTokens按轮次累加（含压缩、标题生成等消耗）
	CompletionTokensCount int       `gorm:"column:completion_tokens_count"`                                                               //累计completionTokensCount：会话级总量，累加方式同上
	PromptCachedTokens    int       `gorm:"column:prompt_cached_tokens;default:0"`                                                        //累计缓存promptTokens：会话级总量，累加方式同上；与message表prompt_cached_tokens_count为同一指标，命名不同但勿改列名
	ContextTokens         int       `gorm:"column:context_tokens"`                                                                        //当前上下文长度
	Status                uint8     `gorm:"column:status"`                                                                                //会话状态 0正常 1进行中
	AIModelId             int       `gorm:"index;column:ai_model_id"`                                                                     //使用模型的id
	LastChatTime          time.Time `gorm:"not null;column:last_chat_time"`                                                               //最后的聊天时间
	PersonaId             int       `gorm:"column:persona_id;index;default:0"`                                                            //用户角色ID 0表示未选择
	McpIds                string    `gorm:"column:mcp_ids;type:text"`                                                                     //MCP配置ID列表（JSON数组：[1,2,3]），仅无角色时手动选择；有角色时为空，读persona_tool表
	SkillIds              string    `gorm:"column:skill_ids;type:text"`                                                                   //技能ID列表（JSON数组：[1,2,3]），仅无角色时手动选择；有角色时为空，读persona_tool表
	Project               string    `gorm:"column:project;type:varchar(255);default:''"`                                                  //项目目录名称（共享项目时为 owner 的真实目录名）
	SharedProjectId       int       `gorm:"column:shared_project_id;index;default:0"`                                                     //团队项目ID，0表示个人项目
	AutomationTaskId      int       `gorm:"index;column:automation_task_id;default:0"`                                                    //自动化任务ID，0表示普通会话；非0为自动化任务产生的会话，不在侧边栏会话列表展示
	Deleted               uint8     `gorm:"index:idx_user_del_arch,priority:2;index:idx_del_created,priority:1;column:deleted;default:0"` //软删除：0正常 1删除
	Created               time.Time `gorm:"index:idx_del_created,priority:2;not null;column:created"`
	Updated               time.Time `gorm:"not null;column:updated"`
}

// TableName 设置表名
func (session *Session) TableName() string {
	return "session"
}

// BeforeCreate 创建时设置时间戳
func (session *Session) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	session.Created = now
	session.Updated = now
	session.LastChatTime = now
	return nil
}

// BeforeSave 更新时间戳
func (session *Session) BeforeSave(tx *gorm.DB) error {
	session.Updated = time.Now()
	return nil
}
