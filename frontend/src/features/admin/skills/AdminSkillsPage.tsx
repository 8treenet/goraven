import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import {
  AlertCircle,
  Archive,
  Check,
  Download,
  FileArchive,
  Loader2,
  Plus,
  Search,
  Settings2,
  Trash2,
  Upload,
  X,
} from 'lucide-react'
import { Icon, renderIcon } from '@/components/common/Icon'
import { IconPickerTrigger } from '@/components/common/IconPicker'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import { adminSkillsApi } from '@/api'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { useT, t as translate } from '@/i18n'

/* ============================================
   Types
   ============================================ */

type TabKey = 'global' | 'market'
type SourceType = 'clawhub' | 'custom_upload'
type InstallStatus = 0 | 1 | 2 | 3
type ClawHubSort = 'newest' | 'updated' | 'downloads' | 'stars' | 'installs' | 'trending'
type DrawerKey = 'system-editor' | 'market-editor' | 'clawhub' | 'publish' | 'categories' | 'users' | null

interface SystemSkill {
  skillId: number
  name: string
  description: string
  content: string
  status: 0 | 1
  updated: string
}

interface SkillCategory {
  categoryId: number
  name: string
  icon: string
  isDefault: 0 | 1
  skillCount: number
  updated: string
}

interface MarketSkill {
  skillId: number
  name: string
  description: string
  icon: string
  source: SourceType
  sourceUrl: string
  categoryId: number
  categoryName: string
  categoryIcon: string
  installedCount: number
  status: 0 | 1
  sortOrder: number
  remark: string
  content: string
  updated: string
}

interface InstalledSkill {
  recordId: number
  userId: string
  skillId: number
  skillName: string
  categoryName: string
  categoryIcon: string
  source: SourceType
  installStatus: InstallStatus
  reason: string
  created: string
}

interface ClawHubItem {
  slug: string
  displayName: string
  summary: string
  version: string
  score: number
  downloads: number
  installs: number
  stars: number
  updatedAt: string
  content: string
}

interface DeleteTarget {
  type: 'system' | 'market' | 'category'
  id: number
  label: string
}

interface SystemForm {
  skillId?: number
  content: string
}

interface MarketForm {
  skillId: number
  icon: string
  categoryId: number
  sortOrder: number
  remark: string
}

/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string): string {
  const date = new Date(iso)
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

function getSourceLabel(source: SourceType): string {
  return source === 'clawhub' ? 'ClawHub' : translate('adminSkills.customUpload')
}

function getInstallStatusLabel(status: InstallStatus): string {
  if (status === 0) return translate('common.notInstalled')
  if (status === 1) return translate('common.installing')
  if (status === 2) return translate('common.installed')
  return translate('common.failed')
}

function parseSkillContent(content: string): { name: string; description: string; errors: string[] } {
  const errors: string[] = []
  const match = content.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/)

  if (!match) {
    return { name: '', description: '', errors: [translate('adminSkills.frontmatterError')] }
  }

  const frontmatter = match[1]
  const name = frontmatter.match(/^name:[ \t]+(.+)$/m)?.[1]?.trim() ?? ''
  const description = frontmatter.match(/^description:[ \t]+(.+)$/m)?.[1]?.trim() ?? ''

  if (!name) errors.push(translate('adminSkills.errMissingName'))
  if (!description) errors.push(translate('adminSkills.errMissingDescription'))
  if (name && !name.startsWith('goraven-')) errors.push(translate('adminSkills.errNamePrefix'))
  if (name && !/^goraven-[a-z0-9][a-z0-9-]*$/.test(name)) errors.push(translate('adminSkills.errNameFormat'))

  return { name, description, errors }
}

const GLOBAL_TEMPLATE = `---
name: goraven-
description: fill skill description here
---

`

/* ============================================
   Small components
   ============================================ */

function Drawer({
  open,
  onClose,
  title,
  description,
  width = 'w-[680px]',
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  description?: string
  width?: string
  children: React.ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <div className={cn('fixed inset-0 z-50', open ? 'visible' : 'invisible')}>
        <div
          className={cn('absolute inset-0 bg-black/60 transition-opacity duration-200', open ? 'opacity-100' : 'opacity-0')}
          onClick={onClose}
        />
        <div
          className={cn(
            'absolute right-0 top-0 z-50 flex h-full flex-col border-l border-border bg-bg-layer-1 shadow-pop transition-transform duration-200 ease-out',
            width,
            open ? 'translate-x-0' : 'translate-x-full',
          )}
        >
          <div className="flex min-h-12 shrink-0 items-center justify-between border-b border-border px-4">
            <div>
              <h2 className="text-sm font-semibold text-text-1">{title}</h2>
              {description && <p className="mt-0.5 text-xs text-text-3">{description}</p>}
            </div>
            <button
              onClick={onClose}
              className="rounded-sm p-0.5 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
            >
              <X className="size-4" />
            </button>
          </div>
          <div className="flex-1 overflow-hidden">{children}</div>
        </div>
      </div>
    </Dialog>
  )
}

function SelectField({ value, onChange, children, className }: {
  value: string
  onChange: (value: string) => void
  children: React.ReactNode
  className?: string
}) {
  return (
    <select
      value={value}
      onChange={(event) => onChange(event.target.value)}
      className={cn(
        'h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm text-text-1 outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30',
        className,
      )}
    >
      {children}
    </select>
  )
}

function SearchBox({ value, onChange, placeholder }: { value: string; onChange: (value: string) => void; placeholder: string }) {
  return (
    <div className="relative w-[260px]">
      <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-text-3" />
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="pl-8 pr-8 text-sm"
      />
      {value && (
        <button
          onClick={() => onChange('')}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-text-3 hover:bg-bg-hover hover:text-text-1"
        >
          <X className="size-3.5" />
        </button>
      )}
    </div>
  )
}

function SkillIcon({ icon }: { icon: string }) {
  return (
    <div className="inline-flex size-8 shrink-0 items-center justify-center rounded bg-bg-layer-3 text-interactive">
      {renderIcon(icon || 'bot', 'size-4')}
    </div>
  )
}

function ScopeLabel({ children }: { children: React.ReactNode }) {
  return <span className="rounded bg-interactive/10 px-1.5 py-0.5 text-xs text-interactive">{children}</span>
}

function StatusText({ children, tone = 'neutral' }: { children: React.ReactNode; tone?: 'neutral' | 'good' | 'warn' }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded px-1.5 py-0.5 text-xs',
        tone === 'neutral' && 'bg-bg-layer-3 text-text-2',
        tone === 'good' && 'bg-bg-layer-3 text-text-1',
        tone === 'warn' && 'bg-bg-hover text-text-1',
      )}
    >
      {children}
    </span>
  )
}

