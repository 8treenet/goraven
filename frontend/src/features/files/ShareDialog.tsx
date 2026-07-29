import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'

export type ShareDialogMode = 'create' | 'editShare' | 'delete' | null

interface ShareDialogProps {
  mode: ShareDialogMode
  projectName: string
  description: string
  onDescriptionChange: (value: string) => void
  onProjectNameChange?: (value: string) => void
  onClose: () => void
  onConfirm: () => void
}

export function ShareDialog({
  mode,
  projectName,
  description,
  onDescriptionChange,
  onProjectNameChange,
  onClose,
  onConfirm,
}: ShareDialogProps) {
  const t = useT()
  const nameInvalid = mode === 'create' && projectName.length > 0 && !/^[a-zA-Z0-9\-_]+$/.test(projectName)
  const canSubmit = mode === 'create' ? projectName.trim().length > 0 && !nameInvalid : true

  return (
    <Dialog open={mode !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-sm">
        {(mode === 'create' || mode === 'editShare') && (
          <>
            <DialogHeader>
              <DialogTitle>
                {mode === 'create' ? t('files.createProjectTitle') : t('files.editShareDialogTitle')}
              </DialogTitle>
              <DialogDescription>
                {mode === 'create' ? t('files.createProjectDesc') : t('files.editDescription')}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-text-muted">{t('files.name')}</label>
                {mode === 'create' ? (
                  <>
                    <input
                      value={projectName}
                      onChange={(e) => onProjectNameChange?.(e.target.value)}
                      placeholder={t('files.projectNamePlaceholder')}
                      className={`mt-1 w-full rounded-md border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:ring-2 focus:ring-ring/30 ${nameInvalid ? 'border-destructive focus:border-destructive' : 'border-border focus:border-ring'}`}
                      autoFocus
                    />
                    {nameInvalid && (
                      <p className="mt-1 text-xs text-destructive">{t('files.projectNameInvalid')}</p>
                    )}
                  </>
                ) : (
                  <div className="mt-1 rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                    {projectName}
                  </div>
                )}
              </div>
              <div>
                <label className="text-xs text-text-muted">{t('files.descriptionLabel')}</label>
                <textarea
                  value={description}
                  onChange={(e) => onDescriptionChange(e.target.value)}
                  placeholder={t('files.descriptionPlaceholder')}
                  className="mt-1 h-20 w-full resize-none rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                  autoFocus={mode === 'editShare'}
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onConfirm} disabled={!canSubmit}>
                  {mode === 'create' ? t('common.create') : t('common.save')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'delete' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.confirmDeleteProject')}</DialogTitle>
              <DialogDescription>{t('files.confirmDeleteProjectDesc')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div className="rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                {projectName}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="destructive" size="default" onClick={onConfirm}>
                  {t('common.delete')}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
