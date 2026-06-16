import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { SortField, SortOrder } from './file-helpers'

interface ContextMenuProps {
  x: number
  y: number
  onClose: () => void
  children: React.ReactNode
}

export function ContextMenu({ x, y, onClose, children }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState({ x, y })

  useEffect(() => {
    if (menuRef.current) {
      const rect = menuRef.current.getBoundingClientRect()
      const vw = window.innerWidth
      const vh = window.innerHeight
      setPos({
        x: x + rect.width > vw ? vw - rect.width - 8 : x,
        y: y + rect.height > vh ? vh - rect.height - 8 : y,
      })
    }
  }, [x, y])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    const keyHandler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('mousedown', handler)
    document.addEventListener('keydown', keyHandler)
    return () => {
      document.removeEventListener('mousedown', handler)
      document.removeEventListener('keydown', keyHandler)
    }
  }, [onClose])

  return createPortal(
    <div
      ref={menuRef}
      className="fixed z-50 min-w-[140px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop"
      style={{ left: pos.x, top: pos.y }}
    >
      {children}
    </div>,
    document.body,
  )
}

interface ContextMenuItemProps {
  onClick: () => void
  children: React.ReactNode
  danger?: boolean
}

export function ContextMenuItem({ onClick, children, danger }: ContextMenuItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-2 px-3 py-1.5 text-sm transition-colors text-left',
        danger
          ? 'text-text-2 hover:bg-bg-hover hover:text-text-1'
          : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
      )}
    >
      {children}
    </button>
  )
}

interface ColumnHeaderProps {
  label: string
  field: SortField
  currentField: SortField
  order: SortOrder
  onClick: (field: SortField) => void
  className?: string
}

export function ColumnHeader({ label, field, currentField, order, onClick, className }: ColumnHeaderProps) {
  const active = currentField === field
  return (
    <button
      onClick={() => onClick(field)}
      className={cn(
        'flex items-center gap-0.5 text-sm transition-colors select-none',
        active ? 'text-text-1' : 'text-text-3 hover:text-text-2',
        className,
      )}
    >
      {label}
      {active && (
        order === 'asc'
          ? <ChevronUp className="size-3.5" />
          : <ChevronDown className="size-3.5" />
      )}
    </button>
  )
}
