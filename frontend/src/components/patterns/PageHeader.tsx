import * as React from 'react'
import { cn } from '@/lib/utils'

interface PageHeaderProps {
  title: React.ReactNode
  description?: React.ReactNode
  actions?: React.ReactNode
  className?: string
  contained?: boolean
}

export function PageHeader({ title, description, actions, className, contained }: PageHeaderProps) {
  return (
    <header className={cn('shrink-0 border-b border-border bg-bg-base', className)}>
      <div
        className={cn(
          'flex min-h-16 w-full flex-wrap items-center justify-between gap-x-4 gap-y-2 px-6',
          contained && 'mx-auto max-w-[1180px]',
        )}
      >
        <div className="min-w-0">
          <h1 className="truncate text-title font-semibold text-text-1">{title}</h1>
          {description && <p className="mt-0.5 truncate text-meta text-text-3">{description}</p>}
        </div>
        {actions && <div className="flex min-w-0 flex-wrap items-center justify-end gap-2">{actions}</div>}
      </div>
    </header>
  )
}
