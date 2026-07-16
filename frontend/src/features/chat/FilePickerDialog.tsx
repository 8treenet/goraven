import { useState, useCallback, useEffect, useMemo } from 'react'
import {
  AtSign,
  ArrowLeft,
  RefreshCw,
  Users,
  FolderOpen,
  Loader2,
  AlertCircle,
  ChevronRight,
  Folder,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { listFiles, getFileUrl } from '@/api/files'
import {
  listTeamProjects,
  listTeamFiles,
  getTeamDownloadUrl,
} from '@/api/team-projects'
import type { FileItem, TeamProjectItem } from '@/api/types'
import { getFileIcon, formatSize } from '../files/file-helpers'

export interface PickedFile {
  /** Source of the file: 'mine' or 'team' */
  source: 'mine' | 'team'
  /** Display name */
  name: string
  /** Full path from the user space root */
  path: string
  /** File size in bytes (0 for directories) */
  size: number
  /** Whether the picked entry is a directory */
  isDir: boolean
  /** For team files, the shared project this file belongs to */
  teamProjectId?: number
  /** Final URL the frontend can fetch to stream/download the file */
  url: string
}

export interface FilePickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Called when the user picks a file/dir. */
  onConfirm?: (file: PickedFile | null) => void
}

type Tab = 'mine' | 'team'

function buildUrl(source: 'mine' | 'team', path: string, teamProjectId?: number): string {
  if (source === 'mine') return getFileUrl(path)
  if (teamProjectId !== undefined) return getTeamDownloadUrl(teamProjectId, path)
  return path
}

/** Join a directory path and a name into a full path. Leading slash included. */
function joinPath(dir: string, name: string): string {
  if (dir === '/' || dir === '') return `/${name}`
  return `${dir}${dir.endsWith('/') ? '' : '/'}${name}`
}

