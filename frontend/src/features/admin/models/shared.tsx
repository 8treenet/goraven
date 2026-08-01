import { useMemo } from 'react'
import { useT } from '@/i18n'
import {
  Zap,
  Plus,
  AlertCircle,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  X,
} from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

/* ============================================
   Types
   ============================================ */

export interface RecommendedModel {
  id: string
  object: string
  ownedBy: string
}

export interface ModelFormData {
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  remark: string
}

export interface ModelEditFormData {
  providerDisplayName: string
  displayName: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  status: number
  remark: string
}

export type DrawerMode = 'add' | 'edit'
export type DialogMode = 'delete' | null

export type PageState = 'loading' | 'data' | 'empty' | 'error'
export type TestState = 'idle' | 'testing' | 'success' | 'fail'

/* ============================================
   Helpers
   ============================================ */

export function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

/* ============================================
   Status Toggle (reused from users page)
   ============================================ */

export function StatusToggle({
  checked,
  onChange,
}: {
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label className="relative inline-flex cursor-pointer items-center">
      <input
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <div
        className={cn(
          'h-4 w-8 rounded-full border transition-colors',
          'peer-checked:border-highlight peer-checked:bg-highlight',
          'border-border-strong bg-bg-layer-2',
          'peer-hover:border-text-2',
        )}
      />
      <div
        className={cn(
          'absolute left-0.5 top-0.5 size-3 rounded-full bg-white shadow-sm transition-transform',
          'peer-checked:translate-x-4',
        )}
      />
    </label>
  )
}

/* ============================================
   Drawer
   ============================================ */

export function Drawer({
  open,
  onClose,
  title,
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <div
        className={cn(
          'fixed inset-0 z-50',
          open ? 'visible' : 'invisible',
        )}
      >
        <div
          className={cn(
            'absolute inset-0 bg-black/60 transition-opacity duration-200',
            open ? 'opacity-100' : 'opacity-0',
          )}
          onClick={onClose}
        />
        <div
          className={cn(
            'absolute right-0 top-0 z-50 flex h-full w-[400px] flex-col border-l border-border bg-bg-layer-1 shadow-pop transition-transform duration-200 ease-out',
            open ? 'translate-x-0' : 'translate-x-full',
          )}
        >
          <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
            <h2 className="text-sm font-semibold text-text-1">{title}</h2>
            <button
              onClick={onClose}
              className="rounded-sm p-0.5 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
            >
              <X className="size-4" />
            </button>
          </div>
          <div className="flex-1 overflow-auto p-4">
            {children}
          </div>
        </div>
      </div>
    </Dialog>
  )
}

/* ============================================
   Skeleton
   ============================================ */

export function TableSkeleton() {
  const t = useT()
  return (
    <div className="flex-1 overflow-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border text-left text-xs text-text-3">
            <th className="pb-2 pl-4 pr-2 font-normal">{t('adminModels.provider')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.modelDisplayName')}</th>
            <th className="pb-2 pr-2 font-normal">{t('common.model')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.labels')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.context')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.updatedAt')}</th>
            <th className="pb-2 pr-4 font-normal" />
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: 8 }).map((_, i) => (
            <tr key={i}>
              <td className="py-2.5 pl-4 pr-2">
                <div className="flex items-center gap-2">
                  <div className="size-7 animate-pulse rounded-sm bg-bg-layer-3" />
                  <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
                </div>
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/* ============================================
   Empty State
   ============================================ */

export function EmptyState({ hasFilter, onClearFilter, onAdd }: {
  hasFilter: boolean
  onClearFilter: () => void
  onAdd: () => void
}) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3">
      <div className="flex size-10 items-center justify-center rounded-full bg-bg-layer-2">
        <Zap className="size-5 text-text-muted" />
      </div>
      <div className="text-center">
        <p className="text-sm text-text-2">
          {hasFilter ? t('adminModels.noMatch') : t('adminModels.noModels')}
        </p>
      </div>
      {hasFilter ? (
        <button
          onClick={onClearFilter}
          className="text-xs text-interactive transition-colors hover:text-interactive-hover"
        >
          {t('adminUsers.clearFilter')}
        </button>
      ) : (
        <Button variant="default" size="sm" onClick={onAdd}>
          <Plus className="size-3.5" />
          {t('adminModels.addFirst')}
        </Button>
      )}
    </div>
  )
}

/* ============================================
   Error State
   ============================================ */

export function ErrorState({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">{t('adminModels.fetchFailedList')}</p>
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
    <div className="flex items-center justify-between border-t border-border px-4 py-2">
      <span className="text-xs text-text-3">{t('common.total')} {totalCount} {t('adminModels.totalModels')}</span>
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
