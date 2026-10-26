import { useMemo } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { AutomationExecType, AutomationStatus } from '@/api/types'
import type { AutomationTaskItem } from '@/api/types'
import type { TranslationKey } from '@/i18n'
import { useT } from '@/i18n'
import { cn } from '@/lib/utils'

/* ============================================
   自动化任务共享展示组件与格式化工具
   ============================================ */

type T = (key: TranslationKey) => string

export function formatDateTime(iso: string | null): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return '-'
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatDate(iso: string | null): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return '-'
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function formatDuration(startIso: string, endIso: string, t: T): string {
  const ms = new Date(endIso).getTime() - new Date(startIso).getTime()
  if (isNaN(ms) || ms < 0) return '-'
  const totalSec = Math.round(ms / 1000)
  const min = Math.floor(totalSec / 60)
  const sec = totalSec % 60
  return t('automation.durationValue')
    .replace('{min}', String(min))
    .replace('{sec}', String(sec))
}

function isZeroTime(iso: string): boolean {
  const d = new Date(iso)
  return isNaN(d.getTime()) || d.getFullYear() < 2000
}

/** 列表行的下次执行文案（今天/明天/具体时间） */
export function nextRunText(task: AutomationTaskItem, t: T): { label: string; value: string } {
  if (task.status === AutomationStatus.Done) {
    return { label: t('automation.lastRun'), value: formatDateTime(task.nextRunAt) }
  }
  if (isZeroTime(task.nextRunAt)) {
    return { label: t('automation.nextRun'), value: t('automation.noNextRun') }
  }
  const next = new Date(task.nextRunAt)
  const now = new Date()
  const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate())
  const diffDays = Math.round((startOfDay(next).getTime() - startOfDay(now).getTime()) / 86_400_000)
  const pad = (n: number) => String(n).padStart(2, '0')
  const hm = `${pad(next.getHours())}:${pad(next.getMinutes())}`
  let prefix = ''
  if (diffDays === 0) prefix = t('automation.today')
  else if (diffDays === 1) prefix = t('automation.tomorrow')
  else if (diffDays > 1 && diffDays < 7) prefix = formatDate(task.nextRunAt).slice(5)
  else prefix = formatDate(task.nextRunAt)
  return { label: t('automation.nextRun'), value: `${prefix} ${hm}` }
}

const WEEKDAY_KEYS = [
  'automation.weekdaySun',
  'automation.weekdayMon',
  'automation.weekdayTue',
  'automation.weekdayWed',
  'automation.weekdayThu',
  'automation.weekdayFri',
  'automation.weekdaySat',
] as const

/** 执行类型标签 */
export function execTypeLabel(execType: number, t: T): string {
  switch (execType) {
    case AutomationExecType.Once: return t('automation.execTypeOnce')
    case AutomationExecType.Interval: return t('automation.execTypeInterval')
    case AutomationExecType.Daily: return t('automation.execTypeDaily')
    case AutomationExecType.Weekly: return t('automation.execTypeWeekly')
    default: return '-'
  }
}

/** 调度规则简述（列表 meta 标签 / 详情页执行时间） */
export function scheduleLabel(task: AutomationTaskItem, t: T): string {
  switch (task.execType) {
    case AutomationExecType.Once:
      return t('automation.onceLabel').replace('{time}', formatDateTime(task.runAt))
    case AutomationExecType.Interval:
      return t('automation.intervalLabel').replace('{n}', String(task.intervalMinutes))
    case AutomationExecType.Daily:
      return t('automation.dailyLabel').replace('{time}', task.fixedTime)
    case AutomationExecType.Weekly: {
      const weekday = t(WEEKDAY_KEYS[task.weekday % 7])
      return t('automation.weeklyLabel').replace('{weekday}', weekday).replace('{time}', task.fixedTime)
    }
    default:
      return '-'
  }
}

/** 详情页调度信息中的具体时间值 */
export function scheduleTimeValue(task: AutomationTaskItem, t: T): string {
  switch (task.execType) {
    case AutomationExecType.Once: return formatDateTime(task.runAt)
    case AutomationExecType.Interval: return t('automation.intervalLabel').replace('{n}', String(task.intervalMinutes))
    case AutomationExecType.Daily: return task.fixedTime
    case AutomationExecType.Weekly: {
      const weekday = t(WEEKDAY_KEYS[task.weekday % 7])
      return t('automation.weeklyLabel').replace('{weekday}', weekday).replace('{time}', task.fixedTime)
    }
    default: return '-'
  }
}

export function StatusBadge({ status, t, className }: { status: number; t: T; className?: string }) {
  const config = {
    [AutomationStatus.Enabled]: {
      className: 'bg-emerald-500/10 text-emerald-600',
      dot: 'bg-emerald-500',
      label: t('automation.statusEnabled'),
    },
    [AutomationStatus.Disabled]: {
      className: 'bg-bg-layer-3 text-text-3',
      dot: 'bg-text-3',
      label: t('automation.statusDisabled'),
    },
    [AutomationStatus.Done]: {
      className: 'bg-interactive-soft text-interactive',
      dot: 'bg-interactive',
      label: t('automation.statusDone'),
    },
  }[status] ?? {
    className: 'bg-bg-layer-3 text-text-3',
    dot: 'bg-text-3',
    label: '-',
  }

  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-medium',
        config.className,
        className,
      )}
    >
      <span className={cn('size-1.5 rounded-full', config.dot)} />
      {config.label}
    </span>
  )
}

/* ============================================
   Pagination
   ============================================ */

export function Pagination({
  page,
  totalPages,
  totalCount,
  onPageChange,
}: {
  page: number
  totalPages: number
  totalCount: number
  onPageChange: (p: number) => void
}) {
  const t = useT()
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
    <div className="flex items-center justify-between border-t border-border-custom px-4 py-2">
      <span className="hidden text-xs text-text-3 md:inline">{t('automation.totalTasks').replace('{n}', String(totalCount))}</span>
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
