import { useState, useCallback, useEffect } from 'react'
import { AlertCircle, RefreshCw, Star, Zap, Image } from 'lucide-react'
import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Icon } from '@/components/common/Icon'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { useT } from '@/i18n'
import { personasApi, providersApi, mcpApi, skillsApi } from '@/api'
import type { PersonaDetail, McpInfo, SkillSimple, ModelInfo } from '@/api/types'

function formatK(n: number) {
  if (!n) return '0K'
  const v = n % 1 === 0 ? n : Number(n.toFixed(1))
  return `${v}K`
}

function formatTokensInK(n: number) {
  if (!n) return '0k'
  const v = n / 1024
  return formatK(v)
}

function ModelIcon({ icon }: { icon?: string }) {
  const [err, setErr] = useState(false)
  if (!icon || err) return <Icon name="brain" className="size-4 shrink-0 text-text-2" />
  return <img src={icon} alt="" className="size-4 shrink-0 rounded object-cover" onError={() => setErr(true)} />
}

export function PersonaInfoDialog({
  personaId,
  contextTokens,
  promptTokensCount,
  completionTokensCount,
  open,
  onOpenChange,
}: {
  personaId: number | null
  contextTokens: number
  promptTokensCount: number
  completionTokensCount: number
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const t = useT()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)
  const [persona, setPersona] = useState<PersonaDetail | null>(null)
  const [modelInfo, setModelInfo] = useState<ModelInfo | null>(null)
  const [modelFailed, setModelFailed] = useState(false)
  const [mcps, setMcps] = useState<McpInfo[]>([])
  const [skills, setSkills] = useState<SkillSimple[]>([])

  const fetchData = useCallback(async () => {
    if (!personaId) return
    setLoading(true)
    setError(false)
    try {
      const detail = await personasApi.getPersonaDetail(personaId)
      if (!detail) { setError(true); return }
      setPersona(detail)

      if (detail.aiModelId > 0) {
        try {
          const m = await providersApi.getModel(detail.aiModelId)
          if (m) {
            setModelInfo(m)
            setModelFailed(false)
          } else {
            setModelInfo(null)
            setModelFailed(true)
          }
        } catch {
          setModelInfo(null)
          setModelFailed(true)
        }
      } else {
        setModelInfo(null)
        setModelFailed(false)
      }

      if (detail.mcpIds?.length > 0) {
        try {
          const mcpList = await mcpApi.getMcpEndpointsByIDs(detail.mcpIds)
          setMcps(mcpList ?? [])
        } catch { setMcps([]) }
      } else {
        setMcps([])
      }

      if (detail.skillIds?.length > 0) {
        try {
          const skillList = await skillsApi.getSimpleSkillsByIDs(detail.skillIds)
          setSkills(skillList ?? [])
        } catch { setSkills([]) }
      } else {
        setSkills([])
      }

      setLoading(false)
    } catch {
      setError(true)
      setLoading(false)
    }
  }, [personaId, t])

  useEffect(() => {
    if (open && personaId) {
      fetchData()
    } else {
      setPersona(null)
      setModelInfo(null)
      setModelFailed(false)
      setMcps([])
      setSkills([])
      setError(false)
      setLoading(false)
    }
  }, [open, personaId, fetchData])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl gap-0 p-0">
        {loading ? (
          <div className="space-y-5 p-6">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i}>
                <div className="mb-2 h-3 w-16 rounded bg-bg-layer-3 animate-pulse" />
                <div className="h-10 w-full rounded bg-bg-layer-3 animate-pulse" />
              </div>
            ))}
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-14">
            <AlertCircle className="size-7 text-text-3" />
            <p className="mt-2 text-sm text-text-3">{t('common.loadFailed')}</p>
            <Button variant="outline" size="default" className="mt-4" onClick={fetchData}>
              <RefreshCw className="size-4" />
              {t('common.retry')}
            </Button>
          </div>
        ) : persona ? (
          <div className="border-b border-border">
            <div className="flex items-center gap-3 border-b border-border px-6 py-4">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-bg-layer-2 text-text-1">
                <Icon name={persona.icon} className="size-4.5" />
              </span>
              <div>
                <h2 className="text-[15px] font-semibold text-text-1 leading-tight">{persona.name}</h2>
                <p className="text-xs text-text-3 mt-0.5">{persona.categoryName}</p>
              </div>
            </div>

            <div className="px-6 py-5 space-y-6 overflow-y-auto max-h-[65vh] [scrollbar-width:thin] [scrollbar-color:var(--color-bg-layer-3)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-bg-layer-3 [&::-webkit-scrollbar-track]:bg-transparent">
              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">{t('personas.basicInfo')}</h3>
                <div className="rounded-lg border border-border bg-bg-layer-2 px-4 py-3 space-y-2">
                  <div className="flex items-center">
                    <span className="w-24 shrink-0 text-xs text-text-3">{t('personas.personaName')}</span>
                    <span className="text-sm text-text-1">{persona.name}</span>
                  </div>

                  <div className="flex items-center">
                    <span className="w-24 shrink-0 text-xs text-text-3">{t('common.category')}</span>
                    <span className="flex items-center gap-1.5">
                      <Icon name={persona.categoryIcon} className="size-3.5 text-text-2" />
                      <span className="text-sm text-text-1">{persona.categoryName}</span>
                    </span>
                  </div>
                </div>
              </section>

              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">{t('personas.personaSettings')}</h3>
                <div className="rounded-lg border border-border bg-bg-layer-2 px-4 py-3">
                  <p className="text-sm text-text-1 whitespace-pre-wrap leading-relaxed">{persona.roleInfo}</p>
                </div>
              </section>

              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">{t('personas.modelConfig')}</h3>
                <div className="rounded-lg border border-border bg-bg-layer-2 px-4 py-3 space-y-2">
                  {modelInfo ? (
                    <>
                      <div className="flex items-center">
                        <span className="w-32 shrink-0 text-xs text-text-3">{t('common.name')}</span>
                        <span className="flex items-center gap-2 text-sm text-text-1">
                          <ModelIcon icon={modelInfo.icon} />
                          {modelInfo.displayName}
                        </span>
                      </div>
                      <div className="flex items-center">
                        <span className="w-32 shrink-0 text-xs text-text-3">{t('personas.modelConfig')}</span>
                        <span className="text-sm text-text-1">{modelInfo.modelName}</span>
                      </div>
                      <div className="flex items-center">
                        <span className="w-32 shrink-0 text-xs text-text-3">{t('personas.contextLength')}</span>
                        <span className="text-sm text-text-1">{formatK(modelInfo.contextLen)}</span>
                      </div>
                      {(!!modelInfo.isDefault || !!modelInfo.isFlash || !!modelInfo.isVisual) && (
                        <div className="flex items-center">
                          <span className="w-32 shrink-0 text-xs text-text-3">{t('adminModels.labels')}</span>
                          <span className="flex items-center gap-1">
                            {!!modelInfo.isDefault && (
                              <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-highlight/15 text-highlight">
                                <Star className="size-2.5" />
                                {t('common.default')}
                              </span>
                            )}
                            {!!modelInfo.isFlash && (
                              <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-interactive text-white">
                                <Zap className="size-2.5" />
                                {t('adminModels.flash')}
                              </span>
                            )}
                            {!!modelInfo.isVisual && (
                              <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-interactive/15 text-interactive">
                                <Image className="size-2.5" />
                                {t('adminModels.multimodal')}
                              </span>
                            )}
                          </span>
                        </div>
                      )}
                    </>
                  ) : persona.aiModelId > 0 && modelFailed ? (
                    <span className="text-sm text-text-3">{t('personas.modelUnavailable')}</span>
                  ) : persona.aiModelId === 0 ? (
                    <span className="text-sm text-text-3">{t('personas.useDefaultModel')}</span>
                  ) : null}
                </div>
              </section>

              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">{t('personas.tokenUsage')}</h3>
                <div className="rounded-lg border border-border bg-bg-layer-2 px-4 py-3 space-y-2">
                  <div className="flex items-center">
                    <span className="w-32 shrink-0 text-xs text-text-3">{t('personas.currentContext')}</span>
                    <span className="text-sm text-highlight tabular-nums">
                      {formatTokensInK(contextTokens)}
                      {modelInfo?.contextLen ? (
                        <span> / {formatK(modelInfo.contextLen)}</span>
                      ) : null}
                    </span>
                  </div>
                  <div className="flex items-center">
                    <span className="w-32 shrink-0 text-xs text-text-3">{t('personas.totalPrompt')}</span>
                    <span className="text-sm text-text-1 tabular-nums">{formatTokensInK(promptTokensCount)}</span>
                  </div>
                  <div className="flex items-center">
                    <span className="w-32 shrink-0 text-xs text-text-3">{t('personas.totalCompletion')}</span>
                    <span className="text-sm text-text-1 tabular-nums">{formatTokensInK(completionTokensCount)}</span>
                  </div>
                </div>
              </section>

              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">
                  {t('personas.mcpTools')}
                  <span className="ml-1 font-normal text-text-muted">({mcps.length})</span>
                </h3>
                {mcps.length > 0 ? (
                  <div className="rounded-lg border border-border bg-bg-layer-2 divide-y divide-border px-1">
                    {mcps.map((m) => (
                      <Tooltip key={m.mcpId}>
                        <TooltipTrigger asChild>
                          <div className="flex items-center gap-2.5 px-3 py-2.5">
                            <Icon name={m.icon} className="size-3.5 text-text-3" />
                            <span className="text-sm text-text-1">{m.displayName || m.name}</span>
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="top" align="start" className="max-w-xs">
                          {m.description}
                        </TooltipContent>
                      </Tooltip>
                    ))}
                  </div>
                ) : (
                  <p className="text-xs text-text-muted">{t('personas.notConfigured')}</p>
                )}
              </section>

              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-text-3">
                  {t('personas.skillSelection')}
                  <span className="ml-1 font-normal text-text-muted">({skills.length})</span>
                </h3>
                {skills.length > 0 ? (
                  <div className="rounded-lg border border-border bg-bg-layer-2 divide-y divide-border px-1">
                    {skills.map((s) => (
                      <Tooltip key={s.userSkillId}>
                        <TooltipTrigger asChild>
                          <div className="flex items-center gap-2.5 px-3 py-2.5">
                            <Icon name={s.icon} className="size-3.5 text-text-3" />
                            <span className="text-sm text-text-1">{s.skillName}</span>
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="top" align="start" className="max-w-xs">
                          {s.description}
                        </TooltipContent>
                      </Tooltip>
                    ))}
                  </div>
                ) : (
                  <p className="text-xs text-text-muted">{t('personas.notConfigured')}</p>
                )}
              </section>
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
