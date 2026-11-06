import { useState, useCallback, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import {
  ArrowLeft,
  Menu,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Tag } from '@/components/ui/tag'
import { TONE_TEXT, toneFromKey } from '@/lib/tone'
import { cn } from '@/lib/utils'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { Icon } from '@/components/common/Icon'
import { useSidebarStore } from '@/stores/sidebar-store'

import { personasApi, mcpApi, skillsApi } from '@/api'
import type { PersonaDetail, McpInfo, SkillSimple } from '@/api/types'

type PageState = 'loading' | 'data' | 'error'

function MetaRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-4 py-1">
      <dt className="w-24 shrink-0 text-meta text-text-3">{label}</dt>
      <dd className="min-w-0 flex-1 text-[13px] leading-[1.65] text-text-1">{children}</dd>
    </div>
  )
}

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const t = useT()
  const openMobile = useSidebarStore((s) => s.openMobile)
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
      <PageHeader
        title={data?.name || t('personas.title')}
        className="sticky top-0 z-10"
        actions={
          <>
            <Button
              variant="ghost"
              size="icon"
              onClick={openMobile}
              aria-label={t('sidebar.menu')}
              className="md:hidden"
            >
              <Menu className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={handleBack}
              title={t('common.back')}
              aria-label={t('common.back')}
            >
              <ArrowLeft className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleEdit}
              aria-label={t('common.edit')}
              className="h-8 bg-transparent px-1.5 text-[13px] font-medium text-accent hover:bg-transparent hover:text-accent-hover hover:underline hover:underline-offset-4"
            >
              {t('common.edit')}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setDeleteOpen(true)}
              aria-label={t('common.delete')}
              className="h-8 bg-transparent px-1.5 text-[13px] font-medium text-destructive hover:bg-transparent hover:underline hover:underline-offset-4"
            >
              {t('common.delete')}
            </Button>
          </>
        }
      />

      {/* Content */}
      {state === 'loading' && (
        <PageContainer className="mx-auto max-w-[680px]">
          <PersonaDetailSkeleton />
        </PageContainer>
      )}
      {state === 'error' && (
        <ErrorState description={translate('personas.checkNetwork')} onRetry={fetchDetail} />
      )}
      {state === 'data' && data && (
        <PageContainer className="mx-auto max-w-[680px] space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle>{t('personas.basicInfo')}</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <dl>
                <MetaRow label={t('personas.personaName')}>{data.name}</MetaRow>
                <MetaRow label={t('common.icon')}>
                  <span className="flex items-center gap-2">
                    <Icon name={data.icon} className="size-4 text-text-3" />
                    <span>{data.icon}</span>
                  </span>
                </MetaRow>
                <MetaRow label={t('common.category')}>
                  <span className="flex items-center gap-2">
                    <Icon name={data.categoryIcon} className={cn('size-4', TONE_TEXT[toneFromKey(data.categoryId)])} />
                    <Tag tone={toneFromKey(data.categoryId)}>{data.categoryName}</Tag>
                  </span>
                </MetaRow>
              </dl>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle>{t('personas.personaSettings')}</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <p className="whitespace-pre-wrap text-body leading-[1.7] text-text-2">{data.roleInfo}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle>{t('personas.modelConfig')}</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="flex items-center gap-2.5">
                <ModelIcon icon={data.aiModelId === 0 ? '/favicon.svg' : data.modelIcon} />
                <span className="text-[13px] text-text-1">
                  {data.aiModelId === 0 ? t('chat.autoModel') : data.modelName}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle>
                {t('personas.mcpTools')}
                <span className="ml-1 text-meta font-normal text-text-3">({data.mcpIds?.length ?? 0})</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              {data.mcpIds?.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {data.mcpIds.map(id => {
                    const m = mcps.find(x => x.mcpId === id)
                    if (!m) return null
                    return (
                      <Tag key={id} tone={2} className="gap-1.5 px-2.5 py-1 text-meta">
                        <Icon name={m.icon} className="size-3.5" />
                        {m.displayName || m.name}
                      </Tag>
                    )
                  })}
                </div>
              ) : (
                <p className="text-meta text-text-muted">{t('personas.notConfigured')}</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle>
                {t('personas.skillSelection')}
                <span className="ml-1 text-meta font-normal text-text-3">({data.skillIds?.length ?? 0})</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              {data.skillIds?.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {data.skillIds.map(id => {
                    const s = skills.find(x => x.userSkillId === id)
                    if (!s) return null
                    return (
                      <Tag key={id} tone={4} className="gap-1.5 px-2.5 py-1 text-meta">
                        <Icon name={s.icon} className="size-3.5" />
                        {s.skillName}
                      </Tag>
                    )
                  })}
                </div>
              ) : (
                <p className="text-meta text-text-muted">{t('personas.notConfigured')}</p>
              )}
            </CardContent>
          </Card>
        </PageContainer>
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

function ModelIcon({ icon }: { icon?: string }) {
  const [err, setErr] = useState(false)
  if (!icon || err) return <Icon name="brain" className="size-4 shrink-0 text-text-3" />
  return <img src={icon} alt="" className="size-4 shrink-0 rounded object-cover" onError={() => setErr(true)} />
}

function PersonaDetailSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3, 4].map(i => (
        <div key={i} className="rounded-xl border border-border bg-bg-layer-1 px-5 py-4 shadow-soft">
          <Skeleton className="mb-3 block h-4 w-20 rounded" />
          <Skeleton className="block h-4 w-full max-w-[420px] rounded" />
          <Skeleton className="mt-2 block h-4 w-2/3 rounded" />
        </div>
      ))}
    </div>
  )
}
