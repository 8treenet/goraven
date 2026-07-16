import { useState, useCallback, useMemo, useEffect, useRef } from 'react'
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
  FolderOpen,
  AlertCircle,
  RefreshCw,
  Eye,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import {
  listTeamFiles,
  commitTeamUpload,
  teamMkdir,
  teamRename,
  teamDeleteFiles,
  teamCompress,
  teamDecompress,
  getTeamDownloadUrl,
  createTeamTempAccess,
} from '@/api/team-projects'
import type { FileItem, TeamProjectItem } from '@/api/types'
import { useChunkUpload } from '@/hooks/useChunkUpload'
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

interface TeamProjectFilesProps {
  project: TeamProjectItem
  onBack: () => void
  onBackToMine: () => void
}

export function TeamProjectFiles({ project, onBack, onBackToMine }: TeamProjectFilesProps) {
  const [pageState, setPageState] = useState<PageState>('loading')
  const [currentDir, setCurrentDir] = useState('/')
  const [items, setItems] = useState<FileItem[]>([])
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc')
  const [selectedNames, setSelectedNames] = useState<Set<string>>(new Set())
  const [lastSelected, setLastSelected] = useState<string | null>(null)
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null)
  const [renamingItem, setRenamingItem] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [isDragOver, setIsDragOver] = useState(false)
  const t = useT()

  const { upload: chunkUpload, progress: uploadProgress, isUploading } = useChunkUpload()

  const projectRootPath = useMemo(
    () => `projects/${project.projectName}`,
    [project.projectName],
  )

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
    buildDownloadUrl: (filePath) => getTeamDownloadUrl(project.id, filePath),
    createAccess: (path, type) => createTeamTempAccess(project.id, path, type),
    buildAkPath: (filePath) => `${projectRootPath}${filePath}`,
  })

  const [dialogMode, setDialogMode] = useState<FileDialogMode>(null)
  const [dialogValue, setDialogValue] = useState('')
  const [dialogError, setDialogError] = useState<string | null>(null)

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

  const isRoot = currentDir === '/'
  const currentDirName = isRoot ? project.projectName : (currentDir.split('/').pop() || null)
  const canDelete = selectedItems.length > 0
  const canCompress = selectedItems.length > 0
  const canDecompress =
    selectedItems.length === 1 &&
    !selectedItems[0].isDir &&
    selectedItems[0].name.endsWith('.zip')

  const loadDir = useCallback((dir: string) => {
    setPageState('loading')
    setSelectedNames(new Set())
    setLastSelected(null)
    setContextMenu(null)
    setRenamingItem(null)

    listTeamFiles(project.id, dir === '/' ? '' : dir)
      .then((data) => {
        setItems(data.items)
        setPageState(data.items.length === 0 ? 'empty' : 'data')
      })
      .catch(() => {
        setPageState('error')
      })
  }, [project.id])

  useEffect(() => {
    loadDir(currentDir)
  }, [currentDir, loadDir])

  const navigateTo = useCallback((dir: string) => {
    setCurrentDir(dir)
  }, [])

  const goBack = useCallback(() => {
    if (isRoot) {
      onBack()
      return
    }
    const parent = currentDir.split('/').slice(0, -1).join('/') || '/'
    setCurrentDir(parent)
  }, [currentDir, isRoot, onBack])

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

  const handleSelect = useCallback(
    (name: string, e: React.MouseEvent) => {
      setSelectedNames((prev) => {
        const next = new Set(prev)
        if (e.metaKey || e.ctrlKey) {
          if (next.has(name)) next.delete(name)
          else next.add(name)
        } else if (e.shiftKey && lastSelected) {
          const allNames = sortedItems.map((item) => item.name)
          const start = allNames.indexOf(lastSelected)
          const end = allNames.indexOf(name)
          if (start !== -1 && end !== -1) {
            const range = allNames.slice(Math.min(start, end), Math.max(start, end) + 1)
            for (const n of range) next.add(n)
          }
        } else {
          if (next.has(name) && next.size === 1) next.clear()
          else {
            next.clear()
            next.add(name)
          }
        }
        return next
      })
      setLastSelected(name)
    },
    [sortedItems, lastSelected],
  )

  const handleCheckboxToggle = useCallback((name: string) => {
    setSelectedNames((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
    setLastSelected(name)
  }, [])

  const handleSelectAll = useCallback(() => {
    if (allSelected) setSelectedNames(new Set())
    else setSelectedNames(new Set(items.map((item) => item.name)))
  }, [allSelected, items])

  const handleContextMenu = useCallback((item: FileItem, e: React.MouseEvent) => {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, item })
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
    teamRename(project.id, oldPath, newPath)
      .then(() => {
        setRenamingItem(null)
        toast.success(translate('files.renameSuccess'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.renameFailed'))
      })
  }, [renamingItem, renameValue, items, currentDir, project.id, loadDir])

  const cancelRename = useCallback(() => setRenamingItem(null), [])

  const handleUploadFiles = useCallback(
    async (files: FileList | File[]) => {
      const fileArray = Array.from(files)
      for (const file of fileArray) {
        try {
          const mergeResult = await chunkUpload(file)
          await commitTeamUpload(project.id, mergeResult.uploadId, currentDir === '/' ? '' : currentDir)
        } catch {
        }
      }
      loadDir(currentDir)
    },
    [chunkUpload, project.id, currentDir, loadDir],
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

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setIsDragOver(false)
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        handleUploadFiles(e.dataTransfer.files)
      }
    },
    [handleUploadFiles],
  )

  const openNewFolderDialog = useCallback(() => {
    setDialogMode('newFolder')
    setDialogValue('')
    setDialogError(null)
  }, [])

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
    const error = validateName(dialogValue, items.map((i) => i.name))
    if (error) {
      setDialogError(error)
      return
    }
    const dirPath = `${currentDir === '/' ? '' : currentDir}/${dialogValue.trim()}`
    teamMkdir(project.id, dirPath)
      .then(() => {
        closeDialog()
        toast.success(translate('files.dirCreated'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.dirCreateFailed'))
      })
  }, [dialogValue, items, currentDir, project.id, closeDialog, loadDir])

  const handleDelete = useCallback(() => {
    const names = selectedItems.map((i) => i.name)
    const paths = names.map((n) => `${currentDir === '/' ? '' : currentDir}/${n}`)
    teamDeleteFiles(project.id, paths)
      .then(() => {
        setSelectedNames(new Set())
        closeDialog()
        toast.success(translate('files.deletedCount').replace('{n}', String(names.length)))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [selectedItems, currentDir, project.id, closeDialog, loadDir])

  const handleCompress = useCallback(() => {
    const name = dialogValue.trim() || 'archive'
    const zipName = name.endsWith('.zip') ? name : `${name}.zip`
    const error = validateName(zipName, items.map((i) => i.name))
    if (error) {
      setDialogError(error)
      return
    }
    const paths = selectedItems.map((i) => `${currentDir === '/' ? '' : currentDir}/${i.name}`)
    const outputName = name.endsWith('.zip') ? name.replace(/\.zip$/i, '') : name
    teamCompress(project.id, { paths, outputName })
      .then(() => {
        closeDialog()
        toast.success(translate('files.compressComplete'))
        loadDir(currentDir)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.compressFailed'))
      })
  }, [dialogValue, items, selectedItems, currentDir, project.id, closeDialog, loadDir])

  const handleDecompress = useCallback(
    (toSubDir: boolean) => {
      const zipItem = selectedItems[0]
      const zipPath = `${currentDir === '/' ? '' : currentDir}/${zipItem.name}`
      teamDecompress(project.id, { path: zipPath, toSubDir })
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
    [selectedItems, currentDir, project.id, closeDialog, loadDir],
  )

  const handleRefresh = useCallback(() => {
    loadDir(currentDir)
  }, [currentDir, loadDir])

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
      if (e.key === 'F2' && selectedItems.length === 1) {
        e.preventDefault()
        startRename(selectedItems[0])
      }
      if ((e.key === 'Delete' || e.key === 'Backspace') && canDelete) {
        e.preventDefault()
        openDeleteDialog()
      }
      if (e.key === 'Escape') setContextMenu(null)
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [dialogMode, renamingItem, selectedItems, canDelete, startRename, openDeleteDialog])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <div className="flex h-10 shrink-0 items-center gap-1 border-b border-border px-3">
        <Button
          variant="ghost"
          size="icon"
          onClick={goBack}
          title={t('files.backTooltip')}
        >
          <ArrowLeft className="size-4" />
        </Button>

        {currentDirName && (
          <>
            <span className="text-sm font-semibold text-text-1 tabular-nums">
              {currentDirName}
            </span>
            <div className="mx-1 h-4 w-px bg-border" />
          </>
        )}

        <Button variant="ghost" size="default" onClick={handleUploadClick} className="text-highlight hover:text-highlight">
          <Upload className="size-4" />
          {t('files.upload')}
        </Button>
        <Button variant="ghost" size="default" onClick={openNewFolderDialog} className="text-highlight hover:text-highlight">
          <FolderPlus className="size-4" />
          {t('files.newFolder')}
        </Button>

        <div className="mx-1 h-4 w-px bg-border" />

        <Button
          variant="ghost"
          size="default"
          onClick={openDeleteDialog}
          disabled={!canDelete}
          className={cn(!canDelete ? '' : 'text-highlight hover:text-highlight')}
        >
          <Trash2 className="size-4" />
          {t('files.delete')}
        </Button>
        <Button
          variant="ghost"
          size="default"
          onClick={openCompressDialog}
          disabled={!canCompress}
          className={cn(!canCompress ? '' : 'text-highlight hover:text-highlight')}
        >
          <Archive className="size-4" />
          {t('files.compress')}
        </Button>
        <Button
          variant="ghost"
          size="default"
          onClick={openDecompressDialog}
          disabled={!canDecompress}
          className={cn(!canDecompress ? '' : 'text-highlight hover:text-highlight')}
        >
          <FolderArchive className="size-4" />
          {t('files.decompress')}
        </Button>

        <div className="flex-1" />

        <Button
          variant="ghost"
          size="default"
          onClick={onBackToMine}
          className="text-highlight hover:text-highlight"
        >
          <FolderOpen className="size-4" />
          {t('files.myFiles')}
        </Button>

        <Button
          variant="ghost"
          size="icon"
          onClick={handleRefresh}
          title={t('common.refresh')}
          className="text-highlight hover:text-highlight"
        >
          <RefreshCw className="size-4" />
        </Button>

        <input
          ref={fileInputRef}
          type="file"
          multiple
          className="hidden"
          onChange={handleFileInputChange}
        />
      </div>

      <div
        className="flex-1 overflow-auto"
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        {pageState === 'loading' && (
          <div>
            {[1, 2, 3, 4, 5, 6, 7].map((i) => (
              <div
                key={i}
                className="flex h-10 items-center gap-3 border-b border-border px-4"
              >
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
              <p className="text-sm text-text-2">{t('files.projectNotFound')}</p>
            </div>
            <button
              onClick={onBack}
              className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
            >
              <ArrowLeft className="size-4" />
              {t('files.teamProjects')}
            </button>
          </div>
        )}

        {pageState === 'empty' && !isDragOver && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <FolderOpen className="size-12 text-text-3" />
            <div className="text-center">
              <p className="text-sm text-text-2">{t('files.emptyFolder')}</p>
              <p className="mt-1 text-sm text-text-3">{t('files.emptyHint')}</p>
            </div>
          </div>
        )}

        {(pageState === 'data' || (pageState === 'empty' && isDragOver)) && (
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

            {sortedItems.map((item, i) => {
              const isSelected = selectedNames.has(item.name)
              const isRenaming = renamingItem === item.name
              const Icon = getFileIcon(item)
              const rowH = sortedItems.length <= 5 ? 'h-11' : 'h-9'

              return (
                <div
                  key={item.name}
                  onClick={(e) => handleSelect(item.name, e)}
                  onDoubleClick={() => {
                    if (item.isDir) navigateTo(`${currentDir === '/' ? '' : currentDir}/${item.name}`)
                  }}
                  onContextMenu={(e) => handleContextMenu(item, e)}
                  className={cn(
                    'flex cursor-default items-center gap-3 px-4 transition-colors select-none',
                    rowH,
                    isSelected
                      ? 'bg-bg-layer-3'
                      : i % 2 === 0
                        ? 'bg-interactive/10 dark:bg-bg-layer-1 hover:bg-bg-hover'
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

            {isDragOver && (
              <div className="flex h-20 items-center justify-center border-2 border-dashed border-highlight mx-3 my-2 rounded-lg">
                <p className="text-sm text-highlight">{t('files.dropHint')}</p>
              </div>
            )}

            {!isDragOver && sortedItems.length < 8 && sortedItems.length > 0 && (
              <div className="px-4 py-4 text-xs text-text-muted">
                <span className="text-text-3">{t('files.dragPrompt')}</span>{' '}
                <span className="text-text-muted">| {t('files.orClick')}</span>
              </div>
            )}
          </>
        )}
      </div>

      {items.length >= 10 && (
        <div className="flex h-8 shrink-0 items-center border-t border-border px-4">
          <span className="text-sm text-text-3 tabular-nums">
            {selectedNames.size > 0
              ? t('files.selectedCount').replace('{n}', String(selectedNames.size)).replace('{total}', String(items.length))
              : t('files.totalCount').replace('{n}', String(items.length))}
          </span>
        </div>
      )}

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
          <ContextMenuItem onClick={() => startRename(contextMenu.item)}>
            <Pencil className="size-3.5" />
            {t('files.rename')}
          </ContextMenuItem>
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
        </ContextMenu>
      )}

      <FileDialogs
        mode={dialogMode}
        value={dialogValue}
        error={dialogError}
        selectedItems={selectedItems}
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
}
