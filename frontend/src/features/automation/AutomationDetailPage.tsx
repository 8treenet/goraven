import { useState, useCallback, useEffect, useLayoutEffect, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { ChevronLeft, ChevronDown, Clock, FileText, History, Menu, Play, Pencil, Power, SlidersHorizontal, Trash2, Lock, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Tag } from '@/components/ui/tag'
import { Skeleton } from '@/components/ui/skeleton'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { ErrorState } from '@/components/patterns/ErrorState'
import { Icon } from '@/components/common/Icon'
import { useSidebarStore } from '@/stores/sidebar-store'
import { providersApi } from '@/api'
import { useT, t as translate } from '@/i18n'
import { cn } from '@/lib/utils'
import { TONE_SOFT, TONE_TEXT } from '@/lib/tone'
import { AutomationStatus } from '@/api/types'
import type { AutomationTaskDetail, AutomationExecutionItem, AutomationAnswerRsp, ModelInfo } from '@/api/types'
import {
  getAutomationTask,
  getTaskExecutions,
  getExecutionAnswer,
  updateTaskStatus,
  updateTaskRequirement,
  updateTaskModel,
  deleteTask,
  executeTask,
} from '@/api/automation'
import { StatusBadge, execTypeLabel, scheduleTimeValue, formatDateTime, formatDate, formatDuration, ExecTypeIcon, execTypeTone } from './shared'
import { Markdown } from '@/components/common/markdown'

type PageState = 'loading' | 'data' | 'error'

/** 需求与执行记录共用一个展开态：同一时刻最多展开一项 */
type ExpandedKey = number | 'requirement' | null

const EXEC_PAGE_SIZE = 10

const MARKDOWN_TEXT_CLASS = 'text-body leading-[1.7] text-text-1 [&_p]:m-0 [&_p]:whitespace-pre-wrap'

/* ============================================
   Page
   ============================================ */

export function Component() {
  const { id } = useParams()
  const taskId = Number(id)
  const navigate = useNavigate()
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)

  const [state, setState] = useState<PageState>('loading')
  const [detail, setDetail] = useState<AutomationTaskDetail | null>(null)
  const [executions, setExecutions] = useState<AutomationExecutionItem[]>([])
  const [execPage, setExecPage] = useState(1)
  const [execTotalPage, setExecTotalPage] = useState(1)
  const [execTotalCount, setExecTotalCount] = useState(0)
  const [expandedKey, setExpandedKey] = useState<ExpandedKey>(null)
  const [reqOverflow, setReqOverflow] = useState(false)
  const [qaCache, setQaCache] = useState<Record<number, AutomationAnswerRsp>>({})
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [reqEditOpen, setReqEditOpen] = useState(false)
  const [reqValue, setReqValue] = useState('')
  const [reqSaving, setReqSaving] = useState(false)
  const reqPreviewRef = useRef<HTMLDivElement>(null)
  const requirementExpanded = expandedKey === 'requirement'

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

  // 收起态下测量需求是否超出预览高度，决定是否显示折叠箭头
  useLayoutEffect(() => {
    if (requirementExpanded) return
    const el = reqPreviewRef.current
    if (!el) return
    setReqOverflow(el.scrollHeight > el.clientHeight + 1)
  }, [detail?.requirement, requirementExpanded, state])

  const handleToggleRequirement = useCallback(() => {
    setExpandedKey((prev) => (prev === 'requirement' ? null : 'requirement'))
  }, [])

  const handleToggleExpand = useCallback(async (exec: AutomationExecutionItem) => {
    if (expandedKey === exec.id) {
      setExpandedKey(null)
      return
    }
    setExpandedKey(exec.id)
    if (!qaCache[exec.id]) {
      try {
        const qa = await getExecutionAnswer(taskId, exec.id)
        setQaCache((prev) => ({ ...prev, [exec.id]: qa }))
      } catch {
        toast.error(translate('common.loadFailed'))
      }
    }
  }, [expandedKey, qaCache, taskId])

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
      setExpandedKey(null)
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
    <div className="flex h-full min-h-0 flex-col bg-bg-base">
      <PageHeader
        title={
          <span className="flex min-w-0 items-center gap-2">
            {detail && (
              <span className={cn('flex size-8 shrink-0 items-center justify-center rounded-lg', TONE_SOFT[execTypeTone(detail.execType)], TONE_TEXT[execTypeTone(detail.execType)])}>
                <ExecTypeIcon execType={detail.execType} className="size-4" />
              </span>
            )}
            <span className="truncate">{detail?.title || t('automation.title')}</span>
          </span>
        }
        className="sticky top-0 z-10"
        actions={
          <>
            <Button
              variant="ghost"
              size="icon"
              onClick={openMobile}
              aria-label={t('sidebar.menu')}
              className="md:hidden"
            >
              <Menu className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate('/automation')}
              title={t('common.back')}
              aria-label={t('common.back')}
            >
              <ChevronLeft className="size-4" />
            </Button>
            {detail && (
              <>
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
                    className={
                      detail.status === AutomationStatus.Enabled
                        ? undefined
                        : 'border-success/30 bg-success-soft text-success hover:border-success/40 hover:bg-success/20 hover:text-success'
                    }
                  >
                    <Power className="size-3.5" />
                    <span className="hidden md:inline">
                      {detail.status === AutomationStatus.Enabled ? t('automation.disable') : t('automation.enable')}
                    </span>
                  </Button>
                )}
                <Button size="sm" variant="outline" onClick={openReqEdit} title={t('common.edit')}>
                  <Pencil className="size-3.5" />
                  <span className="hidden md:inline">{t('common.edit')}</span>
                </Button>
                <Button size="sm" variant="destructive" onClick={() => setDeleteOpen(true)} title={t('automation.delete')}>
                  <Trash2 className="size-3.5" />
                  <span className="hidden md:inline">{t('automation.delete')}</span>
                </Button>
              </>
            )}
          </>
        }
      />

      {/* Body */}
      <div className="min-h-0 flex-1 overflow-y-auto">
        {state === 'loading' && (
          <PageContainer>
            <DetailSkeleton />
          </PageContainer>
        )}
        {state === 'error' && (
          <ErrorState description={t('automation.loadDetailFailed')} onRetry={fetchDetail} />
        )}
        {state === 'data' && detail && (
          <PageContainer className="flex flex-col gap-4 md:flex-row">
            {/* Left main */}
            <div className="flex min-w-0 flex-1 flex-col gap-4">
              <Card>
                <div className="flex items-center justify-between gap-3 border-b border-border px-5 py-3">
                  <button
                    onClick={reqOverflow || requirementExpanded ? handleToggleRequirement : undefined}
                    aria-expanded={requirementExpanded}
                    className="flex min-w-0 items-center gap-1.5 rounded-md text-left"
                  >
                    <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[3], TONE_TEXT[3])}>
                      <FileText className="size-4" />
                    </span>
                    <span className="text-section font-semibold text-text-1">{t('automation.requirement')}</span>
                    {(reqOverflow || requirementExpanded) && (
                      <ChevronDown
                        className={cn(
                          'size-3.5 shrink-0 text-text-3 transition-transform',
                          requirementExpanded && 'rotate-180',
                        )}
                      />
                    )}
                  </button>
                  <button
                    onClick={openReqEdit}
                    title={t('common.edit')}
                    className="flex items-center gap-1 rounded-md px-1.5 py-0.5 text-meta text-accent transition-colors hover:bg-accent-soft hover:text-accent"
                  >
                    <Pencil className="size-3.5" />
                    {t('common.edit')}
                  </button>
                </div>
                <div className="px-5 py-4">
                  <div
                    ref={reqPreviewRef}
                    onClick={!requirementExpanded && reqOverflow ? handleToggleRequirement : undefined}
                    className={cn(
                      'relative',
                      !requirementExpanded && 'max-h-[72px] overflow-hidden',
                      !requirementExpanded && reqOverflow && 'cursor-pointer',
                    )}
                  >
                    <div className={cn(!requirementExpanded && 'pointer-events-none')}>
                      <Markdown mode="static" className={MARKDOWN_TEXT_CLASS}>
                        {detail.requirement || t('automation.none')}
                      </Markdown>
                    </div>
                    {!requirementExpanded && reqOverflow && (
                      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-gradient-to-t from-bg-layer-1 to-transparent" />
                    )}
                  </div>
                </div>
              </Card>

              <Card className="flex min-h-[280px] flex-1 flex-col">
                <div className="flex items-center justify-between border-b border-border px-5 py-3">
                  <div className="flex items-center gap-2">
                    <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[1], TONE_TEXT[1])}>
                      <History className="size-4" />
                    </span>
                    <span className="text-section font-semibold text-text-1">{t('automation.executions')}</span>
                  </div>
                  <span className="text-meta text-text-2">{t('automation.runs').replace('{n}', String(execTotalCount))}</span>
                </div>
                <div className="flex-1 overflow-y-auto">
                  {executions.length === 0 && (
                    <p className="py-10 text-center text-meta text-text-3">{t('automation.emptyExecutions')}</p>
                  )}
                  {executions.map((exec) => {
                    const expanded = expandedKey === exec.id
                    const qa = qaCache[exec.id]
                    return (
                      <div
                        key={exec.id}
                        className="border-b border-border last:border-b-0"
                      >
                        <button
                          className="flex min-h-[44px] w-full items-center justify-between gap-2 px-5 py-2.5 text-left"
                          onClick={() => handleToggleExpand(exec)}
                        >
                          <span className="flex items-center gap-1.5 text-meta font-medium tabular-nums text-text-1">
                            <Clock className="size-3.5 shrink-0 text-cat-1" />
                            {formatDateTime(exec.startedAt)}
                          </span>
                          <span className="flex items-center gap-3 text-meta text-text-2">
                            {t('automation.duration')} {formatDuration(exec.startedAt, exec.finishedAt, t)}
                            <ChevronDown className={cn('size-3.5 transition-transform', expanded && 'rotate-180')} />
                          </span>
                        </button>
                        {expanded && (
                          <div className="border-t border-dashed border-border px-5 py-3">
                            {!qa ? (
                              <span className="block py-2 text-center text-meta text-text-3">{t('common.loading')}</span>
                            ) : qa.answer ? (
                              <Markdown
                                mode="static"
                                className={MARKDOWN_TEXT_CLASS}
                              >
                                {qa.answer}
                              </Markdown>
                            ) : (
                              <p className="py-2 text-center text-meta text-text-3">{t('automation.emptyAnswer')}</p>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
                <div className="flex items-center justify-end gap-3 border-t border-border px-5 py-2.5">
                  <span className="text-meta text-text-3">
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
              </Card>
            </div>

            {/* Right sidebar */}
            <div className="flex w-full flex-col gap-4 md:w-80 md:shrink-0">
              <Card>
                <CardHeader className="pb-3">
                  <div className="flex items-center gap-2">
                    <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[execTypeTone(detail.execType)], TONE_TEXT[execTypeTone(detail.execType)])}>
                      <ExecTypeIcon execType={detail.execType} className="size-4" />
                    </span>
                    <CardTitle>{t('automation.schedule')}</CardTitle>
                    <StatusBadge status={detail.status} t={t} />
                  </div>
                </CardHeader>
                <CardContent className="space-y-2 pt-0">
                  <KV k="TaskID" v={String(detail.id)} />
                  <KV k={t('automation.execType')} v={execTypeLabel(detail.execType, t)} />
                  <KV k={t('automation.execTime')} v={scheduleTimeValue(detail, t)} />
                  <KV k={t('automation.nextRunAt')} v={formatDateTime(detail.nextRunAt)} />
                  <KV k={t('automation.createdAt')} v={formatDate(detail.created)} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-3">
                  <div className="flex items-center gap-2">
                    <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[2], TONE_TEXT[2])}>
                      <SlidersHorizontal className="size-4" />
                    </span>
                    <CardTitle>{t('automation.runConfig')}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2 pt-0">
                  <ModelRow
                    detail={detail}
                    onChanged={(aiModelId, aiModelName) =>
                      setDetail((prev) => (prev ? { ...prev, aiModelId, aiModelName } : prev))
                    }
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
                    <div className="flex flex-wrap gap-1.5 pt-1.5">
                      {detail.mcpNames.map((n) => (
                        <Tag key={`mcp-${n}`} tone={2} className="text-meta">@{n}</Tag>
                      ))}
                      {detail.skillNames.map((n) => (
                        <Tag key={`skill-${n}`} tone={4} className="text-meta">{n}</Tag>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </PageContainer>
        )}
      </div>

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
            className="w-full resize-none rounded-lg border border-border bg-transparent px-3 py-2 text-[13px] leading-[1.65] transition-colors outline-none placeholder:text-text-muted focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
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
      <span className="shrink-0 text-meta text-text-3">{k}</span>
      <span className="text-right text-meta text-text-1">{v}</span>
    </div>
  )
}

function ModelIcon({ icon }: { icon?: string }) {
  const [err, setErr] = useState(false)
  if (!icon || err) return <Icon name="brain" className="size-3.5 shrink-0 text-text-2" />
  return <img src={icon} alt="" className="size-3.5 shrink-0 rounded object-cover" onError={() => setErr(true)} />
}

/** 模型行：未选角色且未完成时可下拉修改，否则展示模型名并锁定（模型由角色/状态决定） */
function ModelRow({
  detail,
  onChanged,
}: {
  detail: AutomationTaskDetail
  onChanged: (aiModelId: number, aiModelName: string) => void
}) {
  const t = useT()
  const lockedByPersona = detail.personaId > 0
  const lockedDone = detail.status === AutomationStatus.Done
  const locked = lockedByPersona || lockedDone
  const [models, setModels] = useState<ModelInfo[]>([])
  const [open, setOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (models.length > 0) return
    providersApi.getAvailableModels().then(setModels).catch(() => {})
  }, [models.length])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const displayName = detail.aiModelName || t('automation.defaultModel')
  const modelIcon =
    models.find((m) => m.aiModelId === detail.aiModelId)?.icon ??
    models.find((m) => m.displayName === detail.aiModelName)?.icon ??
    (detail.aiModelId === 0 ? '/favicon.svg' : undefined)
  const lockHint = lockedByPersona ? t('automation.modelLockedByPersona') : t('automation.modelLockedDone')

  if (locked) {
    return (
      <div className="flex items-start justify-between gap-3 py-0.5">
        <span className="shrink-0 text-meta text-text-3">{t('automation.model')}</span>
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="flex min-w-0 flex-1 items-center justify-end gap-1 text-meta text-text-1">
              <ModelIcon icon={modelIcon} />
              <span className="min-w-0 truncate" title={displayName}>{displayName}</span>
              <Lock className="size-3 shrink-0 text-text-muted" />
            </span>
          </TooltipTrigger>
          <TooltipContent className="max-w-56">{lockHint}</TooltipContent>
        </Tooltip>
      </div>
    )
  }

  const handleSelect = async (aiModelId: number) => {
    setOpen(false)
    if (aiModelId === detail.aiModelId) return
    setSaving(true)
    try {
      await updateTaskModel(detail.id, aiModelId)
      const model = models.find((m) => m.aiModelId === aiModelId)
      onChanged(aiModelId, model ? model.displayName : '')
      toast.success(t('automation.modelUpdated'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.failed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex items-start justify-between gap-3 py-0.5">
      <span className="shrink-0 text-meta text-text-3">{t('automation.model')}</span>
      <div ref={ref} className="relative min-w-0 flex-1">
        <button
          onClick={() => setOpen(!open)}
          disabled={saving}
          title={t('automation.changeModel')}
          className={cn(
            'flex w-full min-w-0 items-center justify-end gap-1 rounded px-1 -mx-1 text-meta text-text-1 transition-colors hover:bg-bg-layer-2',
            open && 'bg-bg-layer-2',
            saving && 'opacity-60',
          )}
        >
          <ModelIcon icon={modelIcon} />
          <span className="min-w-0 truncate" title={displayName}>{displayName}</span>
          {saving ? (
            <Loader2 className="size-3 shrink-0 animate-spin" />
          ) : (
            <ChevronDown className={cn('size-3 shrink-0 transition-transform', open && 'rotate-180')} />
          )}
        </button>
        {open && (
          <div className="absolute right-0 top-full z-40 mt-1 max-h-72 w-64 max-w-[calc(100vw-2rem)] overflow-y-auto rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
            <button
              onClick={() => handleSelect(0)}
              className={cn(
                'flex w-full items-center gap-2 px-3 py-1.5 text-left text-meta transition-colors hover:bg-bg-layer-2',
                detail.aiModelId === 0 ? 'text-text-1' : 'text-text-2',
              )}
            >
              <span className={cn('shrink-0 text-sm', detail.aiModelId === 0 ? 'text-accent' : 'text-text-muted opacity-0')}>
                {'✓'}
              </span>
              <span className="truncate">{t('automation.defaultModel')}</span>
            </button>
            {models.map((m) => (
              <button
                key={m.aiModelId}
                onClick={() => handleSelect(m.aiModelId)}
                className={cn(
                  'flex w-full items-center gap-2 px-3 py-1.5 text-left text-meta transition-colors hover:bg-bg-layer-2',
                  detail.aiModelId === m.aiModelId ? 'text-text-1' : 'text-text-2',
                )}
              >
                <span className={cn('shrink-0 text-sm', detail.aiModelId === m.aiModelId ? 'text-accent' : 'text-text-muted opacity-0')}>
                  {'✓'}
                </span>
                <ModelIcon icon={m.icon} />
                <span className="truncate">{m.displayName}</span>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-4 md:flex-row">
      <div className="flex flex-1 flex-col gap-4">
        <Skeleton className="h-32 rounded-xl" />
        <Skeleton className="h-80 rounded-xl" />
      </div>
      <div className="flex w-full flex-col gap-4 md:w-72">
        <Skeleton className="h-44 rounded-xl" />
        <Skeleton className="h-48 rounded-xl" />
      </div>
    </div>
  )
}
