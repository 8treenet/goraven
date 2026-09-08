import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'
import { validateDirName } from './file-helpers'

export type ShareDialogMode = 'create' | 'editShare' | 'delete' | null

interface ShareDialogProps {
  mode: ShareDialogMode
  projectName: string
  description: string
  onDescriptionChange: (value: string) => void
  onProjectNameChange?: (value: string) => void
  /** 编辑模式下允许改名（个人项目）；默认只读展示项目名（团队项目） */
  allowRenameInEdit?: boolean
  /** 自定义标题（默认 files.createProjectTitle / files.editShareDialogTitle） */
  createTitle?: string
  editTitle?: string
  /** 新建对话框描述（默认 files.createProjectDesc） */
  createDesc?: string
  /** 删除确认描述（默认 files.confirmDeleteProjectDesc） */
  deleteDesc?: string
  onClose: () => void
  onConfirm: () => void
}

export function ShareDialog({
  mode,
  projectName,
  description,
  onDescriptionChange,
  onProjectNameChange,
  allowRenameInEdit = false,
  createTitle,
  editTitle,
  createDesc,
  deleteDesc,
  onClose,
  onConfirm,
}: ShareDialogProps) {
  const t = useT()
  const nameEditable = mode === 'create' || (mode === 'editShare' && allowRenameInEdit)
  const nameError = nameEditable && projectName.length > 0 ? validateDirName(projectName) : null
  const nameInvalid = nameError !== null
  const canSubmit = nameEditable ? projectName.trim().length > 0 && !nameInvalid : true

  return (
    <Dialog open={mode !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-sm">
        {(mode === 'create' || mode === 'editShare') && (
          <>
            <DialogHeader>
              <DialogTitle>
                {mode === 'create' ? (createTitle || t('files.createProjectTitle')) : (editTitle || t('files.editShareDialogTitle'))}
              </DialogTitle>
              <DialogDescription>
                {mode === 'create' ? (createDesc || t('files.createProjectDesc')) : t('files.editDescription')}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-text-muted">{t('files.name')}</label>
                {nameEditable ? (
                  <>
                    <input
                      value={projectName}
                      onChange={(e) => onProjectNameChange?.(e.target.value)}
                      spellCheck={false}
                      placeholder={t('files.projectNamePlaceholder')}
                      className={`mt-1 w-full rounded-md border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:ring-2 focus:ring-ring/30 ${nameInvalid ? 'border-destructive focus:border-destructive' : 'border-border focus:border-ring'}`}
                      autoFocus
                    />
                    {nameInvalid && (
                      <p className="mt-1 text-xs text-destructive">{nameError}</p>
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
                  spellCheck={false}
                  placeholder={t('files.descriptionPlaceholder')}
                  className="mt-1 h-20 w-full resize-none rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                  autoFocus={mode === 'editShare' && !allowRenameInEdit}
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
              <DialogDescription>{deleteDesc || t('files.confirmDeleteProjectDesc')}</DialogDescription>
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
