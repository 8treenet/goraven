import { useState, useCallback, useMemo, useEffect } from 'react'
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
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { useT, type TranslationKey } from '@/i18n'
import { formatNumber, formatPercent } from '@/lib/format'
import { cn } from '@/lib/utils'
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
  promptCachedTokens: number
  completionTokens: number
}

interface ModelUsageItem {
  modelName: string
  tokenCount: number
  percentage: number
  promptTokens: number
  promptCachedTokens: number
  completionTokens: number
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
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
}

/* ============================================
   Chart colors
   ============================================ */

const CHART_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
] as const

function useChartColors() {
  return CHART_COLORS
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

const AXIS_STYLE = { fontSize: 11 }

/* ============================================
   Skeleton
   ============================================ */

function AdminDashboardSkeleton() {
  return (
    <PageContainer className="flex flex-1 flex-col gap-4 overflow-hidden">
      {/* Row 1: Pulse */}
      <div className="flex min-h-[120px] gap-4">
        <Card className="flex flex-1 flex-col justify-center gap-3 px-6 py-4">
          <Skeleton className="h-3 w-20 rounded bg-bg-layer-2" />
          <div className="flex gap-8">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i}>
                <Skeleton className="h-7 w-14 rounded bg-bg-layer-2" />
                <Skeleton className="mt-1.5 h-2.5 w-10 rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
          <Skeleton className="h-12 rounded bg-bg-layer-2" />
        </Card>
      </div>

      {/* Row 2: Token Trend */}
      <div className="flex flex-[2] min-h-[220px] gap-4">
        <Card className="flex flex-1 flex-col px-6 py-4">
          <div className="flex items-center justify-between">
            <Skeleton className="h-3 w-28 rounded bg-bg-layer-2" />
            <Skeleton className="h-6 w-40 rounded bg-bg-layer-2" />
          </div>
          <div className="mt-2 flex gap-6">
            <Skeleton className="h-3 w-24 rounded bg-bg-layer-2" />
            <Skeleton className="h-3 w-32 rounded bg-bg-layer-2" />
          </div>
          <Skeleton className="mt-3 flex-1 rounded bg-bg-layer-2" />
        </Card>
      </div>

      {/* Row 3: Model Usage + User Rank */}
      <div className="flex flex-[2] min-h-[220px] gap-4">
        <Card className="flex flex-col px-6 py-4" style={{ flex: 3 }}>
          <Skeleton className="h-3 w-20 rounded bg-bg-layer-2" />
          <div className="mt-4 flex-1 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-4 rounded bg-bg-layer-2" style={{ width: `${100 - i * 15}%` }} />
            ))}
          </div>
        </Card>
        <Card className="flex flex-col px-6 py-4" style={{ flex: 2 }}>
          <Skeleton className="h-3 w-28 rounded bg-bg-layer-2" />
          <div className="mt-4 flex-1 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-4 rounded bg-bg-layer-2" style={{ width: `${100 - i * 10}%` }} />
            ))}
          </div>
        </Card>
      </div>

      {/* Row 4: Active Trend */}
      <div className="flex min-h-[140px] gap-4">
        <Card className="flex flex-1 flex-col px-6 py-4">
          <div className="flex items-center justify-between">
            <Skeleton className="h-3 w-28 rounded bg-bg-layer-2" />
            <Skeleton className="h-6 w-40 rounded bg-bg-layer-2" />
          </div>
          <Skeleton className="mt-3 flex-1 rounded bg-bg-layer-2" />
        </Card>
      </div>

      {/* Row 5: Rankings */}
      <div className="flex flex-[2] min-h-[200px] gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="flex flex-1 flex-col px-6 py-4">
            <Skeleton className="h-3 w-16 rounded bg-bg-layer-2" />
            <div className="mt-4 flex-1 space-y-2">
              {[1, 2, 3, 4, 5].map((j) => (
                <Skeleton key={j} className="h-3 rounded bg-bg-layer-2" style={{ width: `${100 - j * 12}%` }} />
              ))}
            </div>
          </Card>
        ))}
      </div>
    </PageContainer>
  )
}

