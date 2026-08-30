import { useState, useEffect, useRef } from 'react'
import { useT } from '@/i18n'
import {
  MoreHorizontal,
  Pencil,
  Copy,
  Users,
  Star,
  Zap,
  Image,
  Trash2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { AdminModelItem } from '@/api'
import { formatDate } from './shared'

/* ============================================
   Row Actions Dropdown
   ============================================ */

function RowActionsMenu({
  onEdit,
  onDelete,
  onDuplicate,
  onMembers,
  onSetDefault,
  onSetFlash,
  onSetVisual,
  isDefault,
  isFlash,
  isVisual,
}: {
  onEdit: () => void
  onDelete: () => void
  onDuplicate: () => void
  onMembers: () => void
  onSetDefault: () => void
  onSetFlash: () => void
  onSetVisual: () => void
  isDefault: boolean
  isFlash: boolean
  isVisual: boolean
}) {
  const t = useT()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
      >
        <MoreHorizontal className="size-3.5" />
      </button>
      {open && (
        <div
          className={cn(
            'absolute right-0 top-full z-30 mt-1 w-44 overflow-hidden rounded-md border border-border bg-bg-layer-1 shadow-pop',
          )}
        >
          <button
            onClick={() => { onEdit(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Pencil className="size-3.5 text-text-3" />
            {t('common.edit')}
          </button>
          <button
            onClick={() => { onDuplicate(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Copy className="size-3.5 text-text-3" />
            {t('common.duplicate')}
          </button>
          <button
            onClick={() => { onMembers(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Users className="size-3.5 text-text-3" />
            {t('adminModels.members')}
          </button>
          <div className="border-t border-border" />
          <button
            onClick={() => { onSetDefault(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Star className={cn('size-3.5', isDefault ? 'text-highlight' : 'text-text-3')} />

            {isDefault ? t('adminModels.unsetDefault') : t('adminModels.setAsDefault')}
          </button>
          {!isFlash && (
            <button
              onClick={() => { onSetFlash(); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
            >
              <Zap className="size-3.5 text-text-3" />

              {t('adminModels.setAsFlash')}
            </button>
          )}
          {!isVisual && (
            <button
              onClick={() => { onSetVisual(); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
            >
              <Image className="size-3.5 text-text-3" />

              {t('adminModels.setAsMultimodal')}
            </button>
          )}
          <div className="border-t border-border" />
          <button
            onClick={() => { onDelete(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-destructive transition-colors hover:bg-bg-hover"
          >
            <Trash2 className="size-3.5" />
            {t('common.delete')}
          </button>
        </div>
      )}
    </div>
  )
}

/* ============================================
   Table Row
   ============================================ */

function ModelIcon({ icon, name }: { icon: string; name: string }) {
  const [error, setError] = useState(false)

  if (!icon || error) {
    return (
      <div className="flex size-7 shrink-0 items-center justify-center rounded-sm bg-bg-layer-3 text-xs font-medium text-interactive">
        {name.charAt(0).toUpperCase()}
      </div>
    )
  }

  return (
    <span className="inline-flex size-7 shrink-0 items-center justify-center rounded-sm bg-white/20">
      <img
        src={icon}
        alt={name}
        className="size-5 rounded-sm object-contain"
        onError={() => setError(true)}
      />
    </span>
  )
}

export function ModelRow({
  model,
  onEdit,
  onDelete,
  onDuplicate,
  onMembers,
  onSetDefault,
  onSetFlash,
  onSetVisual,
}: {
  model: AdminModelItem
  onEdit: () => void
  onDelete: () => void
  onDuplicate: () => void
  onMembers: () => void
  onSetDefault: () => void
  onSetFlash: () => void
  onSetVisual: () => void
}) {
  const t = useT()
  return (
    <tr className="transition-colors hover:bg-bg-hover">
      <td className="py-2.5 pl-4 pr-2">
        <div className="flex items-center gap-2">
          <ModelIcon icon={model.icon} name={model.providerDisplayName} />
          <span className="text-sm font-medium text-text-1">{model.providerDisplayName}</span>
        </div>
      </td>
      <td className="py-2.5 pr-2 text-sm font-medium text-text-1">{model.displayName}</td>
      <td className="py-2.5 pr-2 text-sm font-mono text-text-1">{model.modelName}</td>
      <td className="py-2.5 pr-2">
        <div className="flex items-center gap-1">
          {model.isDefault === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-highlight/15 text-highlight">
              <Star className="size-2.5" />
              {t('common.default')}
            </span>
          )}
          {model.isFlash === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-interactive text-white">
              <Zap className="size-2.5" />
              {t('adminModels.flash')}
            </span>
          )}
          {model.isVisual === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-interactive/15 text-interactive">
              <Image className="size-2.5" />
              {t('adminModels.multimodal')}
            </span>
          )}
          {model.isDefault === 0 && model.isFlash === 0 && model.isVisual === 0 && (
            <span className="text-xs text-text-muted">—</span>
          )}
        </div>
      </td>
      <td className="py-2.5 pr-2">
        <span className={cn(
          'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs',
          model.access === 1 ? 'bg-interactive/10 text-interactive' : 'bg-highlight/15 text-highlight',
        )}>
          <Users className="size-3" />
          {model.access === 1 ? t('adminModels.accessMembers') : t('adminModels.accessAll')}
        </span>
      </td>
      <td className="py-2.5 pr-2 text-sm tabular-nums text-text-2">{model.contextLen}K</td>
      <td className="py-2.5 pr-2 text-sm text-text-3">{formatDate(model.updated)}</td>
      <td className="py-2.5 pr-4">
        <RowActionsMenu
          onEdit={onEdit}
          onDelete={onDelete}
          onDuplicate={onDuplicate}
          onMembers={onMembers}
          onSetDefault={onSetDefault}
          onSetFlash={onSetFlash}
          onSetVisual={onSetVisual}
          isDefault={model.isDefault === 1}
          isFlash={model.isFlash === 1}
          isVisual={model.isVisual === 1}
        />
      </td>
    </tr>
  )
}
