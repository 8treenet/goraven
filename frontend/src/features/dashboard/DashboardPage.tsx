import { useState, useCallback, useMemo, useEffect } from 'react'
import {
  RefreshCw,
  AlertCircle,
  Zap,
  CalendarRange,
  Database,
  MessageSquare,
  Sparkles,
  HardDrive,
  Inbox,
  type LucideIcon,
} from 'lucide-react'
import { useT } from '@/i18n'
import { formatNumber, formatBytes } from '@/lib/format'
import { SegmentedControl } from '@/components/charts/SegmentedControl'
import { ChartTooltip } from '@/components/charts/ChartTooltip'
import { RankingPanel } from '@/components/charts/RankingPanel'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Cell,
} from 'recharts'
import { getDashboard, getTokenTrend, getModelUsage } from '@/api/dashboard'
import type { DashboardData, TokenTrendItem, DashboardOverview, StorageStats, ModelUsageItem } from '@/api'

/* ============================================
   Helpers
   ============================================ */

function computeTrendStats(data: TokenTrendItem[]) {
  if (data.length === 0) {
    return { avg: 0, peakDate: '', peakTotal: 0 }
  }
  const total = data.reduce((s, d) => s + d.promptTokens + d.completionTokens, 0)
  const avg = Math.round(total / data.length)
  let peak = data[0]
  for (const d of data) {
    if (d.promptTokens + d.completionTokens > (peak.promptTokens + peak.completionTokens)) {
      peak = d
    }
  }
  return { avg, peakDate: peak.date, peakTotal: peak.promptTokens + peak.completionTokens }
}

function useChartColors() {
  return useMemo(() => {
    const style = getComputedStyle(document.documentElement)
    return [
      style.getPropertyValue('--chart-1').trim() || '#d99a2c',
      style.getPropertyValue('--chart-2').trim() || '#4a6fa5',
      style.getPropertyValue('--chart-3').trim() || '#c9782e',
      style.getPropertyValue('--chart-4').trim() || '#3a8a8a',
      style.getPropertyValue('--chart-5').trim() || '#6b9e6b',
    ]
  }, [])
}

type TrendDays = 7 | 30 | 90

/* ============================================
   Shared chart axis styles
   ============================================ */

const axisStyle = {
  fontSize: 10,
}

/* ============================================
   Dashboard Skeleton
   ============================================ */

