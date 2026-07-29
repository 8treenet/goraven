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
  Plus,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { listFiles, mkdir, rename, deleteFiles, compress, decompress } from '@/api/files'
import { createTempAccess, getDownloadUrl } from '@/api/files'
import type { FileItem, TeamProjectItem } from '@/api/types'
import { useFileUpload } from '@/hooks/useFileUpload'
import { evictCachedBlob } from '@/lib/file-blob-cache'
import {
  ContextMenu,
  ContextMenuItem,
  ColumnHeader,
} from './FileContextMenu'
import { FileDialogs, type FileDialogMode } from './FileDialogs'
import { PreviewDialog } from './PreviewDialog'
import { TeamProjectView, type TeamProjectViewHandle } from './TeamProjectView'
import { TeamProjectFiles, type TeamProjectFilesHandle } from './TeamProjectFiles'
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

export function Component() {
  const [tab, setTab] = useState<'files' | 'projects' | 'team'>('files')
  const [activeTeamProject, setActiveTeamProject] = useState<TeamProjectItem | null>(null)
  const [plusOpen, setPlusOpen] = useState(false)
  const plusRef = useRef<HTMLDivElement>(null)
  const mineRef = useRef<MineFilesHandle>(null)
  const teamFilesRef = useRef<TeamProjectFilesHandle>(null)
  const teamViewRef = useRef<TeamProjectViewHandle>(null)
  const t = useT()

  const switchTab = useCallback((newTab: 'files' | 'projects' | 'team') => {
    setTab(newTab)
    setPlusOpen(false)
    if (newTab !== 'team') setActiveTeamProject(null)
  }, [])

  useEffect(() => {
    if (!plusOpen) return
    const handler = (e: MouseEvent) => {
      if (plusRef.current && !plusRef.current.contains(e.target as Node)) setPlusOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [plusOpen])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Tab navigation bar */}
      <div className="flex h-10 shrink-0 items-center gap-1 border-b border-border px-3">
        <div className="flex items-center gap-0">
          {([
            ['files', t('files.myFiles')],
            ['projects', t('files.myProjects')],
            ['team', t('files.teamProjects')],
          ] as const).map(([key, label]) => (
            <button
              key={key}
              onClick={() => switchTab(key)}
              className={cn(
                'relative px-3 py-1 text-sm transition-colors',
                tab === key
                  ? 'text-text-1 font-medium after:absolute after:bottom-[-9px] after:left-0 after:right-0 after:h-[2px] after:bg-highlight after:rounded-full'
                  : 'text-text-3 hover:text-text-2',
              )}
            >
              {label}
            </button>
          ))}
        </div>

        <div className="flex-1" />

        {tab === 'files' && (
          <div className="relative" ref={plusRef}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setPlusOpen((v) => !v)}
              title={t('files.newFolderTitle')}
              className="text-highlight hover:text-highlight/80"
            >
              <Plus className="size-4" />
            </Button>
            {plusOpen && (
              <div className="absolute right-0 top-full z-50 mt-1 min-w-[140px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop">
                <button
                  onClick={() => { setPlusOpen(false); mineRef.current?.upload() }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                >
                  <Upload className="size-3.5" />
                  {t('files.upload')}
                </button>
                <button
                  onClick={() => { setPlusOpen(false); mineRef.current?.newFolder() }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                >
                  <FolderPlus className="size-3.5" />
                  {t('files.newFolderTitle')}
                </button>
              </div>
            )}
          </div>
        )}

        {tab === 'projects' && (
          <div className="relative" ref={plusRef}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => { setPlusOpen(false); mineRef.current?.newFolder() }}
              title={t('files.newProject')}
              className="text-highlight hover:text-highlight/80"
            >
              <Plus className="size-4" />
            </Button>
          </div>
        )}

        {tab === 'team' && !activeTeamProject && (
          <div className="relative" ref={plusRef}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => teamViewRef.current?.createProject()}
              title={t('files.newTeamProject')}
              className="text-highlight hover:text-highlight/80"
            >
              <Plus className="size-4" />
            </Button>
          </div>
        )}

        {tab === 'team' && activeTeamProject && (
          <div className="relative" ref={plusRef}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setPlusOpen((v) => !v)}
              title={t('files.newFolderTitle')}
              className="text-highlight hover:text-highlight/80"
            >
              <Plus className="size-4" />
            </Button>
            {plusOpen && (
              <div className="absolute right-0 top-full z-50 mt-1 min-w-[140px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop">
                <button
                  onClick={() => { setPlusOpen(false); teamFilesRef.current?.upload() }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                >
                  <Upload className="size-3.5" />
                  {t('files.upload')}
                </button>
                <button
                  onClick={() => { setPlusOpen(false); teamFilesRef.current?.newFolder() }}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                >
                  <FolderPlus className="size-3.5" />
                  {t('files.newFolderTitle')}
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Content area */}
      <div className="flex-1 min-h-0">
        {tab === 'files' && (
          <MineFiles ref={mineRef} key="root" initialDir="/" />
        )}
        {tab === 'projects' && (
          <MineFiles ref={mineRef} key="projects" initialDir="/projects" rootDir="/projects" />
        )}
        {tab === 'team' && !activeTeamProject && (
          <TeamProjectView
            ref={teamViewRef}
            onEnterProject={setActiveTeamProject}
          />
        )}
        {tab === 'team' && activeTeamProject && (
          <TeamProjectFiles
            ref={teamFilesRef}
            key={activeTeamProject.id}
            project={activeTeamProject}
            onBack={() => setActiveTeamProject(null)}
          />
        )}
      </div>
    </div>
  )
}

interface MineFilesHandle {
  upload: () => void
  newFolder: () => void
}

const MineFiles = forwardRef<MineFilesHandle, { initialDir?: string; rootDir?: string }>(function MineFiles({ initialDir = '/', rootDir = '/' }, ref) {
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

  const { upload, progress: uploadProgress, isUploading } = useFileUpload()

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
    buildDownloadUrl: getDownloadUrl,
    createAccess: createTempAccess,
    buildAkPath: (filePath) => filePath,
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
  const currentDirName = isRoot ? null : currentDir.split('/').pop() || null
  const inProjectsDir = currentDir === '/projects'

  const canDelete = selectedItems.length > 0 && !selectedItems.some((item) => item.isDefault)
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

    listFiles(dir === '/' ? '' : dir)
      .then((data) => {
        const filtered = dir === '/'
          ? data.items.filter((item) => !(item.isDir && item.name === 'projects'))
          : data.items
        setItems(filtered)
        setPageState(filtered.length === 0 ? 'empty' : 'data')
      })
      .catch(() => {
        setPageState('error')
      })
  }, [])

  useEffect(() => {
    loadDir(currentDir)
  }, [currentDir, loadDir])

  const navigateTo = useCallback((dir: string) => {
    setCurrentDir(dir)
  }, [])

  const goBack = useCallback(() => {
    if (isRoot) return
    const parent = currentDir.split('/').slice(0, -1).join('/') || '/'
    setCurrentDir(parent)
  }, [currentDir, isRoot])

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
    if (item.isDir && item.isDefault) return
    setContextMenu({ x: e.clientX, y: e.clientY, item })
  }, [])

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
    rename(oldPath, newPath)
      .then(() => {
        evictCachedBlob(getDownloadUrl(oldPath))
        setRenamingItem(null)
        toast.success(translate('files.renameSuccess'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.renameFailed'))
      })
  }, [renamingItem, renameValue, items, currentDir, loadDir])

  const cancelRename = useCallback(() => {
    setRenamingItem(null)
  }, [])

  const handleUploadFiles = useCallback(
    async (files: FileList | File[]) => {
      const fileArray = Array.from(files)
      for (const file of fileArray) {
        try {
          await upload(file, currentDir === '/' ? '' : currentDir)
          const filePath = `${currentDir === '/' ? '' : currentDir}/${file.name}`
          evictCachedBlob(getDownloadUrl(filePath))
        } catch {
        }
      }
      loadDir(currentDir)
    },
    [upload, currentDir, loadDir],
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
    mkdir(dirPath)
      .then(() => {
        closeDialog()
        toast.success(translate('files.dirCreated'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.dirCreateFailed'))
      })
  }, [dialogValue, items, currentDir, closeDialog, loadDir])

  const handleDelete = useCallback(() => {
    const names = selectedItems.map((i) => i.name)
    const paths = names.map((n) => `${currentDir === '/' ? '' : currentDir}/${n}`)
    deleteFiles(paths)
      .then(() => {
        for (const p of paths) {
          evictCachedBlob(getDownloadUrl(p))
        }
        setSelectedNames(new Set())
        closeDialog()
        toast.success(translate('files.deletedCount').replace('{n}', String(names.length)))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [selectedItems, currentDir, closeDialog, loadDir])

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
    compress({ paths, outputName })
      .then(() => {
        closeDialog()
        toast.success(translate('files.compressComplete'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.compressFailed'))
      })
  }, [dialogValue, items, selectedItems, currentDir, closeDialog, loadDir])

  const handleDecompress = useCallback(
    (toSubDir: boolean) => {
      const zipItem = selectedItems[0]
      const zipPath = `${currentDir === '/' ? '' : currentDir}/${zipItem.name}`
      decompress({ path: zipPath, toSubDir })
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
    [selectedItems, currentDir, closeDialog, loadDir],
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

      if (e.key === 'F2' && selectedItems.length === 1 && !selectedItems[0].isDefault) {
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
  }, [dialogMode, renamingItem, selectedItems, canDelete, startRename, openDeleteDialog])

  const hasSelection = selectedNames.size > 0

  return (
    <div className="flex h-full flex-col">
      {/* Back / directory indicator */}
      {!isRoot && !hasSelection && (
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
            <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" onClick={() => handleDownload(selectedItems[0])}>
              <Download className="size-3.5" />
              {t('files.download')}
            </Button>
          )}
          <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" disabled={!canDelete} onClick={openDeleteDialog}>
            <Trash2 className="size-3.5" />
            {t('files.delete')}
          </Button>
          <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" onClick={openCompressDialog}>
            <Archive className="size-3.5" />
            {t('files.compress')}
          </Button>
          {canDecompress && (
            <Button variant="ghost" size="default" className="h-6 px-2 text-xs text-text-2 hover:text-text-1" onClick={openDecompressDialog}>
              <FolderArchive className="size-3.5" />
              {t('files.decompress')}
            </Button>
          )}
          <div className="flex-1" />
          <Button variant="ghost" size="icon" className="size-6 text-text-3 hover:text-text-1" onClick={() => setSelectedNames(new Set())}>
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
              <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
              <p className="mt-1 text-sm text-text-3">{t('files.errReadDir')}</p>
            </div>
            <button
              onClick={() => loadDir(currentDir)}
              className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
            >
              <RefreshCw className="size-4" />
              {t('common.retry')}
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
            <div className="flex h-8 items-center gap-3 border-b border-border px-4">
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
                <div className="w-[120px]" />
                <div className="w-[160px]" />
              </div>
            )}

            {sortedItems.map((item) => {
              const isSelected = selectedNames.has(item.name)
              const isRenaming = renamingItem === item.name
              const Icon = getFileIcon(item)
              const rowH = sortedItems.length <= 5 ? 'h-11' : 'h-9'

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

                  <div className="flex flex-1 items-center gap-2.5 min-w-0">
                    <Icon className={cn('size-4 shrink-0', isSelected ? 'text-highlight' : 'text-text-3')} />
                    {isRenaming ? (
                      <input
                        ref={renameInputRef}
                        value={renameValue}
                        onChange={(e) => setRenameValue(e.target.value)}
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
                    {item.isDefault && (
                      <Lock className="size-3.5 shrink-0 text-interactive" />
                    )}
                  </div>

                  <div className="w-[120px] text-right text-sm text-text-3 tabular-nums">
                    {item.isDir ? '--' : formatSize(item.size)}
                  </div>

                  <div className="w-[160px] text-right text-sm text-text-3 tabular-nums">
                    {formatTime(item.modTime)}
                  </div>
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
          {!contextMenu.item.isDefault && (
            <ContextMenuItem onClick={() => startRename(contextMenu.item)}>
              <Pencil className="size-3.5" />
              {t('files.rename')}
            </ContextMenuItem>
          )}
          {!contextMenu.item.isDefault && (
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
      {emptyContextMenu && !inProjectsDir && (
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
        createTitle={inProjectsDir ? t('files.newProject') : undefined}
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
