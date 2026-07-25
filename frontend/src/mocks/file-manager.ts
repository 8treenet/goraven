/* ============================================
   File Manager — mock data and async functions
   ============================================ */

import type { FileItem, FileListResponse, StorageUsage } from './types'
import { listDelay, mutationDelay, heavyDelay } from './delay'

/* ============================================
   Mock Filesystem
   ============================================ */

const MOCK_FS: Record<string, FileItem[]> = {
  '/': [
    { name: 'documents', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'temp', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'downloads', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'images', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'videos', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'projects', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'skills', isDir: true, size: 0, modTime: '2025-01-15T08:00:00Z', isDefault: true },
    { name: 'report.pdf', isDir: false, size: 1_258_291, modTime: '2025-05-26T09:30:00Z' },
    { name: 'data.csv', isDir: false, size: 314_572, modTime: '2025-05-25T16:45:00Z' },
    { name: 'avatar.png', isDir: false, size: 2_097_152, modTime: '2025-05-24T11:20:00Z' },
    { name: 'demo.mp4', isDir: false, size: 47_185_920, modTime: '2025-05-20T08:00:00Z' },
    { name: 'archive.zip', isDir: false, size: 3_670_016, modTime: '2025-05-19T14:30:00Z' },
    { name: 'notes', isDir: true, size: 0, modTime: '2025-05-27T10:00:00Z' },
  ],
  '/documents': [
    { name: 'readme.txt', isDir: false, size: 2_048, modTime: '2025-03-10T14:30:00Z' },
    { name: 'proposal.pdf', isDir: false, size: 6_081_740, modTime: '2025-05-22T09:15:00Z' },
    { name: 'meeting-notes.md', isDir: false, size: 15_360, modTime: '2025-05-26T17:00:00Z' },
    { name: 'archive', isDir: true, size: 0, modTime: '2025-05-20T10:00:00Z' },
    { name: 'budget-2025.xlsx', isDir: false, size: 248_832, modTime: '2025-05-18T14:00:00Z' },
    { name: 'contract-v3.pdf', isDir: false, size: 3_145_728, modTime: '2025-05-15T09:30:00Z' },
    { name: 'screencast.mp4', isDir: false, size: 28_311_552, modTime: '2025-05-12T16:00:00Z' },
    { name: 'backup', isDir: true, size: 0, modTime: '2025-05-10T08:00:00Z' },
    { name: 'release-notes.txt', isDir: false, size: 8_192, modTime: '2025-05-08T11:45:00Z' },
    { name: 'logo.svg', isDir: false, size: 12_288, modTime: '2025-05-05T20:00:00Z' },
    { name: 'email-template.html', isDir: false, size: 12_288, modTime: '2025-02-05T09:00:00Z' },
    { name: 'terms-of-service.md', isDir: false, size: 24_576, modTime: '2025-02-01T10:00:00Z' },
    { name: 'database-dump.sql.gz', isDir: false, size: 34_603_008, modTime: '2025-01-15T10:00:00Z' },
    { name: 'notes-2024.txt', isDir: false, size: 1_024, modTime: '2024-12-31T23:00:00Z' },
  ],
  '/projects': [
    { name: 'goraven', isDir: true, size: 0, modTime: '2025-05-20T10:00:00Z' },
    { name: 'website', isDir: true, size: 0, modTime: '2025-04-15T08:30:00Z' },
    { name: '.gitconfig', isDir: false, size: 1_024, modTime: '2025-05-10T12:00:00Z' },
  ],
  '/images': [
    { name: 'screenshot-1.png', isDir: false, size: 450_560, modTime: '2025-05-26T20:00:00Z' },
    { name: 'screenshot-2.png', isDir: false, size: 380_928, modTime: '2025-05-26T19:45:00Z' },
    { name: 'logo.svg', isDir: false, size: 8_192, modTime: '2025-05-15T10:00:00Z' },
  ],
  '/notes': [
    { name: 'todo.md', isDir: false, size: 4_096, modTime: '2025-05-27T08:00:00Z' },
    { name: 'ideas.txt', isDir: false, size: 2_560, modTime: '2025-05-25T22:00:00Z' },
  ],
  '/videos': [],
  '/temp': [],
  '/downloads': [],
  '/skills': [],
}

