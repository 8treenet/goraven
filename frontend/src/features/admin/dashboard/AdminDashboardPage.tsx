import { useState, useCallback, useMemo, useEffect } from 'react'
import {
  RefreshCw,
  AlertCircle,
  TrendingUp,
  TrendingDown,
  Minus,
} from 'lucide-react'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Cell,
  AreaChart,
  Area,
  ReferenceLine,
} from 'recharts'
import { useT, type TranslationKey } from '@/i18n'
import { formatNumber, formatDiff } from '@/lib/format'
import { SegmentedControl } from '@/components/charts/SegmentedControl'
import { ChartTooltip } from '@/components/charts/ChartTooltip'
import { RankingPanel } from '@/components/charts/RankingPanel'
import { adminSystemApi } from '@/api'

/* ============================================
   Types (aligned with /api/admin/dashboard)
   ============================================ */

interface SparklineItem {
  date: string
  tokens: number
}

interface TokenTrendItem {
  date: string
  promptTokens: number
  completionTokens: number
}

interface ModelUsageItem {
  modelName: string
  tokenCount: number
  percentage: number
}

interface UserTokenRankItem {
  userId: string
  username: string
  tokenCount: number
  percentage: number
}

interface ActiveTrendItem {
  date: string
  count: number
}

interface RankItem {
  name: string
  count: number
}

interface OverviewData {
  activeUsers: number
  activeUsersDiff: number
  totalSessions: number
  newSessions: number
  weekTokens: number
  todayTokens: number
  enabledModels: number
  sparkline: SparklineItem[]
}

interface DashboardData {
  overview: OverviewData
  tokenTrend: TokenTrendItem[]
  modelUsage: ModelUsageItem[]
  userTokenRank: UserTokenRankItem[]
  activeTrend: ActiveTrendItem[]
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
}

/* ============================================
   Chart colors
   ============================================ */

const CHART_PALETTE = [
  'var(--highlight)',
  'var(--color-interactive)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
  'var(--color-bg-hover)',
]

function useChartColors() {
  return useMemo(() => {
    const style = getComputedStyle(document.documentElement)
    return CHART_PALETTE.map((c) => {
      if (c.startsWith('var(')) {
        const name = c.slice(4, -1).trim()
        return style.getPropertyValue(name).trim() || c
      }
      return c
    })
  }, [])
}

type TrendDays = 7 | 30 | 90

function getTrendOptions(t: (key: TranslationKey) => string) {
  return [
    { label: t('adminDashboard.trend7Days'), value: 7 as TrendDays },
    { label: t('adminDashboard.trend30Days'), value: 30 as TrendDays },
    { label: t('adminDashboard.trend90Days'), value: 90 as TrendDays },
  ]
}

/* ============================================
   Shared axis style
   ============================================ */

const AXIS_STYLE = { fontSize: 10 }

/* ============================================
   Skeleton
   ============================================ */

