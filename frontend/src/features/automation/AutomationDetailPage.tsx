import { useState, useCallback, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { ChevronLeft, ChevronDown, Play, AlertCircle, RefreshCw, Pencil, Power, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { useT, t as translate } from '@/i18n'
import { cn } from '@/lib/utils'
import { AutomationStatus } from '@/api/types'
import type { AutomationTaskDetail, AutomationExecutionItem, AutomationQARsp } from '@/api/types'
import {
  getAutomationTask,
  getTaskExecutions,
  getExecutionQA,
  updateTaskStatus,
  updateTaskRequirement,
  deleteTask,
  executeTask,
} from '@/api/automation'
import { StatusBadge, execTypeLabel, scheduleTimeValue, formatDateTime, formatDate, formatDuration } from './shared'
import { Markdown } from '@/components/common/markdown'

type PageState = 'loading' | 'data' | 'error'

const EXEC_PAGE_SIZE = 4

/* ============================================
   Page
   ============================================ */

export function Component() {
  const { id } = useParams()
  const taskId = Number(id)
  const navigate = useNavigate()
  const t = useT()

  const [state, setState] = useState<PageState>('loading')
  const [detail, setDetail] = useState<AutomationTaskDetail | null>(null)
  const [executions, setExecutions] = useState<AutomationExecutionItem[]>([])
  const [execPage, setExecPage] = useState(1)
  const [execTotalPage, setExecTotalPage] = useState(1)
  const [execTotalCount, setExecTotalCount] = useState(0)
  const [expandedId, setExpandedId] = useState<number | null>(null)
  const [qaCache, setQaCache] = useState<Record<number, AutomationQARsp>>({})
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [reqEditOpen, setReqEditOpen] = useState(false)
  const [reqValue, setReqValue] = useState('')
  const [reqSaving, setReqSaving] = useState(false)

  const fetchDetail = useCallback(() => {
    getAutomationTask(taskId)
      .then((d) => { setDetail(d); setState('data') })
      .catch(() => setState('error'))
  }, [taskId])

  const fetchExecutions = useCallback((page: number) => {
    getTaskExecutions(taskId, { page, pageSize: EXEC_PAGE_SIZE })
      .then((rsp) => {
        setExecutions(rsp.list)
        setExecPage(rsp.page)
        setExecTotalPage(rsp.totalPage)
        setExecTotalCount(rsp.totalCount)
      })
      .catch(() => toast.error(translate('common.loadFailed')))
  }, [taskId])

  useEffect(() => {
    setState('loading')
    fetchDetail()
    fetchExecutions(1)
  }, [fetchDetail, fetchExecutions])

  const handleToggleExpand = useCallback(async (exec: AutomationExecutionItem) => {
    if (expandedId === exec.id) {
      setExpandedId(null)
      return
    }
    setExpandedId(exec.id)
    if (!qaCache[exec.id]) {
      try {
        const qa = await getExecutionQA(taskId, exec.id)
        setQaCache((prev) => ({ ...prev, [exec.id]: qa }))
      } catch {
        toast.error(translate('common.loadFailed'))
      }
    }
  }, [expandedId, qaCache, taskId])

  const handleToggleStatus = useCallback(async () => {
    if (!detail) return
    if (detail.status === AutomationStatus.Done) {
      toast.error(translate('automation.doneCannotToggle'))
      return
    }
    const target = detail.status === AutomationStatus.Enabled
      ? AutomationStatus.Disabled
      : AutomationStatus.Enabled
    setBusy(true)
    try {
      await updateTaskStatus(detail.id, target)
      setDetail({ ...detail, status: target })
      toast.success(target === AutomationStatus.Enabled ? translate('automation.taskEnabled') : translate('automation.taskDisabled'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusy(false)
    }
  }, [detail])

  const handleExecute = useCallback(async () => {
    if (!detail) return
    setBusy(true)
    try {
      await executeTask(detail.id)
      toast.success(translate('automation.executionStarted'))
      fetchDetail()
      fetchExecutions(1)
      setExpandedId(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusy(false)
    }
  }, [detail, fetchDetail, fetchExecutions])

  const handleDeleteConfirm = useCallback(async () => {
    if (!detail) return
    try {
      await deleteTask(detail.id)
      toast.success(translate('automation.deleted'))
      navigate('/automation')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.deleteFailed'))
    }
  }, [detail, navigate])

  const openReqEdit = useCallback(() => {
    if (!detail) return
    setReqValue(detail.requirement)
    setReqEditOpen(true)
  }, [detail])

  const handleSaveRequirement = useCallback(async () => {
    if (!detail) return
    if (!reqValue.trim()) {
      toast.error(translate('automation.requirementRequired'))
      return
    }
    setReqSaving(true)
    try {
      await updateTaskRequirement(detail.id, reqValue)
      setDetail((prev) => (prev ? { ...prev, requirement: reqValue } : prev))
      setReqEditOpen(false)
      toast.success(translate('automation.requirementUpdated'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setReqSaving(false)
    }
  }, [detail, reqValue])

  const done = detail?.status === AutomationStatus.Done

  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* Toolbar */}
      <div className="flex h-11 shrink-0 items-center gap-3 border-b border-border-custom bg-bg-base px-3">
        <button
          onClick={() => navigate('/automation')}
          aria-label={t('common.back')}
          className="flex shrink-0 items-center gap-0.5 rounded-md px-1.5 py-1 text-[13px] text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
        >
          <ChevronLeft className="size-4" />
          <span className="hidden md:inline">{t('common.back')}</span>
        </button>
        {detail && (
          <>
            <span className="min-w-0 truncate text-[15px] font-semibold text-text-1">{detail.title}</span>
            <StatusBadge status={detail.status} t={t} />
            <div className="ml-auto flex shrink-0 items-center gap-2">
              <Button size="sm" disabled={done || busy} onClick={handleExecute} title={t('automation.executeNow')}>
                <Play className="size-3.5" />
                <span className="hidden md:inline">{t('automation.executeNow')}</span>
              </Button>
              {!done && (
                <Button
                  size="sm"
                  variant="outline"
                  disabled={busy}
                  onClick={handleToggleStatus}
                  title={detail.status === AutomationStatus.Enabled ? t('automation.disable') : t('automation.enable')}
                >
                  <Power className="size-3.5" />
                  <span className="hidden md:inline">
                    {detail.status === AutomationStatus.Enabled ? t('automation.disable') : t('automation.enable')}
                  </span>
                </Button>
              )}
              <Button size="sm" variant="destructive" onClick={() => setDeleteOpen(true)} title={t('automation.delete')}>
                <Trash2 className="size-3.5" />
                <span className="hidden md:inline">{t('automation.delete')}</span>
              </Button>
            </div>
          </>
        )}
      </div>

      {/* Body */}
      {state === 'loading' && <DetailSkeleton />}
      {state === 'error' && (
        <div className="flex flex-1 flex-col items-center justify-center py-24">
          <AlertCircle className="size-8 text-text-muted" />
          <p className="mt-3 text-sm text-text-2">{t('automation.loadDetailFailed')}</p>
          <Button variant="outline" size="default" className="mt-4" onClick={fetchDetail}>
            <RefreshCw className="size-4" />
            {t('common.retry')}
          </Button>
        </div>
      )}
      {state === 'data' && detail && (
        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-4 md:flex-row">
          {/* Left main */}
          <div className="flex min-w-0 flex-1 flex-col gap-4">
            <div className="rounded-lg border border-border-custom bg-bg-base">
              <div className="flex items-center justify-between border-b border-border-custom px-3.5 py-2.5">
                <span className="text-[13px] font-semibold text-text-1">{t('automation.requirement')}</span>
                <button
                  onClick={openReqEdit}
                  title={t('common.edit')}
                  className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
                >
                  <Pencil className="size-3.5" />
                  {t('common.edit')}
                </button>
              </div>
              <div className="px-3.5 py-3">
                <p className="text-[13px] leading-relaxed text-text-2">
                  {detail.requirement || t('automation.none')}
                </p>
              </div>
            </div>

            <div className="flex min-h-0 flex-1 flex-col rounded-lg border border-border-custom bg-bg-base">
              <div className="flex items-center justify-between border-b border-border-custom px-3.5 py-2.5">
                <span className="text-[13px] font-semibold text-text-1">{t('automation.executions')}</span>
                <span className="text-xs text-text-3">{t('automation.runs').replace('{n}', String(execTotalCount))}</span>
              </div>
              <div className="flex-1 overflow-y-auto p-3">
                {executions.length === 0 && (
                  <p className="py-10 text-center text-sm text-text-3">{t('automation.emptyExecutions')}</p>
                )}
                {executions.map((exec) => {
                  const expanded = expandedId === exec.id
                  const qa = qaCache[exec.id]
                  return (
                    <div
                      key={exec.id}
                      className={cn(
                        'mb-2 rounded-lg border px-3 py-2.5 transition-colors',
                        expanded ? 'border-interactive/40 bg-interactive-soft/40' : 'border-border-custom',
                      )}
                    >
                      <button
                        className="flex w-full items-center justify-between gap-2 text-left"
                        onClick={() => handleToggleExpand(exec)}
                      >
                        <span className="text-[12.5px] font-medium text-text-1">{formatDateTime(exec.startedAt)}</span>
                        <span className="flex items-center gap-3 text-xs text-text-3">
                          {t('automation.duration')} {formatDuration(exec.startedAt, exec.finishedAt, t)}
                          <ChevronDown className={cn('size-3.5 transition-transform', expanded && 'rotate-180')} />
                        </span>
                      </button>
                      {expanded && (
                        <div className="mt-2.5 flex flex-col gap-2 border-t border-dashed border-border-custom pt-2.5">
                          {!qa ? (
                            <span className="py-2 text-center text-xs text-text-3">{t('common.loading')}</span>
                          ) : (
                            <>
                              <div className="max-w-[92%] self-end rounded-lg bg-interactive-soft px-2.5 py-1.5 text-[12.5px] leading-relaxed text-text-1">
                                {qa.question ? (
                                  <Markdown mode="static" className="[&_p]:m-0 [&_p]:whitespace-pre-wrap">
                                    {qa.question}
                                  </Markdown>
                                ) : (
                                  t('automation.emptyQA')
                                )}
                              </div>
                              <div className="max-w-[92%] self-start rounded-lg border border-border-custom bg-bg-layer-1 px-2.5 py-1.5 text-[12.5px] leading-relaxed text-text-2">
                                {qa.answer ? (
                                  <Markdown mode="static" className="[&_p]:m-0 [&_p]:whitespace-pre-wrap">
                                    {qa.answer}
                                  </Markdown>
                                ) : (
                                  t('automation.emptyQA')
                                )}
                              </div>
                            </>
                          )}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
              <div className="flex items-center justify-end gap-3 border-t border-border-custom px-3 py-2">
                <span className="text-xs text-text-3">
                  {t('automation.page').replace('{page}', String(execPage)).replace('{total}', String(execTotalPage))}
                </span>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={execPage <= 1}
                  onClick={() => fetchExecutions(execPage - 1)}
                >
                  {t('automation.prevPage')}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={execPage >= execTotalPage}
                  onClick={() => fetchExecutions(execPage + 1)}
                >
                  {t('automation.nextPage')}
                </Button>
              </div>
            </div>
          </div>

          {/* Right sidebar */}
          <div className="flex w-full flex-col gap-4 md:w-64 md:shrink-0">
            <div className="rounded-lg border border-border-custom bg-bg-base">
              <div className="border-b border-border-custom px-3.5 py-2.5 text-[13px] font-semibold text-text-1">
                {t('automation.schedule')}
              </div>
              <div className="space-y-1 px-3.5 py-3">
                <KV k="TaskID" v={String(detail.id)} />
                <KV k={t('automation.execType')} v={execTypeLabel(detail.execType, t)} />
                <KV k={t('automation.execTime')} v={scheduleTimeValue(detail, t)} />
                <KV k={t('automation.nextRunAt')} v={formatDateTime(detail.nextRunAt)} />
                <KV k={t('automation.createdAt')} v={formatDate(detail.created)} />
              </div>
            </div>

            <div className="rounded-lg border border-border-custom bg-bg-base">
              <div className="border-b border-border-custom px-3.5 py-2.5 text-[13px] font-semibold text-text-1">
                {t('automation.runConfig')}
              </div>
              <div className="space-y-1 px-3.5 py-3">
                <KV
                  k={t('automation.model')}
                  v={detail.aiModelName || t('automation.defaultModel')}
                />
                <KV
                  k={t('automation.persona')}
                  v={detail.personaName || t('automation.none')}
                />
                {(detail.sharedProjectId > 0 || detail.project) && (
                  <KV
                    k={detail.sharedProjectId > 0 ? t('automation.teamProject') : t('automation.project')}
                    v={
                      detail.sharedProjectId > 0
                        ? detail.sharedProjectName || t('automation.none')
                        : detail.project
                    }
                  />
                )}
                {(detail.mcpNames.length > 0 || detail.skillNames.length > 0) && (
                  <div className="flex flex-wrap gap-1 pt-1.5">
                    {detail.mcpNames.map((n) => (
                      <span key={`mcp-${n}`} className="rounded-sm bg-interactive/10 px-1.5 py-0.5 text-[11px] text-interactive">
                        @{n}
                      </span>
                    ))}
                    {detail.skillNames.map((n) => (
                      <span key={`skill-${n}`} className="rounded-sm bg-highlight-soft px-1.5 py-0.5 text-[11px] text-highlight">
                        {n}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Requirement edit dialog */}
      <Dialog open={reqEditOpen} onOpenChange={setReqEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('automation.editRequirement')}</DialogTitle>
            <DialogDescription>{t('automation.editRequirementDesc')}</DialogDescription>
          </DialogHeader>
          <textarea
            value={reqValue}
            onChange={(e) => setReqValue(e.target.value)}
            rows={8}
            spellCheck={false}
            className="w-full resize-none rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
          />
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" size="default" onClick={() => setReqEditOpen(false)} disabled={reqSaving}>
              {t('common.cancel')}
            </Button>
            <Button size="default" onClick={handleSaveRequirement} disabled={reqSaving}>
              {t('common.save')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('automation.confirmDeleteTitle')}</DialogTitle>
            <DialogDescription>
              {t('automation.confirmDeleteDesc').replace('{name}', detail?.title ?? '')}
            </DialogDescription>
          </DialogHeader>
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" size="default" onClick={() => setDeleteOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" size="default" onClick={handleDeleteConfirm}>
              {t('automation.delete')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="flex items-start justify-between gap-3 py-0.5">
      <span className="shrink-0 text-xs text-text-3">{k}</span>
      <span className="text-right text-xs text-text-1">{v}</span>
    </div>
  )
}

function DetailSkeleton() {
  return (
    <div className="flex flex-1 flex-col gap-4 overflow-hidden p-4 md:flex-row">
      <div className="flex-1 space-y-4">
        <div className="h-28 rounded-lg bg-bg-layer-3 animate-pulse" />
        <div className="h-72 rounded-lg bg-bg-layer-3 animate-pulse" />
      </div>
      <div className="w-full space-y-4 md:w-64">
        <div className="h-40 rounded-lg bg-bg-layer-3 animate-pulse" />
        <div className="h-44 rounded-lg bg-bg-layer-3 animate-pulse" />
      </div>
    </div>
  )
}
