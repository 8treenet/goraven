import { useState, useRef, useCallback, useEffect } from 'react'
import { FileText, Play, File, Download, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { t, useT } from '@/i18n'
import { getFileUrl } from '@/api/files'
import { useUserStore } from '@/stores/user-store'

interface RavenFileProps {
  kind?: string
  path?: string
  name?: string
  description?: string
}

const KIND_CONFIG = {
  doc: { icon: FileText, label: t('common.file') },
} as const

function getFileName(path: string): string {
  return path.split('/').pop() || path
}

function getFileExtension(path: string): string {
  const name = getFileName(path)
  const dot = name.lastIndexOf('.')
  return dot > 0 ? name.slice(dot + 1).toUpperCase() : ''
}

/**
 * Resolve a path from <raven-file> into a usable URL.
 * - Already a full URL (http/https) → use as-is
 * - Static asset (/assets/...) → use as-is (no auth required)
 * - Filesystem absolute path (e.g. /raven/data/users/<user>/documents/foo.pdf)
 *   → use /api/hfs/file/<abs-path>; cross-user sharing supported.
 */
function resolveUrl(path: string): string {
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  if (path.startsWith('/assets/')) return path
  return getFileUrl(path)
}

/**
 * Fetch a private file with auth and trigger a browser download.
 */
async function downloadFile(url: string, filename: string) {
  const token = useUserStore.getState().token
  try {
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!res.ok) throw new Error('download failed')
    const blob = await res.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch {
    // Fallback: open in new tab
    window.open(url, '_blank')
  }
}

export function RavenFile({ kind = 'doc', path = '', name, description }: RavenFileProps) {
  const t = useT()

  if (kind === 'image') {
    return <ImagePreview path={path} name={name} description={description} />
  }

  if (kind === 'video') {
    return <VideoPlayer path={path} name={name} description={description} />
  }

  // kind === 'doc' (default)
  const config = KIND_CONFIG[kind as keyof typeof KIND_CONFIG] || { icon: File, label: t('common.file') }
  const Icon = config.icon
  const displayName = name || getFileName(path)
  const ext = getFileExtension(path)
  const url = resolveUrl(path)
  const isPrivate = !path.startsWith('http://') && !path.startsWith('https://') && !path.startsWith('/assets/')

  // 兜底：path 无文件后缀视为目录（LLM 幻觉），不触发下载
  const isDownloadable = isPrivate && !!ext

  const handleDownload = useCallback(() => {
    if (!ext) return
    downloadFile(url, displayName)
  }, [url, displayName, ext])

  return (
    <div
      onClick={isDownloadable ? handleDownload : undefined}
      className={cn(
        'group my-2.5 flex items-start gap-3 rounded-md bg-bg-layer-2 px-3 py-2',
        isDownloadable && 'cursor-pointer hover:bg-bg-hover',
      )}
    >
      <Icon className="mt-px size-4 shrink-0 text-text-muted" />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium text-text-1">{displayName}</span>
          {ext && (
            <span className="shrink-0 rounded-sm bg-bg-layer-3 px-1.5 py-px text-xs text-text-muted">
              {ext}
            </span>
          )}
          {isDownloadable && (
            <Download className="size-3 shrink-0 text-text-muted opacity-0 transition-opacity group-hover:opacity-100" />
          )}
        </div>
        {description ? (
          <p className="mt-1 text-xs leading-relaxed text-text-3">{description}</p>
        ) : path && !isPrivate ? (
          <p className="mt-1 text-xs text-text-muted truncate">{path}</p>
        ) : null}
      </div>
    </div>
  )
}

/* ============================================
   Image Preview (handles auth for sandbox files)
   ============================================ */

function ImagePreview({ path, name, description }: { path: string; name?: string; description?: string }) {
  const url = resolveUrl(path)
  const isPrivate = !path.startsWith('http://') && !path.startsWith('https://') && !path.startsWith('/assets/')
  const [blobUrl, setBlobUrl] = useState<string | null>(null)
  const [loading, setLoading] = useState(isPrivate)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!isPrivate) return
    let cancelled = false
    const token = useUserStore.getState().token
    fetch(url, { headers: { Authorization: `Bearer ${token}` } })
      .then((res) => {
        if (!res.ok) throw new Error('failed')
        return res.blob()
      })
      .then((blob) => {
        if (cancelled) return
        const objectUrl = URL.createObjectURL(blob)
        setBlobUrl(objectUrl)
        setLoading(false)
      })
      .catch(() => {
        if (cancelled) return
        setError(true)
        setLoading(false)
      })
    return () => {
      cancelled = true
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path])

  // Cleanup blob URL on unmount
  useEffect(() => {
    return () => {
      if (blobUrl) URL.revokeObjectURL(blobUrl)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blobUrl])

  if (loading) {
    return (
      <figure className="my-3">
        <div className="flex items-center justify-center rounded-md bg-bg-layer-2 py-10">
          <Loader2 className="size-5 animate-spin text-text-muted" />
        </div>
      </figure>
    )
  }

  if (error) {
    return (
      <figure className="my-3">
        <div className="flex items-center justify-center rounded-md bg-bg-layer-2 py-10">
          <File className="size-5 text-text-muted" />
          <span className="ml-2 text-sm text-text-3">{name || getFileName(path)}</span>
        </div>
      </figure>
    )
  }

  return (
    <figure className="my-3">
      <img
        src={blobUrl || url}
        alt={name || getFileName(path)}
        className="mx-auto block max-h-80 max-w-full rounded-md object-contain"
      />
      {(name || description) && (
        <figcaption className="mt-1.5">
          {name && <p className="text-xs text-text-2">{name}</p>}
          {description && <p className="mt-0.5 text-xs text-text-muted">{description}</p>}
        </figcaption>
      )}
    </figure>
  )
}