function DashboardSkeleton() {
  return (
    <div className="flex flex-1 flex-col">
      <div className="flex h-10 shrink-0 items-center border-b border-border px-4">
        <div className="h-4 w-28 animate-pulse rounded bg-bg-layer-2" />
      </div>

      <div className="flex flex-1 flex-col gap-2 p-2">
        {/* Row 1 */}
        <div className="flex min-h-[100px] gap-2">
          <div className="flex w-1/2 flex-col gap-3 rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
            <div className="flex flex-1 gap-4">
              <div className="flex flex-1 flex-col justify-center gap-2.5 rounded-lg bg-bg-layer-2/50 px-4 py-3">
                <div className="h-3 w-14 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-7 w-20 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-1.5 w-full animate-pulse rounded-full bg-bg-layer-3" />
              </div>
              <div className="grid grid-cols-2 content-center gap-x-5 gap-y-3">
                {[1, 2, 3, 4].map((i) => (
                  <div key={i} className="flex items-center gap-2.5">
                    <div className="size-7 animate-pulse rounded-md bg-bg-layer-2" />
                    <div className="h-3 w-12 animate-pulse rounded bg-bg-layer-2" />
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="flex w-1/2 flex-col gap-3 rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
            <div className="flex flex-1 gap-4">
              <div className="flex flex-1 flex-col justify-center gap-2.5 rounded-lg bg-bg-layer-2/50 px-4 py-3">
                <div className="h-3 w-10 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-7 w-20 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-1.5 w-full animate-pulse rounded-full bg-bg-layer-3" />
              </div>
              <div className="flex flex-col justify-center gap-3">
                {[1, 2].map((i) => (
                  <div key={i} className="flex items-center gap-2.5">
                    <div className="size-7 animate-pulse rounded-md bg-bg-layer-2" />
                    <div className="h-3 w-12 animate-pulse rounded bg-bg-layer-2" />
                  </div>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-4 gap-x-3 gap-y-2 border-t border-border pt-2.5">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="h-3 animate-pulse rounded bg-bg-layer-2" />
              ))}
            </div>
          </div>
        </div>

        {/* Row 2 */}
        <div className="flex flex-1 gap-2 min-h-[220px]">
          <div className="flex w-2/3 flex-col rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="flex items-center justify-between">
              <div className="h-3 w-28 animate-pulse rounded bg-bg-layer-2" />
              <div className="h-6 w-40 animate-pulse rounded bg-bg-layer-2" />
            </div>
            <div className="mt-4 flex-1 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="flex w-1/3 flex-col rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
            <div className="mt-4 flex-1 space-y-2">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-4 animate-pulse rounded bg-bg-layer-2" style={{ width: `${100 - i * 15}%` }} />
              ))}
            </div>
          </div>
        </div>

        {/* Row 3 */}
        <div className="flex flex-1 min-h-[200px] gap-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-6 py-4"
            >
              <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
              <div className="mt-4 flex-1 space-y-2">
                {[1, 2, 3, 4, 5].map((j) => (
                  <div
                    key={j}
                    className="h-3 animate-pulse rounded bg-bg-layer-2"
                    style={{ width: `${100 - j * 12}%` }}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

/* ============================================
   Dashboard Error
   ============================================ */

function DashboardError({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">
          {t('dashboard.errLoad')}
        </p>
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
   Panel: Token Overview
   ============================================ */

function MiniStat({
  icon: Icon,
  label,
  value,
  color,
}: {
  icon: LucideIcon
  label: string
  value: string
  color: string
}) {
  return (
    <div className="flex items-center gap-2.5">
      <div
        className="flex size-7 shrink-0 items-center justify-center rounded-md"
        style={{
          color,
          backgroundColor: `color-mix(in srgb, ${color} 10%, transparent)`,
        }}
      >
        <Icon className="size-3.5" />
      </div>
      <div className="min-w-0">
        <p className="text-sm font-medium leading-none text-text-1 tabular-nums">{value}</p>
        <p className="mt-1 truncate text-xs text-text-3">{label}</p>
      </div>
    </div>
  )
}

function TokenOverviewPanel({ data }: { data: DashboardOverview }) {
  const { todayTokens, weekTokens, totalTokens, totalSessions, newSessions, dailyTokenLimit } = data
  const t = useT()

  const hasLimit = dailyTokenLimit > 0
  const limitTokens = dailyTokenLimit * 1_000_000
  const pct = hasLimit ? Math.min((todayTokens / limitTokens) * 100, 100) : 0
  const pctLabel = pct < 10 ? pct.toFixed(1) : String(Math.round(pct))

  return (
    <div className="flex flex-1 flex-col gap-3 px-6 py-4">
      <div className="flex items-center gap-1.5">
        <Zap className="size-3.5 text-text-3" />
        <h3 className="text-xs font-semibold text-text-2">{t('dashboard.tokenUsage')}</h3>
      </div>

      <div className="flex flex-1 items-center gap-4">
        {/* Hero: today's usage against daily limit */}
        <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
          <p className="text-xs text-text-3">{t('dashboard.todayUsed')}</p>
          <p className="mt-1.5 text-2xl font-semibold leading-none text-highlight tabular-nums">
            {formatNumber(todayTokens)}
          </p>
          {hasLimit ? (
            <>
              <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-bg-layer-3">
                <div
                  className="h-full rounded-full bg-highlight transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <div className="mt-1.5 flex items-center justify-between text-xs text-text-3 tabular-nums">
                <span>{pctLabel}%</span>
                <span>{t('dashboard.limit')} {dailyTokenLimit}M</span>
              </div>
            </>
          ) : (
            <p className="mt-3 text-xs text-text-3">{t('dashboard.unlimited')}</p>
          )}
        </div>

        {/* Secondary metrics */}
        <div className="grid grid-cols-2 content-center gap-x-5 gap-y-3">
          <MiniStat icon={CalendarRange} label={t('dashboard.thisWeek')} value={formatNumber(weekTokens)} color="var(--chart-2)" />
          <MiniStat icon={Database} label={t('dashboard.totalTokens')} value={formatNumber(totalTokens)} color="var(--chart-3)" />
          <MiniStat icon={MessageSquare} label={t('dashboard.sessions')} value={formatNumber(totalSessions)} color="var(--chart-4)" />
          <MiniStat icon={Sparkles} label={t('dashboard.newSessions')} value={formatNumber(newSessions)} color="var(--chart-5)" />
        </div>
      </div>
    </div>
  )
}

/* ============================================
   Panel: Storage
   ============================================ */

function StoragePanel({ data }: { data: StorageStats }) {
  const { usedBytes, freeBytes, totalBytes, items } = data
  const t = useT()

  const hasTotal = totalBytes > 0
  const pct = hasTotal ? Math.min((usedBytes / totalBytes) * 100, 100) : 0
  const pctLabel = pct < 10 ? pct.toFixed(1) : String(Math.round(pct))

  return (
    <div className="flex flex-1 flex-col gap-3 px-6 py-4">
      <div className="flex items-center gap-1.5">
        <HardDrive className="size-3.5 text-text-3" />
        <h3 className="text-xs font-semibold text-text-2">{t('dashboard.storage')}</h3>
      </div>

      <div className="flex items-stretch gap-4">
        {/* Hero: used space against total */}
        <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
          <p className="text-xs text-text-3">{t('dashboard.used')}</p>
          <p className="mt-1.5 text-2xl font-semibold leading-none text-highlight tabular-nums">
            {formatBytes(usedBytes)}
          </p>
          {hasTotal ? (
            <>
              <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-bg-layer-3">
                <div
                  className="h-full rounded-full bg-highlight transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <p className="mt-1.5 text-xs text-text-3 tabular-nums">{pctLabel}%</p>
            </>
          ) : (
            <p className="mt-3 text-xs text-text-3">-</p>
          )}
        </div>

        {/* Secondary metrics */}
        <div className="grid grid-cols-1 content-center gap-y-3">
          <MiniStat icon={Inbox} label={t('dashboard.free')} value={hasTotal ? formatBytes(freeBytes) : '-'} color="var(--chart-2)" />
          <MiniStat icon={Database} label={t('dashboard.total')} value={hasTotal ? formatBytes(totalBytes) : '-'} color="var(--chart-3)" />
        </div>
      </div>

      {/* Directory breakdown */}
      {items.length > 0 && (
        <div className="grid grid-cols-4 gap-x-3 gap-y-2 border-t border-border pt-2.5">
          {items.map((item) => (
            <div key={item.name} className="min-w-0">
              <p className="text-xs font-medium text-text-1 tabular-nums">{formatBytes(item.bytesSize)}</p>
              <p className="mt-0.5 truncate text-xs text-text-3">{item.name}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

/* ============================================
   Panel: Token Trend Chart
   ============================================ */

function TrendChartPanel({
  data,
  trendDays,
  onTrendDaysChange,
}: {
  data: TokenTrendItem[]
  trendDays: TrendDays
  onTrendDaysChange: (v: TrendDays) => void
}) {
  const colors = useChartColors()
  const t = useT()
  const stats = useMemo(() => computeTrendStats(data), [data])
  const maxY = useMemo(() => {
    let m = 0
    for (const d of data) {
      const total = d.promptTokens + d.completionTokens
      if (total > m) m = total
    }
    return Math.ceil(m / 1000) * 1000
  }, [data])

  const trendOptions = useMemo(() => [
    { label: t('dashboard.last7d'), value: 7 as TrendDays },
    { label: t('dashboard.last30d'), value: 30 as TrendDays },
    { label: t('dashboard.last90d'), value: 90 as TrendDays },
  ], [t])

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-xs font-semibold text-text-2">
          {t('dashboard.tokenTrend')}
        </h3>
        <SegmentedControl
          value={trendDays}
          options={trendOptions}
          onChange={onTrendDaysChange}
        />
      </div>

      {/* Summary line */}
      <div className="mt-2 flex shrink-0 gap-6">
        <span className="text-xs text-text-3">
          {t('dashboard.dailyAvg')} {formatNumber(stats.avg)}
        </span>
        <span className="text-xs text-text-3">
          {t('dashboard.peak')} {stats.peakDate} {formatNumber(stats.peakTotal)}
        </span>
      </div>

      {/* Chart */}
      <div className="mt-3 flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} barCategoryGap="20%">
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--color-bg-layer-3)"
              vertical={false}
            />
            <XAxis
              dataKey="date"
              tick={{ ...axisStyle, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={{ stroke: 'var(--color-border-custom)' }}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ ...axisStyle, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={false}
              domain={[0, maxY]}
              tickFormatter={(v: number) => formatNumber(v)}
              width={45}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ fill: 'var(--color-bg-hover)' }} />
            <Bar dataKey="promptTokens" name="Prompt" stackId="a" fill={colors[0]} radius={0} />
            <Bar dataKey="completionTokens" name="Completion" stackId="a" fill={colors[1]} radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

/* ============================================
   Panel: Model Usage
   ============================================ */

function ModelUsagePanel({
  data,
  trendDays,
  onTrendDaysChange,
}: {
  data: ModelUsageItem[]
  trendDays: TrendDays
  onTrendDaysChange: (v: TrendDays) => void
}) {
  const colors = useChartColors()
  const t = useT()
  const trendOptions = useMemo(() => [
    { label: t('dashboard.last7d'), value: 7 as TrendDays },
    { label: t('dashboard.last30d'), value: 30 as TrendDays },
    { label: t('dashboard.last90d'), value: 90 as TrendDays },
  ], [t])

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-xs font-semibold text-text-2">
          {t('dashboard.modelUsage')}
        </h3>
        <SegmentedControl
          value={trendDays}
          options={trendOptions}
          onChange={onTrendDaysChange}
        />
      </div>
      <div className="mt-4 flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart
            data={data}
            layout="vertical"
            margin={{ top: 0, right: 0, bottom: 0, left: 0 }}
            barCategoryGap="40%"
          >
            <XAxis type="number" hide domain={[0, 100]} />
            <YAxis
              type="category"
              dataKey="modelName"
              tick={{ ...axisStyle, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={false}
              width={110}
            />
            {data.length > 0 && (
              <Bar dataKey="percentage" radius={[0, 3, 3, 0]} maxBarSize={12}>
                {data.map((_, i) => (
                  <Cell key={i} fill={colors[i % colors.length]} />
                ))}
              </Bar>
            )}
            <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null
                  const data = payload[0]?.payload as ModelUsageItem | undefined
                  if (!data) return null
                  return (
                    <div className="rounded-lg border border-border bg-bg-layer-2 px-3 py-2 shadow-pop">
                      <p className="text-xs text-text-3">{data.modelName}</p>
                      <p className="text-xs" style={{ color: 'var(--color-interactive)' }}>
                        Prompt: {formatNumber(data.promptTokens)}
                      </p>
                      <p className="text-xs" style={{ color: 'var(--highlight)' }}>
                        Completion: {formatNumber(data.completionTokens)}
                      </p>
                    </div>
                  )
                }}
                cursor={{ fill: 'var(--color-bg-hover)' }}
            />
            {data.length === 0 && (
              <text
                x="50%"
                y="50%"
                textAnchor="middle"
                fill="var(--color-text-3)"
                fontSize={12}
                dy={6}
              >
                {t('common.noData')}
              </text>
            )}
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

type PageState = 'loading' | 'data' | 'empty' | 'error'

export function Component() {
  const [state, setState] = useState<PageState>('loading')
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null)
  const [trendDays, setTrendDays] = useState<TrendDays>(30)
  const [trendData, setTrendData] = useState<TokenTrendItem[]>([])
  const [modelUsageDays, setModelUsageDays] = useState<TrendDays>(30)
  const [modelUsageData, setModelUsageData] = useState<ModelUsageItem[]>([])
  const t = useT()

  const loadData = useCallback(async () => {
    setState('loading')
    try {
      const data = await getDashboard()
      setDashboardData(data)
      setState('data')
    } catch {
      setState('error')
    }
  }, [])

  useEffect(() => {
    loadData()
    getTokenTrend({ days: 30 }).then((t) => setTrendData(t.items))
    getModelUsage({ days: 30 }).then((m) => setModelUsageData(m.items))
  }, [loadData])

  const handleRetry = useCallback(() => {
    loadData()
    getTokenTrend({ days: trendDays }).then((t) => setTrendData(t.items))
    getModelUsage({ days: modelUsageDays }).then((m) => setModelUsageData(m.items))
  }, [loadData, trendDays, modelUsageDays])

  const handleTrendDaysChange = useCallback(
    async (days: TrendDays) => {
      setTrendDays(days)
      try {
        const trend = await getTokenTrend({ days })
        setTrendData(trend.items)
      } catch {
        // keep existing data on error
      }
    },
    [],
  )

  const handleModelUsageDaysChange = useCallback(
    async (days: TrendDays) => {
      setModelUsageDays(days)
      try {
        const modelUsage = await getModelUsage({ days })
        setModelUsageData(modelUsage.items)
      } catch {
        // keep existing data on error
      }
    },
    [],
  )

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">
          {t('dashboard.title')}
        </h1>
      </div>

      {/* Content */}
      {state === 'loading' && <DashboardSkeleton />}

      {state === 'error' && <DashboardError onRetry={handleRetry} />}

      {state === 'data' && dashboardData && (
        <div className="flex flex-1 flex-col gap-2 overflow-hidden p-2">
          {/* Row 1: Overview + Storage */}
          <div className="flex min-h-[140px] gap-2">
            <div className="flex w-1/2 rounded-lg border border-border bg-bg-layer-1">
              <TokenOverviewPanel data={dashboardData.overview} />
            </div>
            <div className="flex w-1/2 rounded-lg border border-border bg-bg-layer-1">
              <StoragePanel data={dashboardData.storageStats} />
            </div>
          </div>

          {/* Row 2: Trend + Model Usage */}
          <div className="flex flex-1 gap-2 min-h-[220px]">
            <div className="flex w-2/3 rounded-lg border border-border bg-bg-layer-1">
              <TrendChartPanel
                data={trendData}
                trendDays={trendDays}
                onTrendDaysChange={handleTrendDaysChange}
              />
            </div>
            <div className="flex w-1/3 rounded-lg border border-border bg-bg-layer-1">
              <ModelUsagePanel
                data={modelUsageData}
                trendDays={modelUsageDays}
                onTrendDaysChange={handleModelUsageDaysChange}
              />
            </div>
          </div>

          {/* Row 3: Rankings */}
          <div className="flex flex-1 min-h-[200px] gap-2">
            <div className="flex flex-1 rounded-lg border border-border bg-bg-layer-1">
              <RankingPanel title={t('dashboard.weeklyToolRank')} data={dashboardData.toolUsageRank} />
            </div>
            <div className="flex flex-1 rounded-lg border border-border bg-bg-layer-1">
              <RankingPanel title={t('dashboard.weeklyMcpRank')} data={dashboardData.mcpUsageRank} />
            </div>
            <div className="flex flex-1 rounded-lg border border-border bg-bg-layer-1">
              <RankingPanel title={t('dashboard.weeklySkillRank')} data={dashboardData.skillUsageRank} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
