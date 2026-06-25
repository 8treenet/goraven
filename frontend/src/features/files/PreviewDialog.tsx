import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, Download, Maximize2, Minimize2, RefreshCw } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useT } from '@/i18n'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Markdown } from '@/components/common/markdown'
import { formatSize, formatTime, type PreviewType, type SheetData } from './file-helpers'
import type { FileItem } from '@/api/types'

interface PreviewDialogProps {
  item: FileItem | null
  type: PreviewType | null
  url: string | null
  text: string | null
  sheets: SheetData[] | null
  loading: boolean
  error: boolean
  onClose: () => void
  onDownload?: () => void
}

export function PreviewDialog({ item, type, url, text, sheets, loading, error, onClose, onDownload }: PreviewDialogProps) {
  const t = useT()
  const [maximized, setMaximized] = useState(false)
  const [activeSheet, setActiveSheet] = useState(0)

  useEffect(() => {
    setActiveSheet(0)
  }, [sheets])

  const toggleMaximized = useCallback(() => {
    setMaximized((v) => !v)
  }, [])

  const handleEscape = useCallback(
    (e: KeyboardEvent) => {
      if (maximized) {
        e.preventDefault()
        setMaximized(false)
      }
    },
    [maximized],
  )

  return (
    <Dialog open={item !== null} onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent
        className={cn(
          'max-w-4xl',
          maximized &&
            'flex flex-col max-w-none w-screen h-screen rounded-none left-0 top-0 translate-x-0 translate-y-0',
        )}
        onEscapeKeyDown={handleEscape}
      >
        <DialogHeader>
          <DialogTitle className="truncate">{item?.name}</DialogTitle>
          <DialogDescription>
            {item ? `${formatSize(item.size)} · ${formatTime(item.modTime)}` : ''}
          </DialogDescription>
        </DialogHeader>
        <button
          type="button"
          onClick={toggleMaximized}
          className="absolute right-12 top-4 text-text-3 transition-colors hover:text-text-1"
          title={maximized ? t('common.exitFullscreen') : t('common.fullscreen')}
        >
          {maximized ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
          <span className="sr-only">{maximized ? t('common.exitFullscreen') : t('common.fullscreen')}</span>
        </button>
        <div
          className={cn(
            'flex min-h-[200px] items-center justify-center overflow-auto rounded-md border border-border bg-bg-layer-2 p-2',
            maximized ? 'min-h-0 flex-1' : 'max-h-[70vh]',
          )}
        >
          {loading && (
            <div className="flex flex-col items-center gap-2 py-8">
              <RefreshCw className="size-5 animate-spin text-text-3" />
              <p className="text-sm text-text-3">{t('common.loading')}</p>
            </div>
          )}
          {!loading && error && (
            <div className="flex flex-col items-center gap-3 py-8">
              <AlertCircle className="size-6 text-text-3" />
              <p className="text-sm text-text-2">{type === 'office' ? t('files.previewFailedOffice') : t('common.loadFailed')}</p>
              {onDownload && (
                <Button variant="outline" size="sm" onClick={onDownload}>
                  <Download className="size-3.5" />
                  {t('files.download')}
                </Button>
              )}
            </div>
          )}
          {!loading && !error && url && type === 'image' && (
            <img
              src={url}
              alt={item?.name}
              className={cn('max-w-full object-contain', maximized ? 'max-h-full' : 'max-h-[65vh]')}
            />
          )}
          {!loading && !error && url && type === 'video' && (
            <video
              src={url}
              controls
              className={cn('max-w-full', maximized ? 'max-h-full' : 'max-h-[65vh]')}
            />
          )}
          {!loading && !error && url && type === 'audio' && (
            <div className="w-full py-8">
              <audio src={url} controls className="w-full" />
            </div>
          )}
          {!loading && !error && text !== null && type === 'text' && (
            <pre
              className={cn(
                'w-full overflow-auto text-xs font-mono text-text-1 whitespace-pre-wrap break-all leading-relaxed',
                maximized ? 'flex-1' : 'max-h-[65vh]',
              )}
            >
              {text}
            </pre>
          )}
          {!loading && !error && text !== null && type === 'markdown' && (
            <div
              className={cn(
                'w-full overflow-auto text-sm text-text-1 leading-relaxed',
                maximized ? 'flex-1' : 'max-h-[65vh]',
              )}
            >
              <Markdown mode="static">{text}</Markdown>
            </div>
          )}
          {!loading && !error && url && type === 'pdf' && (
            <iframe
              src={url}
              title={item?.name}
              className={cn('w-full rounded-sm bg-white', maximized ? 'h-full' : 'h-[65vh]')}
            />
          )}
          {!loading && !error && url && type === 'html' && (
            <iframe
              src={url}
              title={item?.name}
              sandbox="allow-scripts"
              className={cn('w-full rounded-sm bg-white', maximized ? 'h-full' : 'h-[65vh]')}
            />
          )}
          {!loading && !error && url && type === 'office' && (
            <iframe
              src={url}
              title={item?.name}
              className={cn('w-full rounded-sm bg-white', maximized ? 'h-full' : 'h-[65vh]')}
            />
          )}
          {!loading && !error && sheets !== null && type === 'xlsx' && (
            <div className={cn('flex w-full flex-col', maximized ? 'flex-1 min-h-0' : 'max-h-[65vh]')}>
              {sheets.length > 1 && (
                <div className="mb-2 flex flex-wrap gap-3 border-b border-border pb-2">
                  {sheets.map((s, i) => (
                    <button
                      key={s.name}
                      type="button"
                      onClick={() => setActiveSheet(i)}
                      className={cn(
                        'text-xs transition-colors',
                        i === activeSheet
                          ? '-mb-[9px] border-b-2 border-interactive pb-1 text-interactive'
                          : 'text-text-3 hover:text-text-1',
                      )}
                    >
                      {s.name}
                    </button>
                  ))}
                </div>
              )}
              <div
                className="raven-xlsx flex-1 overflow-auto"
                dangerouslySetInnerHTML={{ __html: sheets[activeSheet]?.html ?? '' }}
              />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