export function FilePickerDialog({
  open,
  onOpenChange,
  onConfirm,
}: FilePickerDialogProps) {
  const t = useT()
  const [tab, setTab] = useState<Tab>('mine')
  const [teamProject, setTeamProject] = useState<TeamProjectItem | null>(null)
  const [mineBreadcrumbs, setMineBreadcrumbs] = useState<string[]>([])
  const [teamBreadcrumbs, setTeamBreadcrumbs] = useState<string[]>([])

  const reset = useCallback(() => {
    setTab('mine')
    setTeamProject(null)
    setMineBreadcrumbs([])
    setTeamBreadcrumbs([])
  }, [])

  // ---------- Mine files ----------
  const [mineItems, setMineItems] = useState<FileItem[]>([])
  const [mineLoading, setMineLoading] = useState(false)
  const [mineError, setMineError] = useState(false)

  const mineDir = useMemo(
    () => (mineBreadcrumbs.length === 0 ? '/' : mineBreadcrumbs.join('/')),
    [mineBreadcrumbs],
  )

  const loadMine = useCallback((dir: string) => {
    setMineLoading(true)
    setMineError(false)
    listFiles(dir === '/' ? '' : dir)
      .then((data) => {
        setMineItems(data.items)
        setMineLoading(false)
      })
      .catch(() => {
        setMineError(true)
        setMineLoading(false)
      })
  }, [])

  // ---------- Team projects ----------
  const [teamProjects, setTeamProjects] = useState<TeamProjectItem[]>([])
  const [teamProjectsLoading, setTeamProjectsLoading] = useState(false)
  const [teamProjectsError, setTeamProjectsError] = useState(false)

  const loadTeamProjects = useCallback(() => {
    setTeamProjectsLoading(true)
    setTeamProjectsError(false)
    listTeamProjects()
      .then((data) => {
        setTeamProjects((data.items ?? []).filter((item) => !item.isOwner))
        setTeamProjectsLoading(false)
      })
      .catch(() => {
        setTeamProjectsError(true)
        setTeamProjectsLoading(false)
      })
  }, [])

  /** Current directory system absolute path (derived from first file item) */
  const mineDirAbs = useMemo(() => {
    const first = mineItems.find((item) => item.path)
    if (!first) return null
    const segs = first.path.split('/')
    segs.pop()
    return segs.join('/') || '/'
  }, [mineItems])

  // ---------- Team project files ----------
  const [teamItems, setTeamItems] = useState<FileItem[]>([])
  const [teamLoading, setTeamLoading] = useState(false)
  const [teamError, setTeamError] = useState(false)

  const teamDir = useMemo(
    () => (teamBreadcrumbs.length === 0 ? '/' : teamBreadcrumbs.join('/')),
    [teamBreadcrumbs],
  )

  const teamDirAbs = useMemo(() => {
    const first = teamItems.find((item) => item.path)
    if (!first) return null
    const segs = first.path.split('/')
    segs.pop()
    return segs.join('/') || '/'
  }, [teamItems])

  const loadTeamFiles = useCallback((projectId: number, dir: string) => {
    setTeamLoading(true)
    setTeamError(false)
    listTeamFiles(projectId, dir === '/' ? '' : dir)
      .then((data) => {
        setTeamItems(data.items)
        setTeamLoading(false)
      })
      .catch(() => {
        setTeamError(true)
        setTeamLoading(false)
      })
  }, [])

  // ---------- Load on open ----------
  useEffect(() => {
    if (!open) return
    reset()
    loadMine('/')
    loadTeamProjects()
  }, [open, reset, loadMine, loadTeamProjects])

  // ---------- Navigation: mine ----------
  const enterMineDir = useCallback(
    (name: string) => {
      setMineBreadcrumbs((prev) => [...prev, name])
      loadMine(joinPath(mineDir, name))
    },
    [mineDir, loadMine],
  )

  const navigateMineTo = useCallback(
    (index: number) => {
      const next = index < 0 ? [] : mineBreadcrumbs.slice(0, index + 1)
      setMineBreadcrumbs(next)
      loadMine(next.length === 0 ? '/' : next.join('/'))
    },
    [mineBreadcrumbs, loadMine],
  )

  // ---------- Navigation: team ----------
  const enterTeamProject = useCallback(
    (project: TeamProjectItem) => {
      setTeamProject(project)
      setTeamBreadcrumbs([])
      loadTeamFiles(project.id, '/')
    },
    [loadTeamFiles],
  )

  const exitTeamProject = useCallback(() => {
    setTeamProject(null)
    setTeamBreadcrumbs([])
    setTeamItems([])
  }, [])

  const enterTeamDir = useCallback(
    (name: string) => {
      setTeamBreadcrumbs((prev) => [...prev, name])
      if (teamProject) loadTeamFiles(teamProject.id, joinPath(teamDir, name))
    },
    [teamDir, teamProject, loadTeamFiles],
  )

  const navigateTeamTo = useCallback(
    (index: number) => {
      const next = index < 0 ? [] : teamBreadcrumbs.slice(0, index + 1)
      setTeamBreadcrumbs(next)
      if (teamProject)
        loadTeamFiles(teamProject.id, next.length === 0 ? '/' : next.join('/'))
    },
    [teamBreadcrumbs, teamProject, loadTeamFiles],
  )

  // ---------- Pick file: immediate confirm-close ----------
  const pickFile = useCallback(
    (file: PickedFile) => {
      onConfirm?.(file)
      onOpenChange(false)
    },
    [onConfirm, onOpenChange],
  )

  // ---------- Pick current directory (confirm button) ----------
  const pickCurrentMineDir = useCallback(() => {
    const name =
      mineBreadcrumbs.length > 0 ? mineBreadcrumbs[mineBreadcrumbs.length - 1] : t('files.myFiles')
    pickFile({
      source: 'mine',
      name,
      path: mineDirAbs || mineDir,
      size: 0,
      isDir: true,
      url: buildUrl('mine', mineDirAbs || mineDir),
    })
  }, [mineBreadcrumbs, mineDir, mineDirAbs, t, pickFile])

  const pickCurrentTeamDir = useCallback(() => {
    if (!teamProject) return
    const name =
      teamBreadcrumbs.length > 0
        ? teamBreadcrumbs[teamBreadcrumbs.length - 1]
        : teamProject.projectName
    pickFile({
      source: 'team',
      name,
      path: teamDirAbs || teamDir,
      size: 0,
      isDir: true,
      teamProjectId: teamProject.id,
      url: buildUrl('team', teamDirAbs || teamDir, teamProject.id),
    })
  }, [teamProject, teamBreadcrumbs, teamDir, teamDirAbs, pickFile])

  const handleConfirm = useCallback(() => {
    if (tab === 'mine') pickCurrentMineDir()
    else if (teamProject) pickCurrentTeamDir()
  }, [tab, pickCurrentMineDir, pickCurrentTeamDir])

  // Confirm button: mine root (/) is not selectable; everything else is.
  const canConfirmDir =
    tab === 'mine'
      ? mineBreadcrumbs.length > 0
      : !!teamProject

  // Current directory display name for footer
  const currentDirLabel =
    tab === 'mine'
      ? mineBreadcrumbs.length > 0
        ? mineBreadcrumbs[mineBreadcrumbs.length - 1]
        : t('files.myFiles')
      : teamProject
        ? teamBreadcrumbs.length > 0
          ? teamBreadcrumbs[teamBreadcrumbs.length - 1]
          : teamProject.projectName
        : t('files.teamProjects')

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="!max-w-xl p-0 gap-0 overflow-hidden" hideClose>
        <DialogTitle className="sr-only">{t('chat.filePickerTitle')}</DialogTitle>
        <DialogDescription className="sr-only">{t('chat.filePickerDesc')}</DialogDescription>

        {/* Header: title + tabs */}
        <div className="flex items-center gap-2 border-b border-border px-4 py-3">
          <AtSign className="size-4 text-highlight" />
          <span className="text-sm font-semibold text-text-1">{t('chat.filePickerTitle')}</span>
          <div className="ml-auto flex items-center gap-1 rounded-md border border-border bg-bg-layer-2 p-0.5">
            <button
              onClick={() => {
                setTab('mine')
                exitTeamProject()
              }}
              className={cn(
                'rounded px-2.5 py-1 text-xs transition-colors',
                tab === 'mine' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3 hover:text-text-2',
              )}
            >
              <span className="inline-flex items-center gap-1">
                <FolderOpen className="size-3.5" />
                {t('files.myFiles')}
              </span>
            </button>
            <button
              onClick={() => {
                setTab('team')
                exitTeamProject()
              }}
              className={cn(
                'rounded px-2.5 py-1 text-xs transition-colors',
                tab === 'team' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3 hover:text-text-2',
              )}
            >
              <span className="inline-flex items-center gap-1">
                <Users className="size-3.5" />
                {t('files.teamProjects')}
              </span>
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="flex flex-col max-h-[28rem] overflow-hidden">
          {/* Breadcrumbs / back nav */}
          <div className="flex min-h-9 items-center gap-1 border-b border-border px-4 py-1.5">
            {tab === 'mine' ? (
              <MineBreadcrumbs
                breadcrumbs={mineBreadcrumbs}
                onNavigate={navigateMineTo}
                onRefresh={() => loadMine(mineDir)}
              />
            ) : teamProject ? (
              <TeamBreadcrumbs
                project={teamProject}
                breadcrumbs={teamBreadcrumbs}
                onBack={exitTeamProject}
                onNavigate={navigateTeamTo}
                onRefresh={() => loadTeamFiles(teamProject.id, teamDir)}
              />
            ) : (
              <span className="px-1 text-xs text-text-muted">{t('chat.filePickerTeamHint')}</span>
            )}
          </div>

          {/* File list */}
          <div className="flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent">
            {tab === 'mine' ? (
              <FileList
                loading={mineLoading}
                error={mineError}
                items={mineItems}
                onPickFile={(item) =>
                  pickFile({
                    source: 'mine',
                    name: item.name,
                    path: item.path,
                    size: item.size,
                    isDir: false,
                    url: buildUrl('mine', item.path),
                  })
                }
                onEnterDir={(item) => enterMineDir(item.name)}
              />
            ) : teamProject ? (
              <FileList
                loading={teamLoading}
                error={teamError}
                items={teamItems}
                onPickFile={(item) => {
                  pickFile({
                    source: 'team',
                    name: item.name,
                    path: item.path,
                    size: item.size,
                    isDir: false,
                    teamProjectId: teamProject.id,
                    url: buildUrl('team', item.path, teamProject.id),
                  })
                }}
                onEnterDir={(item) => enterTeamDir(item.name)}
              />
            ) : (
              <TeamProjectGrid
                loading={teamProjectsLoading}
                error={teamProjectsError}
                projects={teamProjects}
                onEnter={enterTeamProject}
                onRetry={loadTeamProjects}
              />
            )}
          </div>
        </div>

        {/* Footer: current dir hint + confirm (select current directory) */}
        <div className="flex items-center justify-between border-t border-border px-4 py-3">
          <span className="truncate text-xs text-text-3 max-w-64">
            <span className="text-text-muted">{t('chat.filePickerCurrentDir')}: </span>
            {currentDirLabel}
          </span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>
              {translate('common.cancel')}
            </Button>
            <Button size="sm" onClick={handleConfirm} disabled={!canConfirmDir}>
              {t('chat.filePickerConfirmDir')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Breadcrumbs
   ============================================ */

function MineBreadcrumbs({
  breadcrumbs,
  onNavigate,
  onRefresh,
}: {
  breadcrumbs: string[]
  onNavigate: (index: number) => void
  onRefresh: () => void
}) {
  const t = useT()
  return (
    <>
      <button
        onClick={() => onNavigate(-1)}
        className={cn(
          'rounded px-1.5 py-0.5 text-xs transition-colors',
          breadcrumbs.length === 0 ? 'text-text-1' : 'text-text-3 hover:bg-bg-hover hover:text-text-1',
        )}
      >
        <span className="inline-flex items-center gap-1">
          <FolderOpen className="size-3.5" />
          {t('files.myFiles')}
        </span>
      </button>
      {breadcrumbs.map((seg, i) => (
        <span key={i} className="inline-flex items-center gap-1">
          <ChevronRight className="size-3 text-text-muted" />
          <button
            onClick={() => onNavigate(i)}
            className={cn(
              'rounded px-1.5 py-0.5 text-xs transition-colors',
              i === breadcrumbs.length - 1 ? 'text-text-1' : 'text-text-3 hover:bg-bg-hover hover:text-text-1',
            )}
          >
            {seg}
          </button>
        </span>
      ))}
      <button
        onClick={onRefresh}
        className="ml-auto rounded p-1 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
        aria-label={t('common.refresh')}
      >
        <RefreshCw className="size-3.5" />
      </button>
    </>
  )
}

function TeamBreadcrumbs({
  project,
  breadcrumbs,
  onBack,
  onNavigate,
  onRefresh,
}: {
  project: TeamProjectItem
  breadcrumbs: string[]
  onBack: () => void
  onNavigate: (index: number) => void
  onRefresh: () => void
}) {
  const t = useT()
  return (
    <>
      <button
        onClick={onBack}
        className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
        aria-label={t('files.backTooltip')}
      >
        <ArrowLeft className="size-3.5" />
      </button>
      <Users className="size-3.5 text-highlight" />
      <button
        onClick={onBack}
        className={cn(
          'truncate max-w-40 rounded px-1.5 py-0.5 text-xs transition-colors',
          breadcrumbs.length === 0 ? 'text-text-1' : 'text-text-3 hover:bg-bg-hover hover:text-text-1',
        )}
      >
        {project.projectName}
      </button>
      {breadcrumbs.map((seg, i) => (
        <span key={i} className="inline-flex items-center gap-1">
          <ChevronRight className="size-3 text-text-muted" />
          <button
            onClick={() => onNavigate(i)}
            className={cn(
              'rounded px-1.5 py-0.5 text-xs transition-colors',
              i === breadcrumbs.length - 1 ? 'text-text-1' : 'text-text-3 hover:bg-bg-hover hover:text-text-1',
            )}
          >
            {seg}
          </button>
        </span>
      ))}
      <button
        onClick={onRefresh}
        className="ml-auto rounded p-1 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
        aria-label={t('common.refresh')}
      >
        <RefreshCw className="size-3.5" />
      </button>
    </>
  )
}

/* ============================================
   File list (shared by mine & team)
   - Click on file: pick & close
   - Click on dir: navigate into
   ============================================ */

function FileList({
  loading,
  error,
  items,
  onPickFile,
  onEnterDir,
}: {
  loading: boolean
  error: boolean
  items: FileItem[]
  onPickFile: (item: FileItem) => void
  onEnterDir: (item: FileItem) => void
}) {
  const t = useT()

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="size-5 animate-spin text-text-muted" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-8">
        <AlertCircle className="size-6 text-text-3" />
        <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-8">
        <FolderOpen className="size-6 text-text-3" />
        <p className="text-sm text-text-3">{t('files.emptyFolder')}</p>
      </div>
    )
  }

  return (
    <ul className="py-0.5">
      {items.map((item, i) => (
        <li
          key={item.name}
          onClick={() => (item.isDir ? onEnterDir(item) : onPickFile(item))}
          className={cn(
            'flex cursor-default items-center gap-3 px-4 transition-colors select-none h-9',
            i % 2 === 0 ? 'bg-interactive/5 hover:bg-bg-hover' : 'hover:bg-bg-hover',
          )}
        >
          {item.isDir ? (
            <Folder className="size-4 shrink-0 text-highlight" />
          ) : (
            (() => {
              const Icon = getFileIcon(item)
              return <Icon className="size-4 shrink-0 text-text-3" />
            })()
          )}
          <span className="flex-1 truncate text-sm text-text-1">
            {item.name}
            {item.isDir && '/'}
          </span>
          {!item.isDir && (
            <span className="w-20 shrink-0 text-right text-xs text-text-muted tabular-nums">
              {formatSize(item.size)}
            </span>
          )}
          {item.isDir && (
            <ChevronRight className="size-4 shrink-0 text-text-muted" />
          )}
        </li>
      ))}
    </ul>
  )
}

