import { useState, useCallback, useEffect, useMemo } from 'react'
import { toast } from 'sonner'
import { useT, t as translate } from '@/i18n'
import { Search, X, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { adminModelsApi, adminProvidersApi } from '@/api'
import type { AdminModelItem, ProviderItem } from '@/api'
import {
  EmptyState,
  ErrorState,
  Pagination,
  TableSkeleton,
} from './shared'
import type { DialogMode, DrawerMode, ModelEditFormData, ModelFormData, PageState } from './shared'
import { AddModelDrawer, DeleteModelDialog, EditModelDrawer } from './drawers'
import { ModelMembersDialog } from './ModelMembersDialog'
import { ModelRow } from './table'

/* ============================================
   Main Component
   ============================================ */

export function Component() {
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [models, setModels] = useState<AdminModelItem[]>([])
  const [providers, setProviders] = useState<ProviderItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [search, setSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('all')
  const [page, setPage] = useState(1)

  const [drawerMode, setDrawerMode] = useState<DrawerMode | null>(null)
  const [editTarget, setEditTarget] = useState<AdminModelItem | null>(null)
  const [duplicateTarget, setDuplicateTarget] = useState<AdminModelItem | null>(null)
  const [dialogMode, setDialogMode] = useState<DialogMode>(null)
  const [dialogTarget, setDialogTarget] = useState<AdminModelItem | null>(null)
  const [membersTarget, setMembersTarget] = useState<AdminModelItem | null>(null)

  const loadData = useCallback(() => {
    setState('loading')
    const provPromise = providers.length === 0
      ? adminProvidersApi.getProviders().then((r) => { setProviders(r.list); return r.list })
      : Promise.resolve(providers)

    Promise.all([
      adminModelsApi.getModels({
        page,
        pageSize: 20,
        search: search || undefined,
        providerId: providerFilter === 'all' ? undefined : providerFilter,
      }),
      provPromise,
    ]).then(([res]) => {
      const mapped = res.list.map((m) => ({ ...m, apiKey: m.apiKeyMasked }))
      setModels(mapped)
      setTotalCount(res.totalCount)
      setTotalPages(res.totalPage)
      setState(mapped.length > 0 ? 'data' : 'empty')
    }).catch(() => {
      setState('error')
    })
  }, [page, search, providerFilter, providers.length])

  useEffect(() => {
    loadData()
  }, [loadData])

  const safePage = Math.min(page, totalPages)

  useEffect(() => {
    setPage(1)
  }, [search, providerFilter])

  const handleSearch = useCallback((val: string) => {
    setSearch(val)
  }, [])

  const handleClearFilter = useCallback(() => {
    setSearch('')
    setProviderFilter('all')
  }, [])

  const handleAddModel = useCallback(
    async (data: ModelFormData): Promise<void> => {
      await adminModelsApi.createModel({ ...data, status: 1 } as any)
      setDrawerMode(null)
      setDuplicateTarget(null)
      toast.success(translate('adminModels.added'))
      loadData()
    },
    [loadData],
  )

  const handleEditModel = useCallback(
    async (data: ModelEditFormData): Promise<void> => {
      if (!editTarget) return
      await adminModelsApi.updateModel(editTarget.aiModelId, { ...data, status: 1 })
      setDrawerMode(null)
      setEditTarget(null)
      toast.success(translate('adminModels.updated'))
      loadData()
    },
    [editTarget, loadData],
  )

  const handleDeleteModel = useCallback(() => {
    if (!dialogTarget) return
    adminModelsApi.deleteModel(dialogTarget.aiModelId).then(() => {
      setDialogMode(null)
      setDialogTarget(null)
      toast.success(translate('adminModels.deleted'))
      loadData()
    }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
  }, [dialogTarget, loadData])

  const handleSetDefault = useCallback(
    (modelId: number, currentlyDefault: boolean) => {
      adminModelsApi.setDefaultModel(modelId).then(() => {
        setModels((prev) =>
          prev.map((m) => ({
            ...m,
            isDefault: (m.aiModelId === modelId ? (currentlyDefault ? 0 : 1) : m.isDefault) as number,
          })),
        )
        toast.success(translate(currentlyDefault ? 'adminModels.unsetDefaultToast' : 'adminModels.setDefault'))
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [],
  )

  const handleSetFlash = useCallback(
    (modelId: number) => {
      adminModelsApi.setFlashModel(modelId).then(() => {
        setModels((prev) =>
          prev.map((m) => ({
            ...m,
            isFlash: (m.aiModelId === modelId ? 1 : 0) as number,
          })),
        )
        toast.success(translate('adminModels.setFlash'))
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [],
  )

  const handleSetVisual = useCallback(
    (modelId: number) => {
      adminModelsApi.setVisualModel(modelId).then(() => {
        setModels((prev) =>
          prev.map((m) => ({
            ...m,
            isVisual: (m.aiModelId === modelId ? 1 : 0) as number,
          })),
        )
        toast.success(translate('adminModels.setMultimodal'))
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [],
  )

  const hasFilter = search.trim().length > 0 || providerFilter !== 'all'

  const uniqueProviders = useMemo(
    () => providers.map((p) => [p.providerId, p.providerDisplayNameZh] as [string, string]),
    [providers],
  )

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminModels.title')}</h1>
        <div className="flex items-center gap-2">
          <div className="relative">
            {/* Hidden dummy fields to prevent Chrome autofill from targeting the search input */}
            <input type="text" name="username" style={{ position: 'absolute', left: -9999 }} tabIndex={-1} autoComplete="username" />
            <input type="password" name="password" style={{ position: 'absolute', left: -9999 }} tabIndex={-1} autoComplete="current-password" />
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-3" />
            <input
              type="text"
              name="search"
              autoComplete="off"
              spellCheck={false}
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              placeholder={t('adminModels.searchPlaceholder')}
              className="h-7 w-44 rounded-lg border border-border-strong bg-transparent pl-7 pr-2 text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-ring focus:ring-2 focus:ring-ring/30"
            />
            {search && (
              <button
                onClick={() => handleSearch('')}
                className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded-sm p-0.5 text-text-3 hover:text-text-1"
              >
                <X className="size-3" />
              </button>
            )}
          </div>
          <select
            value={providerFilter}
            onChange={(e) => setProviderFilter(e.target.value)}
            className="h-7 rounded-lg border border-border-strong bg-transparent px-2 text-xs text-text-1 outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="all">{t('adminModels.allProviders')}</option>
            {uniqueProviders.map(([pid, pname]) => (
              <option key={pid} value={pid}>
                {pname}
              </option>
            ))}
          </select>
          <Button
            variant="default"
            size="sm"
            onClick={() => setDrawerMode('add')}
            className="text-highlight hover:text-highlight"
          >
            <Plus className="size-3.5" />
            {t('adminModels.addModel')}
          </Button>
        </div>
      </div>

      {/* Content */}
      {state === 'loading' && <TableSkeleton />}

      {state === 'error' && <ErrorState onRetry={loadData} />}

      {state === 'empty' && (
        <EmptyState
          hasFilter={false}
          onClearFilter={handleClearFilter}
          onAdd={() => setDrawerMode('add')}
        />
      )}

      {state === 'data' && models.length === 0 && (
        <EmptyState
          hasFilter={hasFilter}
          onClearFilter={handleClearFilter}
          onAdd={() => setDrawerMode('add')}
        />
      )}

      {state === 'data' && models.length > 0 && (
        <>
          <div className="flex-1 overflow-auto">
            <table className="w-full">
              <thead>
                <tr className="sticky top-0 z-10 border-b border-border bg-bg-base text-left text-xs text-text-3">
                  <th className="pb-2 pl-4 pr-2 font-normal">{t('adminModels.provider')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.modelDisplayName')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('common.model')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.labels')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.accessLabel')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.contextLength')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.updatedAt')}</th>
                  <th className="pb-2 pr-4 font-normal" />
                </tr>
              </thead>
              <tbody>
                {models.map((model) => (
                  <ModelRow
                    key={model.aiModelId}
                    model={model}
                    onEdit={() => {
                      setEditTarget(model)
                      setDrawerMode('edit')
                    }}
                    onDuplicate={() => {
                      setDuplicateTarget(model)
                      setDrawerMode('add')
                    }}
                    onMembers={() => setMembersTarget(model)}
                    onDelete={() => {
                      setDialogTarget(model)
                      setDialogMode('delete')
                    }}
                    onSetDefault={() => handleSetDefault(model.aiModelId, model.isDefault === 1)}
                    onSetFlash={() => handleSetFlash(model.aiModelId)}
                    onSetVisual={() => handleSetVisual(model.aiModelId)}
                  />
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            page={safePage}
            totalPages={totalPages}
            totalCount={totalCount}
            onPageChange={setPage}
          />
        </>
      )}

      {/* Drawers */}
      <AddModelDrawer
        open={drawerMode === 'add'}
        onClose={() => {
          setDrawerMode(null)
          setDuplicateTarget(null)
        }}
        onSave={handleAddModel}
        providers={providers}
        duplicateModel={duplicateTarget}
      />

      <EditModelDrawer
        open={drawerMode === 'edit'}
        onClose={() => {
          setDrawerMode(null)
          setEditTarget(null)
        }}
        onSave={handleEditModel}
        model={editTarget}
      />

      {/* Dialogs */}
      <DeleteModelDialog
        open={dialogMode === 'delete'}
        onClose={() => {
          setDialogMode(null)
          setDialogTarget(null)
        }}
        onConfirm={handleDeleteModel}
        modelName={dialogTarget?.modelName ?? ''}
      />

      {/* Model Members Dialog */}
      <ModelMembersDialog
        model={membersTarget}
        onClose={() => setMembersTarget(null)}
        onSaved={loadData}
      />
    </div>
  )
}
