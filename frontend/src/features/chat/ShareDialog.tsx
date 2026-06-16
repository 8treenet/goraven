import { useState, useMemo, useEffect } from 'react'
import { toast } from 'sonner'
import { TriangleAlert } from 'lucide-react'
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

export type ShareExpiry = '1h' | '24h' | '7d' | '30d'

export interface ShareParams {
  title: string
  expiresIn: ShareExpiry
}

interface ShareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sessionTitle: string
  domainConfigured?: boolean
  onGenerate?: (params: ShareParams) => Promise<string> | string | void
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

export function ShareDialog({ open, onOpenChange, sessionTitle, domainConfigured = true, onGenerate }: ShareDialogProps) {
  const t = useT()
  const expiryLabels: Record<ShareExpiry, string> = {
    '1h': t('share.expire1h'),
    '24h': t('share.expire24h'),
    '7d': t('share.expire7d'),
    '30d': t('share.expire30d'),
  }
  const [title, setTitle] = useState(sessionTitle)
  const [expiry, setExpiry] = useState<ShareExpiry>('24h')

  useEffect(() => {
    if (open) {
      setTitle(sessionTitle)
      setExpiry('24h')
    }
  }, [open, sessionTitle])

  const expiryTime = useMemo(
    () => formatExpiry(EXPIRY_OPTIONS.find((o) => o.value === expiry)!.seconds),
    [expiry],
  )

  const handleGenerate = async () => {
    const trimmed = title.trim()
    if (!trimmed) return

    if (onGenerate) {
      const link = await onGenerate({ title: trimmed, expiresIn: expiry })
      if (link) {
        navigator.clipboard.writeText(link).catch(() => {})
        toast(translate('share.linkCopied'))
        onOpenChange(false)
      }
    }
  }

  if (!domainConfigured) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-md gap-0">
          <DialogHeader className="mb-5">
            <DialogTitle>{t('share.title')}</DialogTitle>
          </DialogHeader>

          <div className="rounded-lg border border-border px-4 py-5">
            <div className="flex items-center gap-2.5 mb-3">
              <TriangleAlert className="size-5 shrink-0 text-highlight" />
              <span className="text-base font-semibold text-text-1">{t('share.noDomain')}</span>
            </div>

            <p className="text-sm text-text-2 leading-relaxed mb-3">
              {t('share.noDomainDesc')}
            </p>

            <ol className="space-y-1.5 text-sm text-text-2">
              <li className="flex gap-2">
                <span className="shrink-0 text-text-muted">1.</span>
                <span>{t('chat.noModelStep1')}</span>
              </li>
              <li className="flex gap-2">
                <span className="shrink-0 text-text-muted">2.</span>
                <span>{t('chat.noModelStep2')}</span>
              </li>
              <li className="flex gap-2">
                <span className="shrink-0 text-text-muted">3.</span>
                <span>{t('chat.noModelStep3')}</span>
              </li>
              <li className="flex gap-2">
                <span className="shrink-0 text-text-muted">4.</span>
                <span>{t('share.noDomainStep4')}</span>
              </li>
            </ol>
          </div>

          <div className="mt-6 flex items-center justify-end">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('chat.gotIt')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md gap-0">
        <DialogHeader className="mb-5">
          <DialogTitle>{t('share.title')}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
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
        </div>

        <div className="mt-6 flex items-center justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={handleGenerate} disabled={!title.trim()}>
            {t('share.generateCopy')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
