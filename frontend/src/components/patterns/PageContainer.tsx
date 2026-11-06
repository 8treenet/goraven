import * as React from 'react'
import { cn } from '@/lib/utils'

interface PageContainerProps {
  children: React.ReactNode
  className?: string
  wide?: boolean
  contained?: boolean
}

export function PageContainer({ children, className, wide, contained }: PageContainerProps) {
  return (
    <div
      className={cn(
        'w-full px-6 py-7',
        contained && (wide ? 'mx-auto max-w-[1320px]' : 'mx-auto max-w-[1180px]'),
        className,
      )}
    >
      {children}
    </div>
  )
}
