import { useState, useCallback, useEffect } from 'react'
import { RefreshCw, AlertCircle, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatBytes, formatRelativeTime, formatNumber } from '@/lib/format'
import { adminSystemApi } from '@/api'
import { useT, t as translate } from '@/i18n'

/* ============================================
   Types (aligned with /api/admin/systemInfo)
   ============================================ */

interface DatabasePool {
  maxOpenConnections: number
  openConnections: number
  inUse: number
  idle: number
  waitCount: number
  waitDurationMs: number
  maxIdleClosed: number
  maxLifetimeClosed: number
}

interface OverviewInfo {
  version: string
  language: string
  cacheType: string
  cacheMemory: string
  timezone: string
  uploadBytes: number
  tempBytes: number
}

interface DatabaseInfo {
  type: string
  version: string
  name: string
  dataSizeBytes: number
  pool: DatabasePool
}

interface DiskInfo {
  mountPoint: string
  fsType: string
  device: string
  totalBytes: number
  usedBytes: number
  freeBytes: number
  usedPercent: number
}

interface McpHealthItem {
  mcpId: number
  name: string
  displayName: string
  icon: string
  healthLatency: number
  healthCheckedAt: string
}

interface EcosystemInfo {
  totalUsers: number
  activeUsers: number
  adminUsers: number
  totalModels: number
  enabledModels: number
  totalMcps: number
  enabledMcps: number
  systemSkills: number
  marketSkills: number
  personaTemplates: number
  totalSessions: number
  totalMessages: number
  totalSharedProjects: number
  totalShareLinks: number
  activeShareLinks: number
  totalShareViews: number
}

interface PluginInfo {
  name: string
  version: string
}

interface SystemInfoData {
  overview: OverviewInfo
  database: DatabaseInfo
  disks: DiskInfo[]
  mcpHealth: McpHealthItem[]
  ecosystem: EcosystemInfo
  plugins: PluginInfo[]
  collectedAt: string
}

/* ============================================
   Shared: Field row
   ============================================ */

function FieldRow({ label, value, muted }: { label: string; value: string; muted?: boolean }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-1">
      <span className="shrink-0 text-xs text-text-3">{label}</span>
      <span className={cn('truncate text-right text-sm tabular-nums', muted ? 'text-text-3' : 'text-text-1')}>
        {value}
      </span>
    </div>
  )
}

/* ============================================
   Panel: Overview
   ============================================ */

function OverviewPanel({ data }: { data: OverviewInfo }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
      <h3 className="text-xs font-semibold text-text-2">{t('adminDashboard.systemOverview')}</h3>
      <div className="mt-3 flex flex-col">
        <FieldRow label={t('adminSystemInfo.version')} value={data.version} />
        <FieldRow label={t('adminSystemInfo.systemLanguage')} value={data.language === 'zh' ? t('adminSystemInfo.chinese') : 'English'} />
        <FieldRow label={t('adminSystemInfo.cacheType')} value={data.cacheType} />
        <FieldRow label={t('adminSystemInfo.cacheUsage')} value={data.cacheMemory} />
        <FieldRow label={t('adminSystemInfo.systemTimezone')} value={data.timezone} />
        <FieldRow label={t('adminSystemInfo.uploadedFiles')} value={formatBytes(data.uploadBytes)} />
        <FieldRow label={t('adminSystemInfo.tempFiles')} value={formatBytes(data.tempBytes)} />
      </div>
    </div>
  )
}

/* ============================================
   Panel: Database
   ============================================ */

function DatabasePanel({ data }: { data: DatabaseInfo }) {
  const t = useT()
  const pool = data.pool
  const hasWait = pool.waitCount > 0 || pool.waitDurationMs > 0

  return (
    <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
      <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.database')}</h3>
      <div className="mt-3 flex flex-col">
        <FieldRow label={t('common.type')} value={data.type} />
        <FieldRow label={t('common.version')} value={data.version} />
        <FieldRow label={t('adminSystemInfo.dbName')} value={data.name} muted />
        <FieldRow label={t('adminSystemInfo.dataSize')} value={formatBytes(data.dataSizeBytes)} />
      </div>

      <div className="mt-3 border-t border-border pt-3">
        <div className="mb-2 text-xs text-text-3">{t('adminSystemInfo.connectionPool')}</div>
        <div className="flex items-center gap-1.5">
          <div className="flex h-3 flex-1 overflow-hidden rounded-sm bg-bg-layer-2">
            {pool.inUse > 0 && (
              <div
                className="h-full bg-interactive transition-all"
                style={{ width: `${(pool.inUse / pool.maxOpenConnections) * 100}%` }}
              />
            )}
            {pool.idle > 0 && (
              <div
                className="h-full bg-bg-layer-3 transition-all"
                style={{ width: `${(pool.idle / pool.maxOpenConnections) * 100}%` }}
              />
            )}
          </div>
          <span className="shrink-0 text-xs tabular-nums text-text-2">
            {pool.inUse + pool.idle}/{pool.maxOpenConnections}
          </span>
        </div>
        <div className="mt-1.5 flex gap-4 text-xs text-text-3">
          <span>{t('adminSystemInfo.open')} {pool.openConnections}</span>
          <span>{t('adminSystemInfo.inUse')} {pool.inUse}</span>
          <span>{t('adminSystemInfo.idle')} {pool.idle}</span>
        </div>
        {hasWait && (
          <div className="mt-1 text-xs text-amber-400">
            {t('adminSystemInfo.waiting')} {pool.waitCount} {t('adminSystemInfo.times')} · {t('adminSystemInfo.duration')} {pool.waitDurationMs}ms
          </div>
        )}
      </div>
    </div>
  )
}

