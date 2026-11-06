import * as React from 'react'
import { Card } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import { TONE_SOFT, TONE_TEXT, type Tone } from '@/lib/tone'

type EmptyStateTone =
  | Tone
  | 'accent'
  | 'amber'
  | 'success'
  | 'warning'
  | 'info'
  | 'destructive'

const TONE_CLASS: Record<EmptyStateTone, string> = {
  1: `${TONE_SOFT[1]} ${TONE_TEXT[1]}`,
  2: `${TONE_SOFT[2]} ${TONE_TEXT[2]}`,
  3: `${TONE_SOFT[3]} ${TONE_TEXT[3]}`,
  4: `${TONE_SOFT[4]} ${TONE_TEXT[4]}`,
  5: `${TONE_SOFT[5]} ${TONE_TEXT[5]}`,
  accent: 'bg-accent-soft text-accent',
  amber: 'bg-amber-soft text-amber-strong',
  success: 'bg-success-soft text-success',
  warning: 'bg-warning-soft text-warning',
  info: 'bg-info-soft text-info',
  destructive: 'bg-destructive-soft text-destructive',
}

interface EmptyStateProps {
  icon?: React.ReactNode
  title: React.ReactNode
  description?: React.ReactNode
  action?: React.ReactNode
  variant?: 'page' | 'card'
  tone?: EmptyStateTone
  className?: string
}

export function EmptyState({
  icon,
  title,
  description,
  action,
  variant = 'page',
  tone,
  className,
}: EmptyStateProps) {
  if (variant === 'card') {
    return (
      <Card
        className={cn(
          'flex min-h-[280px] flex-col items-center justify-center px-6 text-center',
          className,
        )}
      >
        {icon && (
          <div
            className={cn(
              'mb-3 flex items-center justify-center',
              tone && 'size-12 rounded-full',
              tone && TONE_CLASS[tone],
            )}
          >
            {icon}
          </div>
        )}
        <h3 className="text-section font-semibold text-text-1">{title}</h3>
        {description && <p className="mt-1 max-w-sm text-meta text-text-3">{description}</p>}
        {action && <div className="mt-4">{action}</div>}
      </Card>
    )
  }

  return (
    <div className={cn('flex flex-1 flex-col items-center justify-center gap-3 py-24', className)}>
      {icon && (
        <div
          className={cn(
            'flex items-center justify-center',
            tone && 'size-12 rounded-full',
            tone && TONE_CLASS[tone],
          )}
        >
          {icon}
        </div>
      )}
      <div className="text-center">
        <p className="text-body text-text-2">{title}</p>
        {description && <p className="mt-1 text-meta text-text-3">{description}</p>}
      </div>
      {action}
    </div>
  )
}
