import { useState, useCallback, useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { toast } from 'sonner'
import { Users, RefreshCw, MoreHorizontal, Pencil, FolderOpen, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { listTeamProjects, unshareProject, updateProjectDescription } from '@/api/team-projects'
import type { TeamProjectItem } from '@/api/types'
import { formatTime } from './file-helpers'
import { ShareDialog, type ShareDialogMode } from './ShareDialog'

interface TeamProjectViewProps {
  onBack: () => void
  onEnterProject: (project: TeamProjectItem) => void
}

export function TeamProjectView({ onBack, onEnterProject }: TeamProjectViewProps) {
  const t = useT()
  const [projects, setProjects] = useState<TeamProjectItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [menuFor, setMenuFor] = useState<{ id: number; x: number; y: number } | null>(null)
  const [dialogMode, setDialogMode] = useState<ShareDialogMode>(null)
  const [dialogProject, setDialogProject] = useState<TeamProjectItem | null>(null)
  const [dialogDesc, setDialogDesc] = useState('')
  const menuRef = useRef<HTMLDivElement>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    setMenuFor(null)
    listTeamProjects()
      .then((data) => {
        setProjects(data.items || [])
        setLoading(false)
      })
      .catch(() => {
        setError(true)
        setLoading(false)
      })
  }, [])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (!menuFor) return
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuFor(null)
      }
    }
    const keyHandler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuFor(null)
    }
    document.addEventListener('mousedown', handler)
    document.addEventListener('keydown', keyHandler)
    return () => {
      document.removeEventListener('mousedown', handler)
      document.removeEventListener('keydown', keyHandler)
    }
  }, [menuFor])

  const openEditDialog = useCallback((project: TeamProjectItem) => {
    setMenuFor(null)
    setDialogProject(project)
    setDialogDesc(project.description || '')
    setDialogMode('editShare')
  }, [])

  const openUnshareDialog = useCallback((project: TeamProjectItem) => {
    setMenuFor(null)
    setDialogProject(project)
    setDialogMode('unshare')
  }, [])

  const closeDialog = useCallback(() => {
    setDialogMode(null)
    setDialogProject(null)
    setDialogDesc('')
  }, [])

  const handleConfirmDialog = useCallback(() => {
    if (!dialogProject || !dialogMode) return
    if (dialogMode === 'editShare') {
      updateProjectDescription(dialogProject.id, dialogDesc)
        .then(() => {
          toast.success(translate('files.descUpdated'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    } else if (dialogMode === 'unshare') {
      unshareProject(dialogProject.id)
        .then(() => {
          toast.success(translate('files.unshareSuccess'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    }
  }, [dialogProject, dialogMode, dialogDesc, closeDialog, load])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <div className="flex h-10 shrink-0 items-center gap-1 border-b border-border px-3">
        <Button
          variant="ghost"
          size="default"
          onClick={onBack}
          className="text-highlight hover:text-highlight"
        >
          <Users className="size-4" />
          {t('files.teamProjects')}
        </Button>

        <div className="flex-1" />

        <Button
          variant="ghost"
          size="default"
          onClick={onBack}
          className="text-highlight hover:text-highlight"
        >
          <FolderOpen className="size-4" />
          {t('files.myFiles')}
        </Button>

        <Button
          variant="ghost"
          size="icon"
          onClick={load}
          title={t('common.refresh')}
          className="text-highlight hover:text-highlight"
        >
          <RefreshCw className="size-4" />
        </Button>
      </div>

      <div className="flex-1 overflow-auto p-4">
        {loading && (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-36 animate-pulse rounded-lg border border-border bg-bg-layer-1"
              />
            ))}
          </div>
        )}

        {error && (
          <div className="flex h-full flex-col items-center justify-center gap-4">
            <AlertCircle className="size-8 text-text-3" />
            <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
            <button
              onClick={load}
              className="inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
            >
              <RefreshCw className="size-4" />
              {t('common.retry')}
            </button>
          </div>
        )}

        {!loading && !error && projects.length === 0 && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <FolderOpen className="size-12 text-text-3" />
            <div className="text-center">
              <p className="text-sm text-text-2">{t('files.noTeamProjects')}</p>
              <p className="mt-1 text-sm text-text-3">{t('files.noTeamProjectsHint')}</p>
            </div>
          </div>
        )}

        {!loading && !error && projects.length > 0 && (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4">
            {projects.map((project) => (
              <div
                key={project.id}
                onClick={() => onEnterProject(project)}
                className="group relative flex cursor-pointer flex-col gap-2 rounded-lg border border-border bg-bg-layer-1 p-4 transition-colors hover:border-highlight/40 hover:bg-bg-hover"
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="text-sm font-semibold text-text-1 truncate">
                    {project.projectName}
                  </span>
                  {project.isOwner && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        setMenuFor({ id: project.id, x: e.clientX, y: e.clientY })
                      }}
                      className="shrink-0 rounded-sm p-0.5 text-text-3 opacity-0 transition-opacity hover:bg-bg-layer-3 hover:text-text-1 group-hover:opacity-100"
                    >
                      <MoreHorizontal className="size-4" />
                    </button>
                  )}
                </div>

                <div className="flex items-center gap-1.5">
                    {project.ownerAvatar ? (
                      <img
                        src={project.ownerAvatar}
                        alt={project.ownerName}
                        className="size-6 shrink-0 rounded-sm object-cover"
                      />
                    ) : (
                      <div className="inline-flex size-6 shrink-0 items-center justify-center rounded-sm bg-interactive text-[11px] font-medium text-white">
                        {project.ownerName.charAt(0).toUpperCase()}
                      </div>
                    )}
<span className="text-base text-text-3 truncate">{project.ownerName}</span>
                </div>

                {project.description && (
                  <p className="text-xs text-text-2 line-clamp-2">{project.description}</p>
                )}

                <span className="mt-auto text-xs text-text-muted tabular-nums">
                  {formatTime(project.updatedAt)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {menuFor && createPortal(
        <div
          ref={menuRef}
          className="fixed z-50 min-w-[140px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop"
          style={{
            left: Math.min(menuFor.x, window.innerWidth - 148),
            top: Math.min(menuFor.y, window.innerHeight - 80),
          }}
        >
          <button
            onClick={() => {
              const project = projects.find((p) => p.id === menuFor.id)
              if (project) openEditDialog(project)
            }}
            className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
          >
            <Pencil className="size-3.5" />
            {t('files.editDescription')}
          </button>
          <button
            onClick={() => {
              const project = projects.find((p) => p.id === menuFor.id)
              if (project) openUnshareDialog(project)
            }}
            className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
          >
            <Users className="size-3.5" />
            {t('files.unshare')}
          </button>
        </div>,
        document.body,
      )}

      <ShareDialog
        mode={dialogMode}
        projectName={dialogProject?.projectName || ''}
        description={dialogDesc}
        onDescriptionChange={setDialogDesc}
        onClose={closeDialog}
        onConfirm={handleConfirmDialog}
      />
    </div>
  )
}
