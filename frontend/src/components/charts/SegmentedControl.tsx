import { cn } from '@/lib/utils'

export function SegmentedControl<T extends string | number>({
  value,
  options,
  onChange,
}: {
  value: T
  options: { label: string; value: T }[]
  onChange: (v: T) => void
}) {
  return (
    <div className="inline-flex items-center rounded-md bg-bg-layer-2 p-0.5">
      {options.map((opt) => (
        <button
          key={String(opt.value)}
          onClick={() => onChange(opt.value)}
          className={cn(
            'rounded-sm px-2 py-0.5 text-xs transition-colors',
            value === opt.value
              ? 'bg-bg-layer-3 text-text-1'
              : 'text-text-3 hover:text-text-2',
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
