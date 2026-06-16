import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Loader2, Eye, TriangleAlert, Clock } from 'lucide-react'
import { shareApi } from '@/api'
import type { PublicShare, Message as ApiMessage } from '@/api/types'
import { ReadonlyMessageBlock, type ReadonlyMessage } from '@/features/chat/ReadonlyMessageView'
import { useT } from '@/i18n'

function formatDateTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toReadonlyMessage(m: ApiMessage): ReadonlyMessage {
  const thinkingSegments = m.reasoningContent?.length
    ? m.reasoningContent.map((item) => {
        if (item.eventType === 'reasoning') {
          return { type: 'reasoning' as const, content: item.content ?? '' }
        }
        if (item.eventType === 'tool' && item.tool) {
          return {
            type: 'tool' as const,
            tool: {
              name: item.tool.name,
              displayName: item.tool.displayName,
              icon: item.tool.icon,
              action: item.tool.action,
            },
          }
        }
        return { type: 'reasoning' as const, content: '' }
      })
    : undefined
  return {
    id: m.msgId,
    role: m.roleType === 'summary' ? 'assistant' : (m.roleType as 'user' | 'assistant'),
    content: m.content,
    thinkingSegments,
    timestamp: m.created,
  }
}

export function Component() {
  const { shareId } = useParams<{ shareId: string }>()
  const t = useT()
  const [data, setData] = useState<PublicShare | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<'notFound' | 'expired' | 'loadFailed' | null>(null)

  useEffect(() => {
    if (!shareId) return
    let cancelled = false
    setLoading(true)
    setError(null)
    setData(null)
    shareApi
      .getSharedSession(shareId)
      .then((rsp) => {
        if (cancelled) return
        setData(rsp)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        if (msg.includes('expired') || msg.includes('过期')) {
          setError('expired')
        } else if (msg.includes('not found') || msg.includes('notFound') || msg.includes('不存在')) {
          setError('notFound')
        } else {
          setError('loadFailed')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [shareId])

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-bg-base">
        <Loader2 className="size-5 animate-spin text-text-muted" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-bg-base">
        <div className="max-w-sm px-4 text-center">
          <TriangleAlert className="mx-auto mb-3 size-6 text-text-3" />
          <p className="text-sm text-text-2">
            {error === 'expired'
              ? t('share.expired')
              : error === 'notFound'
                ? t('share.notFound')
                : t('share.loadFailed')}
          </p>
        </div>
      </div>
    )
  }

  if (!data) {
    return null
  }

  const messages = (data.messages ?? []).map(toReadonlyMessage)

  return (
    <div className="flex h-screen flex-col bg-bg-base">
      <div className="flex h-10 shrink-0 items-center gap-3 border-b border-border px-4">
        <span className="truncate text-sm text-text-1">{data.title}</span>
        <div className="flex-1" />
        {data.creator && (
          <span className="text-xs text-text-3">
            {t('share.sharedBy')} {data.creator}
          </span>
        )}
        <span className="flex items-center gap-1 text-xs text-text-muted tabular-nums">
          <Eye className="size-3" />
          {t('share.viewCount').replace('{count}', String(data.viewCount))}
        </span>
        {data.expiresAt && (
          <span className="flex items-center gap-1 text-xs text-text-muted tabular-nums">
            <Clock className="size-3" />
            {formatDateTime(data.expiresAt)}
          </span>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl space-y-8 px-4 py-6">
          {messages.length === 0 ? (
            <p className="text-center text-sm text-text-muted">—</p>
          ) : (
            messages.map((msg) => <ReadonlyMessageBlock key={msg.id} message={msg} />)
          )}
        </div>
      </div>
    </div>
  )
}