/* ============================================
   Video Player (handles auth for sandbox files)
   ============================================ */

function VideoPlayer({
  path,
  name,
  description,
}: {
  path: string
  name?: string
  description?: string
}) {
  const url = resolveUrl(path)
  const isPrivate = !path.startsWith('http://') && !path.startsWith('https://') && !path.startsWith('/assets/')
  const [playing, setPlaying] = useState(false)
  const [blobUrl, setBlobUrl] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const videoRef = useRef<HTMLVideoElement>(null)

  const handlePlay = useCallback(() => {
    // For private files, we need to fetch with auth first
    if (isPrivate && !blobUrl) {
      setLoading(true)
      const token = useUserStore.getState().token
      fetch(url, { headers: { Authorization: `Bearer ${token}` } })
        .then((res) => {
          if (!res.ok) throw new Error('failed')
          return res.blob()
        })
        .then((blob) => {
          const objectUrl = URL.createObjectURL(blob)
          setBlobUrl(objectUrl)
          setLoading(false)
          // Auto-play after loading
          setTimeout(() => {
            videoRef.current?.play()
            setPlaying(true)
          }, 0)
        })
        .catch(() => {
          setLoading(false)
        })
      return
    }

    const el = videoRef.current
    if (!el) return
    if (el.paused) {
      el.play()
      setPlaying(true)
    } else {
      el.pause()
      setPlaying(false)
    }
  }, [isPrivate, blobUrl, url])

  const handleEnded = useCallback(() => setPlaying(false), [])

  if (loading) {
    return (
      <figure className="my-3">
        <div className="flex items-center justify-center rounded-md bg-bg-layer-2 py-10">
          <Loader2 className="size-5 animate-spin text-text-muted" />
        </div>
        {(name || description) && (
          <figcaption className="mt-1.5">
            {name && <p className="text-xs text-text-2">{name}</p>}
            {description && <p className="mt-0.5 text-xs text-text-muted">{description}</p>}
          </figcaption>
        )}
      </figure>
    )
  }

  return (
    <figure className="my-3">
      <div className="relative overflow-hidden rounded-md border border-border bg-bg-layer-1">
        <video
          ref={videoRef}
          src={blobUrl || (!isPrivate ? url : undefined)}
          controls={playing}
          onEnded={handleEnded}
          className="max-h-80 max-w-full"
        />
        {!playing && (
          <button
            onClick={handlePlay}
            className="absolute inset-0 flex items-center justify-center bg-black/30 transition-colors hover:bg-black/20"
          >
            <span className="flex size-14 items-center justify-center rounded-full bg-black/50 backdrop-blur-sm transition-transform hover:scale-105">
              <Play className="ml-0.5 size-6 text-white" fill="currentColor" />
            </span>
          </button>
        )}
      </div>
      {(name || description) && (
        <figcaption className="mt-1.5">
          {name && <p className="text-xs text-text-2">{name}</p>}
          {description && <p className="mt-0.5 text-xs text-text-muted">{description}</p>}
        </figcaption>
      )}
    </figure>
  )
}
