import { useState, useCallback, useEffect, useRef, forwardRef, useImperativeHandle } from 'react'
import { createPortal } from 'react-dom'
import { toast } from 'sonner'
import { MoreHorizontal, Pencil, Folder, FolderOpen, Trash2, Plus } from 'lucide-react'
import { useT, t as translate } from '@/i18n'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { EmptyState } from '@/components/patterns/EmptyState'
import { ErrorState } from '@/components/patterns/ErrorState'
import { listMyProjects, createMyProject, updateMyProject, deleteMyProject, syncMyProjects, myProjectGitApi } from '@/api/my-projects'
import type { MyProjectItem, GitClonePayload } from '@/api/types'
import { formatTime } from './file-helpers'
import { ShareDialog, type ShareDialogMode } from './ShareDialog'
import { GIT_CLONE_STATE } from './git/git-helpers'
import { GitCardChip } from './git/GitCardChip'
import { GitDialog } from './git/GitDialog'

export interface MyProjectViewHandle {
  createProject: () => void
  syncProjects: () => void
}

interface MyProjectViewProps {
  onEnterProject: (project: MyProjectItem) => void
}

export const MyProjectView = forwardRef<MyProjectViewHandle, MyProjectViewProps>(function MyProjectView({ onEnterProject }, ref) {
  const t = useT()
  const [projects, setProjects] = useState<MyProjectItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [menuFor, setMenuFor] = useState<{ id: number; x: number; y: number } | null>(null)
  const [gitProject, setGitProject] = useState<MyProjectItem | null>(null)
  const [dialogMode, setDialogMode] = useState<ShareDialogMode>(null)
  const [dialogProject, setDialogProject] = useState<MyProjectItem | null>(null)
  const [dialogDesc, setDialogDesc] = useState('')
  const [dialogProjectName, setDialogProjectName] = useState('')
  const menuRef = useRef<HTMLDivElement>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    setMenuFor(null)
    listMyProjects()
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

  // 手动同步：补录会话中新建的项目目录后重载列表
  const syncProjects = useCallback(() => {
    syncMyProjects()
      .then(() => load())
      .catch((err: Error) => toast.error(err.message))
  }, [load])

  // 静默刷新：仅更新数据，不显示骨架屏、不清除菜单（克隆轮询用）
  const silentRefresh = useCallback(() => {
    listMyProjects().then((data) => setProjects(data.items || [])).catch(() => {})
  }, [])

  // 有克隆进行中的项目时每 3 秒静默刷新列表
  useEffect(() => {
    if (!projects.some((p) => p.cloneState === GIT_CLONE_STATE.RUNNING)) return
    const timer = setInterval(silentRefresh, 3000)
    return () => clearInterval(timer)
  }, [projects, silentRefresh])

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

  const openEditDialog = useCallback((project: MyProjectItem) => {
    setMenuFor(null)
    setDialogProject(project)
    setDialogProjectName(project.projectName)
    setDialogDesc(project.description || '')
    setDialogMode('editShare')
  }, [])

  const openDeleteDialog = useCallback((project: MyProjectItem) => {
    setMenuFor(null)
    setDialogProject(project)
    setDialogProjectName(project.projectName)
    setDialogMode('delete')
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

  const handleConfirmDialog = useCallback((git?: GitClonePayload) => {
    if (dialogMode === 'create') {
      if (!dialogProjectName.trim()) return
      createMyProject(dialogProjectName.trim(), dialogDesc, git)
        .then(() => {
          toast.success(translate('files.createProjectSuccess'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    } else if (!dialogProject || !dialogMode) {
      return
    } else if (dialogMode === 'editShare') {
      updateMyProject(dialogProject.id, {
        projectName: dialogProjectName.trim(),
        description: dialogDesc,
      })
        .then(() => {
          toast.success(translate('files.descUpdated'))
          closeDialog()
          load()
        })
        .catch((err: Error) => {
          toast.error(err.message)
        })
    } else if (dialogMode === 'delete') {
      deleteMyProject(dialogProject.id)
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
    syncProjects: () => syncProjects(),
  }), [openCreateDialog, syncProjects])

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-auto p-6">
        {loading && (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(min(100%,340px),1fr))] gap-4">
            {[1, 2, 3].map((i) => (
              <Card key={i} className="h-36 animate-pulse" />
            ))}
          </div>
        )}

        {error && (
          <ErrorState className="h-full" onRetry={load} />
        )}

        {!loading && !error && projects.length === 0 && (
          <EmptyState
            icon={<FolderOpen className="size-10 text-text-muted" />}
            title={t('files.noMyProjects')}
            description={t('files.noMyProjectsHint')}
            action={
              <Button onClick={openCreateDialog}>
                <Plus className="size-4" />
                {t('files.newProject')}
              </Button>
            }
            className="h-full"
          />
        )}

        {!loading && !error && projects.length > 0 && (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(min(100%,340px),1fr))] gap-4">
            {projects.map((project) => (
              <Card
                key={project.id}
                onClick={() => onEnterProject(project)}
                className="group relative flex cursor-pointer flex-col gap-2 p-4 transition-colors hover:border-border-strong hover:bg-bg-layer-2"
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="flex min-w-0 items-center gap-1.5 text-section font-semibold text-text-1">
                    <Folder className="size-4 shrink-0 text-amber-strong" />
                    <span className="truncate">{project.projectName}</span>
                  </span>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t('sidebar.menu')}
                    onClick={(e) => {
                      e.stopPropagation()
                      setMenuFor({ id: project.id, x: e.clientX, y: e.clientY })
                    }}
                    className="shrink-0 text-text-3 opacity-100 md:opacity-0 md:group-hover:opacity-100 md:focus-visible:opacity-100"
                  >
                    <MoreHorizontal className="size-4" />
                  </Button>
                </div>

                {project.description && (
                  <p className="text-meta text-text-2 line-clamp-2">{project.description}</p>
                )}

                <div className="mt-auto flex min-h-6 items-center justify-between gap-2">
                  <span className="text-micro text-text-muted tabular-nums">
                    {formatTime(project.updatedAt)}
                  </span>
                  <GitCardChip
                    gitEnabled={project.gitEnabled}
                    gitHasRemote={project.gitHasRemote}
                    cloneState={project.cloneState}
                    onClick={() => setGitProject(project)}
                  />
                </div>
              </Card>
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
            {t('files.editMyProject')}
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
        projectName={dialogProjectName}
        description={dialogDesc}
        onDescriptionChange={setDialogDesc}
        onProjectNameChange={setDialogProjectName}
        allowRenameInEdit
        createTitle={t('files.createMyProjectTitle')}
        createDesc={t('files.createMyProjectDesc')}
        editTitle={t('files.editMyProject')}
        deleteDesc={t('files.confirmDeleteMyProjectDesc')}
        gitTestApi={myProjectGitApi.testRemoteDirect}
        onClose={closeDialog}
        onConfirm={handleConfirmDialog}
      />

      <GitDialog
        open={gitProject !== null}
        onOpenChange={(v) => { if (!v) setGitProject(null) }}
        projectName={gitProject?.projectName ?? ''}
        projectId={gitProject?.id ?? 0}
        gitApi={myProjectGitApi}
        canConfig
      />
    </div>
  )
})