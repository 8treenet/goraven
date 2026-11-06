import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Play, Ellipsis, Trash2, CalendarClock, HelpCircle, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
import { EmptyState } from '@/components/patterns/EmptyState'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { PageTabs } from '@/components/patterns/PageTabs'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { useT, t as translate } from '@/i18n'
import { cn } from '@/lib/utils'
import { TONE_SOFT, TONE_TEXT } from '@/lib/tone'
import { useSidebarStore } from '@/stores/sidebar-store'
import { AutomationStatus } from '@/api/types'
import type { AutomationTaskItem } from '@/api/types'
import { getAutomationTasks, updateTaskStatus, deleteTask, executeTask } from '@/api/automation'
import { StatusBadge, scheduleLabel, nextRunText, Pagination, ExecTypeIcon, execTypeTone, statusRailClass } from './shared'

const PAGE_SIZE = 10

type PageState = 'loading' | 'data' | 'empty' | 'error'
type TabKey = 'all' | 'enabled' | 'disabled' | 'done'

const TABS: { key: TabKey; labelKey: 'automation.tabAll' | 'automation.tabEnabled' | 'automation.tabDisabled' | 'automation.tabDone' }[] = [
  { key: 'all', labelKey: 'automation.tabAll' },
  { key: 'enabled', labelKey: 'automation.tabEnabled' },
  { key: 'disabled', labelKey: 'automation.tabDisabled' },
  { key: 'done', labelKey: 'automation.tabDone' },
]

