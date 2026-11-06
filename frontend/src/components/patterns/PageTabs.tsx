import * as React from 'react'
import { cn } from '@/lib/utils'
import { nextTabIndex } from './pageTabsKeyboard'

interface PageTabItem<T extends string> {
  value: T
  label: React.ReactNode
}

interface PageTabsProps<T extends string> {
  value: T
  onChange: (value: T) => void
  items: PageTabItem<T>[]
  className?: string
}

export function PageTabs<T extends string>({ value, onChange, items, className }: PageTabsProps<T>) {
  const buttonRefs = React.useRef<Array<HTMLButtonElement | null>>([])

  function handleKeyDown(event: React.KeyboardEvent<HTMLButtonElement>, index: number) {
    const next = nextTabIndex(index, event.key, items.length)
    if (next === null) return
    event.preventDefault()
    onChange(items[next].value)
    buttonRefs.current[next]?.focus()
  }

  return (
    <div role="tablist" aria-orientation="horizontal" className={cn('flex gap-5 border-b border-border', className)}>
      {items.map((item, index) => {
        const active = item.value === value
        return (
          <button
            key={item.value}
            ref={(node) => {
              buttonRefs.current[index] = node
            }}
            type="button"
            role="tab"
            aria-selected={active}
            tabIndex={active ? 0 : -1}
            onClick={() => onChange(item.value)}
            onKeyDown={(event) => handleKeyDown(event, index)}
            className={cn(
              'relative -mb-px pb-2.5 pt-1 text-[13px] transition-colors',
              active
                ? 'font-medium text-text-1 after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:rounded-full after:bg-accent'
                : 'text-text-3 hover:text-text-1',
            )}
          >
            {item.label}
          </button>
        )
      })}
    </div>
  )
}
