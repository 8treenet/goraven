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
  Lock,
  FolderOpen,
  AlertCircle,
  RefreshCw,
  Braces,
  Eye,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { listFiles, mkdir, rename, deleteFiles, compress, decompress, getDownloadUrl, createTempAccess, getAkDownloadUrl } from '@/api/files'
import type { FileItem } from '@/api/types'
import { useFileUpload } from '@/hooks/useFileUpload'
import { useUserStore } from '@/stores/user-store'
import {
  ContextMenu,
  ContextMenuItem,
  ColumnHeader,
} from './FileContextMenu'
import { FileDialogs, type FileDialogMode } from './FileDialogs'
import { PreviewDialog } from './PreviewDialog'
import { EnvVarsDialog } from './EnvVarsDialog'
import {
  canPreview,
  formatSize,
  formatTime,
  getFileIcon,
  getPreviewType,
  MAX_TEXT_PREVIEW_SIZE,
  sortItems,
  validateName,
  type ContextMenuState,
  type PageState,
  type PreviewType,
  type SheetData,
  type SortField,
  type SortOrder,
} from './file-helpers'
import { read as xlsxRead, utils as xlsxUtils } from 'xlsx'

export function Component() {
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

  const { upload, progress: uploadProgress, isUploading } = useFileUpload()

  const [dialogMode, setDialogMode] = useState<FileDialogMode>(null)
  const [dialogValue, setDialogValue] = useState('')
  const [dialogError, setDialogError] = useState<string | null>(null)

  const [envVarsOpen, setEnvVarsOpen] = useState(false)

  const [previewItem, setPreviewItem] = useState<FileItem | null>(null)
  const [previewType, setPreviewType] = useState<PreviewType | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [previewText, setPreviewText] = useState<string | null>(null)
  const [previewSheets, setPreviewSheets] = useState<SheetData[] | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState(false)

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
  const currentDirName = isRoot ? null : currentDir.split('/').pop() || null

  const canDelete = selectedItems.length > 0 && !selectedItems.some((item) => item.isDefault)
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

    listFiles(dir === '/' ? '' : dir)
      .then((data) => {
        setItems(data.items)
        setPageState(data.items.length === 0 ? 'empty' : 'data')
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

  const handleSelect = useCallback(
    (name: string, e: React.MouseEvent) => {
      setSelectedNames((prev) => {
        const next = new Set(prev)
        if (e.metaKey || e.ctrlKey) {
          if (next.has(name)) {
            next.delete(name)
          } else {
            next.add(name)
          }
        } else if (e.shiftKey && lastSelected) {
          const allNames = sortedItems.map((item) => item.name)
          const start = allNames.indexOf(lastSelected)
          const end = allNames.indexOf(name)
          if (start !== -1 && end !== -1) {
            const range = allNames.slice(Math.min(start, end), Math.max(start, end) + 1)
            for (const n of range) next.add(n)
          }
        } else {
          if (next.has(name) && next.size === 1) {
            next.clear()
          } else {
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
      if (next.has(name)) {
        next.delete(name)
      } else {
        next.add(name)
      }
      return next
    })
    setLastSelected(name)
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
    const filePath = `${currentDir === '/' ? '' : currentDir}/${item.name}`
    const url = getDownloadUrl(filePath)
    const token = useUserStore.getState().token
    fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then((res) => {
        if (!res.ok) throw new Error('Download failed')
        return res.blob()
      })
      .then((blob) => {
        const blobUrl = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = blobUrl
        a.download = item.name
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(blobUrl)
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('files.downloadFailed'))
      })
  }, [currentDir])

  const handlePreview = useCallback((item: FileItem) => {
    setContextMenu(null)
    setPreviewItem(item)
    setPreviewType(getPreviewType(item))
    setPreviewUrl(null)
    setPreviewText(null)
    setPreviewSheets(null)
    setPreviewError(false)
    setPreviewLoading(true)
    const filePath = `${currentDir === '/' ? '' : currentDir}/${item.name}`
    const url = getDownloadUrl(filePath)
    const token = useUserStore.getState().token

    if (getPreviewType(item) === 'text' || getPreviewType(item) === 'markdown') {
      fetch(url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
        .then((res) => {
          if (!res.ok) throw new Error('preview failed')
          return res.text()
        })
        .then((text) => {
          setPreviewText(text.length > MAX_TEXT_PREVIEW_SIZE ? text.slice(0, MAX_TEXT_PREVIEW_SIZE) + '\n\n... File too large, preview truncated' : text)
        })
        .catch(() => {
          setPreviewError(true)
        })
        .finally(() => {
          setPreviewLoading(false)
        })
    } else if (getPreviewType(item) === 'xlsx') {
      fetch(url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
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
        .catch(() => {
          setPreviewError(true)
        })
        .finally(() => {
          setPreviewLoading(false)
        })
    } else if (getPreviewType(item) === 'html') {
      createTempAccess(currentDir === '/' ? '/' : currentDir, 'dir')
        .then((res) => {
          setPreviewUrl(getAkDownloadUrl(res.ak, filePath))
        })
        .catch(() => {
          setPreviewError(true)
        })
        .finally(() => {
          setPreviewLoading(false)
        })
    } else if (getPreviewType(item) === 'office') {
      createTempAccess(filePath, 'file')
        .then((res) => {
          const absUrl = `${window.location.origin}/api/hfs/ak/${res.ak}/${filePath.replace(/^\//, '').split('/').map(encodeURIComponent).join('/')}`
          setPreviewUrl(`https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(absUrl)}`)
        })
        .catch(() => {
          setPreviewError(true)
        })
        .finally(() => {
          setPreviewLoading(false)
        })
    } else {
      fetch(url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
        .then((res) => {
          if (!res.ok) throw new Error('preview failed')
          return res.blob()
        })
        .then((blob) => {
          setPreviewUrl(URL.createObjectURL(blob))
        })
        .catch(() => {
          setPreviewError(true)
        })
        .finally(() => {
          setPreviewLoading(false)
        })
    }
  }, [currentDir])

  const closePreview = useCallback(() => {
    setPreviewItem(null)
    setPreviewType(null)
    setPreviewError(false)
    setPreviewLoading(false)
    setPreviewText(null)
    setPreviewSheets(null)
    setPreviewUrl((url) => {
      if (url) URL.revokeObjectURL(url)
      return null
    })
  }, [])

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
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [dialogMode, renamingItem, selectedItems, canDelete, startRename, openDeleteDialog])

  const handleRefresh = useCallback(() => {
    loadDir(currentDir)
  }, [currentDir, loadDir])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <div className="flex h-10 shrink-0 items-center gap-1 border-b border-border px-3">
        <Button
          variant="ghost"
          size="icon"
          onClick={goBack}
          disabled={isRoot}
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
          onClick={() => setEnvVarsOpen(true)}
          className="text-highlight hover:text-highlight"
        >
          <Braces className="size-4" />
          {t('files.envVars')}
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
              <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
              <p className="mt-1 text-sm text-text-3">
                {t('files.errReadDir')}
              </p>
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

        {pageState === 'empty' && !isDragOver && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <FolderOpen className="size-12 text-text-3" />
            <div className="text-center">
               <p className="text-sm text-text-2">{t('files.emptyFolder')}</p>
               <p className="mt-1 text-sm text-text-3">
                 {t('files.emptyHint')}
               </p>
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
                        ? 'bg-bg-layer-1 hover:bg-bg-hover'
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

            {isDragOver && (
              <div className="flex h-20 items-center justify-center border-2 border-dashed border-highlight mx-3 my-2 rounded-lg">
                <p className="text-sm text-highlight">
                  {t('files.dropHint')}
                </p>
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

      <EnvVarsDialog
        open={envVarsOpen}
        onOpenChange={setEnvVarsOpen}
      />
    </div>
  )
}
