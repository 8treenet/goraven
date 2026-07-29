import { useState, useCallback } from 'react'
import { toast } from 'sonner'
import { t as translate } from '@/i18n'
import { useUserStore } from '@/stores/user-store'
import { getAkDownloadUrl } from '@/api/files'
import { getCachedBlobUrl, peekCachedBlobUrl } from '@/lib/file-blob-cache'
import type { FileItem } from '@/api/types'
import {
  getPreviewType,
  MAX_TEXT_PREVIEW_SIZE,
  type PreviewType,
  type SheetData,
} from './file-helpers'
import { read as xlsxRead, utils as xlsxUtils } from 'xlsx'

/**
 * 可复用的文件预览/下载 Hook。
 *
 * 通过参数化下载 URL 构建和临时凭证创建，同时服务于「我的文件」和「团队项目」。
 */
export interface UseFilePreviewOptions {
  /** 构建需 Bearer token 的下载 URL（fetch 用） */
  buildDownloadUrl: (filePath: string) => string
  /** 创建临时访问凭证（用于 html iframe / office 在线预览） */
  createAccess: (path: string, type: 'file' | 'dir') => Promise<{ ak: string; expiresAt: number }>
  /** 将项目相对路径转换为 ak 下载 URL 中使用的路径（mine 视图为 identity，team 视图加 projects/<name>/ 前缀） */
  buildAkPath: (filePath: string) => string
}

export function useFilePreview(options: UseFilePreviewOptions) {
  const { buildDownloadUrl, createAccess, buildAkPath } = options

  const [previewItem, setPreviewItem] = useState<FileItem | null>(null)
  const [previewType, setPreviewType] = useState<PreviewType | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [previewText, setPreviewText] = useState<string | null>(null)
  const [previewSheets, setPreviewSheets] = useState<SheetData[] | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState(false)

  const handleDownload = useCallback(
    async (item: FileItem, currentDir: string) => {
      const filePath = `${currentDir === '/' ? '' : currentDir}/${item.name}`
      const url = buildDownloadUrl(filePath)
      try {
        const { blobUrl } = await getCachedBlobUrl(url)
        const a = document.createElement('a')
        a.href = blobUrl
        a.download = item.name
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
      } catch (err: unknown) {
        toast.error((err instanceof Error ? err.message : '') || translate('files.downloadFailed'))
      }
    },
    [buildDownloadUrl],
  )

  const handlePreview = useCallback(
    (item: FileItem, currentDir: string) => {
      const filePath = `${currentDir === '/' ? '' : currentDir}/${item.name}`
      const ptype = getPreviewType(item)
      const url = buildDownloadUrl(filePath)

      // Non-previewable file: notify instead of opening a broken preview
      if (!ptype) {
        toast.info(translate('files.previewUnsupported'))
        return
      }

      if (ptype === 'image' || ptype === 'video' || ptype === 'audio' || ptype === 'pdf') {
        const cachedUrl = peekCachedBlobUrl(url)
        if (cachedUrl) {
          setPreviewItem(item)
          setPreviewType(ptype)
          setPreviewText(null)
          setPreviewSheets(null)
          setPreviewError(false)
          setPreviewUrl(cachedUrl)
          setPreviewLoading(false)
          getCachedBlobUrl(url).then(({ blobUrl }) => {
            if (blobUrl !== cachedUrl) setPreviewUrl(blobUrl)
          }).catch(() => {})
          return
        }
      }

      const token = useUserStore.getState().token

      setPreviewItem(item)
      setPreviewType(ptype)
      setPreviewUrl(null)
      setPreviewText(null)
      setPreviewSheets(null)
      setPreviewError(false)
      setPreviewLoading(true)

      const bearerHeaders: Record<string, string> = token ? { Authorization: `Bearer ${token}` } : {}

      if (ptype === 'text' || ptype === 'markdown') {
        fetch(url, { headers: bearerHeaders })
          .then((res) => {
            if (!res.ok) throw new Error('preview failed')
            return res.text()
          })
          .then((text) => {
            setPreviewText(text.length > MAX_TEXT_PREVIEW_SIZE ? text.slice(0, MAX_TEXT_PREVIEW_SIZE) + '\n\n... File too large, preview truncated' : text)
          })
          .catch(() => setPreviewError(true))
          .finally(() => setPreviewLoading(false))
      } else if (ptype === 'xlsx') {
        fetch(url, { headers: bearerHeaders })
          .then((res) => {
            if (!res.ok) throw new Error('preview failed')
            return res.arrayBuffer()
          })
          .then((buf) => {
            const wb = xlsxRead(buf)
            const sheets: SheetData[] = wb.SheetNames.map((name) => ({
              name,
              html: xlsxUtils.sheet_to_html(wb.Sheets[name]),
            }))
            setPreviewSheets(sheets)
          })
          .catch(() => setPreviewError(true))
          .finally(() => setPreviewLoading(false))
      } else if (ptype === 'html') {
        createAccess(currentDir === '/' ? '/' : currentDir, 'dir')
          .then((res) => {
            setPreviewUrl(getAkDownloadUrl(res.ak, buildAkPath(filePath)))
          })
          .catch(() => setPreviewError(true))
          .finally(() => setPreviewLoading(false))
      } else if (ptype === 'office') {
        createAccess(filePath, 'file')
          .then((res) => {
            const akPath = buildAkPath(filePath).replace(/^\//, '').split('/').map(encodeURIComponent).join('/')
            const absUrl = `${window.location.origin}/api/hfs/ak/${res.ak}/${akPath}`
            setPreviewUrl(`https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(absUrl)}`)
          })
          .catch(() => setPreviewError(true))
          .finally(() => setPreviewLoading(false))
      } else {
        getCachedBlobUrl(url)
          .then(({ blobUrl }) => {
            setPreviewUrl(blobUrl)
          })
          .catch(() => setPreviewError(true))
          .finally(() => setPreviewLoading(false))
      }
    },
    [buildDownloadUrl, createAccess, buildAkPath],
  )

  const closePreview = useCallback(() => {
    setPreviewItem(null)
    setPreviewType(null)
    setPreviewError(false)
    setPreviewLoading(false)
    setPreviewText(null)
    setPreviewSheets(null)
    setPreviewUrl(null)
  }, [])

  return {
    previewItem,
    previewType,
    previewUrl,
    previewText,
    previewSheets,
    previewLoading,
    previewError,
    handlePreview,
    closePreview,
    handleDownload,
  }
}
