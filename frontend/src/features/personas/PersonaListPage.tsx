import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import {
  Plus,
  Ellipsis,
  Pencil,
  Trash2,
  AlertCircle,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Icon } from '@/components/common/Icon'

import { personasApi } from '@/api'
import type { PersonaListItem } from '@/api/types'

type PageState = 'loading' | 'data' | 'empty' | 'error'

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [personas, setPersonas] = useState<PersonaListItem[]>([])
  const [menuOpenId, setMenuOpenId] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<PersonaListItem | null>(null)

  const fetchPersonas = useCallback(() => {
    setState('loading')
    personasApi.getPersonas()
      .then((list) => {
        setPersonas(list)
        setState(list.length > 0 ? 'data' : 'empty')
      })
      .catch(() => setState('error'))
  }, [])

  useEffect(() => {
    fetchPersonas()
  }, [fetchPersonas])

  const handleCreate = useCallback(() => {
    navigate('/personas/new')
  }, [navigate])

  const handleEdit = useCallback((id: number) => {
    navigate(`/personas/${id}/edit`)
  }, [navigate])

  const handleRowClick = useCallback((id: number) => {
    navigate(`/personas/${id}`)
  }, [navigate])

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return
    try {
      await personasApi.deletePersona(deleteTarget.personaId)
      setPersonas(prev => prev.filter(p => p.personaId !== deleteTarget.personaId))
      setDeleteTarget(null)
      setMenuOpenId(null)
      toast.success(translate('personas.deleted'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    }
  }, [deleteTarget])

  const toggleMenu = useCallback((id: number) => {
    setMenuOpenId(prev => prev === id ? null : id)
  }, [])

  return (
    <div>
      {/* Sticky toolbar */}
      <div className="sticky top-0 z-10 flex h-10 items-center justify-between border-b border-border-custom bg-bg-base px-4">
        <h1 className="text-[18px] font-semibold text-text-1">{t('personas.title')}</h1>
        <Button size="default" onClick={handleCreate} className="text-highlight hover:text-highlight">
          <Plus className="size-4" />
          {t('personas.newPersona')}
        </Button>
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}
      {state === 'error' && <ErrorView onRetry={fetchPersonas} />}
      {state === 'empty' && <EmptyView onCreate={handleCreate} />}
      {state === 'data' && (
        <div>
          {personas.map((p) => {
            const allTags = [
              ...p.mcpNames.map((n) => ({ label: n, type: 'mcp' as const })),
              ...p.skillNames.map((n) => ({ label: n, type: 'skill' as const })),
            ]
            const visibleTags = allTags.slice(0, 4)
            const overflow = allTags.length - visibleTags.length

            return (
            <div
              key={p.personaId}
              className={cn(
                'flex items-start gap-3 px-4 transition-colors',
                personas.length <= 5 ? 'py-4' : 'py-3',
                'hover:bg-bg-hover',
              )}
            >
              <button
                className="flex flex-1 items-start gap-3 min-w-0 text-left"
                onClick={() => handleRowClick(p.personaId)}
              >
                <span className="shrink-0 mt-0.5">
                  <Icon name={p.icon} className="size-4 text-interactive" />
                </span>

                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate text-sm font-semibold text-text-1">{p.name}</span>
                    {p.categoryName && (
                      <span className="shrink-0 rounded px-1.5 py-0.5 text-[11px] text-text-2 border border-border-custom">
                        {p.categoryName}
                      </span>
                    )}
                  </div>

                  <p className="mt-1 text-xs text-text-3 leading-relaxed truncate max-w-[190ch]">
                    {p.roleInfo.replace(/\n/g, ' ')}
                  </p>

                  {allTags.length > 0 && (
                    <div className="mt-1.5 flex items-center gap-1 flex-wrap">
                      {visibleTags.map((t) => (
                        <span
                          key={t.type + t.label}
                          className="inline-flex items-center rounded-sm px-1 text-[11px] text-interactive bg-interactive/10"
                        >
                          {t.type === 'mcp' ? '@' : ''}{t.label}
                        </span>
                      ))}
                      {overflow > 0 && (
                        <span className="text-[11px] text-interactive">
                           +{overflow}
                         </span>
                      )}
                    </div>
                  )}
                </div>
              </button>

              {p.modelName && (
                <span className="hidden shrink-0 text-xs text-text-3 sm:block max-w-48 truncate text-right mt-0.5">
                  {p.modelName}
                </span>
              )}

              <div className="relative shrink-0">
                <button
                  className="flex h-7 w-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
                  onClick={(e) => { e.stopPropagation(); toggleMenu(p.personaId) }}
                >
                  <Ellipsis className="size-4" />
                </button>
                {menuOpenId === p.personaId && (
                  <>
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setMenuOpenId(null)}
                    />
                    <div className="absolute right-0 top-full z-20 mt-1 w-36 rounded-md border border-border-custom bg-bg-layer-2 py-1 shadow-pop">
                      <button
                        className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                        onClick={() => { setMenuOpenId(null); handleEdit(p.personaId) }}
                      >
                        <Pencil className="size-3" />
                        {t('common.edit')}
                      </button>
                      <button
                        className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
                        onClick={() => { setMenuOpenId(null); setDeleteTarget(p) }}
                      >
                        <Trash2 className="size-3" />
                        {t('common.delete')}
                      </button>
                    </div>
                  </>
                )}
              </div>
            </div>
          )})}
        </div>
      )}

      {/* Delete confirmation dialog */}
      <Dialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('personas.deletePersona')}</DialogTitle>
            <DialogDescription>
              {t('personas.deleteConfirm').replace('{name}', deleteTarget?.name ?? '')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" size="default" onClick={() => setDeleteTarget(null)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" size="default" onClick={handleDeleteConfirm}>
              {t('common.delete')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

/* ============================================
   Sub-views
   ============================================ */

function Skeleton() {
  return (
    <div>
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="flex items-start gap-3 border-b border-border-custom px-4 py-4">
          <span className="mt-0.5 size-4 shrink-0 rounded-sm bg-bg-layer-3 animate-pulse" />
          <div className="flex-1 space-y-2">
            <div className="flex items-center gap-2">
              <span className="h-4 w-28 rounded bg-bg-layer-3 animate-pulse" />
              <span className="h-5 w-10 rounded bg-bg-layer-3 animate-pulse" />
            </div>
            <span className="block h-3 w-full max-w-[480px] rounded bg-bg-layer-3 animate-pulse" />
            <div className="flex gap-1">
              <span className="h-4 w-20 rounded-sm bg-bg-layer-3 animate-pulse" />
              <span className="h-4 w-16 rounded-sm bg-bg-layer-3 animate-pulse" />
              <span className="h-4 w-24 rounded-sm bg-bg-layer-3 animate-pulse" />
            </div>
          </div>
          <span className="mt-0.5 h-4 w-24 rounded bg-bg-layer-3 animate-pulse hidden sm:block" />
          <span className="mt-0.5 h-7 w-7 shrink-0 rounded-md bg-bg-layer-3 animate-pulse" />
        </div>
      ))}
    </div>
  )
}

function EmptyView({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <Icon name="bot" className="size-8 text-text-muted" />
      <p className="mt-3 text-sm text-text-2">{translate('personas.noPersonas')}</p>
      <p className="mt-1 text-xs text-text-3">{translate('personas.noPersonasHint')}</p>
      <Button size="default" className="mt-4 text-highlight hover:text-highlight" onClick={onCreate}>
        <Plus className="size-4" />
        {translate('personas.newPersona')}
      </Button>
    </div>
  )
}

function ErrorView({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <AlertCircle className="size-8 text-text-muted" />
      <p className="mt-3 text-sm text-text-2">{translate('common.loadFailed')}</p>
       <p className="mt-1 text-xs text-text-3">{translate('personas.checkNetwork')}</p>
      <Button variant="outline" size="default" className="mt-4" onClick={onRetry}>
        <RefreshCw className="size-4" />
        {translate('common.retry')}
      </Button>
    </div>
  )
}
