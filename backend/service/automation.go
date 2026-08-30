package service

import (
	"strings"
	"time"

	"goraven/backend/dependency"
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/util"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *AutomationService {
			return &AutomationService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *AutomationService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// AutomationService 自动化任务服务
type AutomationService struct {
	Worker             freedom.Worker
	AutomationTaskRepo *repository.AutomationTaskRepository
	AutomationExecRepo *repository.AutomationExecutionRepository
	MsgSessionRepo     *repository.MsgSessionRepository
	ModelRepo          *repository.ProviderRepository
	PersonaRepo        *repository.PersonaRepository
	McpRepo            *repository.MCPRepository
	SkillRepo          *repository.SkillRepository
	TeamProjectRepo    *repository.TeamProjectRepository
}

// newItem 任务 PO 转列表条目 VO（不含 requirement 与展示名称）
func newItem(task *po.AutomationTask) vo.AutomationTaskItem {
	return vo.AutomationTaskItem{
		Id:              task.Id,
		Title:           task.Title,
		UserId:          task.UserId,
		ExecType:        task.ExecType,
		RunAt:           task.RunAt,
		IntervalMinutes: task.IntervalMinutes,
		FixedTime:       task.FixedTime,
		Weekday:         task.Weekday,
		McpIds:          task.McpIds,
		SkillIds:        task.SkillIds,
		Project:         task.Project,
		SharedProjectId: task.SharedProjectId,
		AIModelId:       task.AIModelId,
		PersonaId:       task.PersonaId,
		NextRunAt:       task.NextRunAt,
		Status:          task.Status,
		Deleted:         task.Deleted,
		Created:         task.Created,
		Updated:         task.Updated,
	}
}

// displayLookups 展示字段批量查询结果：各类 ID -> 可展示名称
type displayLookups struct {
	modelNames   map[int]string
	personaNames map[int]string
	projectNames map[int]string
	mcpNames     map[int]string
	skillNames   map[int]string
}

// loadDisplayLookups 汇总任务的各类关联 ID，批量 IN 查询构建名称映射
// 查询失败静默降级为空映射，不影响列表主体展示
func (service *AutomationService) loadDisplayLookups(userId string, tasks []po.AutomationTask) displayLookups {
	modelIds := make([]int, 0)
	personaIds := make([]int, 0)
	projectIds := make([]int, 0)
	mcpIds := make([]int, 0)
	skillIds := make([]int, 0)
	for i := range tasks {
		t := &tasks[i]
		if t.AIModelId > 0 {
			modelIds = append(modelIds, t.AIModelId)
		}
		if t.PersonaId > 0 {
			personaIds = append(personaIds, t.PersonaId)
		}
		if t.SharedProjectId > 0 {
			projectIds = append(projectIds, t.SharedProjectId)
		}
		mcpIds = append(mcpIds, parseJSONInts(t.McpIds)...)
		skillIds = append(skillIds, parseJSONInts(t.SkillIds)...)
	}
	modelIds = util.DedupInts(modelIds)
	personaIds = util.DedupInts(personaIds)
	projectIds = util.DedupInts(projectIds)
	mcpIds = util.DedupInts(mcpIds)
	skillIds = util.DedupInts(skillIds)

	lookups := displayLookups{
		modelNames:   map[int]string{},
		personaNames: map[int]string{},
		projectNames: map[int]string{},
		mcpNames:     map[int]string{},
		skillNames:   map[int]string{},
	}
	if len(modelIds) > 0 {
		if models, err := service.ModelRepo.GetModelsByIDs(modelIds); err == nil {
			for _, m := range models {
				lookups.modelNames[m.AIModelId] = m.DisplayName
			}
		}
	}
	if len(personaIds) > 0 {
		if personas, err := service.PersonaRepo.GetUserPersonasByIDs(personaIds); err == nil {
			for _, p := range personas {
				lookups.personaNames[p.PersonaId] = p.Name
			}
		}
	}
	if len(projectIds) > 0 {
		if projects, err := service.TeamProjectRepo.GetByIDs(projectIds); err == nil {
			for _, p := range projects {
				lookups.projectNames[p.Id] = p.ProjectName
			}
		}
	}
	if len(mcpIds) > 0 {
		if mcps, err := service.McpRepo.GetMCPEndpointsByIDs(mcpIds); err == nil {
			for _, m := range mcps {
				if m.DisplayName != "" {
					lookups.mcpNames[m.McpId] = m.DisplayName
				} else {
					lookups.mcpNames[m.McpId] = m.Name
				}
			}
		}
	}
	if len(skillIds) > 0 {
		if names, err := service.SkillRepo.GetUserSkillNameMapByIDs(userId, skillIds); err == nil {
			lookups.skillNames = names
		}
	}
	return lookups
}

// fillDisplayFields 依据名称映射填充条目的展示字段，缺失的引用留空
func fillDisplayFields(item *vo.AutomationTaskItem, lookups displayLookups) {
	item.AIModelName = lookups.modelNames[item.AIModelId]
	item.PersonaName = lookups.personaNames[item.PersonaId]
	item.SharedProjectName = lookups.projectNames[item.SharedProjectId]
	item.McpNames = make([]string, 0)
	for _, id := range parseJSONInts(item.McpIds) {
		if name, ok := lookups.mcpNames[id]; ok {
			item.McpNames = append(item.McpNames, name)
		}
	}
	item.SkillNames = make([]string, 0)
	for _, id := range parseJSONInts(item.SkillIds) {
		if name, ok := lookups.skillNames[id]; ok {
			item.SkillNames = append(item.SkillNames, name)
		}
	}
}

// ListTasks 分页查询当前用户的自动化任务列表（不含需求描述）
func (service *AutomationService) ListTasks(userId string, req *vo.AutomationTaskListReq) (*infra.PageResponse, error) {
	tasks, pr, err := service.AutomationTaskRepo.ListTasks(userId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]vo.AutomationTaskItem, 0, len(tasks))
	for i := range tasks {
		items = append(items, newItem(&tasks[i]))
	}
	lookups := service.loadDisplayLookups(userId, tasks)
	for i := range items {
		fillDisplayFields(&items[i], lookups)
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// GetTask 获取自动化任务详情（含需求描述）
func (service *AutomationService) GetTask(id int, userId string) (*vo.AutomationTaskDetail, error) {
	task, err := service.AutomationTaskRepo.GetTask(id, userId)
	if err != nil {
		return nil, errs.ErrAutomationTaskNotFound
	}
	item := newItem(task)
	fillDisplayFields(&item, service.loadDisplayLookups(userId, []po.AutomationTask{*task}))
	return &vo.AutomationTaskDetail{
		AutomationTaskItem: item,
		Requirement:        task.Requirement,
	}, nil
}

// ListExecutions 分页查询任务的执行记录（仅 id 与开始/完成时间）
func (service *AutomationService) ListExecutions(taskId int, userId string, req *vo.AutomationExecutionListReq) (*infra.PageResponse, error) {
	if _, err := service.AutomationTaskRepo.GetTask(taskId, userId); err != nil {
		return nil, errs.ErrAutomationTaskNotFound
	}
	records, pr, err := service.AutomationExecRepo.ListByTaskId(taskId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]vo.AutomationExecutionItem, 0, len(records))
	for _, r := range records {
		items = append(items, vo.AutomationExecutionItem{
			Id:         r.Id,
			StartedAt:  r.StartedAt,
			FinishedAt: r.FinishedAt,
		})
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// GetExecutionQA 获取执行记录的问答对（首条用户问题 + 助手最终回复的内容）
func (service *AutomationService) GetExecutionQA(taskId, executionId int, userId string) (*vo.AutomationQARsp, error) {
	if _, err := service.AutomationTaskRepo.GetTask(taskId, userId); err != nil {
		return nil, errs.ErrAutomationTaskNotFound
	}
	execution, err := service.AutomationExecRepo.GetById(executionId)
	if err != nil || execution.AutomationTaskId != taskId {
		return nil, errs.ErrAutomationExecutionNotFound
	}

	messages, err := service.MsgSessionRepo.GetAllMessages(execution.SessionId)
	if err != nil {
		return nil, err
	}
	question, answer := extractQA(messages)
	return &vo.AutomationQARsp{Question: question, Answer: answer}, nil
}

// DeleteTask 删除自动化任务（软删除）
func (service *AutomationService) DeleteTask(id int, userId string) error {
	if _, err := service.AutomationTaskRepo.GetTask(id, userId); err != nil {
		return errs.ErrAutomationTaskNotFound
	}
	return service.AutomationTaskRepo.SoftDelete(id, userId)
}

// UpdateTaskStatus 启用/停用自动化任务
// 停用保留 NextRunAt 原值（靠 status 过滤调度），启用时重算一次；重复设置幂等
func (service *AutomationService) UpdateTaskStatus(id int, userId string, req *vo.AutomationTaskStatusReq) error {
	if req.Status != po.AutomationStatusDisabled && req.Status != po.AutomationStatusEnabled {
		return errs.ErrAutomationTaskStatusInvalid
	}
	task, err := service.AutomationTaskRepo.GetTask(id, userId)
	if err != nil {
		return errs.ErrAutomationTaskNotFound
	}
	if task.Status == po.AutomationStatusDone {
		return errs.ErrAutomationTaskCompleted
	}
	if task.Status == req.Status {
		return nil
	}
	// 启用时校验角色模式引用的角色是否仍存在，已删除则拒绝启用并提示新建任务
	if req.Status == po.AutomationStatusEnabled && task.PersonaId > 0 {
		if _, err := service.PersonaRepo.GetUserPersonaByID(task.PersonaId, userId); err != nil {
			return errs.ErrAutomationTaskPersonaDeleted
		}
	}
	var next *time.Time
	if req.Status == po.AutomationStatusEnabled {
		nextRunAt, err := repository.CalcNextRunAt(task, time.Now())
		if err != nil {
			return err
		}
		next = &nextRunAt
	}
	return service.AutomationTaskRepo.UpdateStatus(id, userId, req.Status, next)
}

// UpdateTaskRequirement 修改任务需求描述，不改执行计划与 NextRunAt；
// 内容与原值相同视为幂等成功
func (service *AutomationService) UpdateTaskRequirement(id int, userId string, req *vo.AutomationTaskRequirementReq) error {
	if strings.TrimSpace(req.Requirement) == "" {
		return errs.ErrAutomationTaskRequirementEmpty
	}
	task, err := service.AutomationTaskRepo.GetTask(id, userId)
	if err != nil {
		return errs.ErrAutomationTaskNotFound
	}
	if task.Requirement == req.Requirement {
		return nil
	}
	task.Requirement = req.Requirement
	return service.AutomationTaskRepo.UpdateTask(task, false)
}

// ExecuteTask 立即执行任务：复用调度器同一执行链路（静默后台执行、防重复推演、
// 成功后写执行记录）。任务执行中拒绝重复触发；单次任务执行后置为已完成。
func (service *AutomationService) ExecuteTask(id int, userId string, chat dependency.Chat) error {
	task, err := service.AutomationTaskRepo.GetTask(id, userId)
	if err != nil {
		return errs.ErrAutomationTaskNotFound
	}
	if task.Status == po.AutomationStatusDone {
		return errs.ErrAutomationTaskCompleted
	}
	// 执行中（调度触发或上次立即执行未结束）拒绝重复触发；
	// 锁查询失败不拦截，由 AutomationTask 内的 SetNX 兜底
	locked, lockErr := service.AutomationTaskRepo.IsTaskLocked(id)
	if lockErr == nil && locked {
		return errs.ErrAutomationTaskRunning
	}
	service.Worker.DeferRecycle()
	go func() {
		// 与调度器一致的触发语义：锁仲裁、单次置完成/周期占位推演、静默执行
		chat.AutomationTask(task, time.Now())
	}()
	return nil
}

// extractQA 从会话消息中提取第一轮问答对：首条用户消息为问题，最后一条助手消息为回复
// 自动化任务不支持连续对话，会话内仅一轮用户交互
func extractQA(messages []po.Message) (question, answer string) {
	for _, m := range messages {
		switch m.RoleType {
		case po.RoleTypeUser:
			if question == "" {
				question = m.Content
			}
		case po.RoleTypeAssistant:
			answer = m.Content
		}
	}
	return
}
