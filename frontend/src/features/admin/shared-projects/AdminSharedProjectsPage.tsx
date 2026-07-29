import { useState, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import {
  Trash2,
  AlertCircle,
  RefreshCw,
  Loader2,
  Lock,
  Share2,
  Pencil,
} from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import { adminSharedProjectsApi } from '@/api'
import type { AdminTeamProjectItem } from '@/api'

type PageState = 'loading' | 'data' | 'empty' | 'error'

/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatRelative(iso: string): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return translate('files.justNow')
  if (mins < 60) return `${mins}${translate('files.minutesAgo')}`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}${translate('files.hoursAgo')}`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}${translate('files.daysAgo')}`
  return formatDate(iso)
}

function Avatar({ name, avatar }: { name: string; avatar: string }) {
  if (avatar) {
    return (
      <img
        src={avatar}
        alt={name}
        className="size-7 shrink-0 rounded-sm object-cover"
      />
    )
  }
  return (
    <div className="inline-flex size-7 shrink-0 items-center justify-center rounded-sm bg-interactive text-xs font-medium text-white">
      {name.charAt(0).toUpperCase()}
    </div>
  )
}

/* ============================================
   Edit Description Dialog
   ============================================ */

function EditDescriptionDialog({
  open,
  onClose,
  onConfirm,
  projectName,
  description,
  loading,
}: {
  open: boolean
  onClose: () => void
  onConfirm: (desc: string) => void
  projectName: string
  description: string
  loading: boolean
}) {
  const t = useT()
  const [value, setValue] = useState(description)

  useEffect(() => {
    if (open) setValue(description)
  }, [open, description])

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminSharedProjects.editDescription')}</DialogTitle>
          <DialogDescription>
            {translate('adminSharedProjects.editDescriptionFor').replace('{name}', projectName)}
          </DialogDescription>
        </DialogHeader>
        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder={t('adminSharedProjects.descriptionPlaceholder')}
          rows={3}
          className="w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 resize-none"
        />
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={loading}>{t('common.cancel')}</Button>
          <Button size="default" onClick={() => onConfirm(value)} disabled={loading}>
            {loading && <Loader2 className="size-3.5 mr-1 animate-spin" />}
            {t('common.save')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

function DeleteProjectDialog({
  open,
  onClose,
  onConfirm,
  projectName,
  loading,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  projectName: string
  loading: boolean
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminSharedProjects.deleteProject')}</DialogTitle>
          <DialogDescription>
            {translate('adminSharedProjects.deleteProjectConfirm').replace('{name}', projectName)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={loading}>{t('common.cancel')}</Button>
          <Button variant="destructive" size="default" onClick={onConfirm} disabled={loading}>
            {loading && <Loader2 className="size-3.5 mr-1 animate-spin" />}
            {t('common.delete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Table Row
   ============================================ */

function ProjectRow({
  project,
  onDelete,
  onEdit,
}: {
  project: AdminTeamProjectItem
  onDelete: () => void
  onEdit: () => void
}) {
  const t = useT()
  return (
    <tr className="transition-colors hover:bg-bg-hover">
      <td className="py-2.5 pl-4 pr-2">
        <div className="flex items-center gap-2">
          <Share2 className="size-4 shrink-0 text-text-3" />
          <span className="text-sm font-medium text-text-1">{project.projectName}</span>
        </div>
      </td>
      <td className="py-2.5 pr-4">
        <div className="flex items-center gap-2">
          <Avatar name={project.creatorName || project.creatorId} avatar={project.creatorAvatar} />
          <span className="text-sm text-text-2">{project.creatorName || project.creatorId}</span>
        </div>
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-2">
        <span className="line-clamp-1 max-w-[260px]">{project.description || '—'}</span>
      </td>
      <td className="py-2.5 pr-4 text-sm tabular-nums text-text-2">{project.visitCount}</td>
      <td className="py-2.5 pr-4 text-sm text-text-3">
        {project.lastActiveAt ? formatRelative(project.lastActiveAt) : '—'}
      </td>
      <td className="py-2.5 pr-4">
        {project.locked ? (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs bg-highlight/15 text-highlight">
            <Lock className="size-3" />
            {t('adminSharedProjects.inUse')}
          </span>
        ) : (
          <span className="text-xs text-text-3">{t('adminSharedProjects.idle')}</span>
        )}
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-3">{formatDate(project.created)}</td>
      <td className="py-2.5 pr-4">
        <div className="flex items-center gap-0.5">
          <button
            onClick={onEdit}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-interactive"
            title={t('adminSharedProjects.editDescription')}
          >
            <Pencil className="size-3.5" />
          </button>
          <button
            onClick={onDelete}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-destructive"
            title={t('adminSharedProjects.deleteProject')}
          >
            <Trash2 className="size-3.5" />
          </button>
        </div>
      </td>
    </tr>
  )
}

/* ============================================
   Skeleton
   ============================================ */

function TableSkeleton() {
  const t = useT()
  return (
    <div className="flex-1 overflow-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border text-left text-xs text-text-3">
            <th className="pb-2 pl-4 pr-2 font-normal">{t('adminSharedProjects.projectName')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.creator')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.description')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.visitCount')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.lastActive')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.status')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.createTime')}</th>
            <th className="pb-2 pr-4 font-normal" />
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: 8 }).map((_, i) => (
            <tr key={i}>
              <td className="py-2.5 pl-4 pr-2">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="flex items-center gap-2">
                  <div className="size-7 animate-pulse rounded-sm bg-bg-layer-3" />
                  <div className="h-3.5 w-16 animate-pulse rounded bg-bg-layer-2" />
                </div>
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-32 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-5 w-12 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/* ============================================
   Empty State
   ============================================ */

function EmptyState({ onRefresh }: { onRefresh: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3">
      <Share2 className="size-10 text-text-muted" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('adminSharedProjects.noProjects')}</p>
      </div>
      <button
        onClick={onRefresh}
        className="text-xs text-interactive transition-colors hover:text-interactive-hover"
      >
        {t('adminUsers.clearFilter')}
      </button>
    </div>
  )
}

/* ============================================
   Error State
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('adminSharedProjects.fetchFailed')}</p>
        <p className="mt-1 text-xs text-text-3">{t('adminSharedProjects.fetchFailed')}</p>
      </div>
      <button
        onClick={onRetry}
        className="inline-flex items-center gap-1 rounded-md bg-bg-layer-2 px-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-layer-3"
      >
        <RefreshCw className="size-3" />
        {t('common.retry')}
      </button>
    </div>
  )
}

/* ============================================
   Pagination
   ============================================ */

interface PaginateProps {
  page: number
  totalPages: number
  totalCount: number
  onPageChange: (p: number) => void
}

import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useMemo } from 'react'

function Pagination({ page, totalPages, totalCount, onPageChange }: PaginateProps) {
  const pages = useMemo(() => {
    const result: (number | '...')[] = []
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) result.push(i)
    } else {
      result.push(1)
      if (page > 3) result.push('...')
      for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
        result.push(i)
      }
      if (page < totalPages - 2) result.push('...')
      result.push(totalPages)
    }
    return result
  }, [page, totalPages])

  if (totalPages <= 1) return null

  return (
    <div className="flex items-center justify-between border-t border-border px-4 py-2">
      <span className="text-xs text-text-3">共 {totalCount} 个项目</span>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={page === 1}
          className="inline-flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronLeft className="size-3.5" />
        </button>
        {pages.map((p, i) =>
          p === '...' ? (
            <span key={`dots-${i}`} className="px-1 text-xs text-text-3">
              ...
            </span>
          ) : (
            <button
              key={p}
              onClick={() => onPageChange(p)}
              className={cn(
                'inline-flex size-7 items-center justify-center rounded-md text-xs tabular-nums transition-colors',
                p === page
                  ? 'bg-bg-layer-3 text-text-1 font-medium'
                  : 'text-text-3 hover:bg-bg-layer-2 hover:text-text-1',
              )}
            >
              {p}
            </button>
          ),
        )}
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={page === totalPages}
          className="inline-flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronRight className="size-3.5" />
        </button>
      </div>
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

export function Component() {
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [projects, setProjects] = useState<AdminTeamProjectItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [page, setPage] = useState(1)
  const [deleteTarget, setDeleteTarget] = useState<AdminTeamProjectItem | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [editTarget, setEditTarget] = useState<AdminTeamProjectItem | null>(null)
  const [saving, setSaving] = useState(false)

  const pageSize = 20

  const loadData = useCallback(() => {
    setState('loading')
    adminSharedProjectsApi.getTeamProjects({ page, pageSize })
      .then((res) => {
        const list = res.list ?? []
        // 当本页为空且不在第一页时（通常是删除最后一项后），回退一页
        if (list.length === 0 && page > 1) {
          setPage(page - 1)
          return
        }
        setProjects(list)
        setTotalCount(res.totalCount)
        setTotalPages(Math.max(1, res.totalPage))
        setState(list.length > 0 ? 'data' : 'empty')
      })
      .catch(() => {
        setState('error')
      })
  }, [page])

  useEffect(() => {
    loadData()
  }, [loadData])

  const safePage = Math.min(page, totalPages)

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return
    setDeleting(true)
    adminSharedProjectsApi.deleteTeamProject(deleteTarget.id)
      .then(() => {
        setDeleteTarget(null)
        toast.success(translate('adminSharedProjects.deleteSuccess'))
        loadData()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.failed'))
      })
      .finally(() => {
        setDeleting(false)
      })
  }, [deleteTarget, loadData])

  const handleEditDescription = useCallback((desc: string) => {
    if (!editTarget) return
    setSaving(true)
    adminSharedProjectsApi.updateTeamProject(editTarget.id, { description: desc })
      .then(() => {
        setEditTarget(null)
        toast.success(translate('adminSharedProjects.updateSuccess'))
        loadData()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.failed'))
      })
      .finally(() => {
        setSaving(false)
      })
  }, [editTarget, loadData])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('sidebar.sharedProjects')}</h1>
      </div>

      {/* Content */}
      {state === 'loading' && <TableSkeleton />}

      {state === 'error' && <ErrorState onRetry={loadData} />}

      {state === 'empty' && <EmptyState onRefresh={() => loadData()} />}

      {state === 'data' && projects.length === 0 && (
        <EmptyState onRefresh={() => loadData()} />
      )}

      {state === 'data' && projects.length > 0 && (
        <>
          <div className="flex-1 overflow-auto">
            <table className="w-full">
              <thead>
                <tr className="sticky top-0 z-10 border-b border-border bg-bg-base text-left text-xs text-text-3">
                  <th className="pb-2 pl-4 pr-2 font-normal">{t('adminSharedProjects.projectName')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.creator')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.description')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.visitCount')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.lastActive')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.status')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminSharedProjects.createTime')}</th>
                  <th className="pb-2 pr-4 font-normal" />
                </tr>
              </thead>
              <tbody>
                {projects.map((project) => (
                  <ProjectRow
                    key={project.id}
                    project={project}
                    onDelete={() => setDeleteTarget(project)}
                    onEdit={() => setEditTarget(project)}
                  />
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            page={safePage}
            totalPages={totalPages}
            totalCount={totalCount}
            onPageChange={setPage}
          />
        </>
      )}

      {/* Dialogs */}
      <EditDescriptionDialog
        open={!!editTarget}
        onClose={() => { if (!saving) setEditTarget(null) }}
        onConfirm={handleEditDescription}
        projectName={editTarget?.projectName ?? ''}
        description={editTarget?.description ?? ''}
        loading={saving}
      />
      <DeleteProjectDialog
        open={!!deleteTarget}
        onClose={() => { if (!deleting) setDeleteTarget(null) }}
        onConfirm={handleDelete}
        projectName={deleteTarget?.projectName ?? ''}
        loading={deleting}
      />
    </div>
  )
}
