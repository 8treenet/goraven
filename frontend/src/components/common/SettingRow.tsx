import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface SettingRowProps {
  label: string
  description?: string
  isDirty: boolean
  children: ReactNode
  className?: string
}

export function SettingRow({
  label,
  description,
  isDirty,
  children,
  className,
}: SettingRowProps) {
  return (
    <div
      className={cn(
        'grid grid-cols-[1fr_auto] items-center gap-4 px-5 py-3.5',
        className,
      )}
    >
      <div className="min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-sm text-text-1">{label}</span>
          {isDirty && (
            <span className="size-1.5 shrink-0 rounded-full bg-highlight" />
          )}
        </div>
        {description && (
          <p className="mt-0.5 text-xs text-text-3 leading-relaxed">
            {description}
          </p>
        )}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        {children}
      </div>
    </div>
  )
}
