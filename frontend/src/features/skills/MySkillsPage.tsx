import { useState, useCallback, useMemo, useEffect, useRef } from 'react'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import {
  Search,
  Sparkles,
  ChevronDown,
  ChevronRight,
  X,
  Loader2,
  Trash2,
  Eye,
  Package,
  Copy,
  Check,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Icon } from '@/components/common/Icon'
import { IconPicker } from '@/components/common/IconPicker'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { getSkillCategories, getUserSkills, getMarketSkills, installSkill, deleteUserSkill, refreshSkills, toggleAlwaysOn, updateUserSkill } from '@/api/skills'
import { getUserSkillDetail, getMarketSkillDetail } from '@/api/skills'
import type { SkillCategory, UserSkill, MarketSkill } from '@/api/types'

/* ============================================
   Types
   ============================================ */

type TabValue = 'installed' | 'market'
type PageState = 'loading' | 'data' | 'empty'

/* ============================================
   Helpers
   ============================================ */

function getSourceLabel(source: string): string {
  switch (source) {
    case 'clawhub':
      return 'ClawHub'
    case 'custom_upload':
      return translate('skills.adminUpload')
    case 'custom':
      return translate('skills.userCreated')
    case 'market':
      return translate('skills.marketplaceSource')
    default:
      return source
  }
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

/* ============================================
   SegmentedControl
   ============================================ */

function SegmentedControl({
  options,
  value,
  onChange,
}: {
  options: { value: string; label: string }[]
  value: string
  onChange: (v: string) => void
}) {
  return (
    <div className="flex rounded-md bg-bg-layer-2 p-0.5 gap-0.5">
      {options.map((opt) => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value)}
          className={cn(
            'rounded-sm px-3 py-1 text-xs transition-colors',
            value === opt.value
              ? 'bg-bg-layer-3 text-text-1'
              : 'text-text-3 hover:text-text-2',
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}

/* ============================================
   FilterDropdown
   ============================================ */

function FilterDropdown<T extends string | number | null>({
  label,
  options,
  value,
  onChange,
}: {
  label: string
  options: { value: T; label: string }[]
  value: T
  onChange: (v: T) => void
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const current = options.find((o) => o.value === value)

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 rounded-md px-2.5 py-1.5 text-xs text-text-3 transition-colors hover:bg-bg-hover hover:text-text-2"
      >
        {current ? current.label : label}
        <ChevronDown className="size-4" />
      </button>
      {open && (
        <div className="absolute top-full left-0 z-30 mt-1 min-w-[120px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop">
          {options.map((opt) => (
            <button
              key={String(opt.value)}
              onClick={() => {
                onChange(opt.value)
                setOpen(false)
              }}
              className={cn(
                'w-full px-3 py-1.5 text-left text-xs transition-colors',
                value === opt.value
                  ? 'bg-bg-layer-3 text-text-1'
                  : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

export function Component() {
  const t = useT()
  const [pageState, setPageState] = useState<PageState>('loading')
  const [activeTab, setActiveTab] = useState<TabValue>('installed')

  const [installedSkills, setInstalledSkills] = useState<UserSkill[]>([])
  const [marketSkills, setMarketSkills] = useState<MarketSkill[]>([])
  const [categories, setCategories] = useState<SkillCategory[]>([])

  const [searchQuery, setSearchQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<number | null>(null)

  const [drawerSkillId, setDrawerSkillId] = useState<number | null>(null)
  const [drawerSource, setDrawerSource] = useState<'installed' | 'market'>( 'installed')
  const [showSkillMd, setShowSkillMd] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<UserSkill | null>(null)
  const [installSuccessTarget, setInstallSuccessTarget] = useState<MarketSkill | null>(null)
  const [installCopied, setInstallCopied] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [drawerContent, setDrawerContent] = useState<string | null>(null)
  const [drawerContentLoading, setDrawerContentLoading] = useState(false)

  const [iconPickerOpen, setIconPickerOpen] = useState(false)
  const [categoryPickerOpen, setCategoryPickerOpen] = useState(false)
  const iconPickerRef = useRef<HTMLDivElement>(null)
  const categoryPickerRef = useRef<HTMLDivElement>(null)

  const searchRef = useRef<HTMLInputElement>(null)

  /* ---- Load data ---- */

  useEffect(() => {
    const pageParams = { pageSize: 1000 }
    Promise.all([getSkillCategories(), getUserSkills(pageParams), getMarketSkills(pageParams)])
      .then(([cats, userResult, marketResult]) => {
        setCategories(cats)
        setInstalledSkills(userResult.list)
        setMarketSkills(marketResult.list)
        setPageState('data')
      })
      .catch(() => {
        setPageState('data')
      })
  }, [])

  /* ---- Filtered lists ---- */

  const filteredInstalled = useMemo(() => {
    return installedSkills.filter((s) => {
      if (searchQuery) {
        const q = searchQuery.toLowerCase()
        if (
          !s.skillName.toLowerCase().includes(q) &&
          !s.skillName.toLowerCase().includes(q)
        )
          return false
      }
      if (categoryFilter !== null && s.categoryId !== categoryFilter)
        return false
      return true
    })
  }, [installedSkills, searchQuery, categoryFilter])

  const filteredMarket = useMemo(() => {
    return marketSkills.filter((s) => {
      if (searchQuery) {
        const q = searchQuery.toLowerCase()
        if (!s.name.toLowerCase().includes(q)) return false
      }
      if (categoryFilter !== null && s.categoryId !== categoryFilter)
        return false
      return true
    })
  }, [marketSkills, searchQuery, categoryFilter])

  /* ---- Tab counts ---- */

  const installedCount = installedSkills.length
  const marketCount = marketSkills.filter((s) => !s.userInstalled).length

  /* ---- Actions ---- */

  const reloadData = useCallback(() => {
    const pageParams = { pageSize: 1000 }
    return Promise.all([getUserSkills(pageParams), getMarketSkills(pageParams)]).then(
      ([userResult, marketResult]) => {
        setInstalledSkills(userResult.list)
        setMarketSkills(marketResult.list)
      },
    )
  }, [])

  const handleInstall = useCallback(
    (skill: MarketSkill) => {
      installSkill(skill.skillId)
        .then(() => {
          setInstallSuccessTarget(skill)
          reloadData()
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('skills.installFail'))
        })
    },
    [reloadData],
  )

  const handleCopyInstallText = useCallback(() => {
    if (!installSuccessTarget) return
    const dirName = installSuccessTarget.name.toLowerCase().replace(/\s+/g, '-')
    const text = translate('skills.installSuccessContent')
      .replace('{name}', installSuccessTarget.name)
      .replace('{dir}', dirName)
    navigator.clipboard.writeText(text)
    setInstallCopied(true)
    setTimeout(() => setInstallCopied(false), 1000)
  }, [installSuccessTarget])

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return
    const name = deleteTarget.skillName
    deleteUserSkill(deleteTarget.userSkillId)
      .then(() => {
        setDeleteTarget(null)
        if (drawerSkillId === deleteTarget.userSkillId) setDrawerSkillId(null)
        toast.success(translate('skills.deletedToast').replace('{name}', name))
        reloadData()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [deleteTarget, drawerSkillId, reloadData])

  const handleRefresh = useCallback(() => {
    setIsRefreshing(true)
    refreshSkills()
      .then((result) => {
        setIsRefreshing(false)
        if (result.added > 0) {
          toast.success(translate('skills.syncedToast').replace('{count}', String(result.added)))
        }
        reloadData()
      })
      .catch(() => {
        setIsRefreshing(false)
      })
  }, [reloadData])

  const handleToggleAlwaysOn = useCallback(
    (skill: UserSkill) => {
      const newVal = skill.alwaysOn === 1 ? 0 : 1
      toggleAlwaysOn(skill.userSkillId, newVal)
        .then(() => {
          reloadData()
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('common.saveFailed'))
        })
    },
    [reloadData],
  )

  const openDrawer = useCallback(
    (id: number, source: 'installed' | 'market') => {
      setDrawerSkillId(id)
      setDrawerSource(source)
      setShowSkillMd(false)
      setIconPickerOpen(false)
      setCategoryPickerOpen(false)
    },
    [],
  )

  const closeDrawer = useCallback(() => {
    setDrawerSkillId(null)
    setShowSkillMd(false)
    setIconPickerOpen(false)
    setCategoryPickerOpen(false)
  }, [])

  /* ---- Drawer data ---- */

  const drawerInstalledSkill = useMemo(() => {
    if (drawerSource !== 'installed' || drawerSkillId === null) return null
    return installedSkills.find((s) => s.userSkillId === drawerSkillId) || null
  }, [drawerSource, drawerSkillId, installedSkills])

  const drawerMarketSkill = useMemo(() => {
    if (drawerSource !== 'market' || drawerSkillId === null) return null
    return marketSkills.find((s) => s.skillId === drawerSkillId) || null
  }, [drawerSource, drawerSkillId, marketSkills])

  const drawerSkill = drawerInstalledSkill || drawerMarketSkill

  /* ---- Drawer edit handlers ---- */

  const handleIconChange = useCallback(
    (name: string) => {
      if (!drawerInstalledSkill) return
      setIconPickerOpen(false)
      updateUserSkill(drawerInstalledSkill.userSkillId, { icon: name })
        .then(() => reloadData())
        .catch((err: Error) => {
          toast.error(err.message || translate('common.saveFailed'))
        })
    },
    [drawerInstalledSkill, reloadData],
  )

  const handleCategoryChange = useCallback(
    (categoryId: number) => {
      if (!drawerInstalledSkill) return
      setCategoryPickerOpen(false)
      updateUserSkill(drawerInstalledSkill.userSkillId, { categoryId })
        .then(() => reloadData())
        .catch((err: Error) => {
          toast.error(err.message || translate('common.saveFailed'))
        })
    },
    [drawerInstalledSkill, reloadData],
  )

  /* ---- Category filter options ---- */

  const categoryOptions = [
    { value: null, label: t('skills.allCategories') },
    ...categories.map((c) => ({
      value: c.categoryId,
      label: c.name,
    })),
  ]

  /* ---- Keyboard shortcut ---- */

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        searchRef.current?.focus()
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [])

  /* ---- Click-outside for pickers ---- */

  useEffect(() => {
    if (!iconPickerOpen) return
    const handler = (e: MouseEvent) => {
      if (iconPickerRef.current && !iconPickerRef.current.contains(e.target as Node)) {
        setIconPickerOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [iconPickerOpen])

  useEffect(() => {
    if (!categoryPickerOpen) return
    const handler = (e: MouseEvent) => {
      if (categoryPickerRef.current && !categoryPickerRef.current.contains(e.target as Node)) {
        setCategoryPickerOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [categoryPickerOpen])

  /* ---- Fetch drawer content ---- */

  useEffect(() => {
    if (drawerSkillId === null) {
      setDrawerContent(null)
      return
    }
    setDrawerContent(null)
    setDrawerContentLoading(true)
    const promise =
      drawerSource === 'installed'
        ? getUserSkillDetail(drawerSkillId)
        : getMarketSkillDetail(drawerSkillId)
    promise
      .then((detail) => {
        setDrawerContent(detail.content)
        setDrawerContentLoading(false)
      })
      .catch(() => {
        setDrawerContent(null)
        setDrawerContentLoading(false)
      })
  }, [drawerSkillId, drawerSource])

  /* ---- Render helpers ---- */

  const renderInstalledRow = (skill: UserSkill, i: number, total: number) => {
    return (
      <div
        key={skill.userSkillId}
        className={cn(
          'group flex items-center gap-3 px-4 transition-colors',
          i % 2 === 0
            ? 'bg-bg-layer-1 hover:bg-bg-hover'
            : 'hover:bg-bg-hover',
        )}
      >
        <div className={cn('shrink-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <Icon name={skill.icon} className="size-4 text-text-3" />
        </div>

        <div className={cn('flex-1 min-w-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <span className="text-sm font-semibold text-text-1 truncate">
            {skill.skillName}
          </span>
          <div className="mt-0.5 flex items-center gap-1.5 min-w-0">
            <span className="shrink-0 text-xs text-text-muted">
              {skill.skillName}
            </span>
            <span className="shrink-0 text-text-muted">·</span>
            <span className="truncate text-xs text-text-3">
              {skill.description}
            </span>
          </div>
        </div>

        <div className={cn('shrink-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          {skill.categoryName && (
            <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-text-1 bg-bg-layer-3">
              {skill.categoryName}
            </span>
          )}
        </div>

        <div className={cn('flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => openDrawer(skill.userSkillId, 'installed')}
              >
                <Eye className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('skills.details')}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDeleteTarget(skill)}
              >
                <Trash2 className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('common.delete')}</TooltipContent>
          </Tooltip>
        </div>
      </div>
    )
  }

  const renderMarketRow = (skill: MarketSkill, i: number, total: number) => {
    return (
      <div
        key={skill.skillId}
        className={cn(
          'group flex items-center gap-3 px-4 transition-colors',
          i % 2 === 0
            ? 'bg-bg-layer-1 hover:bg-bg-hover'
            : 'hover:bg-bg-hover',
        )}
      >
        <div className={cn('shrink-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <Icon name={skill.icon} className="size-4 text-text-3" />
        </div>

        <div className={cn('flex-1 min-w-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <span className="text-sm font-semibold text-text-1 truncate">
            {skill.name}
          </span>
          <div className="mt-0.5 flex items-center gap-1.5 min-w-0">
            <span className="truncate text-xs text-text-3">
              {skill.description}
            </span>
          </div>
        </div>

        <div className={cn('shrink-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          {skill.categoryName && (
            <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-text-1 bg-bg-layer-3">
              {skill.categoryName}
            </span>
          )}
        </div>

        <div className={cn('shrink-0', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <span className="text-xs text-text-muted tabular-nums">
            {skill.installedCount}次
          </span>
        </div>

        <div className={cn('flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100', total <= 5 ? 'py-3.5' : 'py-2.5')}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => openDrawer(skill.skillId, 'market')}
              >
                <Eye className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('skills.details')}</TooltipContent>
          </Tooltip>
          {skill.userInstalled ? (
            <span className="px-1.5 text-xs text-text-muted">
              {t('common.installed')}
            </span>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleInstall(skill)}
                >
                  <Package className="size-3.5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t('skills.install')}</TooltipContent>
            </Tooltip>
          )}
        </div>
      </div>
    )
  }

  const renderDrawer = () => {
    if (!drawerSkill) return null

    const isInstalled = drawerSource === 'installed'
    const installedSkill = drawerSkill as UserSkill

    return (
      <>
        <div
          className="fixed inset-0 z-40 bg-black/60 animate-in fade-in duration-150"
          onClick={closeDrawer}
        />
        <div className="fixed inset-y-0 right-0 z-50 w-[400px] animate-in slide-in-from-right duration-200 border-l border-border bg-bg-layer-1 shadow-pop">
          <div className="flex h-full flex-col">
            {/* Header */}
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <h2 className="text-base font-semibold text-text-1">
                {t('skills.details')}
              </h2>
              <button
                onClick={closeDrawer}
                className="text-text-3 transition-colors hover:text-text-1"
              >
                <X className="size-4" />
              </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-auto p-4 space-y-4">
              {/* Icon + Name */}
              <div className="flex items-start gap-3">
                {isInstalled ? (
                  <div ref={iconPickerRef} className="relative shrink-0 mt-0.5">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <button
                          onClick={() => setIconPickerOpen(!iconPickerOpen)}
                          className={cn(
                            'inline-flex size-8 items-center justify-center rounded-md transition-colors',
                            iconPickerOpen
                              ? 'bg-bg-layer-3 text-text-1'
                              : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                          )}
                        >
                          <Icon name={drawerSkill.icon} className="size-5" />
                        </button>
                      </TooltipTrigger>
                      <TooltipContent>{t('skills.editIcon')}</TooltipContent>
                    </Tooltip>
                    {iconPickerOpen && (
                      <div className="absolute left-0 top-full z-50 mt-1 w-72 rounded-md border border-border bg-bg-layer-1 p-3 shadow-pop">
                        <IconPicker
                          value={drawerSkill.icon}
                          onChange={handleIconChange}
                        />
                      </div>
                    )}
                  </div>
                ) : (
                  <Icon name={drawerSkill.icon} className="size-5 shrink-0 mt-0.5 text-text-2" />
                )}
                <div className="min-w-0">
                  <h3 className="text-base font-semibold text-text-1">
                    {'skillName' in drawerSkill ? drawerSkill.skillName : drawerSkill.name}
                  </h3>
                  <p className="text-xs text-text-muted">
                    {'skillName' in drawerSkill
                      ? drawerSkill.skillName
                      : drawerSkill.name.toLowerCase().replace(/\s+/g, '-')}
                  </p>
                </div>
              </div>

              {/* Description */}
              <p className="text-sm text-text-2 leading-relaxed">
                {drawerSkill.description}
              </p>

              {/* Install error */}
              <div className="border-t border-border" />

              {/* Meta */}
              <div className="space-y-2.5">
                <div className="flex items-center gap-2">
                  <span className="shrink-0 text-xs text-text-muted w-16">
                    {t('skills.source')}
                  </span>
                  <span className="text-xs text-text-2">
                    {getSourceLabel(drawerSkill.source)}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="shrink-0 text-xs text-text-muted w-16">
                    {t('common.category')}
                  </span>
                  {isInstalled ? (
                    <div ref={categoryPickerRef} className="relative">
                      <button
                        onClick={() => setCategoryPickerOpen(!categoryPickerOpen)}
                        className={cn(
                          'flex items-center gap-1 rounded py-0.5 text-xs transition-colors',
                          categoryPickerOpen
                            ? 'bg-bg-layer-3 text-text-1'
                            : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                        )}
                      >
                        {drawerInstalledSkill?.categoryName || t('skills.noCategory')}
                        <ChevronDown className="size-3" />
                      </button>
                      {categoryPickerOpen && (
                        <div className="absolute top-full left-0 z-30 mt-1 min-w-[120px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop">
                          {categories.map(cat => (
                            <button
                              key={cat.categoryId}
                              onClick={() => handleCategoryChange(cat.categoryId)}
                              className={cn(
                                'w-full whitespace-nowrap px-3 py-1.5 text-left text-xs transition-colors',
                                drawerInstalledSkill?.categoryId === cat.categoryId
                                  ? 'bg-bg-layer-3 text-text-1'
                                  : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
                              )}
                            >
                              {cat.name}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                  ) : (
                    <span className="text-xs text-text-2">
                      {drawerSkill.categoryName || t('skills.noCategory')}
                    </span>
                  )}
                </div>
                {isInstalled && (
                  <div className="flex items-center gap-2">
                    <span className="shrink-0 text-xs text-text-muted w-16">
                      {t('skills.installTime')}
                    </span>
                    <span className="text-xs text-text-2 tabular-nums">
                      {formatDate(installedSkill.created)}
                    </span>
                  </div>
                )}
                {isInstalled && (
                  <div className="flex items-center gap-2">
                    <span className="shrink-0 text-xs text-text-muted w-16">
                      {t('skills.alwaysOn')}
                    </span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <button
                          onClick={() => handleToggleAlwaysOn(installedSkill)}
                          className="flex items-center gap-1.5 rounded-md py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover"
                        >
                          <span
                            className={cn(
                              'relative inline-flex h-3.5 w-8 items-center rounded-full transition-colors',
                              installedSkill.alwaysOn === 1 ? 'bg-highlight' : 'bg-bg-hover',
                            )}
                          >
                            <span
                              className={cn(
                                'inline-block h-3 w-3 rounded-full transition-transform',
                                installedSkill.alwaysOn === 1
                                  ? 'bg-highlight-fg translate-x-5'
                                  : 'bg-text-muted',
                              )}
                            />
                          </span>
                          <span>{installedSkill.alwaysOn === 1 ? t('common.enabled') : t('common.disabled')}</span>
                        </button>
                      </TooltipTrigger>
                      <TooltipContent>{t('skills.alwaysOnTip')}</TooltipContent>
                    </Tooltip>
                  </div>
                )}
                {!isInstalled && (drawerSkill as MarketSkill).installedCount !== undefined && (
                  <div className="flex items-center gap-2">
                    <span className="shrink-0 text-xs text-text-muted w-16">
                      {t('skills.installCount')}
                    </span>
                    <span className="text-xs text-text-2 tabular-nums">
                      {(drawerSkill as MarketSkill).installedCount}
                    </span>
                  </div>
                )}
              </div>

              <div className="border-t border-border" />

              {/* SKILL.md collapsible */}
              <div>
                <button
                  onClick={() => setShowSkillMd(!showSkillMd)}
                  className="flex items-center gap-1.5 text-xs text-text-3 transition-colors hover:text-text-2"
                >
                  <ChevronRight
                    className={cn(
                      'size-4 transition-transform',
                      showSkillMd && 'rotate-90',
                    )}
                  />
                  SKILL.md {t('skills.skillMdContent')}
                </button>
                {showSkillMd && (
                  <div className="mt-2 rounded-md border border-border bg-bg-layer-2 p-3">
                    {drawerContentLoading ? (
                      <div className="flex items-center gap-2 py-2">
                        <Loader2 className="size-4 animate-spin text-text-3" />
                        <span className="text-xs text-text-3">{t('common.loading')}</span>
                      </div>
                    ) : drawerContent ? (
                      <pre className="text-xs text-text-3 whitespace-pre-wrap leading-relaxed">
                        {drawerContent}
                      </pre>
                    ) : (
                      <span className="text-xs text-text-3">
                        {'skillName' in drawerSkill ? drawerSkill.skillName : drawerSkill.name}
                      </span>
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Bottom action */}
            {isInstalled && (
              <div className="border-t border-border px-4 py-3">
                <Button
                  variant="destructive"
                  size="default"
                  className="w-full"
                  onClick={() => {
                    setDeleteTarget(installedSkill)
                    closeDrawer()
                  }}
                >
                  <Trash2 className="size-4" />
                  {t('skills.deleteSkill')}
                </Button>
              </div>
            )}
          </div>
        </div>
      </>
    )
  }

  /* ---- Render ---- */

  const currentList = activeTab === 'installed' ? filteredInstalled : filteredMarket

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center gap-2 border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">
          {t('skills.title')}
        </h1>

        <SegmentedControl
          options={[
            { value: 'installed', label: `${t('skills.installed')} (${installedCount})` },
            { value: 'market', label: `${t('skills.marketplace')} (${marketCount})` },
          ]}
          value={activeTab}
          onChange={(v) => {
            setActiveTab(v as TabValue)
            setSearchQuery('')
            setCategoryFilter(null)
          }}
        />

        <div className="flex-1" />

        <div className="flex items-center gap-1.5">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-text-muted" />
            <input
              ref={searchRef}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={t('skills.searchPlaceholder')}
              className="h-7 w-40 rounded-md border border-border bg-transparent pl-7 pr-2 text-xs text-text-1 placeholder:text-text-muted outline-none focus:border-border-strong"
            />
          </div>
          <FilterDropdown
            label={t('skills.allCategories')}
            options={categoryOptions}
            value={categoryFilter}
            onChange={setCategoryFilter}
          />
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="inline-flex size-7 items-center justify-center rounded text-highlight transition-colors hover:bg-bg-layer-2 disabled:opacity-50"
              >
                <Sparkles
                  className={cn('size-3.5', isRefreshing && 'animate-spin')}
                />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('skills.refreshTip')}</TooltipContent>
          </Tooltip>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {/* Loading */}
        {pageState === 'loading' && (
          <div>
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div
                key={i}
                className="flex items-center gap-3 border-b border-border px-4 py-2.5"
              >
                <div className="size-4 shrink-0 animate-pulse rounded-sm bg-bg-layer-3" />
                <div className="flex-1">
                  <div className="h-4 w-32 animate-pulse rounded bg-bg-layer-3" />
                  <div className="mt-1.5 h-3 w-56 animate-pulse rounded bg-bg-layer-3" />
                </div>
                <div className="h-4 w-16 animate-pulse rounded bg-bg-layer-3" />
              </div>
            ))}
          </div>
        )}

        {/* Data */}
        {pageState === 'data' && currentList.length === 0 && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <Icon name="bot" className="size-10 text-text-muted" />
            <div className="text-center">
              <p className="text-sm text-text-2">
                {searchQuery || categoryFilter
                  ? t('skills.noMatch')
                  : activeTab === 'installed'
                    ? t('skills.noInstalled')
                    : t('skills.noMarketplace')}
              </p>
              <p className="mt-1 text-xs text-text-3">
                {activeTab === 'installed'
                  ? t('skills.marketplaceHint')
                  : t('skills.marketplaceEmptyHint')}
              </p>
            </div>
          </div>
        )}

        {pageState === 'data' && currentList.length > 0 && (
          <>
            {activeTab === 'installed'
              ? filteredInstalled.map((s, i) => renderInstalledRow(s, i, filteredInstalled.length))
              : filteredMarket.map((s, i) => renderMarketRow(s, i, filteredMarket.length))}
          </>
        )}
      </div>

      {/* Bottom bar */}
      {currentList.length >= 10 && (
        <div className="flex h-7 shrink-0 items-center border-t border-border px-4">
          <span className="text-xs text-text-muted tabular-nums">
            {t('skills.totalCount').replace('{n}', String(currentList.length))}
          </span>
        </div>
      )}

      {/* Drawer */}
      {drawerSkill && renderDrawer()}

      {/* Delete dialog */}
      <Dialog
        open={deleteTarget !== null}
        onOpenChange={() => setDeleteTarget(null)}
      >
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('skills.confirmDeleteSkill')}</DialogTitle>
            <DialogDescription>
              将删除 &quot;{deleteTarget?.skillName}&quot; ({deleteTarget?.skillName})
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <p className="text-xs text-text-3">
              {t('skills.deleteSkillDesc')}
            </p>
            <div className="flex justify-end gap-2">
              <Button
                variant="ghost"
                size="default"
                onClick={() => setDeleteTarget(null)}
              >
                {t('common.cancel')}
              </Button>
              <Button variant="default" size="default" onClick={handleDelete}>
                {t('common.confirmDelete')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Install success dialog */}
      <Dialog
        open={installSuccessTarget !== null}
        onOpenChange={() => {
          setInstallSuccessTarget(null)
          setInstallCopied(false)
        }}
      >
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t('skills.installSuccessTitle')}</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div className="rounded-md border border-border bg-bg-layer-2 p-4">
              <div className="flex items-start gap-4">
                <Icon
                  name={installSuccessTarget?.icon || 'bot'}
                  className="size-8 shrink-0 mt-0.5 text-text-2"
                />
                <div className="min-w-0">
                  <h3 className="text-base font-semibold text-text-1">
                    {installSuccessTarget?.name}
                  </h3>
                  <p className="mt-1 text-sm text-text-3 leading-relaxed">
                    {installSuccessTarget?.description}
                  </p>
                </div>
              </div>
            </div>

            <p className="text-sm text-text-3 leading-relaxed">
              {t('skills.installSuccessDesc')}
            </p>

            <div className="rounded-md border border-border bg-bg-layer-2 p-4">
              <p className="text-base text-text-2 leading-relaxed">
                {installSuccessTarget
                  ? translate('skills.installSuccessContent')
                      .replace('{name}', installSuccessTarget.name)
                      .replace('{dir}', installSuccessTarget.name.toLowerCase().replace(/\s+/g, '-'))
                  : ''}
              </p>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                variant="ghost"
                size="default"
                onClick={() => {
                  setInstallSuccessTarget(null)
                  setInstallCopied(false)
                }}
              >
                {t('common.close')}
              </Button>
              <Button variant="default" size="default" onClick={handleCopyInstallText}>
                {installCopied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
                {installCopied ? t('common.copied') : t('common.copy')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