/** 各页签对应的后端状态过滤值，all 不筛选 */
const TAB_STATUS: Record<TabKey, number | undefined> = {
  all: undefined,
  enabled: AutomationStatus.Enabled,
  disabled: AutomationStatus.Disabled,
  done: AutomationStatus.Done,
}

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)
  const [state, setState] = useState<PageState>('loading')
  const [tasks, setTasks] = useState<AutomationTaskItem[]>([])
  const [tab, setTab] = useState<TabKey>('all')
  const [page, setPage] = useState(1)
  const [totalPage, setTotalPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [menuOpenId, setMenuOpenId] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AutomationTaskItem | null>(null)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [hintOpen, setHintOpen] = useState(false)

  const fetchTasks = useCallback((targetTab: TabKey, targetPage: number, silent = false) => {
    if (!silent) setState('loading')
    getAutomationTasks({ page: targetPage, pageSize: PAGE_SIZE, status: TAB_STATUS[targetTab] })
      .then((rsp) => {
        setTasks(rsp.list)
        setPage(rsp.page)
        setTotalPage(rsp.totalPage)
        setTotalCount(rsp.totalCount)
        setState(rsp.list.length > 0 ? 'data' : 'empty')
      })
      .catch(() => setState('error'))
  }, [])

  useEffect(() => {
    fetchTasks(tab, page)
  }, [tab, page, fetchTasks])

  const patchTask = useCallback((id: number, patch: Partial<AutomationTaskItem>) => {
    setTasks((prev) => prev.map((x) => (x.id === id ? { ...x, ...patch } : x)))
  }, [])

  const handleTabChange = useCallback((key: TabKey) => {
    setTab(key)
    setPage(1)
  }, [])

  const handleToggle = useCallback(async (task: AutomationTaskItem, checked: boolean) => {
    if (task.status === AutomationStatus.Done) {
      toast.error(translate('automation.doneCannotToggle'))
      return
    }
    const target = checked ? AutomationStatus.Enabled : AutomationStatus.Disabled
    setBusyId(task.id)
    try {
      await updateTaskStatus(task.id, target)
      patchTask(task.id, { status: target })
      toast.success(target === AutomationStatus.Enabled ? translate('automation.taskEnabled') : translate('automation.taskDisabled'))
      if (target === AutomationStatus.Enabled) fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusyId(null)
    }
  }, [patchTask, fetchTasks, tab, page])

  const handleExecute = useCallback(async (task: AutomationTaskItem) => {
    setBusyId(task.id)
    try {
      await executeTask(task.id)
      toast.success(translate('automation.executionStarted'))
      fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    } finally {
      setBusyId(null)
    }
  }, [fetchTasks, tab, page])

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return
    try {
      await deleteTask(deleteTarget.id)
      setDeleteTarget(null)
      setMenuOpenId(null)
      toast.success(translate('automation.deleted'))
      fetchTasks(tab, page, true)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.deleteFailed'))
    }
  }, [deleteTarget, fetchTasks, tab, page])

  const handleRowClick = useCallback((id: number) => {
    navigate(`/automation/${id}`)
  }, [navigate])

  return (
    <div>
      <PageHeader
        title={t('automation.title')}
        className="sticky top-0 z-10"
        actions={
          <>
            <button
              onClick={openMobile}
              aria-label={t('sidebar.menu')}
              className="inline-flex size-9 shrink-0 items-center justify-center rounded-lg text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 md:hidden"
            >
              <Menu className="size-4" />
            </button>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  aria-label={t('automation.createHint')}
                  onClick={() => setHintOpen(true)}
                  className="inline-flex size-9 items-center justify-center rounded-lg text-cat-3 transition-colors hover:bg-cat-3-soft hover:text-cat-3"
                >
                  <HelpCircle className="size-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent className="max-w-72">{t('automation.createHint')}</TooltipContent>
            </Tooltip>
          </>
        }
      />

      <PageContainer className="py-0">
        {/* Status tabs */}
        <PageTabs
          className="pt-4"
          value={tab}
          onChange={handleTabChange}
          items={TABS.map(({ key, labelKey }) => ({ value: key, label: t(labelKey) }))}
        />

        {/* Content */}
        {state === 'loading' && <AutomationListSkeleton />}
      {state === 'error' && <ErrorState onRetry={() => fetchTasks(tab, page)} />}
      {state === 'empty' && (tab === 'all' ? (
        <EmptyState
          variant="card"
          tone={3}
          icon={<CalendarClock className="size-5" />}
          title={t('automation.emptyTasks')}
          description={t('automation.emptyTasksHint')}
        />
      ) : (
        <p className="py-10 text-center text-meta text-text-3">{t('automation.emptyTasks')}</p>
      ))}
      {state === 'data' && (
        <div>
          {tasks.map((task) => {
            const execTone = execTypeTone(task.execType)
            const done = task.status === AutomationStatus.Done
            const next = nextRunText(task, t)
            const meta: string[] = [scheduleLabel(task, t)]
            if (task.aiModelName) meta.push(`${t('automation.model')} ${task.aiModelName}`)
            if (task.personaName) meta.push(`${t('automation.persona')} ${task.personaName}`)
            if (task.mcpNames.length > 0) meta.push(t('automation.mcpCount').replace('{n}', String(task.mcpNames.length)))
            if (task.skillNames.length > 0) meta.push(t('automation.skillCount').replace('{n}', String(task.skillNames.length)))

            return (
              <div
                key={task.id}
                className={cn(
                  'group relative flex min-h-[72px] flex-col gap-2 border-b border-border px-5 py-3 transition-colors hover:bg-bg-layer-2 md:flex-row md:items-center md:gap-4',
                  'before:absolute before:left-0 before:top-3 before:bottom-3 before:w-0.5 before:rounded-r',
                  statusRailClass(task.status),
                )}
              >
                <button
                  className="min-w-0 flex-1 text-left"
                  onClick={() => handleRowClick(task.id)}
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[execTone], TONE_TEXT[execTone])}>
                      <ExecTypeIcon execType={task.execType} className="size-4" />
                    </span>
                    <span className="min-w-0 truncate text-body font-medium text-text-1">{task.title}</span>
                    <StatusBadge status={task.status} t={t} />
                  </span>
                  <span className="mt-1 line-clamp-2 text-meta leading-[18px] text-text-3 md:truncate">
                    {meta.join(' · ')}
                  </span>
                </button>

                <div className="flex shrink-0 items-center gap-2 md:gap-3">
                  <span className="mr-auto whitespace-nowrap text-meta tabular-nums text-text-3 md:mr-0">
                    {next.label}：{next.value}
                  </span>

                  <button
                    title={t('automation.executeNow')}
                    aria-label={t('automation.executeNow')}
                    disabled={done || busyId === task.id}
                    onClick={(e) => { e.stopPropagation(); handleExecute(task) }}
                    className={cn(
                      'flex size-7 items-center justify-center rounded-md text-text-3 opacity-100 transition-all hover:bg-bg-layer-3 hover:text-text-1 md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100',
                      (done || busyId === task.id) && 'cursor-not-allowed opacity-35 hover:bg-transparent hover:text-text-3',
                    )}
                  >
                    <Play className="size-3.5" />
                  </button>

                  <Switch
                    size="default"
                    aria-label={task.status === AutomationStatus.Enabled ? t('automation.disable') : t('automation.enable')}
                    checked={task.status === AutomationStatus.Enabled}
                    disabled={done || busyId === task.id}
                    onCheckedChange={(checked) => handleToggle(task, checked)}
                  />

                  <div className="relative opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
                    <button
                      aria-label={t('sidebar.menu')}
                      aria-haspopup="menu"
                      aria-expanded={menuOpenId === task.id}
                      onClick={(e) => { e.stopPropagation(); setMenuOpenId(menuOpenId === task.id ? null : task.id) }}
                      className="flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-3 hover:text-text-1"
                    >
                      <Ellipsis className="size-4" />
                    </button>
                    {menuOpenId === task.id && (
                      <>
                        <div className="fixed inset-0 z-10" onClick={() => setMenuOpenId(null)} />
                        <div className="absolute right-0 top-full z-20 mt-1 w-32 rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
                          <button
                            className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-destructive transition-colors hover:bg-destructive/10"
                            onClick={(e) => { e.stopPropagation(); setMenuOpenId(null); setDeleteTarget(task) }}
                          >
                            <Trash2 className="size-3.5" />
                            {t('automation.delete')}
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
          <Pagination
            page={page}
            totalPages={totalPage}
            totalCount={totalCount}
            onPageChange={setPage}
          />
        </div>
      )}
      </PageContainer>

      {/* Hint dialog */}
      <Dialog open={hintOpen} onOpenChange={setHintOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <HelpCircle className="size-4 text-text-3" />
              {t('automation.title')}
            </DialogTitle>
            <DialogDescription>{t('automation.createHint')}</DialogDescription>
          </DialogHeader>
          <div className="mt-4 flex justify-end">
            <Button variant="outline" size="default" onClick={() => setHintOpen(false)}>
              {t('common.close')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('automation.confirmDeleteTitle')}</DialogTitle>
            <DialogDescription>
              {t('automation.confirmDeleteDesc').replace('{name}', deleteTarget?.title ?? '')}
            </DialogDescription>
          </DialogHeader>
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" size="default" onClick={() => setDeleteTarget(null)}>
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

/* ============================================
   Sub-views
   ============================================ */

function AutomationListSkeleton() {
  return (
    <Card className="overflow-hidden">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="flex min-h-[72px] flex-col gap-2 border-b border-border px-5 py-3 last:border-b-0 md:flex-row md:items-center md:gap-4">
          <div className="flex-1 space-y-2">
            <Skeleton className="block h-4 w-40 rounded" />
            <Skeleton className="block h-3 w-72 rounded" />
          </div>
          <Skeleton className="h-5 w-14 rounded-full" />
          <Skeleton className="h-5 w-9 rounded-full" />
        </div>
      ))}
    </Card>
  )
}
