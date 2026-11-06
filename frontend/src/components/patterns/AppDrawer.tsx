import * as React from 'react'
import { X } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetDescription,
  SheetClose,
} from '@/components/ui/sheet'
import { cn } from '@/lib/utils'

type AppDrawerWidth = 'sm' | 'md' | 'lg' | 'xl' | '2xl'

const WIDTH_CLASS: Record<AppDrawerWidth, string> = {
  sm: 'w-[400px] sm:max-w-none',
  md: 'w-[480px] sm:max-w-none',
  lg: 'w-[680px] sm:max-w-none',
  xl: 'w-[720px] sm:max-w-none',
  '2xl': 'w-[760px] sm:max-w-none',
}

interface AppDrawerProps {
  open: boolean
  onClose: () => void
  title: React.ReactNode
  description?: React.ReactNode
  width?: AppDrawerWidth
  footer?: React.ReactNode
  flush?: boolean
  bodyClassName?: string
  children: React.ReactNode
}

export function AppDrawer({
  open,
  onClose,
  title,
  description,
  width = 'lg',
  footer,
  flush,
  bodyClassName,
  children,
}: AppDrawerProps) {
  return (
    <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
      <SheetContent hideClose className={WIDTH_CLASS[width]}>
        <div className="flex min-h-12 shrink-0 items-center justify-between border-b border-border px-5 py-3">
          <div className="min-w-0">
            <SheetTitle className="truncate text-section font-semibold text-text-1">{title}</SheetTitle>
            {description && (
              <SheetDescription className="mt-0.5 truncate text-meta text-text-3">
                {description}
              </SheetDescription>
            )}
          </div>
          <SheetClose asChild>
            <button
              type="button"
              className="rounded-md p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
              aria-label="Close"
            >
              <X className="size-4" />
            </button>
          </SheetClose>
        </div>
        <div className={cn('flex-1', flush ? 'overflow-hidden' : 'overflow-auto p-5', bodyClassName)}>
          {children}
        </div>
        {footer && <div className="shrink-0 border-t border-border px-5 py-4">{footer}</div>}
      </SheetContent>
    </Sheet>
  )
}

export type { AppDrawerWidth, AppDrawerProps }
