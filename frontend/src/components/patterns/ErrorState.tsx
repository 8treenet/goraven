import * as React from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'
import { cn } from '@/lib/utils'

interface ErrorStateProps {
  title?: React.ReactNode
  description?: React.ReactNode
  onRetry?: () => void
  retryLabel?: React.ReactNode
  className?: string
}

export function ErrorState({ title, description, onRetry, retryLabel, className }: ErrorStateProps) {
  const t = useT()

  return (
    <div className={cn('flex flex-1 flex-col items-center justify-center gap-4 py-24', className)}>
      <AlertCircle className="size-8 text-destructive" />
      <div className="text-center">
        <p className="text-body text-text-2">{title ?? t('common.loadFailed')}</p>
        {description && <p className="mt-1 text-meta text-text-3">{description}</p>}
      </div>
      {onRetry && (
        <Button variant="outline" size="default" onClick={onRetry}>
          <RefreshCw className="size-4" />
          {retryLabel ?? t('common.retry')}
        </Button>
      )}
    </div>
  )
}
