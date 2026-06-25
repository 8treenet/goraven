import { useState, useCallback, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import {
  ArrowLeft,
  ChevronDown,
  Search,
  Check,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { Icon } from '@/components/common/Icon'
import { IconPickerTrigger } from '@/components/common/IconPicker'
import type { IconName } from '@/components/common/icon-registry'
import { personasApi, mcpApi, providersApi, skillsApi } from '@/api'
import type { PersonaCategory, PersonaTemplateItem, ModelInfo, McpInfo, SkillSimple } from '@/api/types'

/* ============================================
   Types
   ============================================ */

type FormState = 'loading' | 'ready' | 'submitting'

const DEFAULT_MODEL: ModelInfo = {
  aiModelId: 0,
  providerDisplayName: '',
  displayName: '',
  modelName: '',
  icon: '',
  contextLen: 0,
  isDefault: 0,
  isCompress: 0,
  isVisual: 0,
}

/* ============================================
   ModelIcon
   ============================================ */

function ModelIcon({ icon }: { icon?: string }) {
  const [error, setError] = useState(false)

  if (!icon || error) {
    return <Icon name="brain" className="size-4 shrink-0 text-text-2" />
  }

  return (
    <img
      src={icon}
      alt=""
      className="size-4 shrink-0 rounded object-cover"
      onError={() => setError(true)}
    />
  )
}

/* ============================================
   Page
   ============================================ */

export function Component() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isNew = id === undefined
  const isEdit = !isNew
  const t = useT()

  const [formState, setFormState] = useState<FormState>('loading')
  const [name, setName] = useState('')
  const [icon, setIcon] = useState('bot')
  const [categoryId, setCategoryId] = useState<number>(0)
  const [roleInfo, setRoleInfo] = useState('')
  const [aiModelId, setAiModelId] = useState(0)
  const [mcpIds, setMcpIds] = useState<number[]>([])
  const [skillIds, setSkillIds] = useState<number[]>([])
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null)

  const [categoryOpen, setCategoryOpen] = useState(false)
  const [modelOpen, setModelOpen] = useState(false)
  const [templateOpen, setTemplateOpen] = useState(false)

  const [mcpSearch, setMcpSearch] = useState('')
  const [skillSearch, setSkillSearch] = useState('')
  const [templateSearch, setTemplateSearch] = useState('')

  const [errors, setErrors] = useState<Record<string, string>>({})

  const [categories, setCategories] = useState<PersonaCategory[]>([])
  const [templates, setTemplates] = useState<PersonaTemplateItem[]>([])
  const [models, setModels] = useState<ModelInfo[]>([])
  const [mcps, setMcps] = useState<McpInfo[]>([])
  const [skills, setSkills] = useState<SkillSimple[]>([])

  const pageTitle = isEdit ? t('personas.editPersona') : t('personas.newPersonaTitle')

  useEffect(() => {
    const loadRefData = Promise.all([
      personasApi.getPersonaCategories(),
      personasApi.getTemplates(),
      providersApi.getAvailableModels(),
      mcpApi.getMcpEndpoints(),
      skillsApi.getSimpleSkills(),
    ])

    if (isEdit) {
      Promise.all([personasApi.getPersonaDetail(Number(id)), loadRefData])
        .then(([detail, [cats, tpls, mds, mcpList, skillList]]) => {
          setName(detail.name)
          setIcon(detail.icon)
          setCategoryId(detail.categoryId)
          setRoleInfo(detail.roleInfo)
          setAiModelId(detail.aiModelId)
          setMcpIds(detail.mcpIds ?? [])
          setSkillIds(detail.skillIds ?? [])
          setCategories(cats)
          setTemplates(tpls)
          setModels([DEFAULT_MODEL, ...mds])
          setMcps(mcpList)
          setSkills(skillList)
          setFormState('ready')
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('common.loadFailed'))
          navigate('/personas')
        })
    } else {
      loadRefData
        .then(([cats, tpls, mds, mcpList, skillList]) => {
          setCategories(cats)
          setTemplates(tpls)
          setModels([DEFAULT_MODEL, ...mds])
          setMcps(mcpList)
          setSkills(skillList)
          setFormState('ready')
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('common.loadFailed'))
          navigate('/personas')
        })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSelectTemplate = useCallback((t: PersonaTemplateItem) => {
    setSelectedTemplateId(t.templateId)
    setName(t.name)
    setIcon(t.icon)
    setCategoryId(t.categoryId)
    personasApi.getTemplateDetail(t.templateId).then(d => { if (d) setRoleInfo(d.roleInfo) })
    toast.success(translate('personas.templateFilled'))
  }, [])

  const toggleMcp = useCallback((id: number) => {
    setMcpIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }, [])

  const toggleSkill = useCallback((id: number) => {
    setSkillIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }, [])

  const filteredMcps = useMemo(() =>
    mcps.filter(m => m.name.includes(mcpSearch.toLowerCase()) || m.description.includes(mcpSearch)),
    [mcpSearch, mcps])

  const filteredSkills = useMemo(() =>
    skills.filter(s => s.skillName.includes(skillSearch.toLowerCase()) || s.description.includes(skillSearch)),
    [skillSearch, skills])

  const selectedCategory = categories.find(c => c.categoryId === categoryId)
  const selectedModel = models.find(m => m.aiModelId === aiModelId)
  const selectedTemplate = templates.find(t => t.templateId === selectedTemplateId)

  const groupedTemplates = useMemo(() => {
    const filtered = templateSearch
      ? templates.filter(t => t.name.includes(templateSearch) || t.description.includes(templateSearch))
      : templates
    const groupMap = filtered.reduce<Record<number, { categoryId: number; name: string; icon: string; templates: PersonaTemplateItem[] }>>((acc, t) => {
      if (!acc[t.categoryId]) {
        acc[t.categoryId] = { categoryId: t.categoryId, name: t.categoryName, icon: t.categoryIcon, templates: [] }
      }
      acc[t.categoryId].templates.push(t)
      return acc
    }, {})
    return Object.values(groupMap)
  }, [templateSearch, templates])

  const validate = useCallback((): boolean => {
    const errs: Record<string, string> = {}
    if (!name.trim() || name.trim().length < 2 || name.trim().length > 50) {
      errs.name = translate('personas.errNameLen')
    }
    if (categoryId === 0) {
      errs.categoryId = translate('personas.errCategoryRequired')
    }
    if (!roleInfo.trim()) {
      errs.roleInfo = translate('personas.errSettingsRequired')
    } else if (roleInfo.length > 500) {
      errs.roleInfo = translate('personas.errSettingsTooLong')
    }
    setErrors(errs)
    return Object.keys(errs).length === 0
  }, [name, categoryId, roleInfo])

  const handleSave = useCallback(async () => {
    if (!validate()) return
    setFormState('submitting')

    const payload = {
      name: name.trim(),
      icon,
      roleInfo: roleInfo.trim(),
      categoryId,
      mcpIds,
      skillIds,
      aiModelId,
      ...(isNew && selectedTemplateId !== null ? { templateId: selectedTemplateId } : {}),
    }

    try {
      if (isEdit) {
        await personasApi.updatePersona(Number(id), payload)
      } else {
        await personasApi.createPersona(payload)
      }
      toast.success(translate('personas.saved'))
      navigate('/personas')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.saveFailed'))
      setFormState('ready')
    }
  }, [validate, navigate, name, icon, roleInfo, categoryId, mcpIds, skillIds, aiModelId, selectedTemplateId, isNew, isEdit, id])

  const handleCancel = useCallback(() => {
    navigate(-1)
  }, [navigate])

  if (formState === 'loading') {
    return (
      <div>
        <div className="sticky top-0 z-10 flex h-10 items-center border-b border-border-custom bg-bg-base px-4">
          <span className="h-5 w-20 rounded bg-bg-layer-3 animate-pulse" />
        </div>
        <div className="mx-auto max-w-[640px] px-6 py-6 space-y-6">
          {[1, 2, 3, 4].map(i => (
            <div key={i}>
              <span className="mb-2 block h-3 w-16 rounded bg-bg-layer-3 animate-pulse" />
              <span className="block h-8 w-full rounded bg-bg-layer-3 animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div>
      {/* Sticky toolbar */}
      <div className="sticky top-0 z-10 flex h-10 items-center justify-between border-b border-border-custom bg-bg-base px-4">
        <div className="flex items-center gap-2">
          <button
            className="flex items-center justify-center size-6 rounded text-text-3 hover:text-text-1 transition-colors"
            onClick={handleCancel}
          >
            <ArrowLeft className="size-4" />
          </button>
          <h1 className="text-[18px] font-semibold text-text-1">{pageTitle}</h1>
        </div>
        <Button size="default" onClick={handleSave} disabled={formState === 'submitting'} className="text-highlight hover:text-highlight">
          {formState === 'submitting' ? t('personas.saving') : t('personas.save')}
        </Button>
      </div>

      {/* Form body */}
      <div className="mx-auto max-w-[640px] px-6 py-6 space-y-8">

        {/* Template selection (new only) */}
        {isNew && (
          <section>
            <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.presetTemplates')}</h2>
            <Dialog open={templateOpen} onOpenChange={setTemplateOpen}>
              <DialogTrigger asChild>
                <button className="flex h-8 w-full items-center gap-2 rounded-lg border border-input bg-transparent px-2.5 text-sm transition-colors hover:bg-bg-hover">
                  {selectedTemplate ? (
                    <>
                      <Icon name={selectedTemplate.icon} className="size-4 text-text-2" />
                      <span className="text-text-1">{selectedTemplate.name}</span>
                      <span className="ml-auto text-[10px] text-highlight font-medium">{t('personas.selected')}</span>
                    </>
                  ) : (
                    <span className="text-text-3">{t('personas.selectTemplate')}</span>
                  )}
                </button>
              </DialogTrigger>
              <DialogContent className="max-w-lg p-0">
                <div className="flex items-center justify-between px-5 pt-5 pb-3">
                  <div>
                    <DialogTitle className="text-sm">{t('personas.chooseTemplate')}</DialogTitle>
                    <DialogDescription className="text-[11px] mt-0.5">{t('personas.chooseTemplateDesc')}</DialogDescription>
                  </div>
                </div>

                <div className="px-5 pb-3">
                  <div className="relative">
                    <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
                    <input
                      value={templateSearch}
                      onChange={e => setTemplateSearch(e.target.value)}
                      placeholder={t('personas.searchTemplate')}
                      className="h-8 w-full rounded-lg border border-input bg-transparent pl-8 pr-2.5 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-3 focus:ring-ring/50"
                    />
                  </div>
                </div>

                <div className="max-h-[360px] overflow-y-auto px-5 pb-3 space-y-4">
                  {groupedTemplates.map(group => (
                    <div key={group.categoryId}>
                      <div className="mb-1.5 flex items-center gap-1.5 text-[11px] text-text-3">
                        <Icon name={group.icon} className="size-3" />
                        <span>{group.name}</span>
                        <span className="text-text-muted">{group.templates.length}</span>
                      </div>
                      <div className="space-y-0.5">
                        {group.templates.map(t => (
                          <button
                            key={t.templateId}
                            className={cn(
                              'flex w-full items-start gap-2.5 rounded-lg px-2.5 py-2 text-left transition-colors',
                              selectedTemplateId === t.templateId
                                ? 'bg-bg-layer-3'
                                : 'hover:bg-bg-hover'
                            )}
                            onClick={() => { handleSelectTemplate(t); setTemplateOpen(false); setTemplateSearch('') }}
                          >
                            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded bg-bg-layer-3 text-text-2">
                              <Icon name={t.icon} className="size-3.5" />
                            </span>
                            <div className="min-w-0 flex-1">
                              <div className="flex items-center gap-1.5">
                                <span className="text-xs font-medium text-text-1">{t.name}</span>
                                {selectedTemplateId === t.templateId && (
                                  <Check className="size-3 shrink-0 text-highlight" />
                                )}
                              </div>
                              <span className="text-[11px] text-text-3 leading-tight">{t.description}</span>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}

                  {groupedTemplates.length === 0 && (
                    <p className="py-4 text-center text-xs text-text-muted">{t('personas.noMatch')}</p>
                  )}
                </div>

                {selectedTemplateId !== null && (
                  <div className="border-t border-border-custom px-5 py-3">
                    <button
                      className="w-full rounded-md py-1.5 text-xs text-text-2 transition-colors hover:bg-bg-hover"
                      onClick={() => { setSelectedTemplateId(null); setTemplateOpen(false); setTemplateSearch('') }}
                    >
                      {t('personas.noTemplate')}
                    </button>
                  </div>
                )}
              </DialogContent>
            </Dialog>
          </section>
        )}

        {/* Basic info */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.basicInfo')}</h2>
          <div className="space-y-3">

            <div>
              <label className="mb-1 block text-xs text-text-2">
                {t('personas.personaName')} <span className="text-text-3">*</span>
              </label>
              <Input
                value={name}
                onChange={e => { setName(e.target.value); setErrors(prev => ({ ...prev, name: '' })) }}
                placeholder={t('personas.namePlaceholder')}
                maxLength={50}
                className={cn(errors.name && 'border-red-400')}
              />
              {errors.name && <p className="mt-0.5 text-[11px] text-red-400">{errors.name}</p>}
            </div>

            <div>
              <label className="mb-1 block text-xs text-text-2">{t('common.icon')}</label>
              <IconPickerTrigger value={icon} onChange={(name: IconName) => setIcon(name)} />
            </div>

            <div>
              <label className="mb-1 block text-xs text-text-2">
                {t('common.category')} <span className="text-text-3">*</span>
              </label>
              <div className="relative">
                <button
                  className={cn(
                    'flex h-8 w-full items-center gap-2 rounded-lg border px-2.5 text-sm transition-colors hover:bg-bg-hover',
                    errors.categoryId ? 'border-red-400' : 'border-input',
                    !selectedCategory && 'text-text-3'
                  )}
                  onClick={() => setCategoryOpen(!categoryOpen)}
                >
                  {selectedCategory ? (
                    <>
                      <Icon name={selectedCategory.icon} className="size-4 text-text-2" />
                      <span className="text-text-1">{selectedCategory.name}</span>
                    </>
                  ) : (
                    t('personas.categoryPlaceholder')
                  )}
                  <ChevronDown className="ml-auto size-3.5 text-text-3" />
                </button>
                {categoryOpen && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setCategoryOpen(false)} />
                      <div className="absolute left-0 top-full z-20 mt-1 w-full rounded-md border border-border-custom bg-bg-layer-2 py-1 shadow-pop">
                      {categories.map(c => (
                        <button
                          key={c.categoryId}
                          className={cn(
                            'flex w-full items-center gap-2 px-3 py-1.5 text-xs transition-colors hover:bg-bg-hover',
                            categoryId === c.categoryId ? 'text-text-1' : 'text-text-2'
                          )}
                          onClick={() => { setCategoryId(c.categoryId); setCategoryOpen(false); setErrors(prev => ({ ...prev, categoryId: '' })) }}
                        >
                          <Icon name={c.icon} className="size-3.5" />
                          {c.name}
                          {categoryId === c.categoryId && <Check className="ml-auto size-3" />}
                        </button>
                      ))}
                    </div>
                  </>
                )}
              </div>
              {errors.categoryId && <p className="mt-0.5 text-[11px] text-red-400">{errors.categoryId}</p>}
            </div>
          </div>
        </section>

        {/* Role info */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-text-1">
            {t('personas.personaSettings')} <span className="text-text-3">*</span>
          </h2>
          <textarea
            value={roleInfo}
            onChange={e => { setRoleInfo(e.target.value); setErrors(prev => ({ ...prev, roleInfo: '' })) }}
            placeholder={t('personas.settingsPlaceholder')}
             rows={6}
            maxLength={500}
            className={cn(
              'w-full rounded-lg border bg-transparent px-2.5 py-2 text-sm text-text-1 placeholder:text-text-muted transition-colors focus:border-ring focus:ring-3 focus:ring-ring/50 focus:outline-none resize-y',
              errors.roleInfo ? 'border-red-400' : 'border-input'
            )}
          />
          {errors.roleInfo && <p className="mt-0.5 text-[11px] text-red-400">{errors.roleInfo}</p>}
          <p className="mt-1 text-[11px] text-text-muted text-right">{roleInfo.length}/500</p>
        </section>

        {/* Model config */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-text-1">{t('personas.modelConfig')}</h2>
          <div className="relative">
            <button
              className="flex h-8 w-full items-center gap-2 rounded-lg border border-input bg-transparent px-2.5 text-sm transition-colors hover:bg-bg-hover"
              onClick={() => setModelOpen(!modelOpen)}
            >
              <ModelIcon icon={selectedModel?.icon} />
              <span className={selectedModel && selectedModel.aiModelId !== 0 ? 'text-text-1' : 'text-text-3'}>
                {selectedModel
                  ? selectedModel.aiModelId === 0
                    ? t('personas.useDefaultModel')
                    : `${selectedModel.providerDisplayName} - ${selectedModel.displayName}`
                  : t('personas.modelPlaceholder')
                }
              </span>
              <ChevronDown className="ml-auto size-3.5 text-text-3" />
            </button>
            {modelOpen && (
              <>
                <div className="fixed inset-0 z-10" onClick={() => setModelOpen(false)} />
                    <div className="absolute left-0 top-full z-20 mt-1 w-full rounded-md border border-border-custom bg-bg-layer-2 py-1 shadow-pop">
                      {models.map(m => (
                        <button
                          key={m.aiModelId}
                          className={cn(
                            'flex w-full items-center gap-2 px-3 py-1.5 text-xs transition-colors hover:bg-bg-hover',
                            aiModelId === m.aiModelId ? 'text-text-1' : 'text-text-2'
                          )}
                          onClick={() => { setAiModelId(m.aiModelId); setModelOpen(false) }}
                        >
                          <ModelIcon icon={m.icon || undefined} />
                          {m.aiModelId === 0 ? t('personas.useDefaultModel') : `${m.providerDisplayName} - ${m.displayName}`}
                          {aiModelId === m.aiModelId && <Check className="ml-auto size-3" />}
                        </button>
                      ))}
                </div>
              </>
            )}
          </div>
        </section>

        {/* MCP tools */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-text-1">
            {t('personas.mcpTools')}
            {mcpIds.length > 0 && <span className="ml-1 text-xs text-text-3">({mcpIds.length})</span>}
          </h2>
          <div className="mb-2 relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
            <input
              value={mcpSearch}
              onChange={e => setMcpSearch(e.target.value)}
              placeholder={t('personas.searchMcp')}
              className="h-8 w-full rounded-lg border border-input bg-transparent pl-8 pr-2.5 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-3 focus:ring-ring/50"
            />
          </div>
          <div className="space-y-0.5">
            {filteredMcps.map(m => {
              const checked = mcpIds.includes(m.mcpId)
              return (
                <div
                  key={m.mcpId}
                  role="checkbox"
                  aria-checked={checked}
                  tabIndex={0}
                  className={cn(
                    'flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 transition-colors hover:bg-bg-hover',
                    checked && 'bg-bg-layer-3'
                  )}
                  onClick={() => toggleMcp(m.mcpId)}
                  onKeyDown={e => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); toggleMcp(m.mcpId) } }}
                >
                  <span className={cn(
                    'flex size-4 shrink-0 items-center justify-center rounded border transition-colors',
                    checked
                      ? 'border-highlight bg-highlight text-highlight-fg'
                      : 'border-border-strong'
                  )}>
                    {checked && <Check className="size-3" />}
                  </span>
                  <Icon name={m.icon} className="size-3.5 text-text-3" />
                  <div className="min-w-0 flex-1 flex items-baseline">
                    <span className="text-xs text-text-1 shrink-0">{m.name}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="ml-2 text-[11px] text-text-3 truncate">{m.description}</span>
                      </TooltipTrigger>
                      <TooltipContent side="top" align="start" className="max-w-xs">
                        {m.description}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
              )
            })}
            {filteredMcps.length === 0 && (
              <p className="py-2 text-center text-xs text-text-muted">{t('personas.noMatch')}</p>
            )}
          </div>
        </section>

        {/* Skills */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-text-1">
            {t('personas.skillSelection')}
            {skillIds.length > 0 && <span className="ml-1 text-xs text-text-3">({skillIds.length})</span>}
          </h2>
          <div className="mb-2 relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
            <input
              value={skillSearch}
              onChange={e => setSkillSearch(e.target.value)}
              placeholder={t('personas.searchSkills')}
              className="h-8 w-full rounded-lg border border-input bg-transparent pl-8 pr-2.5 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-3 focus:ring-ring/50"
            />
          </div>
          <div className="space-y-0.5">
            {filteredSkills.map(s => {
              const checked = skillIds.includes(s.userSkillId)
              return (
                <div
                  key={s.userSkillId}
                  role="checkbox"
                  aria-checked={checked}
                  tabIndex={0}
                  className={cn(
                    'flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 transition-colors hover:bg-bg-hover',
                    checked && 'bg-bg-layer-3'
                  )}
                  onClick={() => toggleSkill(s.userSkillId)}
                  onKeyDown={e => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); toggleSkill(s.userSkillId) } }}
                >
                  <span className={cn(
                    'flex size-4 shrink-0 items-center justify-center rounded border transition-colors',
                    checked
                      ? 'border-highlight bg-highlight text-highlight-fg'
                      : 'border-border-strong'
                  )}>
                    {checked && <Check className="size-3" />}
                  </span>
                  <Icon name={s.icon} className="size-3.5 text-text-3" />
                  <div className="min-w-0 flex-1 flex items-baseline">
                    <span className="text-xs text-text-1 shrink-0">{s.skillName}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="ml-2 text-[11px] text-text-3 truncate">{s.description}</span>
                      </TooltipTrigger>
                      <TooltipContent side="top" align="start" className="max-w-xs">
                        {s.description}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
              )
            })}
            {filteredSkills.length === 0 && (
              <p className="py-2 text-center text-xs text-text-muted">{t('personas.noMatch')}</p>
            )}
          </div>
        </section>

      </div>
    </div>
  )
}
