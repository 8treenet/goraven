import { useState, useCallback, useMemo, useEffect, useRef, forwardRef } from 'react'
import {
  Upload,
  FolderPlus,
  Plus,
  Menu,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSidebarStore } from '@/stores/sidebar-store'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'
import { listFiles, mkdir, rename, deleteFiles, compress, decompress } from '@/api/files'
import { createTempAccess, getDownloadUrl } from '@/api/files'
import type { FileItem, TeamProjectItem } from '@/api/types'
import { useFileUpload } from '@/hooks/useFileUpload'
import { evictCachedBlob } from '@/lib/file-blob-cache'
import { FileList, type FileListApi } from './FileList'
import { TeamProjectView, type TeamProjectViewHandle } from './TeamProjectView'
import { TeamProjectFiles, type TeamProjectFilesHandle } from './TeamProjectFiles'

export function Component() {
  const [tab, setTab] = useState<'files' | 'projects' | 'team'>('files')
  const [activeTeamProject, setActiveTeamProject] = useState<TeamProjectItem | null>(null)
  const [plusOpen, setPlusOpen] = useState(false)
  const plusRef = useRef<HTMLDivElement>(null)
  const mineRef = useRef<MineFilesHandle>(null)
  const teamFilesRef = useRef<TeamProjectFilesHandle>(null)
  const teamViewRef = useRef<TeamProjectViewHandle>(null)
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)

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
      <div className="flex h-10 shrink-0 items-center gap-1 border-b border-border px-2 md:px-3">
        <button
          onClick={openMobile}
          aria-label={t('sidebar.menu')}
          className="shrink-0 text-text-1 transition-colors md:hidden"
        >
          <Menu className="size-4" />
        </button>
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
                'relative px-2 py-1 text-[13px] transition-colors md:px-3 md:text-sm',
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
  const t = useT()
  const { upload, progress: uploadProgress, isUploading } = useFileUpload()

  const api = useMemo<FileListApi>(() => ({
    list: (dir) => listFiles(dir),
    mkdir: (dir) => mkdir(dir),
    rename: (oldPath, newPath) => rename(oldPath, newPath),
    remove: (paths) => deleteFiles(paths),
    compress: (req) => compress(req),
    decompress: (req) => decompress(req),
    downloadUrl: (path) => getDownloadUrl(path),
    createAccess: (path, type) => createTempAccess(path, type),
  }), [])

  const uploadFile = useCallback((file: File, dir: string) => upload(file, dir), [upload])
  const evictCache = useCallback((path: string) => evictCachedBlob(getDownloadUrl(path)), [])
  const isProtected = useCallback((item: FileItem) => !!item.isDefault, [])
  const filterAtRoot = useCallback((item: FileItem) => !(item.isDir && item.name === 'projects'), [])
  const getCreateTitle = useCallback((dir: string) => (dir === '/projects' ? t('files.newProject') : undefined), [t])
  const allowEmptyContextMenu = useCallback((dir: string) => dir !== '/projects', [])

  return (
    <FileList
      ref={ref}
      initialDir={initialDir}
      rootDir={rootDir}
      api={api}
      uploadFile={uploadFile}
      uploadProgress={uploadProgress}
      isUploading={isUploading}
      evictCache={evictCache}
      isProtected={isProtected}
      filterAtRoot={filterAtRoot}
      getCreateTitle={getCreateTitle}
      allowEmptyContextMenu={allowEmptyContextMenu}
      errorHint={t('files.errReadDir')}
    />
  )
})