/* ============================================
   Error
   ============================================ */

/* ============================================
   Pulse
   ============================================ */

function PulsePanel({ data }: { data: OverviewData }) {
  const t = useT()
  const metrics = [
    {
      label: t('adminDashboard.activeUsers'),
      value: formatNumber(data.activeUsers),
      extra:
        Math.abs(data.activeUsersDiff) < 0.05 ? null : (
          <span className={cn('text-xs tabular-nums', data.activeUsersDiff > 0 ? 'text-success' : 'text-destructive')}>
            {data.activeUsersDiff > 0 ? '+' : ''}{formatPercent(data.activeUsersDiff)}
          </span>
        ),
    },
    { label: t('adminDashboard.totalSessions'), value: formatNumber(data.totalSessions), extra: <span className="text-xs text-text-3">{t('common.total')}</span> },
    { label: t('adminDashboard.newThisWeek'), value: formatNumber(data.newSessions), extra: <span className="text-xs text-text-3">{t('adminDashboard.totalSessions')}</span> },
    { label: t('adminDashboard.weeklyTokens'), value: formatNumber(data.weekTokens), extra: null },
    { label: t('adminDashboard.todayTokens'), value: formatNumber(data.todayTokens), extra: null },
    { label: t('adminDashboard.enabledModels'), value: String(data.enabledModels), extra: <span className="text-xs text-text-3">{t('adminDashboard.units')}</span> },
  ]

  return (
    <div className="flex flex-col gap-2 px-6 py-4">
      <div className="flex items-center gap-2">
        <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.systemOverview')}</h3>
      </div>

      {/* Metrics row */}
      <div className="flex flex-wrap items-baseline gap-x-8 gap-y-3">
        {metrics.map((m) => (
          <div key={m.label}>
            <div className="flex items-baseline gap-1.5">
              <span className="text-display font-semibold leading-none text-text-1 tabular-nums">
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
                <stop offset="0%" stopColor="var(--chart-1)" stopOpacity={0.25} />
                <stop offset="100%" stopColor="var(--chart-1)" stopOpacity={0} />
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
              stroke="var(--chart-1)"
              fill="url(#sparklineGrad)"
              strokeWidth={1.5}
              dot={false}
              activeDot={{ r: 3, fill: 'var(--chart-1)', stroke: 'var(--bg-base)', strokeWidth: 1 }}
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
  const stats = useMemo(() => computeTrendStats(data), [data])
  const maxY = useMemo(() => {
    let m = 0
    for (const d of data) {
      const total = d.promptTokens + d.completionTokens
      if (total > m) m = total
    }
    return Math.ceil(m / 1000) * 1000
  }, [data])
  const hasCached = useMemo(
    () => data.some((d) => d.promptCachedTokens > 0),
    [data],
  )
  const chartData = useMemo(
    () =>
      data.map((d) => ({
        ...d,
        promptTokens: Math.max(0, d.promptTokens - d.promptCachedTokens),
      })),
    [data],
  )

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.tokenTrend')}</h3>
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
          <BarChart data={chartData} barCategoryGap="20%">
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
            <Tooltip content={<ChartTooltip showTotal />} cursor={{ fill: 'var(--color-bg-hover)' }} />
            <ReferenceLine
              y={stats.avg}
              stroke="var(--color-text-muted)"
              strokeDasharray="4 4"
              strokeWidth={1}
            />
            <Bar dataKey="promptTokens" name="Prompt" stackId="a" fill="var(--chart-1)" radius={0} />
            {hasCached && (
              <Bar dataKey="promptCachedTokens" name={t('adminDashboard.cachedTokens')} stackId="a" fill="var(--chart-cached)" radius={0} />
            )}
            <Bar dataKey="completionTokens" name="Completion" stackId="a" fill="var(--chart-4)" radius={[3, 3, 0, 0]} />
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
  const t = useT()
  const colors = useChartColors()

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.modelUsage')}</h3>
        <SegmentedControl value={trendDays} options={getTrendOptions(t)} onChange={onTrendDaysChange} />
      </div>
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
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null
                  const item = payload[0]?.payload as ModelUsageItem | undefined
                  if (!item) return null
                  const itemIndex = data.findIndex((d) => d.modelName === item.modelName)
                  const itemColor = itemIndex >= 0 ? colors[itemIndex % colors.length] : colors[0]
                  return (
                    <div className="rounded-lg border border-border bg-bg-layer-2 px-3 py-2 shadow-pop">
                      <p className="flex items-center text-xs text-text-3">
                        <span aria-hidden="true" className="mr-1 inline-block size-1.5 rounded-full" style={{ backgroundColor: itemColor }} />
                        {item.modelName}
                      </p>
                      <p className="text-xs" style={{ color: 'var(--chart-1)' }}>
                        Prompt: {formatNumber(Math.max(0, item.promptTokens - item.promptCachedTokens))}
                      </p>
                      <p className="text-xs" style={{ color: 'var(--chart-cached)' }}>
                        Cached: {formatNumber(item.promptCachedTokens)}
                      </p>
                      <p className="text-xs" style={{ color: 'var(--chart-4)' }}>
                        Completion: {formatNumber(item.completionTokens)}
                      </p>
                      <p className="mt-1 border-t border-border pt-1 text-xs font-medium text-text-1">
                        Total: {formatNumber(item.tokenCount)}
                      </p>
                    </div>
                  )
                }}
                cursor={{ fill: 'var(--color-bg-hover)' }}
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

function UserTokenRankPanel({
  data,
  trendDays,
  onTrendDaysChange,
}: {
  data: UserTokenRankItem[]
  trendDays: TrendDays
  onTrendDaysChange: (v: TrendDays) => void
}) {
  const t = useT()
  const rankColors = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)']
  const maxToken = data.length > 0 ? data[0].tokenCount : 1

  if (data.length === 0) {
    return (
      <div className="flex flex-1 flex-col px-6 py-4">
        <div className="flex shrink-0 items-center justify-between">
          <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.userTokenRank')}</h3>
          <SegmentedControl value={trendDays} options={getTrendOptions(t)} onChange={onTrendDaysChange} />
        </div>
        <div className="flex flex-1 items-center justify-center">
          <p className="text-xs text-text-3">{t('common.noData')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <div className="flex shrink-0 items-center justify-between">
        <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.userTokenRank')}</h3>
        <SegmentedControl value={trendDays} options={getTrendOptions(t)} onChange={onTrendDaysChange} />
      </div>
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
        <h3 className="text-section font-semibold text-text-1">{t('adminDashboard.activityTrend')}</h3>
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
              stroke="var(--chart-1)"
              strokeWidth={1.5}
              dot={false}
              activeDot={{ r: 3, fill: 'var(--chart-1)', stroke: 'var(--bg-base)', strokeWidth: 1 }}
            />
          <Line
            type="monotone"
            dataKey="ma"
            name={t('adminDashboard.ma7d')}
              stroke="var(--chart-4)"
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
  const [modelUsageDays, setModelUsageDays] = useState<TrendDays>(30)
  const [modelUsageData, setModelUsageData] = useState<ModelUsageItem[]>([])
  const [userTokenRankDays, setUserTokenRankDays] = useState<TrendDays>(30)
  const [userTokenRankData, setUserTokenRankData] = useState<UserTokenRankItem[]>([])

  const loadData = useCallback(() => {
    setState('loading')
    adminSystemApi.getDashboard().then((dash) => {
      setData(dash)
      setState('data')
    }).catch(() => {
      setState('error')
    })
  }, [])

  useEffect(() => {
    loadData()
    adminSystemApi.getTokenTrend({ days: 30 }).then((t) => setTrendData(t.items))
    adminSystemApi.getActiveUsers({ days: 30 }).then((a) => setActiveData(a.items))
    adminSystemApi.getModelUsage({ days: 30 }).then((m) => setModelUsageData(m.items))
    adminSystemApi.getUserTokenRank({ days: 30 }).then((u) => setUserTokenRankData(u.items))
  }, [loadData])

  const handleRetry = useCallback(() => {
    loadData()
    adminSystemApi.getTokenTrend({ days: trendDays }).then((t) => setTrendData(t.items))
    adminSystemApi.getActiveUsers({ days: activeDays }).then((a) => setActiveData(a.items))
    adminSystemApi.getModelUsage({ days: modelUsageDays }).then((m) => setModelUsageData(m.items))
    adminSystemApi.getUserTokenRank({ days: userTokenRankDays }).then((u) => setUserTokenRankData(u.items))
  }, [loadData, trendDays, activeDays, modelUsageDays, userTokenRankDays])

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

  const handleModelUsageDaysChange = useCallback((days: TrendDays) => {
    setModelUsageDays(days)
    adminSystemApi.getModelUsage({ days }).then((modelUsage) => {
      setModelUsageData(modelUsage.items)
    })
  }, [])

  const handleUserTokenRankDaysChange = useCallback((days: TrendDays) => {
    setUserTokenRankDays(days)
    adminSystemApi.getUserTokenRank({ days }).then((userTokenRank) => {
      setUserTokenRankData(userTokenRank.items)
    })
  }, [])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <PageHeader title={t('adminDashboard.title')} />

      {/* Content */}
      {state === 'loading' && <AdminDashboardSkeleton />}

      {state === 'error' && <ErrorState title={t('adminDashboard.errLoad')} description={t('adminDashboard.errLoad')} onRetry={handleRetry} />}

      {state === 'data' && data && (
        <PageContainer className="flex flex-1 flex-col gap-4 overflow-auto">
          {/* Row 1: System Pulse (full width) */}
          <div className="flex min-h-[120px] gap-4">
            <Card className="flex w-full flex-col">
              <PulsePanel data={data.overview} />
            </Card>
          </div>

          {/* Row 2: Token Trend (full width, 2fr) */}
          <div className="flex flex-[2] min-h-[220px] gap-4">
            <Card className="flex flex-1">
              <TokenTrendPanel
                data={trendData}
                trendDays={trendDays}
                onTrendDaysChange={handleTrendDaysChange}
              />
            </Card>
          </div>

          {/* Row 3: Model Usage (3/5) + User Token Rank (2/5) */}
          <div className="flex flex-[2] min-h-[220px] gap-4">
            <Card className="flex" style={{ flex: 3 }}>
              <ModelUsagePanel
                data={modelUsageData}
                trendDays={modelUsageDays}
                onTrendDaysChange={handleModelUsageDaysChange}
              />
            </Card>
            <Card className="flex" style={{ flex: 2 }}>
              <UserTokenRankPanel
                data={userTokenRankData}
                trendDays={userTokenRankDays}
                onTrendDaysChange={handleUserTokenRankDaysChange}
              />
            </Card>
          </div>

          {/* Row 4: Active Trend (full width) */}
          <div className="flex min-h-[140px] gap-4">
            <Card className="flex flex-1">
              <ActiveTrendPanel
                data={activeData}
                trendDays={activeDays}
                onTrendDaysChange={handleActiveDaysChange}
              />
            </Card>
          </div>

          {/* Row 5: Rankings (1:1:1) */}
          <div className="flex flex-[2] min-h-[200px] gap-4">
            <Card className="flex flex-1">
              <RankingPanel title={t('adminDashboard.weeklySkillRank')} data={data.skillUsageRank} tone={4} />
            </Card>
            <Card className="flex flex-1">
              <RankingPanel title={t('adminDashboard.weeklyMcpRank')} data={data.mcpUsageRank} tone={3} />
            </Card>
            <Card className="flex flex-1">
              <RankingPanel title={t('adminDashboard.weeklyToolRank')} data={data.toolUsageRank} tone={2} />
            </Card>
          </div>
        </PageContainer>
      )}
    </div>
  )
}
