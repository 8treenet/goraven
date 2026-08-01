import { useState, useCallback, useEffect, useRef, forwardRef, useImperativeHandle } from 'react'
import { createPortal } from 'react-dom'
import { toast } from 'sonner'
import { RefreshCw, MoreHorizontal, Pencil, FolderOpen, AlertCircle, Trash2, Plus, Users } from 'lucide-react'
import { useT, t as translate } from '@/i18n'
import { listTeamProjects, deleteTeamProject, updateProjectDescription, createTeamProject } from '@/api/team-projects'
import type { TeamProjectItem } from '@/api/types'
import { useUserStore } from '@/stores/user-store'
import { formatTime } from './file-helpers'
import { ShareDialog, type ShareDialogMode } from './ShareDialog'
import { MembersDialog } from './MembersDialog'

export interface TeamProjectViewHandle {
  createProject: () => void
}

interface TeamProjectViewProps {
  onEnterProject: (project: TeamProjectItem) => void
}

export const TeamProjectView = forwardRef<TeamProjectViewHandle, TeamProjectViewProps>(function TeamProjectView({ onEnterProject }, ref) {
  const t = useT()
  const [projects, setProjects] = useState<TeamProjectItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [menuFor, setMenuFor] = useState<{ id: number; x: number; y: number } | null>(null)
  const [dialogMode, setDialogMode] = useState<ShareDialogMode>(null)
  const [dialogProject, setDialogProject] = useState<TeamProjectItem | null>(null)
  const [dialogDesc, setDialogDesc] = useState('')
  const [dialogProjectName, setDialogProjectName] = useState('')
  const [membersProject, setMembersProject] = useState<TeamProjectItem | null>(null)
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

  const openDeleteDialog = useCallback((project: TeamProjectItem) => {
    setMenuFor(null)
    setDialogProject(project)
    setDialogMode('delete')
  }, [])

  const openMembersDialog = useCallback((project: TeamProjectItem) => {
    setMenuFor(null)
    setMembersProject(project)
  }, [])

  const openCreateDialog = useCallback(() => {
    setDialogProject(null)
    setDialogDesc('')
    setDialogProjectName('')
    setDialogMode('create')
  }, [])

  const closeDialog = useCallback(() => {
    setDialogMode(null)
    setDialogProject(null)
    setDialogDesc('')
    setDialogProjectName('')
  }, [])

  const handleConfirmDialog = useCallback(() => {
    if (dialogMode === 'create') {
      if (!dialogProjectName.trim()) return
      createTeamProject(dialogProjectName.trim(), dialogDesc)
        .then((rsp) => {
          toast.success(translate('files.createProjectSuccess'))
          const user = useUserStore.getState().currentUser
          const newProject: TeamProjectItem = {
            id: rsp.id,
            creatorId: user?.userId || '',
            creatorName: user?.nickname || user?.username || '',
            creatorAvatar: user?.avatar || '',
            projectName: dialogProjectName.trim(),
            description: dialogDesc,
            access: 0,
            updatedAt: new Date().toISOString(),
            isCreator: true,
          }
          closeDialog()
          load()
          setMembersProject(newProject)
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    } else if (!dialogProject || !dialogMode) {
      return
    } else if (dialogMode === 'editShare') {
      updateProjectDescription(dialogProject.id, dialogDesc)
        .then(() => {
          toast.success(translate('files.descUpdated'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    } else if (dialogMode === 'delete') {
      deleteTeamProject(dialogProject.id)
        .then(() => {
          toast.success(translate('files.deleteProjectSuccess'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    }
  }, [dialogProject, dialogMode, dialogDesc, dialogProjectName, closeDialog, load])

  useImperativeHandle(ref, () => ({
    createProject: () => openCreateDialog(),
  }), [openCreateDialog])

  return (
    <div className="flex h-full flex-col">
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
            <button
              onClick={openCreateDialog}
              className="mt-1 inline-flex items-center gap-1.5 rounded-md bg-bg-layer-2 px-3 py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-layer-3"
            >
              <Plus className="size-4" />
              {t('files.newTeamProject')}
            </button>
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
                  {project.isCreator && (
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
                    {project.creatorAvatar ? (
                      <img
                        src={project.creatorAvatar}
                        alt={project.creatorName}
                        className="size-6 shrink-0 rounded-sm object-cover"
                      />
                    ) : (
                      <div className="inline-flex size-6 shrink-0 items-center justify-center rounded-sm bg-interactive text-[11px] font-medium text-white">
                        {project.creatorName.charAt(0).toUpperCase()}
                      </div>
                    )}
<span className="text-base text-text-3 truncate">{project.creatorName}</span>
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
            top: Math.min(menuFor.y, window.innerHeight - 120),
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
              if (project) openMembersDialog(project)
            }}
            className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
          >
            <Users className="size-3.5" />
            {t('files.editMembers')}
          </button>
          <button
            onClick={() => {
              const project = projects.find((p) => p.id === menuFor.id)
              if (project) openDeleteDialog(project)
            }}
            className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
          >
            <Trash2 className="size-3.5" />
            {t('common.delete')}
          </button>
        </div>,
        document.body,
      )}

      <ShareDialog
        mode={dialogMode}
        projectName={dialogMode === 'create' ? dialogProjectName : (dialogProject?.projectName || '')}
        description={dialogDesc}
        onDescriptionChange={setDialogDesc}
        onProjectNameChange={setDialogProjectName}
        onClose={closeDialog}
        onConfirm={handleConfirmDialog}
      />

      <MembersDialog
        project={membersProject}
        onClose={() => setMembersProject(null)}
        onSaved={load}
      />
    </div>
  )
})