/* ============================================
   Helpers
   ============================================ */

function normalizePath(path: string): string {
  if (path === '/') return '/'
  return path.replace(/\/+$/, '') || '/'
}

function getParentPath(path: string): string {
  const normalized = normalizePath(path)
  if (normalized === '/') return '/'
  const parts = normalized.split('/')
  parts.pop()
  return parts.join('/') || '/'
}

function getBaseName(path: string): string {
  const normalized = normalizePath(path)
  if (normalized === '/') return ''
  return normalized.split('/').pop() || ''
}

function cloneItems(dir: string): FileItem[] {
  return (MOCK_FS[dir] ?? []).map((item) => ({ ...item }))
}

function computeUsage(): { totalSize: number; fileCount: number } {
  let totalSize = 0
  let fileCount = 0
  for (const items of Object.values(MOCK_FS)) {
    for (const item of items) {
      totalSize += item.size
      fileCount += 1
    }
  }
  return { totalSize, fileCount }
}

/* ============================================
   Async Functions
   ============================================ */

/** List files in a directory. Defaults to root "/". */
export async function listFiles(dir: string = '/'): Promise<FileListResponse> {
  await listDelay()
  const normalized = normalizePath(dir)

  if (!(normalized in MOCK_FS)) {
    throw new Error(`Directory not found: ${normalized}`)
  }

  return {
    items: cloneItems(normalized),
  }
}

/** Create a directory at the given absolute path. */
export async function mkdir(path: string): Promise<void> {
  await mutationDelay()

  const normalized = normalizePath(path)
  const parentPath = getParentPath(normalized)
  const dirName = getBaseName(normalized)

  if (!(parentPath in MOCK_FS)) {
    MOCK_FS[parentPath] = []
  }

  const siblings = MOCK_FS[parentPath]
  if (siblings.some((item) => item.name === dirName)) {
    throw new Error(`Already exists: ${dirName}`)
  }

  siblings.push({
    name: dirName,
    isDir: true,
    size: 0,
    modTime: new Date().toISOString(),
  })

  // Ensure the new directory has an entry in MOCK_FS
  if (!(normalized in MOCK_FS)) {
    MOCK_FS[normalized] = []
  }
}

/** Rename a file or directory. Both oldPath and newPath are absolute. */
export async function rename(oldPath: string, newPath: string): Promise<void> {
  await mutationDelay()

  const oldNormalized = normalizePath(oldPath)
  const newNormalized = normalizePath(newPath)

  if (oldNormalized === '/') {
    throw new Error('Cannot rename root')
  }

  const oldParent = getParentPath(oldNormalized)
  const newParent = getParentPath(newNormalized)
  const oldName = getBaseName(oldNormalized)
  const newName = getBaseName(newNormalized)

  if (oldParent !== newParent) {
    throw new Error('Cross-directory rename not supported in mock')
  }

  const items = MOCK_FS[oldParent]
  if (!items) {
    throw new Error(`Directory not found: ${oldParent}`)
  }

  const target = items.find((item) => item.name === oldName)
  if (!target) {
    throw new Error(`Not found: ${oldName}`)
  }

  if (items.some((item) => item.name === newName && item.name !== oldName)) {
    throw new Error(`Already exists: ${newName}`)
  }

  target.name = newName
  target.modTime = new Date().toISOString()

  // If renaming a directory, also rename the MOCK_FS key
  if (target.isDir) {
    MOCK_FS[newNormalized] = MOCK_FS[oldNormalized]
    delete MOCK_FS[oldNormalized]
  }
}

