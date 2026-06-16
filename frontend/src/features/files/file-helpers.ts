import {
  Archive,
  File,
  FileText,
  FolderOpen,
  Image,
  Table2,
  Video,
  type LucideIcon,
} from 'lucide-react'
import type { FileItem } from '@/api/types'
import { t as translate } from '@/i18n'

export type SortField = 'name' | 'size' | 'time'
export type SortOrder = 'asc' | 'desc'
export type PageState = 'loading' | 'data' | 'empty' | 'error'
export type PreviewType = 'image' | 'video' | 'audio' | 'text' | 'pdf'

export interface ContextMenuState {
  x: number
  y: number
  item: FileItem
}

export const MAX_TEXT_PREVIEW_SIZE = 512 * 1024

export const IMAGE_EXTS = new Set([
  'png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico', 'avif',
])
export const VIDEO_EXTS = new Set(['mp4', 'webm', 'mov', 'avi', 'mkv', 'ogv'])
export const AUDIO_EXTS = new Set(['mp3', 'wav', 'ogg', 'flac', 'aac', 'm4a', 'wma', 'opus'])
export const PDF_EXTS = new Set(['pdf'])
export const TEXT_EXTS = new Set([
  'txt', 'md', 'csv', 'json', 'xml', 'yaml', 'yml', 'toml', 'ini', 'log',
  'html', 'htm', 'css', 'js', 'ts', 'jsx', 'tsx', 'py', 'go', 'rs', 'java',
  'c', 'cpp', 'h', 'hpp', 'sh', 'bash', 'sql', 'env', 'cfg', 'conf',
  'properties', 'diff', 'patch', 'r', 'rb', 'php', 'swift', 'kt', 'scala',
  'vue', 'svelte', 'graphql', 'gql', 'proto',
])

export const ILLEGAL_CHARS = /[\/\\:*?"<>|]/
export const KEY_REGEX = /^[A-Za-z0-9_]+$/

export function formatSize(bytes: number): string {
  if (bytes === 0) return '--'
  if (bytes >= 1_073_741_824) return (bytes / 1_073_741_824).toFixed(1) + ' GB'
  if (bytes >= 1_048_576) return (bytes / 1_048_576).toFixed(1) + ' MB'
  if (bytes >= 1_024) return (bytes / 1_024).toFixed(1) + ' KB'
  return bytes + ' B'
}

export function formatTime(iso: string): string {
  const now = Date.now()
  const t = new Date(iso).getTime()
  const diff = now - t
  const minutes = Math.floor(diff / 60_000)
  const hours = Math.floor(diff / 3_600_000)
  const days = Math.floor(diff / 86_400_000)
  const weeks = Math.floor(days / 7)

  if (minutes < 1) return translate('files.justNow')
  if (minutes < 60) return `${minutes}${translate('files.minutesAgo')}`
  if (hours < 24) return `${hours}${translate('files.hoursAgo')}`
  if (days < 7) return `${days}${translate('files.daysAgo')}`
  if (weeks < 4) return `${weeks}${translate('files.weeksAgo')}`

  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function getFileIcon(item: FileItem): LucideIcon {
  if (item.isDir) return FolderOpen
  const ext = item.name.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'pdf':
    case 'doc':
    case 'docx':
    case 'txt':
    case 'md':
      return FileText
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'svg':
    case 'webp':
      return Image
    case 'mp4':
    case 'mov':
    case 'avi':
    case 'webm':
      return Video
    case 'zip':
    case 'tar':
    case 'gz':
    case 'rar':
    case '7z':
      return Archive
    case 'csv':
    case 'xls':
    case 'xlsx':
      return Table2
    default:
      return File
  }
}

export function canPreview(item: FileItem): boolean {
  if (item.isDir) return false
  const ext = item.name.split('.').pop()?.toLowerCase()
  return !!ext && (IMAGE_EXTS.has(ext) || VIDEO_EXTS.has(ext) || AUDIO_EXTS.has(ext) || TEXT_EXTS.has(ext) || PDF_EXTS.has(ext))
}

export function getPreviewType(item: FileItem): PreviewType | null {
  const ext = item.name.split('.').pop()?.toLowerCase()
  if (!ext) return null
  if (IMAGE_EXTS.has(ext)) return 'image'
  if (VIDEO_EXTS.has(ext)) return 'video'
  if (AUDIO_EXTS.has(ext)) return 'audio'
  if (TEXT_EXTS.has(ext)) return 'text'
  if (PDF_EXTS.has(ext)) return 'pdf'
  return null
}

export function sortItems(items: FileItem[], field: SortField, order: SortOrder): FileItem[] {
  const sorted = [...items].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1

    let cmp = 0
    switch (field) {
      case 'name':
        cmp = a.name.localeCompare(b.name)
        break
      case 'size':
        cmp = a.size - b.size
        break
      case 'time':
        cmp = new Date(a.modTime).getTime() - new Date(b.modTime).getTime()
        break
    }
    return order === 'asc' ? cmp : -cmp
  })
  return sorted
}

export function validateName(name: string, existingNames: string[]): string | null {
  if (!name.trim()) return translate('files.errNameEmpty')
  if (ILLEGAL_CHARS.test(name)) return translate('files.errIllegalChars')
  if (existingNames.includes(name.trim())) return translate('files.errNameExists')
  return null
}
