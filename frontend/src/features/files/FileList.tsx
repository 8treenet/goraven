import { useState, useCallback, useMemo, useEffect, useRef, forwardRef, useImperativeHandle } from 'react'
import { toast } from 'sonner'
import {
  ArrowLeft,
  Upload,
  FolderPlus,
  Trash2,
  Archive,
  FolderArchive,
  Download,
  Pencil,
  Lock,
  FolderOpen,
  AlertCircle,
  RefreshCw,
  Eye,
  X,
  Ellipsis,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import type { FileItem } from '@/api/types'
import type { UploadProgress } from '@/hooks/useChunkUpload'
import {
  ContextMenu,
  ContextMenuItem,
  ColumnHeader,
} from './FileContextMenu'
import { FileDialogs, type FileDialogMode } from './FileDialogs'
import { PreviewDialog } from './PreviewDialog'
import { useFilePreview } from './useFilePreview'
import {
  canPreview,
  formatSize,
  formatTime,
  getFileIcon,
  sortItems,
  validateName,
  type ContextMenuState,
  type PageState,
  type SortField,
  type SortOrder,
} from './file-helpers'

/** 文件操作适配器：由「我的文件」与「团队项目」各自实现 */
export interface FileListApi {
  list: (dir: string) => Promise<{ items: FileItem[] }>
  mkdir: (dir: string) => Promise<unknown>
  rename: (oldPath: string, newPath: string) => Promise<unknown>
  remove: (paths: string[]) => Promise<unknown>
  compress: (req: { paths: string[]; outputName: string }) => Promise<unknown>
  decompress: (req: { path: string; toSubDir: boolean }) => Promise<unknown>
  downloadUrl: (path: string) => string
  createAccess: (path: string, type: 'file' | 'dir') => Promise<{ ak: string; expiresAt: number }>
}

export interface FileListProps {
  initialDir: string
  rootDir?: string
  /** 根目录返回行的目录名（团队项目：项目名） */
  rootLabel?: string
  /** 在根目录点返回时调用（团队项目：离开项目回列表） */
  onBackAtRoot?: () => void
  api: FileListApi
  uploadFile: (file: File, dir: string) => Promise<unknown>
  uploadProgress: UploadProgress | null
  isUploading: boolean
  /** ak 下载路径转换（团队项目加 projects/<name>/ 前缀），默认恒等 */
  buildAkPath?: (path: string) => string
  /** 重命名/上传/删除成功后使 blob 缓存失效（传文件路径） */
  evictCache?: (path: string) => void
  /** 受保护条目（我的文件：默认目录），不允许重命名/删除 */
  isProtected?: (item: FileItem) => boolean
  /** 根目录列表过滤（我的文件：隐藏 projects 目录） */
  filterAtRoot?: (item: FileItem) => boolean
  /** 新建对话框标题（我的文件 /projects 下：新建项目） */
  getCreateTitle?: (dir: string) => string | undefined
  /** 是否允许空白处右键菜单（我的文件 /projects 下禁用） */
  allowEmptyContextMenu?: (dir: string) => boolean
  /** 错误视图主文案，默认 common.loadFailed */
  errorTitle?: string
  /** 错误视图副文案（我的文件：files.errReadDir） */
  errorHint?: string
  /** 错误视图按钮文案（团队项目：files.teamProjects） */
  errorActionLabel?: string
  /** 错误视图按钮动作（团队项目：返回列表）；不传则默认重试 */
  errorAction?: () => void
}

export interface FileListHandle {
  upload: () => void
  newFolder: () => void
}

export const FileList = forwardRef<FileListHandle, FileListProps>(function FileList(props, ref) {
  const {
    initialDir,
    rootDir = '/',
    rootLabel,
    onBackAtRoot,
    api,
    uploadFile,
    uploadProgress,
    isUploading,
    buildAkPath = (p: string) => p,
    evictCache,
    isProtected,
    filterAtRoot,
    getCreateTitle,
    allowEmptyContextMenu = () => true,
    errorTitle,
    errorHint,
    errorActionLabel,
    errorAction,
  } = props

  const [pageState, setPageState] = useState<PageState>('loading')
  const [currentDir, setCurrentDir] = useState(initialDir)
  const [items, setItems] = useState<FileItem[]>([])
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc')
  const [selectedNames, setSelectedNames] = useState<Set<string>>(new Set())
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null)
  const [emptyContextMenu, setEmptyContextMenu] = useState<{ x: number; y: number } | null>(null)
  const [renamingItem, setRenamingItem] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const t = useT()

  const [dialogMode, setDialogMode] = useState<FileDialogMode>(null)
  const [dialogValue, setDialogValue] = useState('')
  const [dialogError, setDialogError] = useState<string | null>(null)

  const {
    previewItem,
    previewType,
    previewUrl,
    previewText,
    previewSheets,
    previewLoading,
    previewError,
    handlePreview: rawHandlePreview,
    closePreview,
    handleDownload: rawHandleDownload,
  } = useFilePreview({
    buildDownloadUrl: (filePath) => api.downloadUrl(filePath),
    createAccess: (path, type) => api.createAccess(path, type),
    buildAkPath,
  })

  const fileInputRef = useRef<HTMLInputElement>(null)
  const renameInputRef = useRef<HTMLInputElement>(null)

  const sortedItems = useMemo(
    () => sortItems(items, sortField, sortOrder),
    [items, sortField, sortOrder],
  )

  const selectedItems = useMemo(
    () => items.filter((item) => selectedNames.has(item.name)),
    [items, selectedNames],
  )

  const allSelected = items.length > 0 && selectedNames.size === items.length
  const someSelected = selectedNames.size > 0 && selectedNames.size < items.length

  const isRoot = currentDir === rootDir
  const currentDirName = isRoot ? (rootLabel ?? null) : (currentDir.split('/').pop() || null)

  const canDelete = selectedItems.length > 0 && !selectedItems.some((item) => isProtected?.(item))
  const canDecompress =
    selectedItems.length === 1 &&
    !selectedItems[0].isDir &&
    selectedItems[0].name.endsWith('.zip')

  const loadDir = useCallback((dir: string) => {
    setPageState('loading')
    setSelectedNames(new Set())
    setContextMenu(null)
    setEmptyContextMenu(null)
    setRenamingItem(null)

    api.list(dir === '/' ? '' : dir)
      .then((data) => {
        const filtered = dir === '/' ? data.items.filter(filterAtRoot ?? (() => true)) : data.items
        setItems(filtered)
        setPageState(filtered.length === 0 ? 'empty' : 'data')
      })
      .catch(() => {
        setPageState('error')
      })
  }, [api, filterAtRoot])

  useEffect(() => {
    loadDir(currentDir)
  }, [currentDir, loadDir])

  const navigateTo = useCallback((dir: string) => {
    setCurrentDir(dir)
  }, [])

  const goBack = useCallback(() => {
    if (isRoot) {
      onBackAtRoot?.()
      return
    }
    const parent = currentDir.split('/').slice(0, -1).join('/') || '/'
    setCurrentDir(parent)
  }, [currentDir, isRoot, onBackAtRoot])

  const handleSort = useCallback(
    (field: SortField) => {
      if (sortField === field) {
        setSortOrder((o) => (o === 'asc' ? 'desc' : 'asc'))
      } else {
        setSortField(field)
        setSortOrder('asc')
      }
    },
    [sortField],
  )

  const handleCheckboxToggle = useCallback((name: string) => {
    setSelectedNames((prev) => {
      const next = new Set(prev)
      if (next.has(name)) {
        next.delete(name)
      } else {
        next.add(name)
      }
      return next
    })
  }, [])

  const handleSelectAll = useCallback(() => {
    if (allSelected) {
      setSelectedNames(new Set())
    } else {
      setSelectedNames(new Set(items.map((item) => item.name)))
    }
  }, [allSelected, items])

  const handleContextMenu = useCallback((item: FileItem, e: React.MouseEvent) => {
    e.preventDefault()
    if (item.isDir && isProtected?.(item)) return
    setContextMenu({ x: e.clientX, y: e.clientY, item })
  }, [isProtected])

  const handleEmptyContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setEmptyContextMenu({ x: e.clientX, y: e.clientY })
  }, [])

  const startRename = useCallback((item: FileItem) => {
    setContextMenu(null)
    setRenamingItem(item.name)
    setRenameValue(item.name)
    setTimeout(() => renameInputRef.current?.select(), 0)
  }, [])

  const commitRename = useCallback(() => {
    if (!renamingItem) return
    const error = validateName(
      renameValue,
      items.filter((i) => i.name !== renamingItem).map((i) => i.name),
    )
    if (error) {
      toast.error(error)
      return
    }
    const oldPath = `${currentDir === '/' ? '' : currentDir}/${renamingItem}`
    const newPath = `${currentDir === '/' ? '' : currentDir}/${renameValue.trim()}`
    api.rename(oldPath, newPath)
      .then(() => {
        evictCache?.(oldPath)
        setRenamingItem(null)
        toast.success(translate('files.renameSuccess'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.renameFailed'))
      })
  }, [renamingItem, renameValue, items, currentDir, api, evictCache, loadDir])

  const cancelRename = useCallback(() => {
    setRenamingItem(null)
  }, [])

  const handleUploadFiles = useCallback(
    async (files: FileList | File[]) => {
      const fileArray = Array.from(files)
      for (const file of fileArray) {
        try {
          await uploadFile(file, currentDir === '/' ? '' : currentDir)
          const filePath = `${currentDir === '/' ? '' : currentDir}/${file.name}`
          evictCache?.(filePath)
        } catch {
        }
      }
      loadDir(currentDir)
    },
    [uploadFile, currentDir, evictCache, loadDir],
  )

  const handleUploadClick = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.files && e.target.files.length > 0) {
        handleUploadFiles(e.target.files)
        e.target.value = ''
      }
    },
    [handleUploadFiles],
  )

  const openNewFolderDialog = useCallback(() => {
    setDialogMode('newFolder')
    setDialogValue('')
    setDialogError(null)
  }, [])

  useImperativeHandle(ref, () => ({
    upload: () => fileInputRef.current?.click(),
    newFolder: () => openNewFolderDialog(),
  }), [openNewFolderDialog])

  const openDeleteDialog = useCallback(() => {
    setDialogMode('delete')
    setDialogError(null)
  }, [])

  const openCompressDialog = useCallback(() => {
    setDialogMode('compress')
    setDialogValue(selectedItems[0]?.name || 'archive')
    setDialogError(null)
  }, [selectedItems])

  const openDecompressDialog = useCallback(() => {
    setDialogMode('decompress')
    setDialogError(null)
  }, [])

  const closeDialog = useCallback(() => {
    setDialogMode(null)
    setDialogValue('')
    setDialogError(null)
  }, [])

  const handleCreateFolder = useCallback(() => {
    const error = validateName(
      dialogValue,
      items.map((i) => i.name),
    )
    if (error) {
      setDialogError(error)
      return
    }
    const dirPath = `${currentDir === '/' ? '' : currentDir}/${dialogValue.trim()}`
    api.mkdir(dirPath)
      .then(() => {
        closeDialog()
        toast.success(translate('files.dirCreated'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.dirCreateFailed'))
      })
  }, [dialogValue, items, currentDir, api, closeDialog, loadDir])

  const handleDelete = useCallback(() => {
    const names = selectedItems.map((i) => i.name)
    const paths = names.map((n) => `${currentDir === '/' ? '' : currentDir}/${n}`)
    api.remove(paths)
      .then(() => {
        for (const p of paths) {
          evictCache?.(p)
        }
        setSelectedNames(new Set())
        closeDialog()
        toast.success(translate('files.deletedCount').replace('{n}', String(names.length)))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [selectedItems, currentDir, api, evictCache, closeDialog, loadDir])

  const handleCompress = useCallback(() => {
    const name = dialogValue.trim() || 'archive'
    const zipName = name.endsWith('.zip') ? name : `${name}.zip`
    const error = validateName(
      zipName,
      items.map((i) => i.name),
    )
    if (error) {
      setDialogError(error)
      return
    }
    const paths = selectedItems.map((i) => `${currentDir === '/' ? '' : currentDir}/${i.name}`)
    const outputName = name.endsWith('.zip') ? name.replace(/\.zip$/i, '') : name
    api.compress({ paths, outputName })
      .then(() => {
        closeDialog()
        toast.success(translate('files.compressComplete'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.compressFailed'))
      })
  }, [dialogValue, items, selectedItems, currentDir, api, closeDialog, loadDir])

  const handleDecompress = useCallback(
    (toSubDir: boolean) => {
      const zipItem = selectedItems[0]
      const zipPath = `${currentDir === '/' ? '' : currentDir}/${zipItem.name}`
      api.decompress({ path: zipPath, toSubDir })
        .then(() => {
          const baseName = zipItem.name.replace(/\.zip$/i, '')
          toast.success(toSubDir ? translate('files.extractedToSubdir').replace('{dir}', baseName) : translate('files.extractedToDir'))
          closeDialog()
          loadDir(currentDir)
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('files.extractFailed'))
        })
    },
    [selectedItems, currentDir, api, closeDialog, loadDir],
  )

  const handleDownload = useCallback((item: FileItem) => {
    setContextMenu(null)
    rawHandleDownload(item, currentDir)
  }, [currentDir, rawHandleDownload])

  const handlePreview = useCallback((item: FileItem) => {
    setContextMenu(null)
    rawHandlePreview(item, currentDir)
  }, [currentDir, rawHandlePreview])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (dialogMode || renamingItem) return

      if (e.key === 'F2' && selectedItems.length === 1 && !isProtected?.(selectedItems[0])) {
        e.preventDefault()
        startRename(selectedItems[0])
      }
      if ((e.key === 'Delete' || e.key === 'Backspace') && canDelete) {
        e.preventDefault()
        openDeleteDialog()
      }
      if (e.key === 'Escape') {
        setContextMenu(null)
        setEmptyContextMenu(null)
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [dialogMode, renamingItem, selectedItems, canDelete, isProtected, startRename, openDeleteDialog])

  const hasSelection = selectedNames.size > 0
  const hasRootBack = !!onBackAtRoot

  return (
    <div className="flex h-full flex-col">
      {/* Back / directory indicator */}
      {(hasRootBack || !isRoot) && !hasSelection && (
        <div className="flex h-8 shrink-0 items-center gap-1 px-3">
          <Button variant="ghost" size="icon" onClick={goBack} title={t('files.backTooltip')} className="size-6 text-text-3 hover:text-text-1">
            <ArrowLeft className="size-3.5" />
          </Button>
          <span className="text-xs font-medium text-folder">{currentDirName}</span>
        </div>
      )}

      {/* Selection action bar */}
      {hasSelection && (
        <div className="flex h-8 shrink-0 items-center gap-1 bg-interactive/5 px-3">
          <span className="mr-2 text-xs font-medium text-interactive tabular-nums">
            {t('files.selectedN').replace('{n}', String(selectedNames.size))}
          </span>
          {!selectedItems.some((i) => i.isDir) && selectedItems.length === 1 && (
            <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" aria-label={t('files.download')} onClick={() => handleDownload(selectedItems[0])}>
              <Download className="size-3.5" />
              <span className="hidden md:inline">{t('files.download')}</span>
            </Button>
          )}
          <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" aria-label={t('files.delete')} disabled={!canDelete} onClick={openDeleteDialog}>
            <Trash2 className="size-3.5" />
            <span className="hidden md:inline">{t('files.delete')}</span>
          </Button>
          <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" aria-label={t('files.compress')} onClick={openCompressDialog}>
            <Archive className="size-3.5" />
            <span className="hidden md:inline">{t('files.compress')}</span>
          </Button>
          {canDecompress && (
            <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" aria-label={t('files.decompress')} onClick={openDecompressDialog}>
              <FolderArchive className="size-3.5" />
              <span className="hidden md:inline">{t('files.decompress')}</span>
            </Button>
          )}
          <div className="flex-1" />
          <Button variant="ghost" size="icon" className="size-6 text-text-3 hover:text-text-1" aria-label={t('sidebar.close')} onClick={() => setSelectedNames(new Set())}>
            <X className="size-3.5" />
          </Button>
        </div>
      )}

      {/* File list area */}
      <div
        className="flex-1 overflow-auto"
        onContextMenu={handleEmptyContextMenu}
      >
        {pageState === 'loading' && (
          <div>
            {[1, 2, 3, 4, 5, 6, 7].map((i) => (
              <div key={i} className="flex h-10 items-center gap-3 border-b border-border px-4">
                <div className="size-4 shrink-0 animate-pulse rounded-sm bg-bg-layer-3" />
                <div className="h-4 flex-1 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-4 w-16 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-4 w-20 animate-pulse rounded bg-bg-layer-3" />
              </div>
            ))}
          </div>
        )}

        {pageState === 'error' && (
          <div className="flex h-full flex-col items-center justify-center gap-4">
            <AlertCircle className="size-8 text-text-3" />
            <div className="text-center">
              <p className="text-sm text-text-2">{errorTitle ?? t('common.loadFailed')}</p>
              {errorHint && (
                <p className="mt-1 text-sm text-text-3">{errorHint}</p>
              )}
            </div>
            <button
              onClick={errorAction ?? (() => loadDir(currentDir))}
              className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
            >
              {errorAction ? (
                <>
                  <ArrowLeft className="size-4" />
                  {errorActionLabel}
                </>
              ) : (
                <>
                  <RefreshCw className="size-4" />
                  {t('common.retry')}
                </>
              )}
            </button>
          </div>
        )}

        {pageState === 'empty' && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <FolderOpen className="size-12 text-text-3" />
            <div className="text-center">
              <p className="text-sm text-text-2">{t('files.emptyFolder')}</p>
              <p className="mt-1 text-sm text-text-3">{t('files.emptyHint')}</p>
            </div>
            <div className="mt-2 flex items-center gap-2">
              <button
                onClick={handleUploadClick}
                className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
              >
                <Upload className="size-4" />
                {t('files.upload')}
              </button>
              <button
                onClick={openNewFolderDialog}
                className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
              >
                <FolderPlus className="size-4" />
                {t('files.newFolder')}
              </button>
            </div>
          </div>
        )}

        {pageState === 'data' && (
          <>
            <div className="hidden h-8 items-center gap-3 border-b border-border px-4 md:flex">
              <div className="flex size-4 shrink-0 items-center justify-center">
                <input
                  type="checkbox"
                  checked={allSelected}
                  ref={(el) => {
                    if (el) el.indeterminate = someSelected
                  }}
                  onChange={handleSelectAll}
                  className="size-4 rounded-sm border-border-strong bg-transparent cursor-pointer"
                />
              </div>
              <div className="flex-1">
                <ColumnHeader
                  label={t('files.name')}
                  field="name"
                  currentField={sortField}
                  order={sortOrder}
                  onClick={handleSort}
                />
              </div>
              <div className="w-[120px] text-right">
                <ColumnHeader
                  label={t('files.size')}
                  field="size"
                  currentField={sortField}
                  order={sortOrder}
                  onClick={handleSort}
                  className="ml-auto"
                />
              </div>
              <div className="w-[160px] text-right">
                <ColumnHeader
                  label={t('files.modified')}
                  field="time"
                  currentField={sortField}
                  order={sortOrder}
                  onClick={handleSort}
                  className="ml-auto"
                />
              </div>
            </div>

            {isUploading && uploadProgress && (
              <div className="flex h-10 items-center gap-3 border-b border-border px-4">
                <div className="size-4 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <Upload className="size-4 shrink-0 text-text-3" />
                    <span className="text-sm text-text-2 truncate">
                      {uploadProgress.fileName}
                    </span>
                    <span className="shrink-0 text-sm text-text-3 tabular-nums">
                      {uploadProgress.percentage}%
                    </span>
                  </div>
                  <div className="mt-0.5 h-0.5 w-full overflow-hidden rounded-full bg-bg-layer-3">
                    <div
                      className="h-full bg-highlight transition-all duration-200"
                      style={{ width: `${uploadProgress.percentage}%` }}
                    />
                  </div>
                </div>
                <div className="hidden w-[120px] md:block" />
                <div className="hidden w-[160px] md:block" />
              </div>
            )}

            {sortedItems.map((item) => {
              const isSelected = selectedNames.has(item.name)
              const isRenaming = renamingItem === item.name
              const Icon = getFileIcon(item)
              const rowH = sortedItems.length <= 5 ? 'h-11' : 'h-11 md:h-9'

              return (
                <div
                  key={item.name}
                  onClick={() => {
                    if (item.isDir) navigateTo(`${currentDir === '/' ? '' : currentDir}/${item.name}`)
                    else if (canPreview(item)) handlePreview(item)
                    else toast.info(t('files.previewUnsupported'))
                  }}
                  onContextMenu={(e) => {
                    e.stopPropagation()
                    handleContextMenu(item, e)
                  }}
                  className={cn(
                    'flex cursor-pointer items-center gap-3 px-4 transition-colors select-none',
                    rowH,
                    isSelected
                      ? 'bg-bg-layer-3'
                      : 'hover:bg-bg-hover',
                  )}
                >
                  <div className="flex size-4 shrink-0 items-center justify-center">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => { }}
                      onClick={(e) => {
                        e.stopPropagation()
                        handleCheckboxToggle(item.name)
                      }}
                      className="size-4 rounded-sm border-border-strong bg-transparent cursor-pointer"
                    />
                  </div>

                  <Icon className={cn('size-4 shrink-0', isSelected ? 'text-highlight' : 'text-text-3')} />

                  <div className="-ml-0.5 flex min-w-0 flex-1 flex-col justify-center">
                    <div className="flex min-w-0 items-center gap-2.5">
                      {isRenaming ? (
                        <input
                          ref={renameInputRef}
                          value={renameValue}
                          onChange={(e) => setRenameValue(e.target.value)}
                          spellCheck={false}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') commitRename()
                            if (e.key === 'Escape') cancelRename()
                          }}
                          onBlur={cancelRename}
                          onClick={(e) => e.stopPropagation()}
                          className="h-6 flex-1 min-w-0 rounded-sm border border-border-strong bg-bg-layer-2 px-1.5 text-sm text-text-1 outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                        />
                      ) : (
                        <span className="text-sm text-text-1 truncate">
                          {item.name}
                          {item.isDir && '/'}
                        </span>
                      )}
                      {isProtected?.(item) && (
                        <Lock className="size-3.5 shrink-0 text-interactive" />
                      )}
                    </div>
                    {!isRenaming && (
                      <div className="mt-0.5 flex items-center gap-1.5 text-xs text-text-3 md:hidden">
                        <span className="tabular-nums">{item.isDir ? '-' : formatSize(item.size)}</span>
                        <span>·</span>
                        <span className="tabular-nums">{formatTime(item.modTime)}</span>
                      </div>
                    )}
                  </div>

                  <div className="hidden w-[120px] text-right text-sm text-text-3 tabular-nums md:block">
                    {item.isDir ? '-' : formatSize(item.size)}
                  </div>

                  <div className="hidden w-[160px] text-right text-sm text-text-3 tabular-nums md:block">
                    {formatTime(item.modTime)}
                  </div>

                  {!(item.isDir && isProtected?.(item)) && !isRenaming && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleContextMenu(item, e)
                      }}
                      aria-label={t('sidebar.menu')}
                      className="flex size-7 shrink-0 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 md:hidden"
                    >
                      <Ellipsis className="size-4" />
                    </button>
                  )}
                </div>
              )
            })}

          </>
        )}
      </div>

      {/* Bottom status bar */}
      {pageState !== 'loading' && pageState !== 'error' && items.length > 0 && (
        <div className="flex h-8 shrink-0 items-center border-t border-border px-4">
          <span className="text-xs text-text-muted tabular-nums">
            {t('files.totalCount').replace('{n}', String(items.length))}
          </span>
        </div>
      )}

      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={handleFileInputChange}
      />

      {/* Item context menu */}
      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          onClose={() => setContextMenu(null)}
        >
          {canPreview(contextMenu.item) && (
            <ContextMenuItem onClick={() => handlePreview(contextMenu.item)}>
              <Eye className="size-3.5" />
              {t('files.preview')}
            </ContextMenuItem>
          )}
          {!contextMenu.item.isDir && (
            <ContextMenuItem onClick={() => handleDownload(contextMenu.item)}>
              <Download className="size-3.5" />
              {t('files.download')}
            </ContextMenuItem>
          )}
          {!isProtected?.(contextMenu.item) && (
            <ContextMenuItem onClick={() => startRename(contextMenu.item)}>
              <Pencil className="size-3.5" />
              {t('files.rename')}
            </ContextMenuItem>
          )}
          {!isProtected?.(contextMenu.item) && (
            <ContextMenuItem
              onClick={() => {
                setContextMenu(null)
                setSelectedNames(new Set([contextMenu.item.name]))
                setTimeout(() => openDeleteDialog(), 0)
              }}
              danger
            >
              <Trash2 className="size-3.5" />
              {t('common.delete')}
            </ContextMenuItem>
          )}
        </ContextMenu>
      )}

      {/* Empty space context menu */}
      {emptyContextMenu && allowEmptyContextMenu(currentDir) && (
        <ContextMenu
          x={emptyContextMenu.x}
          y={emptyContextMenu.y}
          onClose={() => setEmptyContextMenu(null)}
        >
          <ContextMenuItem onClick={() => { setEmptyContextMenu(null); handleUploadClick() }}>
            <Upload className="size-3.5" />
            {t('files.upload')}
          </ContextMenuItem>
          <ContextMenuItem onClick={() => { setEmptyContextMenu(null); openNewFolderDialog() }}>
            <FolderPlus className="size-3.5" />
            {t('files.newFolder')}
          </ContextMenuItem>
        </ContextMenu>
      )}

      <FileDialogs
        mode={dialogMode}
        value={dialogValue}
        error={dialogError}
        selectedItems={selectedItems}
        createTitle={getCreateTitle?.(currentDir)}
        onValueChange={setDialogValue}
        onClearError={() => setDialogError(null)}
        onClose={closeDialog}
        onCreateFolder={handleCreateFolder}
        onDelete={handleDelete}
        onCompress={handleCompress}
        onDecompress={handleDecompress}
      />

      <PreviewDialog
        item={previewItem}
        type={previewType}
        url={previewUrl}
        text={previewText}
        sheets={previewSheets}
        loading={previewLoading}
        error={previewError}
        onClose={closePreview}
        onDownload={previewItem ? () => handleDownload(previewItem) : undefined}
      />
    </div>
  )
})
