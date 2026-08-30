import { useState } from 'react'
import { t as translate } from '@/i18n'
import { useT } from '@/i18n'
import { Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Icon } from '@/components/common/Icon'
import type { UserSkill, MarketSkill } from '@/api/types'

export function DeleteSkillDialog({
  target,
  onClose,
  onConfirm,
}: {
  target: UserSkill | null
  onClose: () => void
  onConfirm: () => void
}) {
  const t = useT()

  return (
    <Dialog open={target !== null} onOpenChange={onClose}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('skills.confirmDeleteSkill')}</DialogTitle>
          <DialogDescription>
            {t('skills.confirmDeleteSkillDesc').replace('{name}', target?.skillName ?? '')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <p className="text-xs text-text-3">{t('skills.deleteSkillDesc')}</p>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" size="default" onClick={onConfirm}>
              {t('common.confirmDelete')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function InstallSuccessDialog({
  target,
  copied,
  onClose,
  onCopy,
}: {
  target: MarketSkill | null
  copied: boolean
  onClose: () => void
  onCopy: () => void
}) {
  const t = useT()

  return (
    <Dialog open={target !== null} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('skills.installSuccessTitle')}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-md border border-border bg-bg-layer-2 p-4">
            <div className="flex items-start gap-4">
              <Icon
                name={target?.icon || 'bot'}
                className="size-8 shrink-0 mt-0.5 text-text-2"
              />
              <div className="min-w-0">
                <h3 className="text-base font-semibold text-text-1">{target?.name}</h3>
                <p className="mt-1 text-sm text-text-3 leading-relaxed">
                  {target?.description}
                </p>
              </div>
            </div>
          </div>

          <p className="text-sm text-text-3 leading-relaxed">
            {t('skills.installSuccessDesc')}
          </p>

          <div className="rounded-md border border-border bg-bg-layer-2 p-4">
            <p className="text-base text-text-2 leading-relaxed">
              {target
                ? translate('skills.installSuccessContent')
                    .replace('{name}', target.name)
                    .replace('{dir}', target.name.toLowerCase().replace(/\s+/g, '-'))
                : ''}
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" size="default" onClick={onClose}>
              {t('common.close')}
            </Button>
            <Button variant="default" size="default" onClick={onCopy}>
              {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
              {copied ? t('common.copied') : t('common.copy')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function ShareSkillDialog({
  target,
  onClose,
  onConfirm,
}: {
  target: UserSkill | null
  onClose: () => void
  onConfirm: (note: string) => void
}) {
  const t = useT()
  const [note, setNote] = useState('')

  return (
    <Dialog open={target !== null} onOpenChange={() => {
      setNote('')
      onClose()
    }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('skills.shareSkill')}</DialogTitle>
          <DialogDescription>
            {t('skills.shareSkillDesc')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="rounded-md border border-border bg-bg-layer-2 p-3">
            <div className="flex items-center gap-2.5">
              <Icon name={target?.icon || 'zap'} className="size-5 shrink-0 text-text-2" />
              <span className="text-sm font-semibold text-text-1 truncate">
                {target?.skillName}
              </span>
            </div>
          </div>
          <div>
            <label className="text-xs text-text-muted">{t('skills.shareNoteLabel')}</label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              spellCheck={false}
              placeholder={t('skills.shareNotePlaceholder')}
              className="mt-1 w-full h-16 rounded-md border border-border bg-transparent px-3 py-2 text-xs text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30 resize-none"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={() => {
              setNote('')
              onClose()
            }}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" size="default" onClick={() => onConfirm(note)}>
              {t('skills.shareToTeam')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function CancelShareDialog({
  target,
  onClose,
  onConfirm,
}: {
  target: { shareId: number; skillName: string } | null
  onClose: () => void
  onConfirm: () => void
}) {
  const t = useT()

  return (
    <Dialog open={target !== null} onOpenChange={onClose}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('skills.cancelShareConfirm')}</DialogTitle>
          <DialogDescription>
            {t('skills.cancelShareDesc').replace('{name}', target?.skillName ?? '')}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button variant="default" size="default" onClick={onConfirm}>
            {t('common.confirmDelete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
