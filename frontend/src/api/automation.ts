import http from './http'
import type {
  PaginationParams,
  PaginatedResponse,
  AutomationTaskItem,
  AutomationTaskDetail,
  AutomationExecutionItem,
  AutomationAnswerRsp,
} from './types'

/** GET /api/automationTasks?page=1&pageSize=20&status=1（status 不传时不过滤） */
export function getAutomationTasks(params?: PaginationParams & { status?: number }) {
  return http.get<PaginatedResponse<AutomationTaskItem>>('/automationTasks', { params })
}

/** GET /api/automationTasks/:id */
export function getAutomationTask(id: number) {
  return http.get<AutomationTaskDetail>(`/automationTasks/${id}`)
}

/** GET /api/automationTasks/:id/executions?page=1&pageSize=20 */
export function getTaskExecutions(id: number, params?: PaginationParams) {
  return http.get<PaginatedResponse<AutomationExecutionItem>>(`/automationTasks/${id}/executions`, { params })
}

/** GET /api/automationTasks/:id/executions/:executionId/messages */
export function getExecutionAnswer(taskId: number, executionId: number) {
  return http.get<AutomationAnswerRsp>(`/automationTasks/${taskId}/executions/${executionId}/messages`)
}

/** DELETE /api/automationTasks/:id */
export function deleteTask(id: number) {
  return http.delete(`/automationTasks/${id}`)
}

/** PUT /api/automationTasks/:id/status */
export function updateTaskStatus(id: number, status: number) {
  return http.put<{ status: string }>(`/automationTasks/${id}/status`, { status })
}

/** PUT /api/automationTasks/:id/requirement */
export function updateTaskRequirement(id: number, requirement: string) {
  return http.put<{ status: string }>(`/automationTasks/${id}/requirement`, { requirement })
}

/** POST /api/automationTasks/:id/execute */
export function executeTask(id: number) {
  return http.post<{ status: string }>(`/automationTasks/${id}/execute`)
}