function Skeleton() {
  return (
    <div className="flex flex-1 flex-col gap-2 overflow-hidden p-2">
      {/* Row 1: Pulse */}
      <div className="flex min-h-[120px] gap-2">
        <div className="flex flex-1 flex-col justify-center gap-3 rounded-lg bg-bg-layer-1 px-6 py-4">
          <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
          <div className="flex gap-8">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i}>
                <div className="h-7 w-14 animate-pulse rounded bg-bg-layer-2" />
                <div className="mt-1.5 h-2.5 w-10 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
          <div className="h-12 animate-pulse rounded bg-bg-layer-2" />
        </div>
      </div>

      {/* Row 2: Token Trend */}
      <div className="flex flex-[2] min-h-[220px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg bg-bg-layer-1 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="h-3 w-28 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-6 w-40 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-2 flex gap-6">
            <div className="h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-3 w-32 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-3 flex-1 animate-pulse rounded bg-bg-layer-2" />
        </div>
      </div>

      {/* Row 3: Model Usage + User Rank */}
      <div className="flex flex-[2] min-h-[220px] gap-2">
        <div className="flex flex-col rounded-lg bg-bg-layer-1 px-6 py-4" style={{ flex: 3 }}>
          <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-4 flex-1 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-4 animate-pulse rounded bg-bg-layer-2" style={{ width: `${100 - i * 15}%` }} />
            ))}
          </div>
        </div>
        <div className="flex flex-col rounded-lg bg-bg-layer-1 px-6 py-4" style={{ flex: 2 }}>
          <div className="h-3 w-28 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-4 flex-1 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-4 animate-pulse rounded bg-bg-layer-2" style={{ width: `${100 - i * 10}%` }} />
            ))}
          </div>
        </div>
      </div>

      {/* Row 4: Active Trend */}
      <div className="flex min-h-[140px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg bg-bg-layer-1 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="h-3 w-28 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-6 w-40 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-3 flex-1 animate-pulse rounded bg-bg-layer-2" />
        </div>
      </div>

      {/* Row 5: Rankings */}
      <div className="flex flex-[2] min-h-[200px] gap-2">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex flex-1 flex-col rounded-lg bg-bg-layer-1 px-6 py-4">
            <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
            <div className="mt-4 flex-1 space-y-2">
              {[1, 2, 3, 4, 5].map((j) => (
                <div key={j} className="h-3 animate-pulse rounded bg-bg-layer-2" style={{ width: `${100 - j * 12}%` }} />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ============================================
   Error
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('adminDashboard.errLoad')}</p>
        <p className="mt-1 text-xs text-text-3">{t('adminDashboard.errLoad')}</p>
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
   Diff arrow
   ============================================ */

function DiffArrow({ diff }: { diff: number }) {
  if (diff > 0) {
    return (
      <span className="inline-flex items-center gap-0.5 text-xs text-[oklch(0.55_0.15_145)]">
        <TrendingUp className="size-3" />
        {formatDiff(diff)}
      </span>
    )
  }
  if (diff < 0) {
    return (
      <span className="inline-flex items-center gap-0.5 text-xs text-[oklch(0.55_0.18_20)]">
        <TrendingDown className="size-3" />
        {formatDiff(diff)}
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-0.5 text-xs text-text-3">
      <Minus className="size-3" />
      0%
    </span>
  )
}

/* ============================================
   Pulse
   ============================================ */

function PulsePanel({ data }: { data: OverviewData }) {
  const t = useT()
  const metrics = [
    { label: t('adminDashboard.activeUsers'), value: formatNumber(data.activeUsers), extra: <DiffArrow diff={data.activeUsersDiff} /> },
    { label: t('adminDashboard.totalSessions'), value: formatNumber(data.totalSessions), extra: <span className="text-xs text-text-3">{t('common.total')}</span> },
    { label: t('adminDashboard.newThisWeek'), value: formatNumber(data.newSessions), extra: <span className="text-xs text-text-3">{t('adminDashboard.totalSessions')}</span> },
    { label: t('adminDashboard.weeklyTokens'), value: formatNumber(data.weekTokens), extra: null },
    { label: t('adminDashboard.todayTokens'), value: formatNumber(data.todayTokens), extra: null },
    { label: t('adminDashboard.enabledModels'), value: String(data.enabledModels), extra: <span className="text-xs text-text-3">{t('adminDashboard.units')}</span> },
  ]

  return (
    <div className="flex flex-col gap-2 px-6 py-4">
      <div className="flex items-center gap-2">
        <h3 className="text-xs font-semibold text-text-2">{t('adminDashboard.systemOverview')}</h3>
      </div>

      {/* Metrics row */}
      <div className="flex items-baseline gap-8">
        {metrics.map((m) => (
          <div key={m.label}>
            <div className="flex items-baseline gap-1.5">
              <span className="text-lg font-semibold text-text-1 leading-none tabular-nums">
                {m.value}
              </span>
              {m.extra}
            </div>
            <p className="mt-1 text-xs text-text-3">{m.label}</p>
          </div>
        ))}
      </div>

      {/* Sparkline */}
      <div className="h-12">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data.sparkline} margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
            <defs>
              <linearGradient id="sparklineGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--highlight)" stopOpacity={0.25} />
                <stop offset="100%" stopColor="var(--highlight)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <Tooltip
              content={({ active, payload }) => {
                if (!active || !payload?.length) return null
                return (
                  <div className="rounded-lg border border-border bg-bg-layer-2 px-3 py-2 shadow-pop">
                    {payload.map((p, i) => (
                      <p key={i} className="text-xs" style={{ color: p.color }}>
                        {formatNumber(typeof p.value === 'number' ? p.value : 0)}
                      </p>
                    ))}
                  </div>
                )
              }}
            />
            <Area
              type="monotone"
              dataKey="tokens"
              stroke="var(--highlight)"
              fill="url(#sparklineGrad)"
              strokeWidth={1.5}
              dot={false}
              activeDot={{ r: 3, fill: 'var(--highlight)', stroke: 'var(--bg-base)', strokeWidth: 1 }}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

/* ============================================
   Panel: Token Trend
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

function TokenTrendPanel({
  data,
  trendDays,
  onTrendDaysChange,
}: {
  data: TokenTrendItem[]
  trendDays: TrendDays
  onTrendDaysChange: (v: TrendDays) => void
}) {
  const t = useT()
  const colors = useChartColors()
  const stats = useMemo(() => computeTrendStats(data), [data])
  const maxY = useMemo(() => {
    let m = 0
    for (const d of data) {
      const total = d.promptTokens + d.completionTokens
      if (total > m) m = total
    }
    return Math.ceil(m / 1000) * 1000
  }, [data])

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-xs font-semibold text-text-2">{t('adminDashboard.tokenTrend')}</h3>
        <SegmentedControl value={trendDays} options={getTrendOptions(t)} onChange={onTrendDaysChange} />
      </div>

      <div className="mt-2 flex shrink-0 gap-6">
        <span className="text-xs text-text-3">
          {t('adminDashboard.dailyAvg')} {formatNumber(stats.avg)}
        </span>
        <span className="text-xs text-text-3">
          {t('adminDashboard.peak')} {stats.peakDate} {formatNumber(stats.peakTotal)}
        </span>
      </div>

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
              tick={{ ...AXIS_STYLE, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={{ stroke: 'var(--color-border-custom)' }}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ ...AXIS_STYLE, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={false}
              domain={[0, maxY]}
              tickFormatter={(v: number) => formatNumber(v)}
              width={45}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ fill: 'var(--color-bg-hover)' }} />
            <ReferenceLine
              y={stats.avg}
              stroke="var(--color-text-muted)"
              strokeDasharray="4 4"
              strokeWidth={1}
            />
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
  const t = useT()
  const colors = useChartColors()

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <h3 className="shrink-0 text-xs font-semibold text-text-2">{t('adminDashboard.modelUsage')}</h3>
      <div className="mt-4 flex-1 min-h-0">
        {data.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <p className="text-xs text-text-3">{t('common.noData')}</p>
          </div>
        ) : (
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
                tick={{ ...AXIS_STYLE, fill: 'var(--color-text-muted)' }}
                tickLine={false}
                axisLine={false}
                width={100}
              />
              <Bar dataKey="percentage" radius={[0, 3, 3, 0]} maxBarSize={12}>
                {data.map((_, i) => (
                  <Cell key={i} fill={colors[i % colors.length]} />
                ))}
              </Bar>
              <Tooltip
                content={<ChartTooltip />}
                cursor={{ fill: 'var(--color-bg-hover)' }}
                formatter={(value: unknown) => [`${(value as number).toFixed(1)}%`]}
              />
            </BarChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  )
}

/* ============================================
   Panel: User Token Rank
   ============================================ */

function UserTokenRankPanel({ data }: { data: UserTokenRankItem[] }) {
  const t = useT()
  const rankColors = useMemo(
    () => ['var(--highlight)', 'var(--color-interactive)', 'var(--chart-4)'],
    [],
  )
  const maxToken = data.length > 0 ? data[0].tokenCount : 1

  if (data.length === 0) {
    return (
      <div className="flex flex-1 flex-col px-6 py-4">
        <h3 className="shrink-0 text-xs font-semibold text-text-2">{t('adminDashboard.userTokenRank')}</h3>
        <div className="flex flex-1 items-center justify-center">
          <p className="text-xs text-text-3">{t('common.noData')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <h3 className="shrink-0 text-xs font-semibold text-text-2">{t('adminDashboard.userTokenRank')}</h3>
      <div className="mt-3 flex-1 space-y-2 overflow-hidden">
        {data.map((item, i) => (
          <div key={item.userId} className="flex items-center gap-2">
            <span className="w-4 shrink-0 text-right text-xs text-text-3 tabular-nums">
              {i + 1}
            </span>
            <span className="w-20 shrink-0 truncate text-xs text-text-2">
              {item.username}
            </span>
            <div className="flex h-2.5 flex-1 overflow-hidden rounded-sm bg-bg-layer-2">
              <div
                className="h-full transition-all duration-300"
                style={{
                  width: `${Math.max((item.tokenCount / maxToken) * 100, 2)}%`,
                  backgroundColor: rankColors[i % rankColors.length],
                }}
              />
            </div>
            <span className="w-12 shrink-0 text-right text-xs text-text-1 tabular-nums">
              {formatNumber(item.tokenCount)}
            </span>
            <span className="w-10 shrink-0 text-right text-xs text-text-3 tabular-nums">
              {item.percentage.toFixed(1)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ============================================
   Panel: Active Trend
   ============================================ */

function computeMA(data: ActiveTrendItem[], window: number): (ActiveTrendItem & { ma: number | null })[] {
  return data.map((d, i) => {
    if (i < window - 1) return { ...d, ma: null }
    let sum = 0
    for (let j = i - window + 1; j <= i; j++) {
      sum += data[j].count
    }
    return { ...d, ma: Math.round(sum / window * 10) / 10 }
  })
}

function ActiveTrendPanel({
  data,
  trendDays,
  onTrendDaysChange,
}: {
  data: ActiveTrendItem[]
  trendDays: TrendDays
  onTrendDaysChange: (v: TrendDays) => void
}) {
  const t = useT()
  const withMA = useMemo(() => computeMA(data, 7), [data])

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-xs font-semibold text-text-2">{t('adminDashboard.activityTrend')}</h3>
        <SegmentedControl value={trendDays} options={getTrendOptions(t)} onChange={onTrendDaysChange} />
      </div>

      <div className="mt-3 flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={withMA} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--color-bg-layer-3)"
              vertical={false}
            />
            <XAxis
              dataKey="date"
              tick={{ ...AXIS_STYLE, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={{ stroke: 'var(--color-border-custom)' }}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ ...AXIS_STYLE, fill: 'var(--color-text-muted)' }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
              width={35}
            />
            <Tooltip content={<ChartTooltip />} />
            <Line
              type="monotone"
              dataKey="count"
              name="DAU"
              stroke="var(--color-interactive)"
              strokeWidth={1.5}
              dot={false}
              activeDot={{ r: 3, fill: 'var(--color-interactive)', stroke: 'var(--bg-base)', strokeWidth: 1 }}
            />
          <Line
            type="monotone"
            dataKey="ma"
            name={t('adminDashboard.ma7d')}
              stroke="var(--highlight)"
              strokeWidth={1.5}
              strokeDasharray="5 3"
              dot={false}
              connectNulls
            />
          </LineChart>
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
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [data, setData] = useState<DashboardData | null>(null)
  const [trendDays, setTrendDays] = useState<TrendDays>(30)
  const [trendData, setTrendData] = useState<TokenTrendItem[]>([])
  const [activeDays, setActiveDays] = useState<TrendDays>(30)
  const [activeData, setActiveData] = useState<ActiveTrendItem[]>([])

  const loadData = useCallback(() => {
    setState('loading')
    Promise.all([
      adminSystemApi.getDashboard(),
      adminSystemApi.getTokenTrend({ days: trendDays }),
      adminSystemApi.getActiveUsers({ days: activeDays }),
    ]).then(([dash, trend, active]) => {
      setData(dash)
      setTrendData(trend.items)
      setActiveData(active.items)
      setState('data')
    }).catch(() => {
      setState('error')
    })
  }, [trendDays, activeDays])

  useEffect(() => {
    const cleanup = loadData()
    return cleanup
  }, [loadData])

  const handleRetry = useCallback(() => {
    loadData()
  }, [loadData])

  const handleTrendDaysChange = useCallback((days: TrendDays) => {
    setTrendDays(days)
    adminSystemApi.getTokenTrend({ days }).then((trend) => {
      setTrendData(trend.items)
    })
  }, [])

  const handleActiveDaysChange = useCallback((days: TrendDays) => {
    setActiveDays(days)
    adminSystemApi.getActiveUsers({ days }).then((active) => {
      setActiveData(active.items)
    })
  }, [])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminDashboard.title')}</h1>
        <button
          onClick={handleRetry}
          className="flex items-center gap-1 rounded-sm px-1.5 py-0.5 text-xs text-text-3 transition-colors hover:bg-bg-hover hover:text-text-2"
        >
          <RefreshCw className="size-3" />
          {t('common.refresh')}
        </button>
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}

      {state === 'error' && <ErrorState onRetry={handleRetry} />}

      {state === 'data' && data && (
        <div className="flex flex-1 flex-col gap-2 overflow-auto p-2">
          {/* Row 1: System Pulse (full width) */}
          <div className="flex min-h-[120px] gap-2">
            <div className="flex w-full flex-col rounded-lg bg-bg-layer-1">
              <PulsePanel data={data.overview} />
            </div>
          </div>

          {/* Row 2: Token Trend (full width, 2fr) */}
          <div className="flex flex-[2] min-h-[220px] gap-2">
            <div className="flex flex-1 rounded-lg bg-bg-layer-1">
              <TokenTrendPanel
                data={trendData}
                trendDays={trendDays}
                onTrendDaysChange={handleTrendDaysChange}
              />
            </div>
          </div>

          {/* Row 3: Model Usage (3/5) + User Token Rank (2/5) */}
          <div className="flex flex-[2] min-h-[220px] gap-2">
            <div className="flex rounded-lg bg-bg-layer-1" style={{ flex: 3 }}>
              <ModelUsagePanel data={data.modelUsage} />
            </div>
            <div className="flex rounded-lg bg-bg-layer-1" style={{ flex: 2 }}>
              <UserTokenRankPanel data={data.userTokenRank} />
            </div>
          </div>

          {/* Row 4: Active Trend (full width) */}
          <div className="flex min-h-[140px] gap-2">
            <div className="flex flex-1 rounded-lg bg-bg-layer-1">
              <ActiveTrendPanel
                data={activeData}
                trendDays={activeDays}
                onTrendDaysChange={handleActiveDaysChange}
              />
            </div>
          </div>

          {/* Row 5: Rankings (1:1:1) */}
          <div className="flex flex-[2] min-h-[200px] gap-2">
            <div className="flex flex-1 rounded-lg bg-bg-layer-1">
              <RankingPanel title={t('adminDashboard.weeklySkillRank')} data={data.skillUsageRank} />
            </div>
            <div className="flex flex-1 rounded-lg bg-bg-layer-1">
              <RankingPanel title={t('adminDashboard.weeklyMcpRank')} data={data.mcpUsageRank} />
            </div>
            <div className="flex flex-1 rounded-lg bg-bg-layer-1">
              <RankingPanel title={t('adminDashboard.weeklyToolRank')} data={data.toolUsageRank} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