/* ============================================
   Panel: MCP Health (paginated table)
   ============================================ */

const MCP_PAGE_SIZE = 7

function mcpStatus(healthLatency: number): { label: string; className: string } {
  if (healthLatency === 0) return { label: translate('common.offline'), className: 'bg-text-muted' }
  if (healthLatency >= 3000) return { label: translate('adminSystemInfo.degraded'), className: 'bg-amber-400' }
  return { label: translate('common.normal'), className: 'bg-emerald-500' }
}

function McpHealthPanel({ data }: { data: McpHealthItem[] }) {
  const t = useT()
  const [page, setPage] = useState(0)
  const totalPages = Math.max(1, Math.ceil(data.length / MCP_PAGE_SIZE))
  const pageData = data.slice(page * MCP_PAGE_SIZE, (page + 1) * MCP_PAGE_SIZE)

  useEffect(() => {
    setPage(0)
  }, [data.length])

  if (data.length === 0) {
    return (
      <div className="flex flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
        <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.mcpStatus')}</h3>
        <p className="mt-3 text-xs text-text-3">{t('adminSystemInfo.noMcp')}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.mcpStatus')}</h3>
        <span className="text-xs text-text-3">{data.length} {t('adminSystemInfo.services')}</span>
      </div>
      <div className="mt-3 overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border text-left text-text-3">
              <th className="pb-2 pr-4 font-normal">{t('common.status')}</th>
              <th className="pb-2 pr-4 font-normal">{t('common.name')}</th>
              <th className="pb-2 pr-4 font-normal">{translate('adminMcp.latency')}</th>
              <th className="pb-2 font-normal">{t('adminSystemInfo.lastCheck')}</th>
            </tr>
          </thead>
          <tbody>
            {pageData.map((item) => {
              const status = mcpStatus(item.healthLatency)
              return (
                <tr key={item.mcpId} className="border-b border-border last:border-0">
                  <td className="py-2.5 pr-4">
                    <span className="inline-flex items-center gap-1.5">
                      <span className={cn('inline-block h-2 w-2 rounded-full', status.className)} />
                      <span className="text-text-2">{status.label}</span>
                    </span>
                  </td>
                  <td className="py-2.5 pr-4 text-text-1">{item.displayName}</td>
                  <td className="py-2.5 pr-4 tabular-nums text-text-2">
                    {item.healthLatency === 0 ? '—' : `${item.healthLatency}ms`}
                  </td>
                  <td className="py-2.5 tabular-nums text-text-3">
                    {formatRelativeTime(item.healthCheckedAt)}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="mt-3 flex items-center justify-end border-t border-border pt-3">
          <div className="flex items-center gap-1">
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1 disabled:opacity-30"
            >
              <ChevronLeft className="size-3" />
              {t('adminSystemInfo.previousPage')}
            </button>
            {Array.from({ length: totalPages }, (_, i) => (
              <button
                key={i}
                onClick={() => setPage(i)}
                className={cn(
                  'min-w-[24px] rounded px-1 py-0.5 text-xs tabular-nums transition-colors',
                  i === page
                    ? 'bg-bg-layer-3 text-text-1'
                    : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
                )}
              >
                {i + 1}
              </button>
            ))}
            <button
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page === totalPages - 1}
              className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1 disabled:opacity-30"
            >
              {t('adminSystemInfo.nextPage')}
              <ChevronRight className="size-3" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

/* ============================================
   Panel: Disks
   ============================================ */

function diskBarColor(usedPercent: number): string {
  if (usedPercent >= 90) return 'bg-[oklch(0.55_0.18_20)]'
  if (usedPercent >= 80) return 'bg-amber-400'
  return 'bg-interactive'
}

function DisksPanel({ data }: { data: DiskInfo[] }) {
  const t = useT()
  if (data.length === 0) {
    return (
      <div className="flex flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
        <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.systemDisks')}</h3>
        <p className="mt-3 text-xs text-text-3">{t('adminSystemInfo.noDisks')}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
      <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.systemDisks')}</h3>
      <div className="mt-3 overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border text-left text-text-3">
              <th className="pb-2 pr-4 font-normal">{t('adminSystemInfo.device')}</th>
              <th className="pb-2 pr-4 font-normal">{t('adminSystemInfo.mountPoint')}</th>
              <th className="pb-2 pr-4 font-normal">{t('adminSystemInfo.fileSystem')}</th>
              <th className="pb-2 pr-4 font-normal">{t('adminSystemInfo.usedTotal')}</th>
              <th className="pb-2 font-normal">{t('adminSystemInfo.usageRate')}</th>
            </tr>
          </thead>
          <tbody>
            {data.map((disk, i) => (
              <tr key={i} className="border-b border-border last:border-0">
                <td className="py-2.5 pr-4 tabular-nums text-text-1">{disk.device}</td>
                <td className="py-2.5 pr-4 tabular-nums text-text-1">{disk.mountPoint}</td>
                <td className="py-2.5 pr-4 tabular-nums text-text-2">{disk.fsType}</td>
                <td className="py-2.5 pr-4 tabular-nums text-text-2">
                  {formatBytes(disk.usedBytes)} / {formatBytes(disk.totalBytes)}
                </td>
                <td className="py-2.5">
                  <div className="flex items-center gap-2">
                    <div className="h-2.5 flex-1 overflow-hidden rounded-sm bg-bg-layer-2">
                      <div
                        className={cn('h-full transition-all', diskBarColor(disk.usedPercent))}
                        style={{ width: `${Math.min(disk.usedPercent, 100)}%` }}
                      />
                    </div>
                    <span className="w-10 shrink-0 text-right tabular-nums text-text-2">
                      {disk.usedPercent.toFixed(1)}%
                    </span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

/* ============================================
   Panel: Ecosystem
   ============================================ */

function EcosystemPanel({ data }: { data: EcosystemInfo }) {
  const t = useT()
  const groups = [
    {
      label: t('common.user'),
      items: [
        { label: t('adminSystemInfo.total'), value: String(data.totalUsers) },
        { label: t('adminSystemInfo.active'), value: String(data.activeUsers) },
        { label: t('common.admin'), value: String(data.adminUsers) },
      ],
    },
    {
      label: t('common.model'),
      items: [
        { label: t('adminSystemInfo.total'), value: String(data.totalModels) },
        { label: t('adminSystemInfo.enabled'), value: String(data.enabledModels) },
      ],
    },
    {
      label: 'MCP',
      items: [
        { label: t('adminSystemInfo.total'), value: String(data.totalMcps) },
        { label: t('adminSystemInfo.enabled'), value: String(data.enabledMcps) },
      ],
    },
    {
      label: t('common.skills'),
      items: [
        { label: translate('adminSkills.global'), value: String(data.systemSkills) },
        { label: t('adminSystemInfo.market'), value: String(data.marketSkills) },
      ],
    },
    {
      label: t('adminSystemInfo.roleTemplates'),
      items: [{ label: t('adminSystemInfo.total'), value: String(data.personaTemplates) }],
    },
    {
      label: t('adminSystemInfo.sessionsMessages'),
      items: [
        { label: t('common.sessions'), value: formatNumber(data.totalSessions) },
        { label: t('adminSystemInfo.messages'), value: formatNumber(data.totalMessages) },
      ],
    },
    {
      label: t('files.teamProjects'),
      items: [
        { label: t('adminSystemInfo.total'), value: formatNumber(data.totalSharedProjects) },
      ],
    },
    {
      label: t('common.share'),
      items: [
        { label: t('adminSystemInfo.total'), value: formatNumber(data.totalShareLinks) },
        { label: t('adminSystemInfo.activeFiles'), value: formatNumber(data.activeShareLinks) },
        { label: t('adminSystemInfo.views'), value: formatNumber(data.totalShareViews) },
      ],
    },
  ]

  return (
    <div className="flex flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
      <h3 className="text-xs font-semibold text-text-2">{t('adminSystemInfo.ecosystemOverview')}</h3>
      <div className="mt-3 grid grid-cols-4 gap-x-8 gap-y-3">
        {groups.map((group) => (
          <div key={group.label}>
            <div className="text-xs text-text-3">{group.label}</div>
            <div className="mt-1 flex items-baseline gap-2">
              {group.items.map((item, i) => (
                <span key={item.label} className="inline-flex items-baseline gap-1">
                  {i > 0 && <span className="text-text-muted">/</span>}
                  <span className="text-sm tabular-nums text-text-1">{item.value}</span>
                  <span className="text-xs text-text-3">{item.label}</span>
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ============================================
   Plugins footer
   ============================================ */

function PluginsFooter({ data }: { data: PluginInfo[] }) {
  return (
    <div className="px-5 py-2 text-xs text-text-3">
      {data.length === 0 ? (
        translate('adminSystemInfo.noPlugins')
      ) : (
        <span>
          {translate('adminSystemInfo.plugins')}{' '}
          {data.map((p, i) => (
            <span key={p.name}>
              {i > 0 && ', '}
              {p.name} {p.version}
            </span>
          ))}
        </span>
      )}
    </div>
  )
}

/* ============================================
   Skeleton
   ============================================ */

function Skeleton() {
  return (
    <div className="flex flex-1 flex-col gap-2 overflow-auto p-2">
      {/* Row 1: Ecosystem */}
      <div className="flex min-h-[120px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
          <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-3 grid grid-cols-4 gap-4">
            {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
              <div key={i}>
                <div className="h-2.5 w-10 animate-pulse rounded bg-bg-layer-2" />
                <div className="mt-1 h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Row 2: Overview + Database */}
      <div className="flex min-h-[160px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
          <div className="h-3 w-14 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-3 flex-1 space-y-2.5">
            {[1, 2, 3, 4, 5, 6, 7].map((j) => (
              <div key={j} className="flex justify-between">
                <div className="h-2.5 w-16 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
        </div>
        <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
          <div className="h-3 w-14 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-3 flex-1 space-y-2.5">
            {[1, 2, 3, 4].map((j) => (
              <div key={j} className="flex justify-between">
                <div className="h-2.5 w-16 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
          <div className="mt-3 h-12 animate-pulse rounded bg-bg-layer-2" />
        </div>
      </div>

      {/* Row 3: MCP Health */}
      <div className="flex min-h-[100px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
          <div className="h-3 w-24 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-3 space-y-2">
            {[1, 2, 3, 4, 5, 6].map((j) => (
              <div key={j} className="flex gap-8">
                <div className="h-2.5 w-12 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-24 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-16 animate-pulse rounded bg-bg-layer-2" />
                <div className="h-2.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Row 4: Disks */}
      <div className="flex min-h-[80px] gap-2">
        <div className="flex flex-1 flex-col rounded-lg border border-border bg-bg-layer-1 px-5 py-4">
          <div className="h-3 w-16 animate-pulse rounded bg-bg-layer-2" />
          <div className="mt-3 flex gap-8">
            <div className="h-2.5 w-28 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2.5 w-16 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2.5 w-12 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2.5 w-32 animate-pulse rounded bg-bg-layer-2" />
            <div className="h-2.5 flex-1 animate-pulse rounded bg-bg-layer-2" />
          </div>
        </div>
      </div>

      <div className="px-5 py-2">
        <div className="h-2.5 w-32 animate-pulse rounded bg-bg-layer-2" />
      </div>
    </div>
  )
}

/* ============================================
   Error
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{translate('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">{translate('adminSystemInfo.fetchFailed')}</p>
      </div>
      <button
        onClick={onRetry}
        className="inline-flex items-center gap-1 rounded-md bg-bg-layer-2 px-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-layer-3"
      >
        <RefreshCw className="size-3" />
        {translate('common.retry')}
      </button>
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

type PageState = 'loading' | 'data' | 'error'

export function Component() {
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [data, setData] = useState<SystemInfoData | null>(null)

  const loadData = useCallback(() => {
    setState('loading')
    adminSystemApi.getSystemInfo().then((info) => {
      setData(info as unknown as SystemInfoData)
      setState('data')
    }).catch(() => {
      setState('error')
    })
  }, [])

  const handleRefresh = useCallback(() => {
    loadData()
  }, [loadData])

  useEffect(() => {
    loadData()
  }, [loadData])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminSystemInfo.title')}</h1>
        <button
          onClick={handleRefresh}
          className="flex items-center gap-1 rounded-sm px-1.5 py-0.5 text-xs text-highlight transition-colors hover:bg-bg-hover"
        >
          <RefreshCw className="size-3" />
          {t('common.refresh')}
        </button>
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}

      {state === 'error' && <ErrorState onRetry={handleRefresh} />}

      {state === 'data' && data && (
        <div className="flex flex-1 flex-col gap-2 overflow-auto p-2">
          {/* Row 1: Ecosystem */}
          <EcosystemPanel data={data.ecosystem} />

          {/* Row 2: Overview + Database */}
          <div className="flex gap-2">
            <OverviewPanel data={data.overview} />
            <DatabasePanel data={data.database} />
          </div>

          {/* Row 3: MCP Health */}
          <McpHealthPanel data={data.mcpHealth} />

          {/* Row 4: Disks */}
          <DisksPanel data={data.disks} />

          {/* Plugins footer */}
          <PluginsFooter data={data.plugins} />
        </div>
      )}
    </div>
  )
}
