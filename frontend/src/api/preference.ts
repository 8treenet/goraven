import http from './http'
import type { PreferenceData } from './types'

/** GET /api/preference */
export function getPreference() {
  return http.get<PreferenceData>('/preference')
}
