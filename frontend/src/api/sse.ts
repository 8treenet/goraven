import { fetchEventSource } from '@microsoft/fetch-event-source'
import { useUserStore } from '@/stores/user-store'
import type { SSEEvent } from './types'

export interface ChatStreamHandlers {
  onConnected?: (sessionId: string) => void
  onReasoning?: (content: string) => void
  onContent?: (content: string) => void
  onTool?: (name: string, displayName: string, icon: string, action: string) => void
  onRetry?: (attempt: number, maxRetries: number, error: string) => void
  onContext?: (tokens: number, limit: number) => void
  onEnd?: () => void
  onError?: (error: Error) => void
}

export function connectChatStream(sessionId: string, handlers: ChatStreamHandlers) {
  const token = useUserStore.getState().token
  const url = `/api/chat/${sessionId}/stream`

  const controller = new AbortController()

  fetchEventSource(url, {
    headers: { Authorization: `Bearer ${token}` },
    signal: controller.signal,
    openWhenHidden: true,
    onmessage(event) {
      const data = JSON.parse(event.data) as SSEEvent
      switch (data.type) {
        case 'connected':
          handlers.onConnected?.(data.sessionId)
          break
        case 'reasoning':
          handlers.onReasoning?.(data.content)
          break
        case 'content':
          handlers.onContent?.(data.content)
          break
        case 'tool':
          handlers.onTool?.(data.tool.name, data.tool.displayName, data.tool.icon, data.tool.action)
          break
        case 'retry':
          handlers.onRetry?.(data.retry.attempt, data.retry.maxRetries, data.retry.error)
          break
        case 'end':
          handlers.onEnd?.()
          controller.abort()
          break
        case 'context':
          handlers.onContext?.(data.context.tokens, data.context.limit)
          break
        case 'heartbeat':
          // Keep-alive signal to prevent nginx/proxy timeout — no action needed
          break
      }
    },
    onerror(error) {
      // AbortError means intentional stop — don't notify or retry
      if (error?.name === 'AbortError') return
      handlers.onError?.(error)
      // Don't throw — throwing triggers automatic retry, which is not desired for SSE
    },
    onclose() {
      handlers.onError?.(new Error('SSE connection closed'))
    },
  })

  return controller
}
