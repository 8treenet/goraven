import { useState, useCallback, useMemo, useEffect } from 'react'
import {
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
import { TONE_SOFT, TONE_TEXT, type Tone } from '@/lib/tone'
import { cn } from '@/lib/utils'
import { SegmentedControl } from '@/components/charts/SegmentedControl'
import { ChartTooltip } from '@/components/charts/ChartTooltip'
import { RankingPanel } from '@/components/charts/RankingPanel'
import { Card } from '@/components/ui/card'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
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
  return useMemo(
    () => [
      'var(--chart-1)',
      'var(--chart-2)',
      'var(--chart-3)',
      'var(--chart-4)',
      'var(--chart-5)',
    ],
    [],
  )
}

type TrendDays = 7 | 30 | 90

/* ============================================
   Shared chart axis styles
   ============================================ */

const axisStyle = {
  fontSize: 11,
}

/* ============================================
   Dashboard Skeleton
   ============================================ */

function DashboardSkeleton() {
  return (
    <PageContainer className="flex flex-1 flex-col gap-4 overflow-y-auto py-6">
      {/* Row 1 */}
      <div className="grid min-h-[140px] shrink-0 grid-cols-1 gap-4 md:grid-cols-2">
        <Card className="flex flex-col gap-3 px-5 py-5">
          <div className="flex items-center gap-1.5">
            <div className="size-3.5 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="flex flex-1 items-center gap-4">
            <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
              <div className="h-3 w-14 animate-pulse rounded bg-bg-layer-3" />
              <div className="mt-1.5 h-7 w-20 animate-pulse rounded bg-bg-layer-3" />
              <div className="mt-3 h-1.5 w-full animate-pulse rounded-full bg-bg-layer-3" />
              <div className="mt-1.5 flex items-center justify-between">
                <div className="h-2.5 w-8 animate-pulse rounded bg-bg-layer-3" />
                <div className="h-2.5 w-10 animate-pulse rounded bg-bg-layer-3" />
              </div>
            </div>
            <div className="grid grid-cols-2 content-center gap-x-5 gap-y-3">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="flex items-center gap-2.5">
                  <div className="size-7 animate-pulse rounded-md bg-bg-layer-2" />
                  <div className="space-y-1.5">
                    <div className="h-3 w-12 animate-pulse rounded bg-bg-layer-2" />
                    <div className="h-2.5 w-10 animate-pulse rounded bg-bg-layer-2" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Card>
        <Card className="flex flex-col gap-3 px-5 py-5">
          <div className="flex items-center gap-1.5">
            <div className="size-3.5 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="flex flex-1 items-stretch gap-4">
            <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
              <div className="h-3 w-10 animate-pulse rounded bg-bg-layer-3" />
              <div className="mt-1.5 h-7 w-20 animate-pulse rounded bg-bg-layer-3" />
              <div className="mt-3 h-1.5 w-full animate-pulse rounded-full bg-bg-layer-3" />
              <div className="mt-1.5 h-2.5 w-8 animate-pulse rounded bg-bg-layer-3" />
            </div>
            <div className="grid grid-cols-1 content-center gap-y-3">
              {[1, 2].map((i) => (
                <div key={i} className="flex items-center gap-2.5">
                  <div className="size-7 animate-pulse rounded-md bg-bg-layer-2" />
                  <div className="space-y-1.5">
                    <div className="h-3 w-12 animate-pulse rounded bg-bg-layer-2" />
                    <div className="h-2.5 w-10 animate-pulse rounded bg-bg-layer-2" />
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className="grid grid-cols-4 gap-x-3 gap-y-2 border-t border-border pt-2.5">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="space-y-1">
                <div className="h-3 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-2/3 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Row 2 */}
      <div className="flex min-h-[260px] flex-1 flex-col gap-4 lg:flex-row">
        <Card className="flex min-h-[260px] w-full flex-col px-5 py-5 lg:w-2/3">
          <div className="flex items-center justify-between gap-3">
            <div className="h-3 w-28 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-7 w-40 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-3 flex gap-6">
            <div className="h-2.5 w-24 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2.5 w-32 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-3 flex-1 animate-pulse rounded bg-bg-layer-2" />
        </Card>
        <Card className="flex min-h-[260px] w-full flex-col px-5 py-5 lg:w-1/3">
          <div className="flex items-center justify-between gap-3">
            <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-7 w-40 animate-pulse rounded bg-bg-layer-2" />
          </div>
          <div className="mt-4 flex flex-1 flex-col justify-around gap-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center gap-3">
                <div className="h-3 w-20 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-1.5 flex-1 animate-pulse rounded-full bg-bg-layer-2" />
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Row 3 */}
      <div className="grid min-h-[220px] flex-1 grid-cols-1 gap-4 lg:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="flex min-h-0 flex-col px-5 py-5">
            <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
            <div className="mt-3 flex flex-1 flex-col gap-1">
              {[1, 2, 3, 4, 5].map((j) => (
                <div key={j} className="flex h-8 items-center gap-2.5">
                  <div className="h-2.5 w-4 animate-pulse rounded bg-bg-layer-2" />
                  <div className="h-2.5 w-24 animate-pulse rounded bg-bg-layer-2" />
                  <div className="h-1.5 flex-1 animate-pulse rounded-full bg-bg-layer-2" />
                  <div className="h-2.5 w-8 animate-pulse rounded bg-bg-layer-2" />
                </div>
              ))}
            </div>
          </Card>
        ))}
      </div>
    </PageContainer>
  )
}

