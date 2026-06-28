import { useState, useCallback, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import {
  ArrowLeft,
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

import { personasApi, mcpApi, skillsApi } from '@/api'
import type { PersonaDetail, McpInfo, SkillSimple } from '@/api/types'

type PageState = 'loading' | 'data' | 'error'

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [data, setData] = useState<PersonaDetail | null>(null)
  const [mcps, setMcps] = useState<McpInfo[]>([])
  const [skills, setSkills] = useState<SkillSimple[]>([])
  const [deleteOpen, setDeleteOpen] = useState(false)

  const fetchDetail = useCallback(() => {
    setState('loading')
    Promise.all([
      personasApi.getPersonaDetail(Number(id)),
      mcpApi.getMcpEndpoints(),
      skillsApi.getSimpleSkills(),
    ])
      .then(([detail, mcpList, skillList]) => {
        if (detail) {
          setData(detail)
          setMcps(mcpList)
          setSkills(skillList)
          setState('data')
        } else {
          setState('error')
        }
      })
      .catch(() => setState('error'))
  }, [id])

  useEffect(() => {
    fetchDetail()
  }, [fetchDetail])

  const handleEdit = useCallback(() => {
    navigate(`/personas/${id}/edit`)
  }, [navigate, id])

  const handleDelete = useCallback(async () => {
    try {
      await personasApi.deletePersona(Number(id))
      toast.success(translate('personas.deleted'))
      setDeleteOpen(false)
      navigate('/personas')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.failed'))
    }
  }, [navigate, id])

  const handleBack = useCallback(() => {
    navigate('/personas')
  }, [navigate])

  return (
    <div>
      {/* Sticky toolbar */}
      <div className="sticky top-0 z-10 flex h-10 items-center justify-between border-b border-border-custom bg-bg-base px-4">
        <div className="flex items-center gap-2">
          <button
            className="flex items-center justify-center size-6 rounded text-text-3 hover:text-text-1 transition-colors"
            onClick={handleBack}
          >
            <ArrowLeft className="size-4" />
          </button>
          {data && (
            <div className="flex items-center gap-2">
              <span className="flex h-6 w-6 items-center justify-center rounded-md bg-interactive/10 text-interactive">
                <Icon name={data.icon} className="size-3.5" />
              </span>
              <h1 className="text-[18px] font-semibold text-text-1">{data.name}</h1>
            </div>
          )}
        </div>
        <div className="flex items-center gap-1.5">
          <Button variant="ghost" size="default" onClick={handleEdit} className="text-highlight hover:text-highlight">
            <Pencil className="size-4" />
            {t('common.edit')}
          </Button>
          <Button variant="ghost" size="default" onClick={() => setDeleteOpen(true)} className="text-highlight hover:text-highlight">
            <Trash2 className="size-4" />
            {t('common.delete')}
          </Button>
        </div>
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}
      {state === 'error' && <ErrorView onRetry={fetchDetail} />}
      {state === 'data' && data && (
        <div className="mx-auto max-w-[640px] px-6 py-6 space-y-8">

          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.basicInfo')}</h2>
            <div className="space-y-2">
              <Row label={t('personas.personaName')} value={data.name} />
              <Row label={t('common.icon')}>
                <span className="flex items-center gap-1.5">
                  <Icon name={data.icon} className="size-3.5 text-interactive" />
                  <span className="text-sm text-text-1">{data.icon}</span>
                </span>
              </Row>
              <Row label={t('common.category')}>
                <span className="flex items-center gap-1.5">
                  <Icon name={data.categoryIcon} className="size-3.5 text-interactive" />
                  <span className="text-sm text-text-1">{data.categoryName}</span>
                </span>
              </Row>
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.personaSettings')}</h2>
            <div className="rounded-lg border border-border-custom bg-bg-layer-2 px-3 py-2.5">
              <p className="text-sm text-text-1 whitespace-pre-wrap leading-relaxed">{data.roleInfo}</p>
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.modelConfig')}</h2>
            <span className="text-sm text-text-1">
              {data.aiModelId === 0 ? t('personas.useDefaultModel') : data.modelName}
            </span>
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">
              {t('personas.mcpTools')}
              <span className="ml-1 text-xs text-text-3">({data.mcpIds?.length ?? 0})</span>
            </h2>
            {data.mcpIds?.length > 0 ? (
              <div className="space-y-1">
                {data.mcpIds.map(id => {
                  const m = mcps.find(x => x.mcpId === id)
                  if (!m) return null
                  return (
                    <div key={id} className="flex items-center gap-2.5 rounded-lg px-2.5 py-1.5">
                      <Icon name={m.icon} className="size-3.5 text-text-3" />
                      <span className="text-xs text-text-1">{m.displayName || m.name}</span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <p className="text-xs text-text-muted">{t('personas.notConfigured')}</p>
            )}
          </section>

          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">
              {t('personas.skillSelection')}
              <span className="ml-1 text-xs text-text-3">({data.skillIds?.length ?? 0})</span>
            </h2>
            {data.skillIds?.length > 0 ? (
              <div className="space-y-1">
                {data.skillIds.map(id => {
                  const s = skills.find(x => x.userSkillId === id)
                  if (!s) return null
                  return (
                    <div key={id} className="flex items-center gap-2.5 rounded-lg px-2.5 py-1.5">
                      <Icon name={s.icon} className="size-3.5 text-text-3" />
                      <span className="text-xs text-text-1">{s.skillName}</span>
                    </div>
                  )
                })}
              </div>
            ) : (
              <p className="text-xs text-text-muted">{t('personas.notConfigured')}</p>
            )}
          </section>
        </div>
      )}

      {/* Delete dialog */}
      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('personas.deletePersona')}</DialogTitle>
            <DialogDescription>
              {t('personas.deleteConfirm').replace('{name}', data?.name ?? '')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" size="default" onClick={() => setDeleteOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" size="default" onClick={handleDelete}>
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

function Row({ label, value, children }: { label: string; value?: string; children?: React.ReactNode }) {
  return (
    <div className="flex items-center py-1">
      <span className="w-24 shrink-0 text-xs text-text-3">{label}</span>
      <div className="flex-1 min-w-0">
        {children || <span className="text-sm text-text-1">{value}</span>}
      </div>
    </div>
  )
}

function Skeleton() {
  return (
    <div className="mx-auto max-w-[640px] px-6 py-6 space-y-8">
      {[1, 2, 3, 4].map(i => (
        <div key={i}>
          <span className="mb-3 block h-3 w-16 rounded bg-bg-layer-3 animate-pulse" />
          <span className="block h-8 w-full rounded bg-bg-layer-3 animate-pulse" />
        </div>
      ))}
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