function EmptyState({ title, description, action }: { title: string; description: string; action?: React.ReactNode }) {
  return (
    <div className="flex min-h-[280px] flex-col items-center justify-center rounded-lg border border-border bg-bg-layer-1 px-6 text-center">
      <Archive className="mb-3 size-8 text-text-muted" />
      <h3 className="text-sm font-semibold text-text-1">{title}</h3>
      <p className="mt-1 max-w-sm text-sm text-text-3">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}

/* ============================================
   Main page
   ============================================ */

export function Component() {
  const t = useT()
  const [activeTab, setActiveTab] = useState<TabKey>('global')
  const [systemSkills, setSystemSkills] = useState<SystemSkill[]>([])
  const [marketSkills, setMarketSkills] = useState<MarketSkill[]>([])
  const [installedSkills, setInstalledSkills] = useState<InstalledSkill[]>([])
  const [categories, setCategories] = useState<SkillCategory[]>([])
  const [clawHubItems, setClawHubItems] = useState<ClawHubItem[]>([])

  const [systemSearch, setSystemSearch] = useState('')
  const [systemStatus, setSystemStatus] = useState('all')
  const [marketSearch, setMarketSearch] = useState('')
  const [marketSource, setMarketSource] = useState('all')
  const [marketStatus, setMarketStatus] = useState('all')

  const [drawer, setDrawer] = useState<DrawerKey>(null)
  const [systemForm, setSystemForm] = useState<SystemForm>({ content: GLOBAL_TEMPLATE })
  const [marketForm, setMarketForm] = useState<MarketForm | null>(null)
  const [selectedMarketSkillId, setSelectedMarketSkillId] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null)
  const [cascadeDelete, setCascadeDelete] = useState(false)
  const [saving, setSaving] = useState(false)

  const [clawHubTokenConfigured, setClawHubTokenConfigured] = useState(true)
  const [clawHubMode, setClawHubMode] = useState<'search' | 'explore'>('explore')
  const [clawHubQuery, setClawHubQuery] = useState('')
  const [clawHubSubmittedQuery, setClawHubSubmittedQuery] = useState('')
  const [clawHubSort, setClawHubSort] = useState<ClawHubSort>('trending')
  const [clawHubSearching, setClawHubSearching] = useState(false)
  const [clawHubExploring, setClawHubExploring] = useState(false)
  const [clawHubSelected, setClawHubSelected] = useState<ClawHubItem | null>(null)
  const [clawHubCategoryId, setClawHubCategoryId] = useState(1)
  const [clawHubIcon, setClawHubIcon] = useState('zap')
  const [importing, setImporting] = useState(false)

  const [publishFile, setPublishFile] = useState('')
  const [publishUploadId, setPublishUploadId] = useState('')
  const [publishCategoryId, setPublishCategoryId] = useState(1)
  const [publishIcon, setPublishIcon] = useState('package')
  const [publishing, setPublishing] = useState(false)

  const [categoryDraft, setCategoryDraft] = useState<{ categoryId?: number; name: string; icon: string }>({ name: '', icon: 'folder' })

  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isDraggingZip, setIsDraggingZip] = useState(false)
  const { upload: uploadZip, progress: uploadProgress, isUploading: isUploadingZip } = useChunkUpload()
  const publishProgress = publishUploadId ? 100 : uploadProgress?.percentage ?? 0

  const systemParsed = useMemo(() => parseSkillContent(systemForm.content), [systemForm.content])

  const filteredSystemSkills = useMemo(() => systemSkills, [systemSkills])

  const filteredMarketSkills = useMemo(() => marketSkills, [marketSkills])

  const selectedMarketSkill = useMemo(
    () => marketSkills.find((skill) => skill.skillId === selectedMarketSkillId) ?? null,
    [marketSkills, selectedMarketSkillId],
  )

  const selectedSkillUsers = useMemo(
    () => installedSkills.filter((record) => record.skillId === selectedMarketSkillId),
    [installedSkills, selectedMarketSkillId],
  )

  const clawHubResults = useMemo(() => {
    if (clawHubMode === 'search') {
      const query = clawHubSubmittedQuery.trim().toLowerCase()
      if (!query || clawHubSearching) return []
      return clawHubItems.filter((item) => item.displayName.toLowerCase().includes(query) || item.summary.toLowerCase().includes(query))
    }

    if (clawHubExploring) return []

    return [...clawHubItems].sort((a, b) => {
      if (clawHubSort === 'newest') return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      if (clawHubSort === 'updated') return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      if (clawHubSort === 'downloads') return b.downloads - a.downloads
      if (clawHubSort === 'stars') return b.stars - a.stars
      if (clawHubSort === 'installs') return b.installs - a.installs
      return b.score - a.score
    })
  }, [clawHubExploring, clawHubItems, clawHubMode, clawHubSearching, clawHubSort, clawHubSubmittedQuery])

  useEffect(() => {
    if (drawer !== 'users' || !selectedMarketSkillId) return

    const loadUsers = async () => {
      try {
        const result = await adminSkillsApi.getMarketSkillUsers(selectedMarketSkillId, { pageSize: 100 })
        setInstalledSkills(result.list.map((record, index) => ({
          recordId: index + 1,
          userId: record.userId,
          skillId: selectedMarketSkillId,
          skillName: selectedMarketSkill?.name ?? '',
          categoryName: selectedMarketSkill?.categoryName ?? '',
          categoryIcon: selectedMarketSkill?.categoryIcon ?? '',
          source: selectedMarketSkill?.source ?? 'custom_upload',
          installStatus: record.installStatus,
          reason: '',
          created: record.created,
        })))
      } catch (error) {
        toast.error(error instanceof Error ? error.message : translate('common.failed'))
      }
    }

    loadUsers()
  }, [drawer, selectedMarketSkill, selectedMarketSkillId])

  const loadSystemSkills = useCallback(async () => {
    const result = await adminSkillsApi.getSystemSkills({
      pageSize: 100,
      search: systemSearch.trim() || undefined,
      status: systemStatus === 'all' ? undefined : Number(systemStatus) as 0 | 1,
    })
    setSystemSkills(result.list.map((skill) => ({ ...skill, content: '' })))
  }, [systemSearch, systemStatus])

  const loadMarketSkills = useCallback(async () => {
    const result = await adminSkillsApi.getMarketSkills({
      pageSize: 100,
      search: marketSearch.trim() || undefined,
      source: marketSource === 'all' ? undefined : marketSource as SourceType,
      status: marketStatus === 'all' ? undefined : Number(marketStatus) as 0 | 1,
    })
    setMarketSkills(result.list.map((skill) => ({ ...skill, sourceUrl: '', remark: '', content: '' })))
  }, [marketSearch, marketSource, marketStatus])

  const loadCategories = useCallback(async () => {
    const result = await adminSkillsApi.getSkillCategories({ pageSize: 100 })
    setCategories(result.list)
  }, [])

  const loadInitialData = useCallback(async () => {
    try {
      await Promise.all([loadSystemSkills(), loadMarketSkills(), loadCategories()])
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [loadCategories, loadMarketSkills, loadSystemSkills])

  useEffect(() => {
    loadInitialData()
  }, [loadInitialData])

  useEffect(() => {
    if (categories.length === 0) return
    const firstCategoryId = categories[0].categoryId
    if (!categories.some((item) => item.categoryId === clawHubCategoryId)) setClawHubCategoryId(firstCategoryId)
    if (!categories.some((item) => item.categoryId === publishCategoryId)) setPublishCategoryId(firstCategoryId)
  }, [categories, clawHubCategoryId, publishCategoryId])

  const loadClawHubExplore = useCallback(async (sort: ClawHubSort) => {
    setClawHubExploring(true)
    try {
      const result = await adminSkillsApi.exploreClawHub({ sort })
      setClawHubItems(result.items.map((item) => ({
        slug: item.slug,
        displayName: item.displayName,
        summary: item.summary,
        version: '',
        score: 0,
        downloads: item.stats.downloads,
        installs: item.stats.installsAllTime,
        stars: item.stats.stars,
        updatedAt: new Date(item.updatedAt).toISOString(),
        content: '',
      })))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setClawHubExploring(false)
    }
  }, [])

  const submitClawHubSearch = useCallback(async () => {
    const query = clawHubQuery.trim()
    if (!query) {
      toast.error(translate('adminSkills.emptySearchQuery'))
      return
    }

    setClawHubSelected(null)
    setClawHubSearching(true)
    setClawHubSubmittedQuery(query)
    try {
      const result = await adminSkillsApi.searchClawHub({ q: query, limit: 25 })
      setClawHubItems(result.results.map((item) => ({
        slug: item.slug,
        displayName: item.displayName,
        summary: item.summary,
        version: item.version,
        score: item.score,
        downloads: 0,
        installs: 0,
        stars: 0,
        updatedAt: new Date(item.updatedAt).toISOString(),
        content: '',
      })))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setClawHubSearching(false)
    }
  }, [clawHubQuery])

  const changeClawHubSort = useCallback((sort: ClawHubSort) => {
    setClawHubSort(sort)
    setClawHubSelected(null)
    loadClawHubExplore(sort)
  }, [loadClawHubExplore])

  const changeClawHubMode = useCallback((mode: 'search' | 'explore') => {
    setClawHubMode(mode)
    setClawHubSelected(null)
    setClawHubSubmittedQuery('')
    setClawHubQuery('')
    setClawHubSearching(false)
    if (mode === 'explore') {
      loadClawHubExplore(clawHubSort)
    } else {
      setClawHubExploring(false)
      setClawHubItems([])
    }
  }, [clawHubSort, loadClawHubExplore])

  const openClawHubDrawer = useCallback(() => {
    setDrawer('clawhub')
    changeClawHubMode('explore')
  }, [changeClawHubMode])

  const openSystemEditor = useCallback(async (skill?: SystemSkill) => {
    if (!skill) {
      setSystemForm({ content: GLOBAL_TEMPLATE })
      setDrawer('system-editor')
      return
    }

    try {
      const detail = await adminSkillsApi.getSystemSkillDetail(skill.skillId)
      setSystemForm({ skillId: detail.skillId, content: detail.content })
      setDrawer('system-editor')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const saveSystemSkill = useCallback(async () => {
    const parsed = parseSkillContent(systemForm.content)
    if (parsed.errors.length > 0) {
      toast.error(parsed.errors[0])
      return
    }

    setSaving(true)
    try {
      if (systemForm.skillId) {
        await adminSkillsApi.updateSystemSkill(systemForm.skillId, { content: systemForm.content })
      } else {
        await adminSkillsApi.createSystemSkill({ content: systemForm.content })
      }
      await loadSystemSkills()
      setDrawer(null)
      toast.success(t('adminSkills.savedEffective'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setSaving(false)
    }
  }, [loadSystemSkills, systemForm, t])

  const openMarketEditor = useCallback(async (skill: MarketSkill) => {
    setSelectedMarketSkillId(skill.skillId)
    try {
      const detail = await adminSkillsApi.getMarketSkillDetail(skill.skillId)
      setMarketSkills((items) => items.map((item) => item.skillId === detail.skillId ? { ...item, ...detail } : item))
      setMarketForm({
        skillId: detail.skillId,
        icon: detail.icon,
        categoryId: detail.categoryId,
        sortOrder: detail.sortOrder,
        remark: detail.remark,
      })
      setDrawer('market-editor')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const saveMarketSkill = useCallback(async () => {
    if (!marketForm) return
    setSaving(true)
    try {
      await adminSkillsApi.updateMarketSkill(marketForm.skillId, {
        icon: marketForm.icon,
        categoryId: marketForm.categoryId,
        sortOrder: marketForm.sortOrder,
        remark: marketForm.remark,
      })
      await Promise.all([loadMarketSkills(), loadCategories()])
      setDrawer(null)
      toast.success(translate('adminSkills.marketUpdated'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setSaving(false)
    }
  }, [loadCategories, loadMarketSkills, marketForm])

  const toggleSystemStatus = useCallback(async (skillId: number, checked: boolean) => {
    try {
      await adminSkillsApi.updateSystemSkillStatus(skillId, checked ? 0 : 1)
      setSystemSkills((items) => items.map((item) => item.skillId === skillId ? { ...item, status: checked ? 0 : 1 } : item))
      toast.success(checked ? translate('adminSkills.globalEnabled') : translate('adminSkills.globalDisabled'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const toggleMarketStatus = useCallback(async (skillId: number, checked: boolean) => {
    try {
      await adminSkillsApi.updateMarketSkillStatus(skillId, checked ? 1 : 0)
      setMarketSkills((items) => items.map((item) => item.skillId === skillId ? { ...item, status: checked ? 1 : 0 } : item))
      toast.success(checked ? translate('adminSkills.publishedToast') : translate('adminSkills.unpublishedToast'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const confirmDelete = useCallback(async () => {
    if (!deleteTarget) return

    try {
      if (deleteTarget.type === 'system') {
        await adminSkillsApi.deleteSystemSkill(deleteTarget.id)
        await loadSystemSkills()
        toast.success(translate('adminSkills.globalDeleted'))
      }

      if (deleteTarget.type === 'market') {
        await adminSkillsApi.deleteMarketSkill(deleteTarget.id, cascadeDelete)
        await Promise.all([loadMarketSkills(), loadCategories()])
        if (cascadeDelete) {
          setInstalledSkills((items) => items.filter((item) => item.skillId !== deleteTarget.id))
        }
        toast.success(cascadeDelete ? translate('adminSkills.skillRecordsDeleted') : translate('adminSkills.marketDeleted'))
      }

      if (deleteTarget.type === 'category') {
        await adminSkillsApi.deleteSkillCategory(deleteTarget.id)
        await Promise.all([loadCategories(), loadMarketSkills()])
        toast.success(translate('adminSkills.catDeletedMoved'))
      }

      setDeleteTarget(null)
      setCascadeDelete(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [cascadeDelete, deleteTarget, loadCategories, loadMarketSkills, loadSystemSkills])

  const selectClawHubItem = useCallback(async (item: ClawHubItem) => {
    setClawHubSelected(item)
    try {
      const detail = await adminSkillsApi.getClawHubSkillDetail(item.slug)
      setClawHubSelected({
        ...item,
        displayName: item.displayName || detail.name,
        summary: item.summary || detail.description,
        version: detail.version,
        content: detail.content,
      })
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const importClawHubSkill = useCallback(async () => {
    if (!clawHubSelected) return
    if (!categories.some((item) => item.categoryId === clawHubCategoryId)) {
      toast.error(translate('adminPersonaTemplates.errCatNameRequired'))
      return
    }

    setImporting(true)
    try {
      await adminSkillsApi.importClawHubSkill({
        slug: clawHubSelected.slug,
        icon: clawHubIcon || undefined,
        categoryId: clawHubCategoryId,
      })
      await Promise.all([loadMarketSkills(), loadCategories()])
      setDrawer(null)
      setClawHubSelected(null)
      toast.success(translate('adminSkills.importSuccess'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setImporting(false)
    }
  }, [categories, clawHubCategoryId, clawHubIcon, clawHubSelected, loadCategories, loadMarketSkills])

  const uploadPublishFile = useCallback(async (file: File) => {
    if (!file.name.toLowerCase().endsWith('.zip')) {
      toast.error(translate('adminSkills.uploadFirst'))
      return
    }

    setPublishFile(file.name)
    setPublishUploadId('')
    try {
      const result = await uploadZip(file)
      setPublishUploadId(result.uploadId)
    } catch (error) {
      setPublishUploadId('')
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [uploadZip])

  const publishSkill = useCallback(async () => {
    if (!publishUploadId) {
      toast.error(translate('adminSkills.uploadFirst'))
      return
    }

    setPublishing(true)
    try {
      await adminSkillsApi.publishMarketSkill({
        uploadId: publishUploadId,
        icon: publishIcon || undefined,
        categoryId: publishCategoryId,
      })
      await Promise.all([loadMarketSkills(), loadCategories()])
      setDrawer(null)
      setPublishFile('')
      setPublishUploadId('')
      toast.success(translate('adminSkills.publishedToast2'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setPublishing(false)
    }
  }, [loadCategories, loadMarketSkills, publishCategoryId, publishIcon, publishUploadId])

  const saveCategory = useCallback(async () => {
    const name = categoryDraft.name.trim()
    if (!name) {
      toast.error(translate('adminPersonaTemplates.errCatNameRequired'))
      return
    }
    const displayWidth = [...name].reduce((w, c) => w + (c.charCodeAt(0) > 127 ? 2 : 1), 0)
    if (displayWidth > 20) {
      toast.error(translate('adminSkills.categoryNameTooLong'))
      return
    }

    try {
      if (categoryDraft.categoryId) {
        await adminSkillsApi.updateSkillCategory(categoryDraft.categoryId, { name, icon: categoryDraft.icon })
        toast.success(translate('adminSkills.catUpdated'))
      } else {
        await adminSkillsApi.createSkillCategory({ name, icon: categoryDraft.icon })
        toast.success(translate('adminSkills.catCreated'))
      }
      await Promise.all([loadCategories(), loadMarketSkills()])
      setCategoryDraft({ name: '', icon: 'folder' })
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [categoryDraft, loadCategories, loadMarketSkills])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <div className="border-b border-border px-6 py-4">
        <div className="flex items-start justify-between gap-6">
          <div>
            <h1 className="text-xl font-semibold text-text-1">{t('adminSkills.title')}</h1>
            <p className="mt-1 text-sm text-text-3">{t('adminSkills.description')}</p>
          </div>
          <Button variant="outline" onClick={() => setDrawer('categories')} className="text-highlight hover:text-highlight">
            <Settings2 className="size-4" />
            {t('adminSkills.categoryManagement')}
          </Button>
        </div>

        <div className="mt-5 inline-flex rounded-lg bg-bg-layer-2 p-1">
          {([
            ['global', t('adminSkills.global'), systemSkills.length],
            ['market', t('adminSkills.market'), marketSkills.length],
          ] as const).map(([key, label, count]) => (
            <button
              key={key}
              onClick={() => setActiveTab(key)}
              className={cn(
                'flex h-8 items-center gap-2 rounded-md px-3 text-sm transition-colors',
                activeTab === key ? 'bg-highlight text-highlight-fg' : 'text-text-3 hover:text-text-1',
              )}
            >
              {label}
              <span className={cn('text-xs', activeTab === key ? 'text-highlight-fg/70' : 'text-text-muted')}>{count}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        {activeTab === 'global' && (
          <section className="space-y-4">
              <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <SearchBox value={systemSearch} onChange={setSystemSearch} placeholder={t('adminSkills.searchPlaceholder')} />
                <SelectField value={systemStatus} onChange={setSystemStatus}>
                  <option value="all">{t('adminSkills.allStatuses')}</option>
                  <option value="0">{t('common.enabled')}</option>
                  <option value="1">{t('common.disabled')}</option>
                </SelectField>
              </div>
              <Button onClick={() => openSystemEditor()} className="text-highlight hover:text-highlight">
                <Plus className="size-4" />
                {t('adminSkills.newGlobalSkill')}
              </Button>
            </div>

            {filteredSystemSkills.length === 0 ? (
              <EmptyState
                title={t('adminSkills.noGlobalSkills')}
                description={t('adminSkills.globalSkillHint')}
                action={<Button onClick={() => openSystemEditor()}>{t('adminSkills.newGlobalSkill')}</Button>}
              />
            ) : (
              <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
                <table className="w-full table-fixed text-left text-sm">
                  <colgroup>
                    <col />
                    <col className="w-36" />
                    <col className="w-28" />
                    <col className="w-32" />
                  </colgroup>
                  <thead className="text-xs text-text-3">
                    <tr>
                      <th className="px-4 py-3 font-normal">{t('common.skills')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('common.status')}</th>
                      <th className="px-4 py-3 font-normal">{t('common.updated')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('adminSkills.actions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredSystemSkills.map((skill) => (
                      <tr key={skill.skillId} className="transition-colors hover:bg-bg-hover">
                        <td className="px-4 py-3">
                          <div className="font-medium text-text-1">{skill.description}</div>
                          <div className="mt-1 flex items-center gap-2">
                            <ScopeLabel>{skill.name}</ScopeLabel>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end gap-2">
                            <Switch size="sm" checked={skill.status === 0} onCheckedChange={(checked) => toggleSystemStatus(skill.skillId, checked)} />
                            <span className="min-w-10 whitespace-nowrap text-right text-xs text-text-2">{skill.status === 0 ? t('common.enabled') : t('common.disabled')}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-text-3">{formatDate(skill.updated)}</td>
                        <td className="px-4 py-3">
                          <div className="flex justify-end gap-1">
                            <Button variant="ghost" size="sm" onClick={() => openSystemEditor(skill)}>{t('common.edit')}</Button>
                            <Button variant="ghost" size="icon-sm" onClick={() => setDeleteTarget({ type: 'system', id: skill.skillId, label: skill.name })}>
                              <Trash2 className="size-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        )}

        {activeTab === 'market' && (
          <section className="space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <SearchBox value={marketSearch} onChange={setMarketSearch} placeholder={t('adminSkills.searchSkillName')} />
                <SelectField value={marketSource} onChange={setMarketSource}>
                  <option value="all">{t('adminSkills.allSources')}</option>
                  <option value="clawhub">ClawHub</option>
                  <option value="custom_upload">{t('adminSkills.customUpload')}</option>
                </SelectField>
                <SelectField value={marketStatus} onChange={setMarketStatus}>
                  <option value="all">{t('adminSkills.allStatuses')}</option>
                  <option value="1">{t('adminSkills.published')}</option>
                  <option value="0">{t('adminSkills.unpublished')}</option>
                </SelectField>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" onClick={() => setDrawer('publish')} className="text-highlight hover:text-highlight">
                  <Upload className="size-4" />
                  {t('adminSkills.publishSkill')}
                </Button>
                <Button onClick={() => openClawHubDrawer()} className="text-highlight hover:text-highlight">
                  <Download className="size-4" />
                  {t('adminSkills.importFromClawhub')}
                </Button>
              </div>
            </div>

            {filteredMarketSkills.length === 0 ? (
              <EmptyState
                title={t('adminSkills.emptyMarket')}
                description={t('adminSkills.marketHint')}
                action={<div className="flex gap-2"><Button onClick={() => openClawHubDrawer()}>{t('adminSkills.importFromClawhub')}</Button><Button variant="outline" onClick={() => setDrawer('publish')}>{t('adminSkills.publishSkill')}</Button></div>}
              />
            ) : (
              <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
                <table className="w-full table-fixed text-left text-sm">
                  <colgroup>
                    <col />
                    <col className="w-28" />
                    <col className="w-36" />
                    <col className="w-24" />
                    <col className="w-28" />
                    <col className="w-20" />
                    <col className="w-36" />
                  </colgroup>
                  <thead className="text-xs text-text-3">
                    <tr>
                      <th className="px-4 py-3 font-normal">{t('common.skills')}</th>
                      <th className="px-4 py-3 font-normal">{translate('skills.source')}</th>
                      <th className="px-4 py-3 font-normal">{t('common.category')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('adminSkills.installs')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('common.status')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('common.sort')}</th>
                      <th className="px-4 py-3 text-right font-normal">{t('adminSkills.actions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredMarketSkills.map((skill) => (
                      <tr key={skill.skillId} className="transition-colors hover:bg-bg-hover">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <Icon name={skill.icon} className="size-5 text-interactive" />
                            <div>
                              <div className="font-medium text-text-1">{skill.name}</div>
                              <div className="mt-1 max-w-xl truncate text-xs text-text-3">{skill.description}</div>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3"><ScopeLabel>{getSourceLabel(skill.source)}</ScopeLabel></td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2 text-text-2">
                            <Icon name={skill.categoryIcon} className="size-3.5 text-interactive" />
                            <span className="max-w-[100px] truncate">{skill.categoryName}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={() => {
                              setSelectedMarketSkillId(skill.skillId)
                              setDrawer('users')
                            }}
                            className="text-text-1 underline-offset-4 hover:underline"
                          >
                            {skill.installedCount}
                          </button>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end gap-2">
                            <Switch size="sm" checked={skill.status === 1} onCheckedChange={(checked) => toggleMarketStatus(skill.skillId, checked)} />
                            <span className="min-w-10 whitespace-nowrap text-right text-xs text-text-2">{skill.status === 1 ? t('adminSkills.published') : t('adminSkills.unpublished')}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-right text-text-3">{skill.sortOrder}</td>
                        <td className="px-4 py-3">
                          <div className="flex justify-end gap-1">
                            <Button variant="ghost" size="sm" onClick={() => openMarketEditor(skill)}>{t('common.edit')}</Button>
                            <Button variant="ghost" size="sm" onClick={() => {
                              setSelectedMarketSkillId(skill.skillId)
                              setDrawer('users')
                            }}>{t('common.user')}</Button>
                            <Button variant="ghost" size="icon-sm" onClick={() => setDeleteTarget({ type: 'market', id: skill.skillId, label: skill.name })}>
                              <Trash2 className="size-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        )}
      </div>

      <Drawer
        open={drawer === 'system-editor'}
        onClose={() => setDrawer(null)}
        title={systemForm.skillId ? t('adminSkills.editGlobalSkill') : t('adminSkills.newGlobalSkill')}
                description={t('adminSkills.editSkillMdHint')}
        width="w-[720px]"
      >
        <div className="flex h-full">
          <div className="flex w-[260px] shrink-0 flex-col gap-3 border-r border-border p-4">
            <div className="rounded-lg bg-bg-layer-2 p-3">
              <div className="text-xs text-text-3">{t('common.name')}</div>
              <div className="mt-1 break-all text-sm font-medium text-text-1">{systemParsed.name || t('adminSkills.unresolved')}</div>
            </div>
            <div className="rounded-lg bg-bg-layer-2 p-3">
              <div className="text-xs text-text-3">{t('common.description')}</div>
              <div className="mt-1 text-sm text-text-1">{systemParsed.description || t('adminSkills.unresolved')}</div>
            </div>
            <div className="rounded-lg bg-bg-layer-2 p-3">
              <div className="text-xs text-text-3">{t('adminSkills.validation')}</div>
              <div className="mt-2 space-y-1">
                {systemParsed.errors.length === 0 ? (
                  <div className="flex items-center gap-1.5 text-sm text-text-1"><Check className="size-4" /> {t('adminSkills.formatValid')}</div>
                ) : systemParsed.errors.map((error) => (
                  <div key={error} className="flex items-center gap-1.5 text-sm text-text-1"><AlertCircle className="size-4" /> {error}</div>
                ))}
              </div>
            </div>
            <p className="text-xs leading-5 text-text-3">{t('adminSkills.globalSkillDesc')}</p>
          </div>
          <div className="flex min-w-0 flex-1 flex-col">
            <textarea
              value={systemForm.content}
              onChange={(event) => setSystemForm((form) => ({ ...form, content: event.target.value }))}
              className="min-h-0 flex-1 resize-none bg-transparent p-4 font-mono text-sm leading-6 text-text-1 outline-none placeholder:text-text-muted"
              spellCheck={false}
            />
            <div className="flex items-center justify-between border-t border-border px-4 py-3">
              <span className="text-xs text-text-3">{t('adminSkills.effectiveInNew')}</span>
              <div className="flex gap-2">
                <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{t('common.cancel')}</Button>
                <Button onClick={saveSystemSkill} disabled={saving} className="text-highlight">
                  {saving && <Loader2 className="size-4 animate-spin" />}
                  {t('common.save')}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </Drawer>

      <Drawer
        open={drawer === 'market-editor'}
        onClose={() => setDrawer(null)}
        title={t('adminSkills.editMarketSkill')}
        description={t('adminSkills.marketEditHint')}
      >
        {selectedMarketSkill && marketForm && (
          <div className="flex h-full flex-col">
            <div className="flex-1 space-y-4 overflow-auto p-4">
              <div className="rounded-lg bg-bg-layer-2 p-3">
                <div className="flex items-center gap-3">
                  <SkillIcon icon={marketForm.icon} />
                  <div>
                    <div className="font-medium text-text-1">{selectedMarketSkill.name}</div>
                    <div className="mt-1 text-xs text-text-3">{getSourceLabel(selectedMarketSkill.source)}，{t('adminSkills.nameReadonly')}</div>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label className="text-xs text-text-2">{t('common.icon')}</Label>
                  <IconPickerTrigger value={marketForm.icon} onChange={(icon) => setMarketForm({ ...marketForm, icon })} />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs text-text-2">{t('adminPersonaTemplates.sortOrder')}</Label>
                  <Input type="number" value={marketForm.sortOrder} onChange={(event) => setMarketForm({ ...marketForm, sortOrder: Number(event.target.value) })} />
                </div>
              </div>

              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{t('common.category')}</Label>
                <SelectField value={String(marketForm.categoryId)} onChange={(value) => setMarketForm({ ...marketForm, categoryId: Number(value) })} className="w-full">
                  {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
                </SelectField>
              </div>

              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{t('adminSkills.adminNotes')}</Label>
                <textarea
                  value={marketForm.remark}
                  onChange={(event) => setMarketForm({ ...marketForm, remark: event.target.value })}
                  className="min-h-24 w-full resize-none rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm text-text-1 outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                  placeholder={t('adminSkills.adminOnly')}
                />
              </div>

              <div className="rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-3">
                {translate('common.installed')} {selectedMarketSkill.installedCount} {t('adminSkills.installedCount')}
              </div>

              {selectedMarketSkill.content && (
                <div className="space-y-1.5">
                  <Label className="text-xs text-text-2">SKILL.md</Label>
                  <pre className="max-h-[360px] overflow-auto rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-2 whitespace-pre-wrap break-words">{selectedMarketSkill.content}</pre>
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2 border-t border-border p-4">
              <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{t('common.cancel')}</Button>
              <Button onClick={saveMarketSkill} disabled={saving} className="text-highlight">{saving && <Loader2 className="size-4 animate-spin" />}{t('common.save')}</Button>
            </div>
          </div>
        )}
      </Drawer>

      <Drawer
        open={drawer === 'clawhub'}
        onClose={() => setDrawer(null)}
        title={t('adminSkills.importTitle')}
        description={t('adminSkills.importDesc')}
        width="w-[760px]"
      >
        {!clawHubTokenConfigured ? (
          <div className="flex h-full items-center justify-center p-6">
            <div className="max-w-sm text-center">
              <AlertCircle className="mx-auto mb-3 size-8 text-text-muted" />
              <h3 className="text-sm font-semibold text-text-1">{t('adminSkills.clawhubNoToken')}</h3>
              <p className="mt-1 text-sm text-text-3">{t('adminSkills.clawhubTokenHint')}</p>
              <div className="mt-4 flex justify-center gap-2">
                <Button variant="outline" onClick={() => setClawHubTokenConfigured(true)}>{t('adminSkills.simulateConfigured')}</Button>
                <Button>{t('adminSkills.goToSettings')}</Button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex h-full">
            <div className="flex w-[48%] flex-col border-r border-border">
              <div className="space-y-3 border-b border-border p-4">
                <div className="inline-flex rounded-lg bg-bg-layer-2 p-1">
                  <button onClick={() => changeClawHubMode('search')} className={cn('h-7 rounded-md px-3 text-sm', clawHubMode === 'search' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3')}>{t('common.search')}</button>
                  <button onClick={() => changeClawHubMode('explore')} className={cn('h-7 rounded-md px-3 text-sm', clawHubMode === 'explore' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3')}>{t('adminSkills.browse')}</button>
                </div>
                {clawHubMode === 'search' && (
                  <div className="space-y-2">
                    <div className="flex gap-2">
                      <div className="relative min-w-0 flex-1">
                        <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-text-3" />
                        <Input
                          value={clawHubQuery}
                          onChange={(event) => setClawHubQuery(event.target.value)}
                          onKeyDown={(event) => {
                            if (event.key === 'Enter') submitClawHubSearch()
                          }}
                          placeholder={t('adminSkills.searchHint')}
                          className="pl-8 text-sm"
                          disabled={clawHubSearching}
                        />
                      </div>
                      <Button onClick={submitClawHubSearch} disabled={clawHubSearching || !clawHubQuery.trim()}>
                        {clawHubSearching && <Loader2 className="size-4 animate-spin" />}
                        {t('common.search')}
                      </Button>
                    </div>
                    <p className="text-xs leading-5 text-text-3">{t('adminSkills.searchSlow')}</p>
                  </div>
                )}
                {clawHubMode === 'explore' && (
                  <div className="grid grid-cols-3 gap-1 rounded-lg bg-bg-layer-2 p-1">
                    {([
                      ['trending', t('adminSkills.trending')],
                      ['newest', t('adminSkills.newest')],
                      ['updated', t('common.updated')],
                      ['downloads', t('adminSkills.downloads')],
                      ['stars', t('adminSkills.stars')],
                      ['installs', t('skills.installCount')],
                    ] as const).map(([sort, label]) => (
                      <button
                        key={sort}
                        onClick={() => changeClawHubSort(sort)}
                        className={cn(
                          'h-7 rounded-md px-2 text-xs transition-colors',
                          clawHubSort === sort ? 'bg-bg-layer-3 text-text-1' : 'text-text-3 hover:text-text-1',
                        )}
                      >
                        {label}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <div className="flex-1 overflow-auto p-2">
                {clawHubSearching || clawHubExploring ? (
                  <div className="space-y-2 p-2">
                    {[0, 1, 2].map((item) => (
                      <div key={item} className="rounded-lg bg-bg-layer-2 p-3">
                        <div className="h-4 w-2/3 rounded bg-bg-layer-3" />
                        <div className="mt-3 h-3 w-full rounded bg-bg-layer-3" />
                        <div className="mt-2 h-3 w-1/2 rounded bg-bg-layer-3" />
                      </div>
                    ))}
                  </div>
                ) : clawHubResults.length === 0 ? (
                  <div className="flex h-full items-center justify-center px-6 text-center">
                    <div>
                      <Search className="mx-auto mb-3 size-8 text-text-muted" />
                      <div className="text-sm font-medium text-text-1">
                        {clawHubMode === 'search' && clawHubSubmittedQuery ? translate('adminPersonaTemplates.noMatch') : translate('adminSkills.searchHint')}
                      </div>
                      <p className="mt-1 text-xs leading-5 text-text-3">
                        {clawHubMode === 'search'
                          ? translate('adminSkills.clawhubSearchTip')
                          : translate('adminSkills.clawhubRankingWait')}
                      </p>
                    </div>
                  </div>
                ) : clawHubResults.map((item) => (
                  <button
                    key={item.slug}
                    onClick={() => selectClawHubItem(item)}
                    className={cn(
                      'mb-1 w-full rounded-lg px-3 py-2 text-left transition-colors hover:bg-bg-hover',
                      clawHubSelected?.slug === item.slug && 'bg-bg-layer-3',
                    )}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="font-medium text-text-1">{item.displayName}</div>
                      <span className="text-xs text-text-3">v{item.version}</span>
                    </div>
                    <div className="mt-1 line-clamp-2 text-xs leading-5 text-text-3">{item.summary}</div>
                    <div className="mt-2 flex gap-3 text-xs text-text-muted">
                      <span>score {item.score}</span>
                      <span>{item.downloads} {t('adminSkills.downloads')}</span>
                      <span>{item.installs} {translate('skills.installCount')}</span>
                      <span>{item.stars} {t('adminSkills.stars')}</span>
                    </div>
                  </button>
                ))}
              </div>
            </div>
            <div className="flex min-w-0 flex-1 flex-col">
              {clawHubSelected ? (
                <>
                  <div className="flex-1 overflow-auto p-4">
                    <div className="mb-4">
                      <h3 className="text-base font-semibold text-text-1">{clawHubSelected.displayName}</h3>
                      <p className="mt-1 text-sm text-text-3">{clawHubSelected.summary}</p>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="space-y-1.5">
                        <Label className="text-xs text-text-2">{t('common.category')}</Label>
                        <SelectField value={String(clawHubCategoryId)} onChange={(value) => setClawHubCategoryId(Number(value))} className="w-full">
                          {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
                        </SelectField>
                      </div>
                      <div className="space-y-1.5">
                        <Label className="text-xs text-text-2">{translate('common.icon')}</Label>
                        <IconPickerTrigger value={clawHubIcon} onChange={setClawHubIcon} />
                      </div>
                    </div>
                    <pre className="mt-4 max-h-[360px] overflow-auto rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-2">{clawHubSelected.content}</pre>
                  </div>
                  <div className="border-t border-border p-4">
                    {importing && <div className="mb-3 flex items-center gap-2 text-xs text-text-3"><Loader2 className="size-4 animate-spin" /> {t('adminSkills.downloadingMd')}</div>}
                    <div className="flex justify-end gap-2">
                      <Button variant="outline" onClick={() => setClawHubSelected(null)} disabled={importing} className="text-highlight hover:text-highlight/80">{t('common.back')}</Button>
                      <Button onClick={importClawHubSkill} disabled={importing} className="text-highlight">{importing && <Loader2 className="size-4 animate-spin" />}{t('adminSkills.confirmImport')}</Button>
                    </div>
                  </div>
                </>
              ) : (
                <div className="flex h-full items-center justify-center p-6 text-center text-sm text-text-3">
                  {t('adminSkills.selectToPreview')}
                </div>
              )}
            </div>
          </div>
        )}
      </Drawer>

      <Drawer open={drawer === 'publish'} onClose={() => setDrawer(null)} title={t('adminSkills.publishTitle')} description={t('adminSkills.publishDesc')}>
        <div className="flex h-full flex-col">
          <div className="flex-1 space-y-4 overflow-auto p-4">
            <input
              ref={fileInputRef}
              type="file"
              accept=".zip,application/zip,application/x-zip-compressed"
              className="hidden"
              onChange={(event) => {
                const file = event.target.files?.[0]
                if (file) uploadPublishFile(file)
                event.target.value = ''
              }}
            />
            <div
              role="button"
              tabIndex={0}
              onClick={() => { if (!isUploadingZip && !publishing) fileInputRef.current?.click() }}
              onKeyDown={(event) => {
                if ((event.key === 'Enter' || event.key === ' ') && !isUploadingZip && !publishing) {
                  event.preventDefault()
                  fileInputRef.current?.click()
                }
              }}
              onDragOver={(event) => {
                event.preventDefault()
                event.stopPropagation()
                if (isUploadingZip || publishing) return
                event.dataTransfer.dropEffect = 'copy'
                if (!isDraggingZip) setIsDraggingZip(true)
              }}
              onDragEnter={(event) => {
                event.preventDefault()
                event.stopPropagation()
                if (isUploadingZip || publishing) return
                setIsDraggingZip(true)
              }}
              onDragLeave={(event) => {
                event.preventDefault()
                event.stopPropagation()
                if (event.currentTarget.contains(event.relatedTarget as Node)) return
                setIsDraggingZip(false)
              }}
              onDrop={(event) => {
                event.preventDefault()
                event.stopPropagation()
                setIsDraggingZip(false)
                if (isUploadingZip || publishing) return
                const file = event.dataTransfer.files?.[0]
                if (file) uploadPublishFile(file)
              }}
              aria-disabled={isUploadingZip || publishing}
              className={cn(
                'flex h-36 w-full flex-col items-center justify-center rounded-lg border border-dashed text-center transition-colors',
                isUploadingZip || publishing
                  ? 'cursor-not-allowed border-border bg-bg-layer-2 opacity-60'
                  : 'cursor-pointer hover:bg-bg-hover',
                isDraggingZip
                  ? 'border-interactive bg-bg-hover'
                  : 'border-border bg-bg-layer-2',
              )}
            >
              {isUploadingZip ? <Loader2 className="mb-2 size-8 animate-spin text-text-3" /> : <FileArchive className="mb-2 size-8 text-text-3" />}
              <span className="text-sm font-medium text-text-1">{publishFile || t('adminSkills.zipDropHint')}</span>
              <span className="mt-1 text-xs text-text-3">{t('adminSkills.zipHint')}</span>
            </div>
            {publishFile && (
              <div className="rounded-lg bg-bg-layer-2 p-3">
                <div className="flex justify-between text-xs text-text-3"><span>{publishUploadId || publishFile}</span><span>{publishProgress}%</span></div>
                <div className="mt-2 h-1.5 overflow-hidden rounded bg-bg-layer-3">
                  <div className="h-full bg-text-1 transition-all duration-200" style={{ width: `${publishProgress}%` }} />
                </div>
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{translate('common.category')}</Label>
                <SelectField value={String(publishCategoryId)} onChange={(value) => setPublishCategoryId(Number(value))} className="w-full">
                  {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
                </SelectField>
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{translate('common.icon')}</Label>
                <IconPickerTrigger value={publishIcon} onChange={setPublishIcon} />
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-border p-4">
            <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{translate('common.cancel')}</Button>
            <Button onClick={publishSkill} disabled={publishing || isUploadingZip} className="text-highlight">{publishing && <Loader2 className="size-4 animate-spin" />}{translate('adminSkills.publishSkill')}</Button>
          </div>
        </div>
      </Drawer>

      <Drawer open={drawer === 'categories'} onClose={() => setDrawer(null)} title={t('adminSkills.categoryManagement')} description={t('adminSkills.catManageHint')}>
        <div className="flex h-full flex-col">
          <div className="border-b border-border p-4">
            <div className="grid grid-cols-[1fr_180px_auto] gap-2">
              <Input placeholder={t('adminPersonaTemplates.categoryName')} maxLength={20} value={categoryDraft.name} onChange={(event) => setCategoryDraft({ ...categoryDraft, name: event.target.value })} />
              <IconPickerTrigger value={categoryDraft.icon} onChange={(icon) => setCategoryDraft({ ...categoryDraft, icon })} />
              <Button onClick={saveCategory}>{categoryDraft.categoryId ? t('common.save') : t('common.create')}</Button>
            </div>
          </div>
          <div className="flex-1 overflow-auto p-2">
            {categories.map((category) => (
              <div key={category.categoryId} className="mb-1 flex items-center justify-between rounded-lg px-3 py-2 transition-colors hover:bg-bg-hover">
                <div className="flex items-center gap-3">
                  <SkillIcon icon={category.icon} />
                  <div>
                    <div className="flex items-center gap-2 text-sm font-medium text-text-1">
                      {category.name}
                      {category.isDefault === 1 && <ScopeLabel>{t('common.default')}</ScopeLabel>}
                    </div>
                    <div className="mt-1 text-xs text-text-3">{category.skillCount} {t('adminSkills.skillNumUpdated')} {formatDate(category.updated)}</div>
                  </div>
                </div>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={category.isDefault === 1}
                    onClick={() => setCategoryDraft({ categoryId: category.categoryId, name: category.name, icon: category.icon })}
                  >{t('common.edit')}</Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    disabled={category.isDefault === 1}
                    onClick={() => setDeleteTarget({ type: 'category', id: category.categoryId, label: category.name })}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Drawer>

      <Drawer
        open={drawer === 'users'}
        onClose={() => setDrawer(null)}
        title={t('adminSkills.installedUsers')}
        description={selectedMarketSkill ? selectedMarketSkill.name : undefined}
      >
        {selectedSkillUsers.length === 0 ? (
          <EmptyState title={t('adminSkills.noInstalledUsers')} description={t('adminSkills.installedUsersHint')} />
        ) : (
          <div className="p-4">
            <div className="overflow-hidden rounded-lg bg-bg-layer-2">
              <table className="w-full text-left text-sm">
                <thead className="text-xs text-text-3">
                  <tr>
                    <th className="px-3 py-2 font-normal">{t('common.user')}</th>
                    <th className="px-3 py-2 font-normal">{t('common.status')}</th>
                    <th className="px-3 py-2 font-normal">{translate('skills.installTime')}</th>
                  </tr>
                </thead>
                <tbody>
                  {selectedSkillUsers.map((record) => (
                    <tr key={record.recordId}>
                      <td className="px-3 py-2 text-text-1">{record.userId}</td>
                      <td className="px-3 py-2"><StatusText tone={record.installStatus === 3 ? 'warn' : record.installStatus === 2 ? 'good' : 'neutral'}>{getInstallStatusLabel(record.installStatus)}</StatusText></td>
                      <td className="px-3 py-2 text-text-3">{formatDate(record.created)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </Drawer>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {deleteTarget?.type === 'market' ? t('adminSkills.deleteMarketSkill') : deleteTarget?.type === 'category' ? t('adminSkills.deleteCategory') : t('adminSkills.deleteGlobalSkill')}
            </DialogTitle>
            <DialogDescription>
              {deleteTarget?.type === 'market'
                ? t('adminSkills.deleteMarketSkillDesc')
                : deleteTarget?.type === 'category'
                  ? t('adminSkills.deleteCatDesc')
                  : t('adminSkills.deleteGlobalSkillDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="rounded-lg bg-bg-layer-2 px-3 py-2 text-sm text-text-1">{deleteTarget?.label}</div>
            {deleteTarget?.type === 'market' && (
              <label className="flex items-center gap-2 text-sm text-text-2">
                <input
                  type="checkbox"
                  checked={cascadeDelete}
                  onChange={(event) => setCascadeDelete(event.target.checked)}
                  className="size-4 rounded border-border bg-transparent"
                />
                {translate('adminSkills.alsoDeleteRecords')}
              </label>
            )}
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>{translate('common.cancel')}</Button>
              <Button variant="destructive" onClick={confirmDelete}>
                {deleteTarget?.type === 'market' && cascadeDelete ? translate('adminSkills.deleteSkillAndRecords') : translate('common.delete')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