/** Delete files and/or directories by absolute paths. */
export async function deleteFiles(paths: string[]): Promise<void> {
  await mutationDelay()

  for (const path of paths) {
    const normalized = normalizePath(path)
    if (normalized === '/') {
      throw new Error('Cannot delete root')
    }

    const parentPath = getParentPath(normalized)
    const name = getBaseName(normalized)

    const items = MOCK_FS[parentPath]
    if (!items) continue

    const idx = items.findIndex((item) => item.name === name)
    if (idx === -1) continue

    const item = items[idx]
    items.splice(idx, 1)

    // If deleting a directory, remove its MOCK_FS entry
    if (item.isDir && normalized in MOCK_FS) {
      delete MOCK_FS[normalized]
    }
  }
}

/** Compress the given paths into a zip file. Returns the zip path. */
export async function compress(
  paths: string[],
  outputName?: string,
): Promise<{ zipPath: string }> {
  await heavyDelay()

  if (paths.length === 0) {
    throw new Error('No paths to compress')
  }

  const firstPath = normalizePath(paths[0])
  const parentPath = getParentPath(firstPath)
  const baseName = outputName || getBaseName(firstPath) || 'archive'
  const zipName = baseName.endsWith('.zip') ? baseName : `${baseName}.zip`

  if (!(parentPath in MOCK_FS)) {
    throw new Error(`Directory not found: ${parentPath}`)
  }

  const siblings = MOCK_FS[parentPath]
  if (siblings.some((item) => item.name === zipName)) {
    throw new Error(`Already exists: ${zipName}`)
  }

  siblings.push({
    name: zipName,
    isDir: false,
    size: Math.floor(Math.random() * 5_242_880) + 102_400,
    modTime: new Date().toISOString(),
  })

  return { zipPath: `${parentPath === '/' ? '' : parentPath}/${zipName}` }
}

/** Decompress a zip file into the current directory or a subdirectory. */
export async function decompress(
  path: string,
  toSubDir?: boolean,
): Promise<void> {
  await heavyDelay()

  const normalized = normalizePath(path)
  if (normalized === '/') {
    throw new Error('Cannot decompress root')
  }

  const parentPath = getParentPath(normalized)
  const zipName = getBaseName(normalized)

  if (!zipName.endsWith('.zip')) {
    throw new Error('Not a zip file')
  }

  const items = MOCK_FS[parentPath]
  if (!items) {
    throw new Error(`Directory not found: ${parentPath}`)
  }

  const zipItem = items.find((item) => item.name === zipName)
  if (!zipItem) {
    throw new Error(`Not found: ${zipName}`)
  }

  const baseName = zipName.replace(/\.zip$/i, '')

  if (toSubDir) {
    const subDirPath = `${parentPath === '/' ? '' : parentPath}/${baseName}`
    if (items.some((item) => item.name === baseName)) {
      throw new Error(`Already exists: ${baseName}`)
    }
    items.push({
      name: baseName,
      isDir: true,
      size: 0,
      modTime: new Date().toISOString(),
    })
    MOCK_FS[subDirPath] = [
      {
        name: 'readme.txt',
        isDir: false,
        size: 2_048,
        modTime: new Date().toISOString(),
      },
      {
        name: 'data.csv',
        isDir: false,
        size: 15_360,
        modTime: new Date().toISOString(),
      },
    ]
  } else {
    const newItems: FileItem[] = [
      {
        name: `${baseName}_readme.txt`,
        isDir: false,
        size: 2_048,
        modTime: new Date().toISOString(),
      },
      {
        name: `${baseName}_data.csv`,
        isDir: false,
        size: 15_360,
        modTime: new Date().toISOString(),
      },
    ]
    items.push(...newItems)
  }
}

/** Get storage usage across all directories. */
export async function getUsage(): Promise<StorageUsage> {
  await listDelay()

  const { totalSize, fileCount } = computeUsage()

  // Simulate a total disk size (100 GB)
  const totalSize_ = 100 * 1024 * 1024 * 1024

  return {
    totalSize: totalSize_,
    usedSize: totalSize,
    fileCount,
  }
}
