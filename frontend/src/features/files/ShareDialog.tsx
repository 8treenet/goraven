import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'

export type ShareDialogMode = 'share' | 'editShare' | 'unshare' | null

interface ShareDialogProps {
  mode: ShareDialogMode
  projectName: string
  description: string
  onDescriptionChange: (value: string) => void
  onClose: () => void
  onConfirm: () => void
}

export function ShareDialog({
  mode,
  projectName,
  description,
  onDescriptionChange,
  onClose,
  onConfirm,
}: ShareDialogProps) {
  const t = useT()

  return (
    <Dialog open={mode !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-sm">
        {(mode === 'share' || mode === 'editShare') && (
          <>
            <DialogHeader>
              <DialogTitle>
                {mode === 'share' ? t('files.shareDialogTitle') : t('files.editShareDialogTitle')}
              </DialogTitle>
              <DialogDescription>
                {mode === 'share' ? t('files.shareDialogDesc') : t('files.editDescription')}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-text-muted">{t('files.name')}</label>
                <div className="mt-1 rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                  {projectName}
                </div>
              </div>
              <div>
                <label className="text-xs text-text-muted">{t('files.descriptionLabel')}</label>
                <textarea
                  value={description}
                  onChange={(e) => onDescriptionChange(e.target.value)}
                  placeholder={t('files.descriptionPlaceholder')}
                  className="mt-1 h-20 w-full resize-none rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                  autoFocus
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onConfirm}>
                  {mode === 'share' ? t('files.shareToTeam') : t('common.save')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'unshare' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.confirmUnshare')}</DialogTitle>
              <DialogDescription>{t('files.confirmUnshareDesc')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div className="rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                {projectName}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onConfirm}>
                  {t('files.unshare')}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
