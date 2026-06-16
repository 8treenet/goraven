import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useT } from '@/i18n'
import type { FileItem } from '@/api/types'

export type FileDialogMode = 'newFolder' | 'delete' | 'compress' | 'decompress' | null

interface FileDialogsProps {
  mode: FileDialogMode
  value: string
  error: string | null
  selectedItems: FileItem[]
  onValueChange: (value: string) => void
  onClearError: () => void
  onClose: () => void
  onCreateFolder: () => void
  onDelete: () => void
  onCompress: () => void
  onDecompress: (toSubDir: boolean) => void
}

export function FileDialogs({
  mode,
  value,
  error,
  selectedItems,
  onValueChange,
  onClearError,
  onClose,
  onCreateFolder,
  onDelete,
  onCompress,
  onDecompress,
}: FileDialogsProps) {
  const t = useT()

  return (
    <Dialog open={mode !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-sm">
        {mode === 'newFolder' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.newFolderTitle')}</DialogTitle>
              <DialogDescription>{t('files.folderNameLabel')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <Input
                value={value}
                onChange={(e) => {
                  onValueChange(e.target.value)
                  onClearError()
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') onCreateFolder()
                }}
                placeholder={t('files.folderNamePlaceholder')}
                autoFocus
              />
              {error && <p className="text-sm text-text-3">{error}</p>}
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onCreateFolder}>
                  {t('common.create')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'delete' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.confirmDeleteTitle')}</DialogTitle>
              <DialogDescription>
                {t('files.confirmDeleteDesc').replace('{n}', String(selectedItems.length))}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div className="max-h-32 overflow-auto rounded-md border border-border bg-bg-layer-2 p-2">
                {selectedItems.map((item) => (
                  <p key={item.name} className="text-sm text-text-2 truncate">
                    {item.isDir ? `${item.name}/` : item.name}
                  </p>
                ))}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onDelete}>
                  {t('common.delete')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'compress' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.compress')}</DialogTitle>
              <DialogDescription>{t('files.archiveNamePlaceholder')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Input
                  value={value}
                  onChange={(e) => {
                    onValueChange(e.target.value)
                    onClearError()
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') onCompress()
                  }}
                  placeholder={t('files.archiveNamePlaceholder')}
                  autoFocus
                />
                <span className="shrink-0 text-sm text-text-3">.zip</span>
              </div>
              {error && <p className="text-sm text-text-3">{error}</p>}
              <p className="text-sm text-text-3">
                {t('files.compressCount').replace('{n}', String(selectedItems.length))}
              </p>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="default" size="default" onClick={onCompress}>
                  {t('files.compress')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'decompress' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.decompress')}</DialogTitle>
              <DialogDescription>{selectedItems[0]?.name}</DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <button
                onClick={() => onDecompress(false)}
                className="w-full rounded-md px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
              >
                {t('files.extractToDir')}
              </button>
              <button
                onClick={() => onDecompress(true)}
                className="w-full rounded-md px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
              >
                {t('files.extractToSubdir').replace('{dir}', selectedItems[0]?.name.replace(/\.zip$/i, '') || '')}
              </button>
              <div className="flex justify-end">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
