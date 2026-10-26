import { useState, useMemo, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { Copy, Check, Globe, Lock } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import type { ShareType } from '@/api/types'

export type ShareExpiry = '1h' | '24h' | '7d' | '30d'

export interface ShareParams {
  title: string
  expiresIn: ShareExpiry
  shareType: ShareType
}

interface ShareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sessionTitle: string
  onGenerate: (params: ShareParams) => Promise<string | void>
}

const EXPIRY_OPTIONS: ReadonlyArray<{ value: ShareExpiry; seconds: number }> = [
  { value: '1h', seconds: 3600 },
  { value: '24h', seconds: 86400 },
  { value: '7d', seconds: 604800 },
  { value: '30d', seconds: 2592000 },
]

function formatExpiry(seconds: number): string {
  const d = new Date(Date.now() + seconds * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function ShareDialog({ open, onOpenChange, sessionTitle, onGenerate }: ShareDialogProps) {
  const t = useT()
  const expiryLabels: Record<ShareExpiry, string> = {
    '1h': t('share.expire1h'),
    '24h': t('share.expire24h'),
    '7d': t('share.expire7d'),
    '30d': t('share.expire30d'),
  }
  const [title, setTitle] = useState(sessionTitle)
  const [expiry, setExpiry] = useState<ShareExpiry>('24h')
  const [shareType, setShareType] = useState<ShareType>('public')
  const [generatedLink, setGeneratedLink] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (open) {
      setTitle(sessionTitle)
      setExpiry('24h')
      setShareType('public')
      setGeneratedLink(null)
      setCopied(false)
    }
  }, [open, sessionTitle])

  const expiryTime = useMemo(
    () => formatExpiry(EXPIRY_OPTIONS.find((o) => o.value === expiry)!.seconds),
    [expiry],
  )

  const copyLink = useCallback(async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.left = '-9999px'
      document.body.appendChild(textarea)
      textarea.focus()
      textarea.select()
      try { document.execCommand('copy') } catch {}
      document.body.removeChild(textarea)
    }
    setCopied(true)
    toast(translate('share.linkCopied'))
    setTimeout(() => setCopied(false), 2000)
  }, [])

  const handleGenerate = async () => {
    const trimmed = title.trim()
    if (!trimmed) return

    const link = await onGenerate({ title: trimmed, expiresIn: expiry, shareType })
    if (link) {
      setGeneratedLink(link)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0">
        <DialogHeader className="mb-5">
          <DialogTitle>{t('share.title')}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 rounded-lg border border-border px-5 py-5">
          <div className="space-y-1.5">
            <label className="text-sm text-text-2">{t('share.shareTitle')}</label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t('share.titlePlaceholder')}
            />
            <p className="text-xs text-text-muted">{t('share.titleHint')}</p>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm text-text-2">{t('share.type')}</label>
            <div className="grid grid-cols-2 gap-2">
              <label
                className={cn(
                  'flex cursor-pointer items-start gap-2 rounded-lg border px-3 py-2.5 transition-colors',
                  shareType === 'public' ? 'border-highlight' : 'border-border hover:border-text-muted',
                )}
              >
                <span
                  className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full border transition-colors',
                    shareType === 'public' ? 'border-highlight' : 'border-border',
                  )}
                >
                  {shareType === 'public' && <span className="size-2 rounded-full bg-highlight" />}
                </span>
                <input
                  type="radio"
                  name="shareType"
                  value="public"
                  checked={shareType === 'public'}
                  onChange={() => setShareType('public')}
                  className="sr-only"
                />
                <span className="min-w-0">
                  <span className="flex items-center gap-1.5 text-sm text-text-1">
                    <Globe className="size-3.5 text-text-3" />
                    {t('share.typePublic')}
                  </span>
                  <span className="mt-0.5 block text-xs text-text-muted leading-relaxed">
                    {t('share.typePublicHint')}
                  </span>
                </span>
              </label>
              <label
                className={cn(
                  'flex cursor-pointer items-start gap-2 rounded-lg border px-3 py-2.5 transition-colors',
                  shareType === 'internal' ? 'border-highlight' : 'border-border hover:border-text-muted',
                )}
              >
                <span
                  className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full border transition-colors',
                    shareType === 'internal' ? 'border-highlight' : 'border-border',
                  )}
                >
                  {shareType === 'internal' && <span className="size-2 rounded-full bg-highlight" />}
                </span>
                <input
                  type="radio"
                  name="shareType"
                  value="internal"
                  checked={shareType === 'internal'}
                  onChange={() => setShareType('internal')}
                  className="sr-only"
                />
                <span className="min-w-0">
                  <span className="flex items-center gap-1.5 text-sm text-text-1">
                    <Lock className="size-3.5 text-text-3" />
                    {t('share.typeInternal')}
                  </span>
                  <span className="mt-0.5 block text-xs text-text-muted leading-relaxed">
                    {t('share.typeInternalHint')}
                  </span>
                </span>
              </label>
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm text-text-2">{t('share.expiration')}</label>
            <div className="rounded-lg border border-border px-4 py-3">
              <div className="grid grid-cols-2 gap-x-8 gap-y-2">
                {EXPIRY_OPTIONS.map(({ value }) => (
                  <label
                    key={value}
                    className={cn(
                      'flex cursor-pointer items-center gap-2 text-sm transition-colors',
                      expiry === value ? 'text-text-1' : 'text-text-3 hover:text-text-2',
                    )}
                  >
                    <span
                      className={cn(
                        'flex size-4 shrink-0 items-center justify-center rounded-full border transition-colors',
                        expiry === value
                          ? 'border-highlight'
                          : 'border-border',
                      )}
                    >
                      {expiry === value && (
                        <span className="size-2 rounded-full bg-highlight" />
                      )}
                    </span>
                    <input
                      type="radio"
                      name="expiry"
                      value={value}
                      checked={expiry === value}
                      onChange={() => setExpiry(value)}
                      className="sr-only"
                    />
                    {expiryLabels[value]}
                  </label>
                ))}
              </div>
              <div className="mt-3 border-t border-border pt-2.5">
                <span className="text-xs text-text-muted">{t('share.expiresAt')} {expiryTime}</span>
              </div>
            </div>
          </div>

          {generatedLink && (
            <div className="space-y-1.5 rounded-lg border border-border p-3">
              <label className="text-sm text-text-2">{t('share.shareLink')}</label>
              <div className="flex gap-2">
                <Input value={generatedLink} readOnly className="flex-1 text-xs font-mono" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => copyLink(generatedLink)}
                  className="shrink-0"
                >
                  {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
                </Button>
              </div>
            </div>
          )}
        </div>

        <div className="mt-6 flex items-center justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.close')}
          </Button>
          {!generatedLink && (
            <Button onClick={handleGenerate} disabled={!title.trim()}>
              {t('share.generate')}
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
