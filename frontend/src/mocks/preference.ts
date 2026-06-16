import { delay } from './delay'
import type { ApiResponse, PreferenceData } from './types'

export async function getPreference(): Promise<ApiResponse<PreferenceData>> {
  await delay(80)
  return {
    code: 0,
    msg: 'ok',
    data: {
      language: 'en',
      domain: 'https://raven.local',
    },
  }
}
