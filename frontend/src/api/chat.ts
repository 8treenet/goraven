import http from './http'
import type { ChatRequest, ChatResponse, Message } from './types'

/** POST /api/chat */
export function createChat(data: ChatRequest) {
  return http.post<ChatResponse>('/chat', data)
}

/** POST /api/chat/stop */
export function stopChat(sessionId: string) {
  return http.post('/chat/stop', { sessionId })
}

/** POST /api/chat/compress */
export function compressChat(sessionId: string) {
  return http.post<{ taskId: string }>('/chat/compress', { sessionId })
}

/** GET /api/chat/compress/:taskId */
export function getCompressStatus(taskId: string) {
  return http.get<{ status: string }>(`/chat/compress/${taskId}`)
}

/** GET /api/sessions/:sessionId/messages */
export function getSessionMessages(sessionId: string) {
  return http.get<Message[]>(`/sessions/${sessionId}/messages`)
}
