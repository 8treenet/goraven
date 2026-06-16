import { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Icon, renderIcon } from './Icon'
import { ICON_CATEGORIES, DEFAULT_ICON, type IconName } from './icon-registry'

/* ============================================
   IconPicker
   ============================================ */

export interface IconPickerProps {
  /** Currently selected icon name */
  value: string
  /** Called when user selects an icon */
  onChange: (name: IconName) => void
  /** Additional class on the root element */
  className?: string
}

/**
 * Icon picker with category tabs. 8 categories × 16 icons = 2 rows each.
 *
 * Two layouts:
 * - **Inline** (default): renders the full picker inline in a form
 * - **Trigger button** via <IconPickerTrigger>: shows a button that opens a popover
 */
export function IconPicker({ value, onChange, className }: IconPickerProps) {
  const [activeCategory, setActiveCategory] = useState(ICON_CATEGORIES[0].key)

  const category = ICON_CATEGORIES.find(c => c.key === activeCategory) ?? ICON_CATEGORIES[0]

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      {/* Category tabs */}
      <div className="flex flex-wrap gap-1">
        {ICON_CATEGORIES.map(cat => (
          <button
            key={cat.key}
            type="button"
            onClick={() => setActiveCategory(cat.key)}
            className={cn(
              'rounded px-2 py-0.5 text-xs transition-colors',
              activeCategory === cat.key
                ? 'bg-bg-layer-3 text-text-1'
                : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
            )}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* Icon grid */}
      <div className="grid grid-cols-8 gap-1">
        {category.icons.map(name => (
          <button
            key={name}
            type="button"
            title={name}
            onClick={() => onChange(name)}
            className={cn(
              'inline-flex size-8 items-center justify-center rounded-md border transition-colors',
              value === name
                ? 'border-interactive bg-interactive/10 text-interactive'
                : 'border-border-strong text-text-3 hover:border-text-3 hover:text-text-2',
            )}
          >
            {renderIcon(name, 'size-4')}
          </button>
        ))}
      </div>
    </div>
  )
}

/* ============================================
   IconPickerTrigger (button + popover)
   ============================================ */

export interface IconPickerTriggerProps {
  value: string
  onChange: (name: IconName) => void
  className?: string
}

/**
 * A trigger button that opens a popover icon picker.
 *
 * <IconPickerTrigger value={icon} onChange={setIcon} />
 */
export function IconPickerTrigger({ value, onChange, className }: IconPickerTriggerProps) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        className="flex h-8 w-full items-center gap-2 rounded-lg border border-input bg-transparent px-2.5 text-sm text-text-1 transition-colors hover:bg-bg-hover"
        onClick={() => setOpen(!open)}
      >
        <Icon name={value} className="size-4 text-text-2" />
        <span className="text-text-3">{value}</span>
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-1 w-72 rounded-md border border-border bg-bg-layer-1 p-3 shadow-pop">
          <IconPicker
            value={value}
            onChange={name => { onChange(name); setOpen(false) }}
          />
        </div>
      )}
    </div>
  )
}

export { DEFAULT_ICON, type IconName }
