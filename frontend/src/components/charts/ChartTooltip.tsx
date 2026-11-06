import { formatNumber } from '@/lib/format'

export function ChartTooltip({
  active,
  payload,
  label,
  showTotal = false,
}: {
  active?: boolean
  payload?: { name: string; value: number; color: string }[]
  label?: string
  showTotal?: boolean
}) {
  if (!active || !payload?.length) return null
  const total = payload.reduce((s, p) => s + (typeof p.value === 'number' ? p.value : 0), 0)
  return (
    <div className="rounded-lg border border-border bg-bg-layer-2 px-3 py-2 shadow-pop">
      {label && <p className="text-micro text-text-3">{label}</p>}
      {payload.map((p, i) => (
        <p key={i} className="flex items-center gap-1.5 text-micro text-text-2 tabular-nums">
          <span aria-hidden="true" className="size-1.5 shrink-0 rounded-full" style={{ backgroundColor: p.color }} />
          {p.name}: {formatNumber(p.value)}
        </p>
      ))}
      {showTotal && (
        <p className="mt-1 border-t border-border pt-1 text-micro font-medium text-text-1 tabular-nums">
          Total: {formatNumber(total)}
        </p>
      )}
    </div>
  )
}