/* ============================================
   Team project grid (when no project selected)
   ============================================ */

function TeamProjectGrid({
  loading,
  error,
  projects,
  onEnter,
  onRetry,
}: {
  loading: boolean
  error: boolean
  projects: TeamProjectItem[]
  onEnter: (project: TeamProjectItem) => void
  onRetry: () => void
}) {
  const t = useT()

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="size-5 animate-spin text-text-muted" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-8">
        <AlertCircle className="size-6 text-text-3" />
        <button
          onClick={onRetry}
          className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
        >
          <RefreshCw className="size-4" />
          {t('common.retry')}
        </button>
      </div>
    )
  }

  if (projects.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-8">
        <Users className="size-6 text-text-3" />
        <p className="text-sm text-text-3">{t('files.noTeamProjects')}</p>
        <p className="text-xs text-text-muted">{t('files.noTeamProjectsHint')}</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-3 p-4">
      {projects.map((project) => (
        <div
          key={project.id}
          onClick={() => onEnter(project)}
          className="group flex cursor-pointer flex-col gap-1.5 rounded-lg border border-border bg-bg-layer-1 p-3 transition-colors hover:border-highlight/40 hover:bg-bg-hover"
        >
          <div className="flex items-center gap-1.5">
            <Users className="size-3.5 shrink-0 text-highlight" />
            <span className="truncate text-sm font-semibold text-text-1">
              {project.projectName}
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            {project.ownerAvatar ? (
              <img
                src={project.ownerAvatar}
                alt={project.ownerName}
                className="size-4 shrink-0 rounded-sm object-cover"
              />
            ) : (
              <div className="inline-flex size-4 shrink-0 items-center justify-center rounded-sm bg-interactive text-[9px] font-medium text-white">
                {project.ownerName.charAt(0).toUpperCase()}
              </div>
            )}
            <span className="truncate text-xs text-text-3">{project.ownerName}</span>
          </div>
          {project.description && (
            <p className="line-clamp-2 text-xs text-text-2">{project.description}</p>
          )}
        </div>
      ))}
    </div>
  )
}