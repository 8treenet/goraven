import { Loader2, X, File } from 'lucide-react'
import { Icon } from '@/components/common/Icon'

interface UploadedFile {
  id: string
  name: string
  size: number
  type: string
  status: 'uploading' | 'done' | 'error'
  uploadId?: string
  previewUrl?: string
}

const IMAGE_TYPES = ['png', 'jpg', 'jpeg', 'gif', 'bmp', 'webp', 'tiff', 'svg']
const VIDEO_TYPES = ['mp4', 'mov', 'avi']
const AUDIO_TYPES = ['wav', 'mp3']
const MAX_FILE_SIZE = 20 * 1024 * 1024

function getFileCategory(name: string) {
  const ext = name.split('.').pop()?.toLowerCase() ?? ''
  if (IMAGE_TYPES.includes(ext)) return 'image'
  if (VIDEO_TYPES.includes(ext)) return 'video'
  if (AUDIO_TYPES.includes(ext)) return 'audio'
  return 'doc'
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
}

function FilePreviews({
  files,
  onRemove,
}: {
  files: UploadedFile[]
  onRemove: (id: string) => void
}) {
  if (files.length === 0) return null

  return (
    <div className="flex flex-wrap gap-2 px-3 pt-3">
      {files.map((f) => (
        <div
          key={f.id}
          className="group relative flex items-center gap-2 rounded-md bg-bg-layer-3 px-2 py-1.5 pr-7 text-xs text-text-2"
        >
          {f.status === 'uploading' ? (
            <Loader2 className="size-3.5 shrink-0 animate-spin text-text-muted" />
          ) : f.status === 'error' ? (
            <X className="size-3.5 shrink-0 text-text-3" />
          ) : getFileCategory(f.name) === 'image' && f.previewUrl ? (
            <img src={f.previewUrl} alt={f.name} className="size-5 shrink-0 rounded object-cover" />
          ) : getFileCategory(f.name) === 'video' ? (
            <File className="size-3.5 shrink-0 text-text-muted" />
          ) : (
            <Icon name="file-text" className="size-3.5 shrink-0 text-text-muted" />
          )}
          <span className="max-w-32 truncate">{f.name}</span>
          <span className="shrink-0 text-text-muted">{formatSize(f.size)}</span>
          <button
            onClick={() => onRemove(f.id)}
            className="absolute right-1 top-1/2 hidden -translate-y-1/2 rounded p-0.5 text-text-muted hover:bg-bg-hover hover:text-text-2 group-hover:block"
          >
            <X className="size-3" />
          </button>
        </div>
      ))}
    </div>
  )
}

export { FilePreviews, getFileCategory, formatSize, MAX_FILE_SIZE }
export type { UploadedFile }
