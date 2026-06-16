import http from './http'
import type { PublicShare } from './types'

/** GET /api/share/:shareId */
export function getSharedSession(shareId: string) {
  return http.get<PublicShare>(`/share/${shareId}`)
}
