import { useState, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import { Skeleton } from '@/components/ui/skeleton'
import { EmptyState } from '@/components/patterns/EmptyState'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import {
  Plus,
  Ellipsis,
  Pencil,
  Trash2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tag } from '@/components/ui/tag'
import { TONE_SOFT, TONE_TEXT, toneFromKey } from '@/lib/tone'
import { cn } from '@/lib/utils'
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
      <PageHeader
        title={t('personas.title')}
        className="sticky top-0 z-10"
        actions={
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCreate}
            title={t('personas.newPersona')}
            className="h-8 bg-transparent px-1.5 text-[13px] font-medium text-accent hover:bg-transparent hover:text-accent-hover hover:underline hover:underline-offset-4"
          >
            {t('personas.newPersona')}
          </Button>
        }
      />

      <PageContainer className={state === 'data' || state === 'loading' ? 'pt-2' : undefined}>
        {/* Content */}
        {state === 'loading' && <PersonaListSkeleton />}
        {state === 'error' && <ErrorState description={translate('personas.checkNetwork')} onRetry={fetchPersonas} />}
        {state === 'empty' && (
          <EmptyState
            variant="card"
            tone={5}
            icon={<Icon name="bot" className="size-5" />}
            title={translate('personas.noPersonas')}
            description={translate('personas.noPersonasHint')}
            action={
              <Button size="default" className="mt-4" onClick={handleCreate}>
                <Plus className="size-4" />
                {translate('personas.newPersona')}
              </Button>
            }
          />
        )}
        {state === 'data' && (
          <div>
            {personas.map((p) => {
              const tone = toneFromKey(p.categoryId)
              const allTags = [
                ...p.mcpNames.map((n) => ({ label: n, type: 'mcp' as const })),
                ...p.skillNames.map((n) => ({ label: n, type: 'skill' as const })),
              ]
              const visibleTags = allTags.slice(0, 4)
              const overflow = allTags.length - visibleTags.length
              const tags = (
                <>
                  {visibleTags.map((tag) => (
                    <Tag
                      key={tag.type + tag.label}
                      tone={tag.type === 'mcp' ? 2 : 4}
                      className="max-w-40 truncate"
                    >
                      {tag.type === 'mcp' ? '@' : ''}{tag.label}
                    </Tag>
                  ))}
                  {overflow > 0 && (
                    <span className="text-micro text-text-muted">
                      +{overflow}
                    </span>
                  )}
                </>
              )

              return (
                <div
                  key={p.personaId}
                  className="group relative flex min-h-[72px] items-center gap-3 border-b border-border py-2.5 transition-colors hover:bg-bg-layer-2"
                >
                  <button
                    className="flex min-w-0 flex-1 items-center gap-3 text-left"
                    onClick={() => handleRowClick(p.personaId)}
                  >
                    <span className={cn('flex size-9 shrink-0 items-center justify-center rounded-lg', TONE_SOFT[tone], TONE_TEXT[tone])}>
                      <Icon name={p.icon} className="size-4.5" />
                    </span>

                    <span className="min-w-0 flex-1">
                      <span className="flex min-w-0 items-center gap-2">
                        <span className="truncate text-body font-medium leading-5 text-text-1">{p.name}</span>
                        {p.categoryName && (
                          <Tag tone={tone} className="hidden shrink-0 sm:inline-flex">
                            {p.categoryIcon && <Icon name={p.categoryIcon} className="size-3" />}
                            {p.categoryName}
                          </Tag>
                        )}
                      </span>

                      <span className="mt-0.5 line-clamp-2 text-meta leading-[18px] text-text-3">
                        {p.roleInfo.replace(/\n/g, ' ')}
                      </span>

                      {(allTags.length > 0 || p.modelName) && (
                        <span className="mt-1 flex min-w-0 flex-wrap items-center gap-1.5 md:hidden">
                          {p.modelName && (
                            <span className="max-w-48 truncate text-meta text-text-muted">
                              {p.modelName}
                            </span>
                          )}
                          {p.modelName && allTags.length > 0 && (
                            <span className="text-micro text-text-muted">·</span>
                          )}
                          {tags}
                        </span>
                      )}
                    </span>
                  </button>

                  {(allTags.length > 0 || p.modelName) && (
                    <div className="hidden max-w-[46%] shrink-0 flex-col items-end gap-1 md:flex">
                      {p.modelName && (
                        <span className="max-w-56 truncate text-meta text-text-muted">{p.modelName}</span>
                      )}
                      {allTags.length > 0 && (
                        <div className="flex flex-wrap justify-end gap-1">{tags}</div>
                      )}
                    </div>
                  )}

                  <div className="relative shrink-0 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
                    <button
                      className="flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-3 hover:text-text-1"
                      aria-haspopup="menu"
                      aria-expanded={menuOpenId === p.personaId}
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
                        <div className="absolute right-0 top-full z-20 mt-1 w-36 rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
                          <button
                            className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
                            onClick={() => { setMenuOpenId(null); handleEdit(p.personaId) }}
                          >
                            <Pencil className="size-3.5" />
                            {t('common.edit')}
                          </button>
                          <button
                            className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-destructive transition-colors hover:bg-destructive/10"
                            onClick={() => { setMenuOpenId(null); setDeleteTarget(p) }}
                          >
                            <Trash2 className="size-3.5" />
                            {t('common.delete')}
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </PageContainer>

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

function PersonaListSkeleton() {
  return (
    <div>
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="flex min-h-[72px] items-center gap-3 border-b border-border py-3">
          <Skeleton className="size-9 shrink-0 rounded-lg" />
          <div className="flex-1 space-y-2">
            <div className="flex items-center gap-2">
              <Skeleton className="h-4 w-28 rounded" />
              <Skeleton className="h-4 w-12 rounded" />
            </div>
            <Skeleton className="block h-3 w-full max-w-[480px] rounded" />
            <div className="flex gap-1.5">
              <Skeleton className="h-4 w-20 rounded-full" />
              <Skeleton className="h-4 w-16 rounded-full" />
              <Skeleton className="h-4 w-24 rounded-full" />
            </div>
          </div>
          <Skeleton className="h-7 w-7 shrink-0 rounded-md" />
        </div>
      ))}
    </div>
  )
}