/* ============================================
   Dashboard Error
   ============================================ */



/* ============================================
   Panel: Token Overview
   ============================================ */

function MiniStat({
  icon: Icon,
  label,
  value,
  labelClassName = 'text-micro',
  tone,
}: {
  icon: LucideIcon
  label: string
  value: string
  labelClassName?: string
  tone?: Tone
}) {
  return (
    <div className="flex items-center gap-2.5">
      <div className={cn(
        'flex size-7 shrink-0 items-center justify-center rounded-md',
        tone ? `${TONE_SOFT[tone]} ${TONE_TEXT[tone]}` : 'bg-bg-layer-2 text-text-3',
      )}>
        <Icon className="size-3.5" />
      </div>
      <div className="min-w-0">
        <p className="text-[13px] font-medium leading-none text-text-1 tabular-nums">{value}</p>
        <p className={`mt-1 truncate text-text-3 ${labelClassName}`}>{label}</p>
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
    <div className="flex flex-1 flex-col gap-3 px-5 py-5">
      <div className="flex items-center gap-1.5">
        <Zap className="size-3.5 text-amber-strong" />
        <h3 className="text-[13px] font-semibold text-text-2">{t('dashboard.tokenUsage')}</h3>
      </div>

      <div className="flex flex-1 items-center gap-4">
        {/* Hero: today's usage against daily limit */}
        <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
          <p className="text-meta text-text-3">{t('dashboard.todayUsed')}</p>
          <p className="mt-1.5 text-[28px] font-semibold leading-none text-accent tabular-nums">
            {formatNumber(todayTokens)}
          </p>
          {hasLimit ? (
            <>
              <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-bg-layer-3">
                <div
                  className="h-full rounded-full bg-accent transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <div className="mt-1.5 flex items-center justify-between text-meta text-text-3 tabular-nums">
                <span>{pctLabel}%</span>
                <span>{t('dashboard.limit')} {dailyTokenLimit}M</span>
              </div>
            </>
          ) : (
            <p className="mt-3 text-meta text-text-3">{t('dashboard.unlimited')}</p>
          )}
        </div>

        {/* Secondary metrics */}
        <div className="grid grid-cols-2 content-center gap-x-5 gap-y-3">
          <MiniStat icon={CalendarRange} label={t('dashboard.thisWeek')} value={formatNumber(weekTokens)} tone={1} />
          <MiniStat icon={Database} label={t('dashboard.totalTokens')} value={formatNumber(totalTokens)} tone={2} />
          <MiniStat icon={MessageSquare} label={t('dashboard.sessions')} value={formatNumber(totalSessions)} tone={4} />
          <MiniStat icon={Sparkles} label={t('dashboard.newSessions')} value={formatNumber(newSessions)} tone={3} />
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
    <div className="flex flex-1 flex-col gap-3 px-5 py-5">
      <div className="flex items-center gap-1.5">
        <HardDrive className="size-3.5 text-cat-2" />
        <h3 className="text-[13px] font-semibold text-text-2">{t('dashboard.storage')}</h3>
      </div>

      <div className="flex items-stretch gap-4">
        {/* Hero: used space against total */}
        <div className="flex flex-1 flex-col justify-center rounded-lg bg-bg-layer-2/40 px-4 py-3">
          <p className="text-meta text-text-3">{t('dashboard.used')}</p>
          <p className="mt-1.5 text-[28px] font-semibold leading-none text-cat-2 tabular-nums">
            {formatBytes(usedBytes)}
          </p>
          {hasTotal ? (
            <>
              <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-bg-layer-3">
                <div
                  className="h-full rounded-full bg-cat-2 transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <p className="mt-1.5 text-meta text-text-3 tabular-nums">{pctLabel}%</p>
            </>
          ) : (
            <p className="mt-3 text-meta text-text-3">-</p>
          )}
        </div>

        {/* Secondary metrics */}
        <div className="grid grid-cols-1 content-center gap-y-3">
          <MiniStat icon={Inbox} label={t('dashboard.free')} value={hasTotal ? formatBytes(freeBytes) : '-'} labelClassName="text-meta" tone={2} />
          <MiniStat icon={Database} label={t('dashboard.total')} value={hasTotal ? formatBytes(totalBytes) : '-'} labelClassName="text-meta" tone={1} />
        </div>
      </div>

      {/* Directory breakdown */}
      {items.length > 0 && (
        <div className="grid grid-cols-4 gap-x-3 gap-y-2 border-t border-border pt-2.5">
          {items.map((item) => (
            <div key={item.name} className="min-w-0">
              <p className="text-[13px] font-medium text-text-1 tabular-nums">{formatBytes(item.bytesSize)}</p>
              <p className="mt-0.5 truncate text-micro text-text-3">{item.name}</p>
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

  const trendOptions = useMemo(() => [
    { label: t('dashboard.last7d'), value: 7 as TrendDays },
    { label: t('dashboard.last30d'), value: 30 as TrendDays },
    { label: t('dashboard.last90d'), value: 90 as TrendDays },
  ], [t])

  return (
    <div className="flex flex-1 flex-col px-5 py-5">
      <div className="flex shrink-0 items-center justify-between gap-3">
        <h3 className="text-[13px] font-semibold text-text-2">
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
        <span className="text-micro text-text-3 tabular-nums">
          {t('dashboard.dailyAvg')} {formatNumber(stats.avg)}
        </span>
        <span className="text-micro text-text-3 tabular-nums">
          {t('dashboard.peak')} {stats.peakDate} {formatNumber(stats.peakTotal)}
        </span>
      </div>

      {/* Chart */}
      <div className="mt-3 flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData} barCategoryGap="20%">
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--color-border-custom)"
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
            <Tooltip content={<ChartTooltip showTotal />} cursor={{ fill: 'var(--color-bg-hover)' }} />
            <Bar dataKey="promptTokens" name="Prompt" stackId="a" fill={colors[0]} radius={0} />
            {hasCached && (
              <Bar
                dataKey="promptCachedTokens"
                name={t('dashboard.cachedTokens')}
                stackId="a"
                fill="var(--chart-cached)"
                radius={0}
              />
            )}
            <Bar dataKey="completionTokens" name="Completion" stackId="a" fill={colors[3]} radius={[3, 3, 0, 0]} />
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
    <div className="flex flex-1 flex-col px-5 py-5">
      <div className="flex shrink-0 items-center justify-between gap-3">
        <h3 className="text-[13px] font-semibold text-text-2">
          {t('dashboard.modelUsage')}
        </h3>
        <SegmentedControl
          value={trendDays}
          options={trendOptions}
          onChange={onTrendDaysChange}
        />
      </div>
      {data.length === 0 ? (
        <div className="flex flex-1 items-center justify-center">
          <p className="text-meta text-text-3">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="mt-4 flex min-h-0 flex-1 flex-col overflow-y-auto pr-0.5">
          <div className="my-auto flex flex-col gap-2.5">
            {data.map((item, i) => (
              <div
                key={item.modelName}
                className="flex items-center gap-3"
                title={`${item.modelName}
Prompt: ${formatNumber(Math.max(0, item.promptTokens - item.promptCachedTokens))}
Cached: ${formatNumber(item.promptCachedTokens)}
Completion: ${formatNumber(item.completionTokens)}
Total: ${formatNumber(item.tokenCount)}`}
              >
                <span className="w-24 shrink-0 truncate text-meta text-text-2">
                  {item.modelName}
                </span>
                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-bg-layer-3">
                  <div
                    className="h-full rounded-full transition-all duration-300"
                    style={{
                      width: `${item.percentage}%`,
                      backgroundColor: colors[i % colors.length],
                    }}
                  />
                </div>
                <span className="w-9 shrink-0 text-right text-meta text-text-3 tabular-nums">
                  {Math.round(item.percentage)}%
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
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
      <PageHeader title={t('dashboard.title')} />

      {/* Content */}
      {state === 'loading' && <DashboardSkeleton />}

      {state === 'error' && <ErrorState description={t('dashboard.errLoad')} onRetry={handleRetry} />}

      {state === 'data' && dashboardData && (
        <PageContainer className="flex flex-1 flex-col gap-4 overflow-y-auto py-6">
          {/* Row 1: Overview + Storage */}
          <div className="grid min-h-[140px] shrink-0 grid-cols-1 gap-4 md:grid-cols-2">
            <Card className="flex">
              <TokenOverviewPanel data={dashboardData.overview} />
            </Card>
            <Card className="flex">
              <StoragePanel data={dashboardData.storageStats} />
            </Card>
          </div>

          {/* Row 2: Trend + Model Usage */}
          <div className="flex min-h-[260px] flex-1 flex-col gap-4 lg:flex-row">
            <Card className="flex min-h-[260px] w-full lg:w-2/3">
              <TrendChartPanel
                data={trendData}
                trendDays={trendDays}
                onTrendDaysChange={handleTrendDaysChange}
              />
            </Card>
            <Card className="flex min-h-[260px] w-full lg:w-1/3">
              <ModelUsagePanel
                data={modelUsageData}
                trendDays={modelUsageDays}
                onTrendDaysChange={handleModelUsageDaysChange}
              />
            </Card>
          </div>

          {/* Row 3: Rankings */}
          <div className="grid min-h-[220px] flex-1 grid-cols-1 gap-4 lg:grid-cols-3">
            <Card className="flex min-h-0">
              <RankingPanel title={t('dashboard.weeklyToolRank')} data={dashboardData.toolUsageRank} tone={2} />
            </Card>
            <Card className="flex min-h-0">
              <RankingPanel title={t('dashboard.weeklyMcpRank')} data={dashboardData.mcpUsageRank} tone={3} />
            </Card>
            <Card className="flex min-h-0">
              <RankingPanel title={t('dashboard.weeklySkillRank')} data={dashboardData.skillUsageRank} tone={4} />
            </Card>
          </div>
        </PageContainer>
      )}
    </div>
  )
}
