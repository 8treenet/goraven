import http from './http'
import type { PublicShare, Message } from './types'

/** GET /api/share/:shareId — 分享信息（无需鉴权） */
export function getShareInfo(shareId: string) {
  return http.get<PublicShare>(`/share/${shareId}`)
}

/** GET /api/share/:shareId/messages — 公开分享的消息（无需鉴权） */
export function getShareMessages(shareId: string) {
  return http.get<Message[]>(`/share/${shareId}/messages`)
}

/** GET /api/share/:shareId/internalMessages — 内部分享的消息（需要鉴权） */
export function getInternalShareMessages(shareId: string) {
  return http.get<Message[]>(`/share/${shareId}/internalMessages`)
}
