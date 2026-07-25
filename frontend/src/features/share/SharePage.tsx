import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Loader2, Eye, TriangleAlert, Clock, Lock } from 'lucide-react'
import { shareApi } from '@/api'
import type { PublicShare, Message as ApiMessage } from '@/api/types'
import { useUserStore } from '@/stores/user-store'
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

type ErrorState = 'notFound' | 'expired' | 'loadFailed' | null

export function Component() {
  const { shareId } = useParams<{ shareId: string }>()
  const t = useT()
  const [info, setInfo] = useState<PublicShare | null>(null)
  const [messages, setMessages] = useState<ApiMessage[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ErrorState>(null)
  const [needsLogin, setNeedsLogin] = useState(false)

  useEffect(() => {
    if (!shareId) return
    let cancelled = false
    setLoading(true)
    setError(null)
    setInfo(null)
    setMessages(null)
    setNeedsLogin(false)

    shareApi
      .getShareInfo(shareId)
      .then((rsp) => {
        if (cancelled) return
        setInfo(rsp)
        if (rsp.shareType === 'internal' && !useUserStore.getState().isAuthenticated) {
          setNeedsLogin(true)
          return
        }
        const msgsPromise =
          rsp.shareType === 'internal'
            ? shareApi.getInternalShareMessages(shareId)
            : shareApi.getShareMessages(shareId)
        return msgsPromise
      })
      .then((msgs) => {
        if (cancelled || !msgs) return
        setMessages(msgs)
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
    const errorStyle =
      error === 'expired'
        ? { icon: 'text-warning', text: 'text-warning' }
        : error === 'notFound'
          ? { icon: 'text-info', text: 'text-info' }
          : { icon: 'text-destructive', text: 'text-destructive' }
    return (
      <div className="flex h-screen items-center justify-center bg-bg-base">
        <div className="max-w-sm px-4 text-center">
          <TriangleAlert className={`mx-auto mb-3 size-6 ${errorStyle.icon}`} />
          <p className={`text-sm ${errorStyle.text}`}>
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

  if (!info) {
    return null
  }

  const readonlyMessages = (messages ?? []).map(toReadonlyMessage)

  return (
    <div className="flex h-screen flex-col bg-bg-base">
      <div className="flex h-10 shrink-0 items-center gap-3 border-b border-border px-4">
        <span className="truncate text-sm text-text-1">{info.title}</span>
        {info.shareType === 'internal' && (
          <span className="inline-flex shrink-0 items-center gap-1 rounded border border-interactive/20 px-1.5 py-0.5 text-[10px] text-interactive">
            <Lock className="size-2.5" />
            {t('share.typeInternal')}
          </span>
        )}
        <div className="flex-1" />
        {info.creator && (
          <span className="hidden text-xs text-text-1 sm:inline">
            {t('share.sharedBy')} {info.creator}
          </span>
        )}
        <span className="hidden items-center gap-1 text-xs text-text-muted tabular-nums sm:flex">
          <Eye className="size-3" />
          {t('share.viewCount').replace('{count}', String(info.viewCount))}
        </span>
        {info.expiresAt && (
          <span className="hidden items-center gap-1 text-xs text-text-muted tabular-nums sm:flex">
            <Clock className="size-3" />
            {formatDateTime(info.expiresAt)}
          </span>
        )}
      </div>

      {needsLogin ? (
        <div className="flex flex-1 items-center justify-center px-4">
          <div className="max-w-sm text-center">
            <Lock className="mx-auto mb-3 size-6 text-interactive" />
            <p className="mb-1.5 text-sm font-semibold text-text-1">
              {t('share.loginRequiredTitle')}
            </p>
            <p className="text-xs text-text-2 leading-relaxed">
              {t('share.loginRequiredDesc')}
            </p>
          </div>
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto">
          <div className="mx-auto max-w-3xl space-y-8 px-4 pt-6 pb-16">
            {readonlyMessages.length === 0 ? (
              <p className="text-center text-sm text-text-muted">—</p>
            ) : (
              readonlyMessages.map((msg) => <ReadonlyMessageBlock key={msg.id} message={msg} />)
            )}

            <div className="flex flex-col items-center gap-3 pt-10">
              <p className="text-center text-xs text-text-muted">{t('share.tagline')}</p>
              <div className="flex items-center gap-6">
                  <a
                    href="https://goraven.dev"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 text-xs text-interactive transition-colors hover:text-interactive-hover"
                  >
                  <svg className="size-3.5" viewBox="0 0 56 56" fill="none" aria-hidden="true">
                    <polygon points="30,12 42,9 30,17" fill="currentColor" />
                    <rect x="20" y="6" width="12" height="12" fill="currentColor" />
                    <polygon points="20,6 24,6 24,2 16,4" fill="currentColor" />
                    <rect x="18" y="16" width="10" height="6" fill="currentColor" />
                    <polygon points="18,20 32,20 30,32 16,32" fill="currentColor" />
                    <polygon points="20,20 32,20 34,28 20,28" fill="currentColor" />
                    <polygon points="20,28 34,28 36,36 20,36" fill="currentColor" />
                    <polygon points="16,32 10,34 8,42 16,40" fill="currentColor" />
                    <polygon points="16,40 8,42 10,46 18,44" fill="currentColor" />
                    <rect x="18" y="36" width="2" height="10" fill="currentColor" />
                    <rect x="22" y="36" width="2" height="10" fill="currentColor" />
                    <polygon points="18,46 20,46 21,48 17,48" fill="currentColor" />
                    <polygon points="22,46 24,46 25,48 21,48" fill="currentColor" />
                    <rect x="24" y="8" width="3" height="3" fill="#0a0a0b" />
                  </svg>
                  goraven.dev
                </a>
                  <a
                    href="https://github.com/8treenet/goraven"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 text-xs text-interactive transition-colors hover:text-interactive-hover"
                  >
                  <svg className="size-3.5" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
                  </svg>
                  {t('share.githubLabel')}
                </a>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
