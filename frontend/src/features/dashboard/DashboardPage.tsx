import { useState, useCallback, useMemo, useEffect } from 'react'
import {
  RefreshCw,
  AlertCircle,
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
import { getDashboard, getTokenTrend } from '@/api/dashboard'
import type { DashboardData, TokenTrendItem, DashboardOverview, StorageStats, ModelUsageItem } from '@/api'

/* ============================================
   Helpers
   ============================================ */

function computeTrendStats(data: TokenTrendItem[]) {
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
          <div className="flex w-1/2 flex-col justify-center gap-3 rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
            <div className="flex gap-8">
              <div className="h-8 w-16 animate-pulse rounded bg-bg-layer-2" />
              <div className="h-8 w-16 animate-pulse rounded bg-bg-layer-2" />
              <div className="h-8 w-16 animate-pulse rounded bg-bg-layer-2" />
            </div>
          </div>
          <div className="flex w-1/2 flex-col justify-center gap-3 rounded-lg border border-border bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2 w-full animate-pulse rounded bg-bg-layer-2" />
            <div className="flex gap-6">
              <div className="h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
              <div className="h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
              <div className="h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
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

function TokenOverviewPanel({ data }: { data: DashboardOverview }) {
  const { todayTokens, weekTokens, totalTokens, totalSessions, newSessions } = data
  const t = useT()

  return (
    <div className="flex flex-1 flex-col justify-center gap-2 px-6 py-4">
      <h3 className="text-xs font-semibold text-text-2">{t('dashboard.tokenUsage')}</h3>
      <div className="flex items-baseline gap-10">
        <div>
          <p className="text-lg font-semibold text-text-1 leading-none">
            {formatNumber(totalTokens)}
          </p>
          <p className="mt-1.5 text-xs text-text-3">{t('dashboard.totalTokens')}</p>
        </div>
        <div>
          <p className="text-base font-semibold text-text-1 leading-none">
            {formatNumber(weekTokens)}
          </p>
          <p className="mt-1.5 text-xs text-text-3">{t('dashboard.thisWeek')}</p>
        </div>
        <div>
          <p className="text-base font-semibold text-text-1 leading-none">
            {formatNumber(todayTokens)}
          </p>
          <p className="mt-1.5 text-xs text-text-3">{t('dashboard.today')}</p>
        </div>
      </div>
      <div className="flex gap-6">
        <span className="text-xs text-text-3">
          {t('dashboard.sessions')} {totalSessions}
        </span>
        <span className="text-xs text-text-3">
          {t('dashboard.newSessions')} {newSessions}
        </span>
      </div>
    </div>
  )
}

/* ============================================
   Panel: Storage
   ============================================ */

function StoragePanel({ data }: { data: StorageStats }) {
  const { usedBytes, freeBytes, totalBytes, items } = data
  const pct = totalBytes > 0 ? (usedBytes / totalBytes) * 100 : 0
  const t = useT()

  return (
    <div className="flex flex-1 flex-col justify-center gap-2 px-6 py-3">
      <h3 className="text-xs font-semibold text-text-2">{t('dashboard.storage')}</h3>

      <div className="flex items-baseline gap-4">
        <div>
          <p className="text-base font-semibold text-text-1">
            {formatBytes(usedBytes)}
          </p>
          <p className="text-xs text-text-3">{t('dashboard.used')}</p>
        </div>
        <div>
          <p className="text-base font-semibold text-text-1">
            {totalBytes > 0 ? formatBytes(freeBytes) : '-'}
          </p>
          <p className="text-xs text-text-3">{t('dashboard.free')}</p>
        </div>
        <div>
          <p className="text-base font-semibold text-text-1">
            {totalBytes > 0 ? formatBytes(totalBytes) : '-'}
          </p>
          <p className="text-xs text-text-3">{t('dashboard.total')}</p>
        </div>
      </div>

      {/* Progress bar */}
      <div className="h-1.5 w-full overflow-hidden rounded-sm bg-bg-layer-3">
        <div
          className="h-full bg-bg-hover transition-all duration-500"
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
      <p className="text-xs text-text-3">
        {totalBytes > 0 ? `${pct.toFixed(1)}%` : '-'}
      </p>

      {/* Directory breakdown */}
      {items.length > 0 && (
        <div className="grid grid-cols-4 gap-x-3 gap-y-1.5">
          {items.map((item) => (
            <div key={item.name}>
              <p className="text-xs text-text-1 tabular-nums">
                {formatBytes(item.bytesSize)}
              </p>
              <p className="text-xs text-text-3 truncate">
                {item.name}
              </p>
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

function ModelUsagePanel({ data }: { data: ModelUsageItem[] }) {
  const colors = useChartColors()
  const t = useT()

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <h3 className="shrink-0 text-xs font-semibold text-text-2">
        {t('dashboard.modelUsage')}
      </h3>
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
              content={<ChartTooltip />}
              cursor={{ fill: 'var(--color-bg-hover)' }}
              formatter={(value: unknown) => [`${(value as number).toFixed(1)}%`]}
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
  const t = useT()

  const loadData = useCallback(async (refresh = false) => {
    setState('loading')
    try {
      const data = await getDashboard(refresh ? { refresh: true } : undefined)
      const trend = await getTokenTrend({ days: 30, refresh })
      setDashboardData(data)
      setTrendData(trend.items)
      setState('data')
    } catch {
      setState('error')
    }
  }, [])

  useEffect(() => {
    loadData(false)
  }, [loadData])

  const handleRetry = useCallback(() => {
    loadData(true)
  }, [loadData])

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

  const trendDataWithDays = useMemo(() => {
    return trendDays ? trendData : dashboardData?.tokenTrend ?? []
  }, [trendDays, trendData, dashboardData])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">
          {t('dashboard.title')}
        </h1>
        <button
          onClick={handleRetry}
          className="flex items-center gap-1 rounded-sm px-1.5 py-0.5 text-xs text-highlight transition-colors hover:bg-bg-hover hover:text-highlight"
        >
          <RefreshCw className="size-3" />
          {t('common.refresh')}
        </button>
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
                data={trendDataWithDays}
                trendDays={trendDays}
                onTrendDaysChange={handleTrendDaysChange}
              />
            </div>
            <div className="flex w-1/3 rounded-lg border border-border bg-bg-layer-1">
              <ModelUsagePanel data={dashboardData.modelUsage} />
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
