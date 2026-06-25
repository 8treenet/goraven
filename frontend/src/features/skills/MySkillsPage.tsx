import { useState, useCallback, useMemo, useEffect, useRef } from 'react'
import { useT, t as translate } from '@/i18n'
import { toast } from 'sonner'
import { Search, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { getSkillCategories, getUserSkills, getMarketSkills, installSkill, deleteUserSkill, refreshSkills, toggleAlwaysOn, updateUserSkill } from '@/api/skills'
import { getUserSkillDetail, getMarketSkillDetail } from '@/api/skills'
import { getShareSkills, shareSkillToTeam, installShareSkill, deleteShareSkill, getShareSkillDetail } from '@/api/skills'
import type { SkillCategory, UserSkill, MarketSkill, ShareSkill } from '@/api/types'

import { InstalledSkillRow, MarketSkillRow, ShareSkillRow } from './components/SkillRow'
import { SkillDrawer } from './components/SkillDrawer'
import type { SkillDrawerMetaField } from './components/SkillDrawer'
import { DeleteSkillDialog, InstallSuccessDialog, ShareSkillDialog, CancelShareDialog } from './components/SkillDialogs'

/* ============================================
   Types
   ============================================ */

type TabValue = 'installed' | 'market' | 'team'
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
    case 'share':
      return translate('skills.installedShareSource')
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
              ? 'bg-highlight/10 text-highlight'
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
        className={cn(
          'flex items-center gap-1 rounded-md px-2.5 py-1.5 text-xs transition-colors hover:bg-bg-hover hover:text-text-2',
          value !== null ? 'text-highlight' : 'text-text-3',
        )}
      >
        {current ? current.label : label}
        <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m6 9 6 6 6-6"/></svg>
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
                  ? 'bg-highlight/10 text-highlight'
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
  const [shareSkills, setShareSkills] = useState<ShareSkill[]>([])
  const [categories, setCategories] = useState<SkillCategory[]>([])

  const [searchQuery, setSearchQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<number | null>(null)

  const [drawerSkillId, setDrawerSkillId] = useState<number | null>(null)
  const [drawerSource, setDrawerSource] = useState<'installed' | 'market' | 'team'>('installed')
  const [showSkillMd, setShowSkillMd] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<UserSkill | null>(null)
  const [shareTarget, setShareTarget] = useState<UserSkill | null>(null)
  const [cancelShareTarget, setCancelShareTarget] = useState<ShareSkill | null>(null)
  const [installSuccessTarget, setInstallSuccessTarget] = useState<MarketSkill | null>(null)
  const [installCopied, setInstallCopied] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [drawerContent, setDrawerContent] = useState<string | null>(null)
  const [drawerContentLoading, setDrawerContentLoading] = useState(false)
  const [drawerIsShared, setDrawerIsShared] = useState(false)

  const [iconPickerOpen, setIconPickerOpen] = useState(false)
  const [categoryPickerOpen, setCategoryPickerOpen] = useState(false)
  const iconPickerRef = useRef<HTMLDivElement>(null)
  const categoryPickerRef = useRef<HTMLDivElement>(null)

  const searchRef = useRef<HTMLInputElement>(null)

  /* ---- Load data ---- */

  useEffect(() => {
    const pageParams = { pageSize: 1000 }
    Promise.all([getSkillCategories(), getUserSkills(pageParams), getMarketSkills(pageParams), getShareSkills(pageParams)])
      .then(([cats, userResult, marketResult, shareResult]) => {
        setCategories(cats)
        setInstalledSkills(userResult.list)
        setMarketSkills(marketResult.list)
        setShareSkills(shareResult.list)
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

  const filteredShares = useMemo(() => {
    return shareSkills.filter((s) => {
      if (searchQuery) {
        const q = searchQuery.toLowerCase()
        if (!s.skillName.toLowerCase().includes(q)) return false
      }
      if (categoryFilter !== null && s.categoryId !== categoryFilter)
        return false
      return true
    })
  }, [shareSkills, searchQuery, categoryFilter])

  /* ---- Tab counts ---- */

  const installedCount = installedSkills.length
  const marketCount = marketSkills.filter((s) => !s.userInstalled).length
  const shareCount = shareSkills.length

  /* ---- Cross-reference for installed shares ---- */

  const installedShareNames = useMemo(() => {
    return new Set(
      installedSkills
        .filter((s) => s.source === 'share')
        .map((s) => s.skillName),
    )
  }, [installedSkills])

  /* ---- Actions ---- */

  const reloadData = useCallback(() => {
    const pageParams = { pageSize: 1000 }
    return Promise.all([getUserSkills(pageParams), getMarketSkills(pageParams), getShareSkills(pageParams)]).then(
      ([userResult, marketResult, shareResult]) => {
        setInstalledSkills(userResult.list)
        setMarketSkills(marketResult.list)
        setShareSkills(shareResult.list)
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

  const handleShare = useCallback(
    (note: string) => {
      if (!shareTarget) return
      shareSkillToTeam(shareTarget.userSkillId, note)
        .then(() => {
          setShareTarget(null)
          toast.success(translate('skills.shareSuccess'))
          reloadData()
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('skills.shareFail'))
        })
    },
    [shareTarget, reloadData],
  )

  const handleInstallShare = useCallback(
    (skill: ShareSkill) => {
      installShareSkill(skill.shareId)
        .then(() => {
          toast.success(translate('skills.installComplete').replace('{name}', skill.skillName))
          reloadData()
        })
        .catch((err: Error) => {
          toast.error(err.message || translate('skills.installFail'))
        })
    },
    [reloadData],
  )

  const handleCancelShare = useCallback(() => {
    if (!cancelShareTarget) return
    deleteShareSkill(cancelShareTarget.shareId)
      .then(() => {
        setCancelShareTarget(null)
        if (drawerSkillId === cancelShareTarget.shareId && drawerSource === 'team') {
          setDrawerSkillId(null)
        }
        toast.success(translate('skills.deletedToast').replace('{name}', cancelShareTarget.skillName))
        reloadData()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [cancelShareTarget, drawerSkillId, drawerSource, reloadData])

  const openDrawer = useCallback(
    (id: number, source: 'installed' | 'market' | 'team') => {
      setDrawerSkillId(id)
      setDrawerSource(source)
      setShowSkillMd(false)
      setIconPickerOpen(false)
      setCategoryPickerOpen(false)
      setDrawerIsShared(false)
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

  const drawerShareSkill = useMemo(() => {
    if (drawerSource !== 'team' || drawerSkillId === null) return null
    return shareSkills.find((s) => s.shareId === drawerSkillId) || null
  }, [drawerSource, drawerSkillId, shareSkills])

  const drawerSkill = drawerInstalledSkill || drawerMarketSkill || drawerShareSkill

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
      setDrawerIsShared(false)
      return
    }
    setDrawerContent(null)
    setDrawerContentLoading(true)
    let promise: Promise<{ content: string; isShared?: boolean }>
    if (drawerSource === 'installed') {
      promise = getUserSkillDetail(drawerSkillId)
    } else if (drawerSource === 'team') {
      promise = getShareSkillDetail(drawerSkillId).then(d => ({ content: d.content, isShared: undefined }))
    } else {
      promise = getMarketSkillDetail(drawerSkillId).then(d => ({ content: d.content, isShared: undefined }))
    }
    promise
      .then((detail) => {
        setDrawerContent(detail.content)
        if (drawerSource === 'installed' && detail.isShared !== undefined) {
          setDrawerIsShared(detail.isShared)
        }
        setDrawerContentLoading(false)
      })
      .catch(() => {
        setDrawerContent(null)
        setDrawerContentLoading(false)
      })
  }, [drawerSkillId, drawerSource])

  /* ---- Drawer props (derived) ---- */

  const drawerMetaFields = useMemo<SkillDrawerMetaField[]>(() => {
    if (!drawerSkill) return []

    if (drawerSource === 'installed') {
      const isk = drawerSkill as UserSkill
      const fields: SkillDrawerMetaField[] = [
        { label: t('skills.installTime'), value: formatDate(isk.created) },
      ]
      return fields
    }

    if (drawerSource === 'team') {
      const shk = drawerSkill as ShareSkill
      const fields: SkillDrawerMetaField[] = [
        { label: t('skills.owner'), value: shk.ownerName },
        { label: t('skills.installCount'), value: String(shk.installCount) },
      ]
      if (shk.note) {
        fields.push({ label: t('skills.shareNote'), value: shk.note })
      }
      return fields
    }

    const msk = drawerSkill as MarketSkill
    if (msk.installedCount !== undefined) {
      return [
        { label: t('skills.installCount'), value: String(msk.installedCount) },
      ]
    }
    return []
  }, [drawerSkill, drawerSource, t])

  const isDrawerInstalled = drawerSource === 'installed'
  const isDrawerTeam = drawerSource === 'team'
  const drawerName = drawerSkill
    ? ('skillName' in drawerSkill ? drawerSkill.skillName : ('name' in drawerSkill ? drawerSkill.name : ''))
    : ''
  const drawerDirName = drawerSkill
    ? ('skillName' in drawerSkill
        ? drawerSkill.skillName
        : ('name' in drawerSkill ? drawerSkill.name.toLowerCase().replace(/\s+/g, '-') : ''))
    : ''

  const drawerSourceLabel = drawerSkill
    ? ('source' in drawerSkill ? getSourceLabel(drawerSkill.source) : isDrawerTeam ? t('skills.teamShared') : '')
    : ''
  const drawerCategoryName = drawerSkill
    ? ('categoryName' in drawerSkill ? drawerSkill.categoryName : null)
    : null
  const drawerIcon = drawerSkill ? drawerSkill.icon : ''
  const drawerDescription = drawerSkill ? drawerSkill.description : ''

  /* ---- Show share button for installed custom skills ---- */

  const showShareButton = isDrawerInstalled && drawerInstalledSkill?.source === 'custom'

  /* ---- Render ---- */

  const currentList = activeTab === 'installed' ? filteredInstalled : activeTab === 'market' ? filteredMarket : filteredShares

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
            { value: 'team', label: `${t('skills.team')} (${shareCount})` },
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

        {/* Empty */}
        {pageState === 'data' && currentList.length === 0 && (
          <div className="flex h-full flex-col items-center justify-center gap-3">
            <svg className="size-10 text-text-muted" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 8V4H8"/><rect width="16" height="12" x="4" y="8" rx="2"/><path d="M2 14l2 2h16l2-2"/></svg>
            <div className="text-center">
              <p className="text-sm text-text-2">
                {searchQuery || categoryFilter
                  ? t('skills.noMatch')
                  : activeTab === 'installed'
                    ? t('skills.noInstalled')
                    : activeTab === 'market'
                      ? t('skills.noMarketplace')
                      : t('skills.noTeamShares')}
              </p>
              <p className="mt-1 text-xs text-text-3">
                {activeTab === 'installed'
                  ? t('skills.marketplaceHint')
                  : activeTab === 'market'
                    ? t('skills.marketplaceEmptyHint')
                    : t('skills.teamSharesHint')}
              </p>
            </div>
          </div>
        )}

        {/* Rows */}
        {pageState === 'data' && currentList.length > 0 && (
          <>
            {activeTab === 'installed'
              ? filteredInstalled.map((s, i) => (
                  <InstalledSkillRow
                    key={s.userSkillId}
                    skill={s}
                    index={i}
                    total={filteredInstalled.length}
                    onOpenDrawer={(id) => openDrawer(id, 'installed')}
                    onDelete={setDeleteTarget}
                  />
                ))
              : activeTab === 'market'
                ? filteredMarket.map((s, i) => (
                    <MarketSkillRow
                      key={s.skillId}
                      skill={s}
                      index={i}
                      total={filteredMarket.length}
                      onOpenDrawer={(id) => openDrawer(id, 'market')}
                      onInstall={handleInstall}
                    />
                  ))
                : filteredShares.map((s, i) => (
                    <ShareSkillRow
                      key={s.shareId}
                      skill={s}
                      index={i}
                      total={filteredShares.length}
                      userInstalled={installedShareNames.has(s.skillName)}
                      onOpenDrawer={(id) => openDrawer(id, 'team')}
                      onInstall={handleInstallShare}
                      onDelete={(sk) => setCancelShareTarget(sk)}
                    />
                  ))}
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
      {drawerSkill && (
        <SkillDrawer
          name={drawerName}
          dirName={drawerDirName}
          icon={drawerIcon}
          description={drawerDescription}
          sourceLabel={drawerSourceLabel}
          categoryName={drawerCategoryName}
          metaFields={drawerMetaFields}
          editableIcon={isDrawerInstalled}
          editableCategory={isDrawerInstalled}
          categories={categories}
          onIconChange={isDrawerInstalled ? handleIconChange : undefined}
          onCategoryChange={isDrawerInstalled ? handleCategoryChange : undefined}
          showAlwaysOn={isDrawerInstalled}
          alwaysOn={isDrawerInstalled ? (drawerSkill as UserSkill).alwaysOn : undefined}
          onToggleAlwaysOn={isDrawerInstalled ? () => handleToggleAlwaysOn(drawerSkill as UserSkill) : undefined}
          showDeleteButton={isDrawerInstalled}
          onDelete={
            isDrawerInstalled
              ? () => {
                  setDeleteTarget(drawerSkill as UserSkill)
                  closeDrawer()
                }
              : undefined
          }
          showShareButton={showShareButton}
          isShared={drawerIsShared}
          onShare={showShareButton ? () => setShareTarget(drawerSkill as UserSkill) : undefined}
          skillMdContent={drawerContent}
          skillMdLoading={drawerContentLoading}
          iconPickerOpen={iconPickerOpen}
          setIconPickerOpen={setIconPickerOpen}
          categoryPickerOpen={categoryPickerOpen}
          setCategoryPickerOpen={setCategoryPickerOpen}
          iconPickerRef={iconPickerRef}
          categoryPickerRef={categoryPickerRef}
          showSkillMd={showSkillMd}
          onToggleSkillMd={() => setShowSkillMd(!showSkillMd)}
          onClose={closeDrawer}
        />
      )}

      {/* Delete dialog */}
      <DeleteSkillDialog
        target={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
      />

      {/* Install success dialog */}
      <InstallSuccessDialog
        target={installSuccessTarget}
        copied={installCopied}
        onClose={() => {
          setInstallSuccessTarget(null)
          setInstallCopied(false)
        }}
        onCopy={handleCopyInstallText}
      />

      {/* Share dialog */}
      <ShareSkillDialog
        target={shareTarget}
        onClose={() => setShareTarget(null)}
        onConfirm={handleShare}
      />

      {/* Cancel share dialog */}
      <CancelShareDialog
        target={cancelShareTarget}
        onClose={() => setCancelShareTarget(null)}
        onConfirm={handleCancelShare}
      />
    </div>
  )
}
