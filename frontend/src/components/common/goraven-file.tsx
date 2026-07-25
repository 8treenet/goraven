import { useState, useRef, useCallback, useEffect } from 'react'
import { FileText, Play, File, Download, Loader2, Maximize2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { t, useT } from '@/i18n'
import { getFileUrl } from '@/api/files'
import { getCachedBlobUrl, peekCachedBlobUrl } from '@/lib/file-blob-cache'
import { PreviewDialog } from '@/features/files/PreviewDialog'
import type { FileItem } from '@/api/types'

interface GoRavenFileProps {
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
 * Resolve a path from <goraven-file> into a usable URL.
 * - Already a full URL (http/https) → use as-is
 * - Static asset (/assets/...) → use as-is (no auth required)
 * - Filesystem absolute path (e.g. /goraven/data/users/<user>/documents/foo.pdf)
 *   → use /api/hfs/file/<abs-path>; cross-user sharing supported.
 */
function resolveUrl(path: string): string {
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  if (path.startsWith('/assets/')) return path
  return getFileUrl(path)
}

async function downloadFile(url: string, filename: string) {
  try {
    const { blobUrl } = await getCachedBlobUrl(url)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  } catch {
    window.open(url, '_blank')
  }
}

export function GoRavenFile({ kind = 'doc', path = '', name, description }: GoRavenFileProps) {
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
  const [dialogOpen, setDialogOpen] = useState(false)

  useEffect(() => {
    if (!isPrivate) return
    let cancelled = false

    const cached = peekCachedBlobUrl(url)
    if (cached) {
      setBlobUrl(cached)
      setLoading(false)
      getCachedBlobUrl(url).then(({ blobUrl }) => {
        if (!cancelled && blobUrl !== cached) setBlobUrl(blobUrl)
      }).catch(() => {})
      return () => { cancelled = true }
    }

    getCachedBlobUrl(url)
      .then(({ blobUrl }) => {
        if (cancelled) return
        setBlobUrl(blobUrl)
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

  const displayName = name || getFileName(path)
  const previewItem: FileItem = {
    name: displayName,
    path,
    isDir: false,
    size: 0,
    modTime: new Date().toISOString(),
  }

  return (
    <>
      <figure className="group relative my-3 cursor-zoom-in" onClick={() => setDialogOpen(true)}>
        <img
          src={blobUrl || url}
          alt={displayName}
          className="mx-auto block max-h-80 max-w-full rounded-md object-contain transition-opacity group-hover:opacity-90"
        />
        <span className="pointer-events-none absolute right-2 top-2 flex size-7 items-center justify-center rounded-md bg-black/40 text-white opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100">
          <Maximize2 className="size-3.5" />
        </span>
        {(name || description) && (
          <figcaption className="mt-1.5">
            {name && <p className="text-xs text-text-2">{name}</p>}
            {description && <p className="mt-0.5 text-xs text-text-muted">{description}</p>}
          </figcaption>
        )}
      </figure>
      <PreviewDialog
        item={dialogOpen ? previewItem : null}
        type="image"
        url={blobUrl || (!isPrivate ? url : null)}
        text={null}
        sheets={null}
        loading={isPrivate && !blobUrl && !error}
        error={error}
        onClose={() => setDialogOpen(false)}
      />
    </>
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
  const [dialogOpen, setDialogOpen] = useState(false)
  const [aspectRatio, setAspectRatio] = useState<number | null>(null)
  const videoRef = useRef<HTMLVideoElement>(null)

  const fetchBlob = useCallback(() => {
    return getCachedBlobUrl(url).then(({ blobUrl }) => blobUrl)
  }, [url])

  const handlePlay = useCallback(() => {
    // For private files, we need to fetch with auth first
    if (isPrivate && !blobUrl) {
      setLoading(true)
      fetchBlob()
        .then((objectUrl) => {
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
  }, [isPrivate, blobUrl, url, fetchBlob])

  const handleExpand = useCallback(() => {
    if (isPrivate && !blobUrl) {
      setLoading(true)
      fetchBlob()
        .then((objectUrl) => {
          setBlobUrl(objectUrl)
          setLoading(false)
          setDialogOpen(true)
        })
        .catch(() => setLoading(false))
      return
    }
    setDialogOpen(true)
  }, [isPrivate, blobUrl, fetchBlob])

  const handleEnded = useCallback(() => setPlaying(false), [])

  useEffect(() => {
    const src = blobUrl || (!isPrivate ? url : null)
    if (!src) return
    const el = videoRef.current
    if (!el) return
    const onLoaded = () => {
      if (el.videoWidth && el.videoHeight) setAspectRatio(el.videoWidth / el.videoHeight)
    }
    el.addEventListener('loadedmetadata', onLoaded)
    return () => el.removeEventListener('loadedmetadata', onLoaded)
  }, [blobUrl, url, isPrivate])

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

  const displayName = name || getFileName(path)
  const previewItem: FileItem = {
    name: displayName,
    path,
    isDir: false,
    size: 0,
    modTime: new Date().toISOString(),
  }

  return (
    <>
      <figure className="my-3 flex justify-center">
        <div
          className="relative max-h-112 w-full overflow-hidden rounded-md border border-border bg-black"
          style={aspectRatio ? { aspectRatio: String(aspectRatio) } : undefined}
        >
          <video
            ref={videoRef}
            src={blobUrl || (!isPrivate ? url : undefined)}
            controls={playing}
            onEnded={handleEnded}
            className="block h-full w-full object-contain"
          />
          {!playing && (
            <button
              onClick={handlePlay}
              className="absolute inset-0 flex items-center justify-center bg-gradient-to-b from-black/20 to-black/40 transition-colors hover:from-black/10 hover:to-black/30"
            >
              <span className="flex size-14 items-center justify-center rounded-full bg-black/50 backdrop-blur-sm transition-transform hover:scale-105">
                <Play className="ml-0.5 size-6 text-white" fill="currentColor" />
              </span>
            </button>
          )}
          <button
            onClick={(e) => { e.stopPropagation(); handleExpand() }}
            title={t('common.fullscreen')}
            className="absolute right-2 top-2 flex size-7 items-center justify-center rounded-md bg-black/40 text-white opacity-0 backdrop-blur-sm transition-opacity hover:bg-black/60 group-hover:opacity-100"
          >
            <Maximize2 className="size-3.5" />
          </button>
        </div>
        {(name || description) && (
          <figcaption className="mt-1.5">
            {name && <p className="text-xs text-text-2">{name}</p>}
            {description && <p className="mt-0.5 text-xs text-text-muted">{description}</p>}
          </figcaption>
        )}
      </figure>
      <PreviewDialog
        item={dialogOpen ? previewItem : null}
        type="video"
        url={blobUrl || (!isPrivate ? url : null)}
        text={null}
        sheets={null}
        loading={isPrivate && !blobUrl && loading}
        error={false}
        onClose={() => setDialogOpen(false)}
      />
    </>
  )
}
