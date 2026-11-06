import { useState, useCallback, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import {
  ArrowLeft,
  ChevronDown,
  Menu,
  Search,
  Check,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { TONE_SOFT, TONE_TEXT, toneFromKey } from '@/lib/tone'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { Icon } from '@/components/common/Icon'
import { ModelIcon } from '@/components/common/ModelIcon'
import { IconPickerTrigger } from '@/components/common/IconPicker'
import { BRAND_ICON } from '@/components/brand/Logo'
import { useSidebarStore } from '@/stores/sidebar-store'
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
  icon: BRAND_ICON,
  contextLen: 0,
  isDefault: 0,
  isFlash: 0,
  isVisual: 0,
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
  const openMobile = useSidebarStore((s) => s.openMobile)

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
        <PageHeader title={pageTitle} className="sticky top-0 z-10" />
        <PageContainer className="mx-auto max-w-[680px] space-y-4">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="rounded-xl border border-border bg-bg-layer-1 px-5 py-4 shadow-soft">
              <span className="mb-3 block h-4 w-20 animate-pulse rounded bg-bg-layer-3" />
              <span className="block h-9 w-full animate-pulse rounded-lg bg-bg-layer-3" />
            </div>
          ))}
        </PageContainer>
      </div>
    )
  }

  return (
    <div>
      <PageHeader
        title={pageTitle}
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
              onClick={handleCancel}
              title={t('common.back')}
              aria-label={t('common.back')}
            >
              <ArrowLeft className="size-4" />
            </Button>
            <Button variant="outline" onClick={handleCancel}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={formState === 'submitting'}>
              {formState === 'submitting' ? t('personas.saving') : t('personas.save')}
            </Button>
          </>
        }
      />

      {/* Form body */}
      <PageContainer className="mx-auto max-w-[680px] space-y-4">

        {/* Template selection (new only) */}
        {isNew && (
          <Card>
            <CardHeader className="pb-3">
              <CardTitle>{t('personas.presetTemplates')}</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <Dialog open={templateOpen} onOpenChange={setTemplateOpen}>
                <DialogTrigger asChild>
                  <button className="flex h-9 w-full items-center gap-2 rounded-lg border border-border bg-bg-layer-1 px-3 text-[13px] transition-colors hover:bg-bg-layer-2">
                    {selectedTemplate ? (
                      <>
                        <Icon name={selectedTemplate.icon} className="size-4 text-text-3" />
                        <span className="text-text-1">{selectedTemplate.name}</span>
                        <span className="ml-auto text-micro font-medium text-accent">{t('personas.selected')}</span>
                      </>
                    ) : (
                      <span className="text-text-3">{t('personas.selectTemplate')}</span>
                    )}
                  </button>
                </DialogTrigger>
                <DialogContent className="max-w-lg p-0">
                  <div className="flex items-center justify-between px-5 pt-5 pb-3">
                    <div>
                      <DialogTitle>{t('personas.chooseTemplate')}</DialogTitle>
                      <DialogDescription>{t('personas.chooseTemplateDesc')}</DialogDescription>
                    </div>
                  </div>

                  <div className="px-5 pb-3">
                    <div className="relative">
                      <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
                      <input
                        value={templateSearch}
                        onChange={e => setTemplateSearch(e.target.value)}
                        spellCheck={false}
                        placeholder={t('personas.searchTemplate')}
                        className="h-9 w-full rounded-lg border border-border bg-transparent pl-8 pr-2.5 text-[13px] text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30"
                      />
                    </div>
                  </div>

                  <div className="max-h-[360px] overflow-y-auto px-5 pb-3 space-y-4">
                    {groupedTemplates.map(group => (
                      <div key={group.categoryId}>
                        <div className="mb-1.5 flex items-center gap-1.5 text-micro text-text-3">
                          <Icon name={group.icon} className={cn('size-3.5', TONE_TEXT[toneFromKey(group.categoryId)])} />
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
                                  ? 'bg-accent-soft'
                                  : 'hover:bg-bg-layer-2'
                              )}
                              onClick={() => { handleSelectTemplate(t); setTemplateOpen(false); setTemplateSearch('') }}
                            >
                              <span className={cn('flex size-7 shrink-0 items-center justify-center rounded-md', TONE_SOFT[toneFromKey(t.categoryId)], TONE_TEXT[toneFromKey(t.categoryId)])}>
                                <Icon name={t.icon} className="size-3.5" />
                              </span>
                              <div className="min-w-0 flex-1">
                                <div className="flex items-center gap-1.5">
                                  <span className="text-meta font-medium text-text-1">{t.name}</span>
                                  {selectedTemplateId === t.templateId && (
                                    <Check className="size-3 shrink-0 text-accent" />
                                  )}
                                </div>
                                <span className="text-micro leading-tight text-text-3">{t.description}</span>
                              </div>
                            </button>
                          ))}
                        </div>
                      </div>
                    ))}

                    {groupedTemplates.length === 0 && (
                      <p className="py-4 text-center text-meta text-text-muted">{t('personas.noMatch')}</p>
                    )}
                  </div>

                  {selectedTemplateId !== null && (
                    <div className="border-t border-border px-5 py-3">
                      <button
                        className="w-full rounded-md py-1.5 text-meta text-text-2 transition-colors hover:bg-bg-layer-2"
                        onClick={() => { setSelectedTemplateId(null); setTemplateOpen(false); setTemplateSearch('') }}
                      >
                        {t('personas.noTemplate')}
                      </button>
                    </div>
                  )}
                </DialogContent>
              </Dialog>
            </CardContent>
          </Card>
        )}

        {/* Basic info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle>{t('personas.basicInfo')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 pt-0">

            <div>
              <Label className="mb-1.5">
                {t('personas.personaName')} <span className="text-text-3">*</span>
              </Label>
              <Input
                value={name}
                onChange={e => { setName(e.target.value); setErrors(prev => ({ ...prev, name: '' })) }}
                placeholder={t('personas.namePlaceholder')}
                maxLength={50}
                className={cn(errors.name && 'border-destructive')}
              />
              {errors.name && <p className="mt-1 text-micro text-destructive">{errors.name}</p>}
            </div>

            <div>
              <Label className="mb-1.5">{t('common.icon')}</Label>
              <IconPickerTrigger value={icon} onChange={(name: IconName) => setIcon(name)} />
            </div>

            <div>
              <Label className="mb-1.5">
                {t('common.category')} <span className="text-text-3">*</span>
              </Label>
              <div className="relative">
                <button
                  className={cn(
                    'flex h-9 w-full items-center gap-2 rounded-lg border bg-bg-layer-1 px-3 text-[13px] transition-colors hover:bg-bg-layer-2',
                    errors.categoryId ? 'border-destructive' : 'border-border',
                    !selectedCategory && 'text-text-3'
                  )}
                  onClick={() => setCategoryOpen(!categoryOpen)}
                >
                  {selectedCategory ? (
                    <>
                      <Icon name={selectedCategory.icon} className={cn('size-4', TONE_TEXT[toneFromKey(selectedCategory.categoryId)])} />
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
                    <div className="absolute left-0 top-full z-20 mt-1 w-full rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
                      {categories.map(c => (
                        <button
                          key={c.categoryId}
                          className={cn(
                            'flex w-full items-center gap-2 px-3 py-1.5 text-meta transition-colors hover:bg-bg-layer-2',
                            categoryId === c.categoryId ? 'text-text-1' : 'text-text-2'
                          )}
                          onClick={() => { setCategoryId(c.categoryId); setCategoryOpen(false); setErrors(prev => ({ ...prev, categoryId: '' })) }}
                        >
                          <Icon name={c.icon} className={cn('size-3.5', TONE_TEXT[toneFromKey(c.categoryId)])} />
                          {c.name}
                          {categoryId === c.categoryId && <Check className="ml-auto size-3.5" />}
                        </button>
                      ))}
                    </div>
                  </>
                )}
              </div>
              {errors.categoryId && <p className="mt-1 text-micro text-destructive">{errors.categoryId}</p>}
            </div>
          </CardContent>
        </Card>

        {/* Role info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle>
              {t('personas.personaSettings')} <span className="text-text-3">*</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <textarea
              value={roleInfo}
              onChange={e => { setRoleInfo(e.target.value); setErrors(prev => ({ ...prev, roleInfo: '' })) }}
              spellCheck={false}
              placeholder={t('personas.settingsPlaceholder')}
              rows={6}
              maxLength={500}
              className={cn(
                'w-full resize-y rounded-lg border bg-bg-layer-1 px-3 py-2 text-[13px] leading-[1.65] text-text-1 placeholder:text-text-muted transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30 focus:outline-none',
                errors.roleInfo ? 'border-destructive' : 'border-border'
              )}
            />
            {errors.roleInfo && <p className="mt-1 text-micro text-destructive">{errors.roleInfo}</p>}
            <p className="mt-1 text-right text-micro text-text-muted">{roleInfo.length}/500</p>
          </CardContent>
        </Card>

        {/* Model config */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle>{t('personas.modelConfig')}</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="relative">
              <button
                className="flex h-9 w-full items-center gap-2 rounded-lg border border-border bg-bg-layer-1 px-3 text-[13px] transition-colors hover:bg-bg-layer-2"
                onClick={() => setModelOpen(!modelOpen)}
              >
                <ModelIcon icon={selectedModel?.icon} />
                <span className={selectedModel ? 'text-text-1' : 'text-text-3'}>
                  {selectedModel
                    ? selectedModel.aiModelId === 0
                      ? t('chat.autoModel')
                      : `${selectedModel.providerDisplayName} - ${selectedModel.displayName}`
                    : t('personas.modelPlaceholder')
                  }
                </span>
                <ChevronDown className="ml-auto size-3.5 text-text-3" />
              </button>
              {modelOpen && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setModelOpen(false)} />
                  <div className="absolute left-0 top-full z-20 mt-1 w-full rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
                    {models.map(m => (
                      <button
                        key={m.aiModelId}
                        className={cn(
                          'flex w-full items-center gap-2 px-3 py-1.5 text-meta transition-colors hover:bg-bg-layer-2',
                          aiModelId === m.aiModelId ? 'text-text-1' : 'text-text-2'
                        )}
                        onClick={() => { setAiModelId(m.aiModelId); setModelOpen(false) }}
                      >
                        <ModelIcon icon={m.icon || undefined} />
                        {m.aiModelId === 0 ? t('chat.autoModel') : `${m.providerDisplayName} - ${m.displayName}`}
                        {aiModelId === m.aiModelId && <Check className="ml-auto size-3.5" />}
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          </CardContent>
        </Card>

        {/* MCP tools */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle>
              {t('personas.mcpTools')}
              {mcpIds.length > 0 && <span className="ml-1 text-meta font-normal text-text-3">({mcpIds.length})</span>}
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="relative mb-2">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
              <input
                value={mcpSearch}
                onChange={e => setMcpSearch(e.target.value)}
                spellCheck={false}
                placeholder={t('personas.searchMcp')}
                className="h-9 w-full rounded-lg border border-border bg-transparent pl-8 pr-2.5 text-[13px] text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30"
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
                      'flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 transition-colors hover:bg-bg-layer-2',
                      checked && 'bg-accent-soft'
                    )}
                    onClick={() => toggleMcp(m.mcpId)}
                    onKeyDown={e => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); toggleMcp(m.mcpId) } }}
                  >
                    <span className={cn(
                      'flex size-4 shrink-0 items-center justify-center rounded border transition-colors',
                      checked
                        ? 'border-accent bg-accent text-accent-foreground'
                        : 'border-border-strong'
                    )}>
                      {checked && <Check className="size-3" />}
                    </span>
                    <Icon name={m.icon} className="size-3.5 text-text-3" />
                    <div className="min-w-0 flex-1 flex items-baseline">
                      <span className="shrink-0 text-meta text-text-1">{m.name}</span>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="ml-2 truncate text-micro text-text-3">{m.description}</span>
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
                <p className="py-2 text-center text-meta text-text-muted">{t('personas.noMatch')}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Skills */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle>
              {t('personas.skillSelection')}
              {skillIds.length > 0 && <span className="ml-1 text-meta font-normal text-text-3">({skillIds.length})</span>}
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="relative mb-2">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
              <input
                value={skillSearch}
                onChange={e => setSkillSearch(e.target.value)}
                spellCheck={false}
                placeholder={t('personas.searchSkills')}
                className="h-9 w-full rounded-lg border border-border bg-transparent pl-8 pr-2.5 text-[13px] text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30"
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
                      'flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 transition-colors hover:bg-bg-layer-2',
                      checked && 'bg-accent-soft'
                    )}
                    onClick={() => toggleSkill(s.userSkillId)}
                    onKeyDown={e => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); toggleSkill(s.userSkillId) } }}
                  >
                    <span className={cn(
                      'flex size-4 shrink-0 items-center justify-center rounded border transition-colors',
                      checked
                        ? 'border-accent bg-accent text-accent-foreground'
                        : 'border-border-strong'
                    )}>
                      {checked && <Check className="size-3" />}
                    </span>
                    <Icon name={s.icon} className="size-3.5 text-text-3" />
                    <div className="min-w-0 flex-1 flex items-baseline">
                      <span className="shrink-0 text-meta text-text-1">{s.skillName}</span>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="ml-2 truncate text-micro text-text-3">{s.description}</span>
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
                <p className="py-2 text-center text-meta text-text-muted">{t('personas.noMatch')}</p>
              )}
            </div>
          </CardContent>
        </Card>

      </PageContainer>
    </div>
  )
}
