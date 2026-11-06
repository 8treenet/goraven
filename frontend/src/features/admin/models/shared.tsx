import { useMemo } from 'react'
import { useT } from '@/i18n'
import {
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'

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
  conversationHeaderKey: string
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
  conversationHeaderKey: string
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
          'peer-checked:border-success peer-checked:bg-success',
          'border-border-strong bg-bg-layer-2',
          'peer-hover:border-text-2',
        )}
      />
      <div
        className={cn(
          'absolute left-0.5 top-0.5 size-3 rounded-full bg-bg-base shadow-sm transition-transform',
          'peer-checked:translate-x-4',
        )}
      />
    </label>
  )
}

/* ============================================
   Drawer
   ============================================ */



/* ============================================
   Skeleton
   ============================================ */

export function TableSkeleton() {
  const t = useT()
  return (
    <div className="flex-1 overflow-auto">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="pl-4 pr-2">{t('adminModels.provider')}</TableHead>
            <TableHead className="pr-2">{t('adminModels.modelDisplayName')}</TableHead>
            <TableHead className="pr-2">{t('common.model')}</TableHead>
            <TableHead className="pr-2">{t('adminModels.labels')}</TableHead>
            <TableHead className="pr-2 text-right">{t('adminModels.context')}</TableHead>
            <TableHead className="pr-2">{t('adminModels.updatedAt')}</TableHead>
            <TableHead className="pr-4" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 8 }).map((_, i) => (
            <TableRow key={i} className="hover:bg-transparent">
              <TableCell className="h-12 py-0 pl-4 pr-2">
                <div className="flex items-center gap-2">
                  <div className="size-6 animate-pulse rounded-sm bg-bg-layer-3" />
                  <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
                </div>
              </TableCell>
              <TableCell className="h-12 py-0 pr-2">
                <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-2">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-2">
                <div className="h-5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-2">
                <div className="ml-auto h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-2">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="ml-auto h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
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
          className="inline-flex size-8 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
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
                'inline-flex size-8 items-center justify-center rounded-md text-xs tabular-nums transition-colors',
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
          className="inline-flex size-8 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronRight className="size-3.5" />
        </button>
      </div>
    </div>
  )
}
