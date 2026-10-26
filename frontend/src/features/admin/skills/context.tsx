import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { adminSkillsApi } from '@/api'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { useT, t as translate } from '@/i18n'
import {
  type SystemSkill, type MarketSkill, type InstalledSkill, type SkillCategory,
  type ClawHubItem, type ClawHubSort, type DeleteTarget, type DrawerKey,
  type SystemForm, type MarketForm, type SourceType,
  parseSkillContent, GLOBAL_TEMPLATE,
} from './shared'

export interface AdminSkillsContextValue {
  systemSkills: SystemSkill[]
  marketSkills: MarketSkill[]
  installedSkills: InstalledSkill[]
  categories: SkillCategory[]
  clawHubItems: ClawHubItem[]
  systemSearch: string
  setSystemSearch: (v: string) => void
  systemStatus: string
  setSystemStatus: (v: string) => void
  marketSearch: string
  setMarketSearch: (v: string) => void
  marketSource: string
  setMarketSource: (v: string) => void
  marketStatus: string
  setMarketStatus: (v: string) => void
  drawer: DrawerKey
  setDrawer: (v: DrawerKey) => void
  saving: boolean
  deleteTarget: DeleteTarget | null
  setDeleteTarget: (v: DeleteTarget | null) => void
  cascadeDelete: boolean
  setCascadeDelete: (v: boolean) => void
  systemForm: SystemForm
  setSystemForm: React.Dispatch<React.SetStateAction<SystemForm>>
  systemParsed: { name: string; description: string; errors: string[] }
  filteredSystemSkills: SystemSkill[]
  openSystemEditor: (skill?: SystemSkill) => Promise<void>
  saveSystemSkill: () => Promise<void>
  toggleSystemStatus: (skillId: number, checked: boolean) => Promise<void>
  marketForm: MarketForm | null
  setMarketForm: React.Dispatch<React.SetStateAction<MarketForm | null>>
  filteredMarketSkills: MarketSkill[]
  selectedMarketSkill: MarketSkill | null
  selectedMarketSkillId: number | null
  setSelectedMarketSkillId: (v: number | null) => void
  selectedSkillUsers: InstalledSkill[]
  openMarketEditor: (skill: MarketSkill) => Promise<void>
  saveMarketSkill: () => Promise<void>
  toggleMarketStatus: (skillId: number, checked: boolean) => Promise<void>
  openUsersDrawer: (skillId: number) => Promise<void>
  clawHubMode: 'search' | 'explore'
  setClawHubMode: (v: 'search' | 'explore') => void
  clawHubQuery: string
  setClawHubQuery: (v: string) => void
  clawHubSubmittedQuery: string
  clawHubSort: ClawHubSort
  setClawHubSort: (v: ClawHubSort) => void
  clawHubSearching: boolean
  clawHubExploring: boolean
  clawHubSelected: ClawHubItem | null
  setClawHubSelected: (v: ClawHubItem | null) => void
  clawHubCategoryId: number
  setClawHubCategoryId: (v: number) => void
  clawHubIcon: string
  setClawHubIcon: (v: string) => void
  importing: boolean
  clawHubResults: ClawHubItem[]
  openClawHubDrawer: () => void
  changeClawHubMode: (mode: 'search' | 'explore') => void
  changeClawHubSort: (sort: ClawHubSort) => void
  submitClawHubSearch: () => Promise<void>
  selectClawHubItem: (item: ClawHubItem) => Promise<void>
  importClawHubSkill: () => Promise<void>
  publishFile: string
  publishUploadId: string
  publishCategoryId: number
  setPublishCategoryId: (v: number) => void
  publishIcon: string
  setPublishIcon: (v: string) => void
  publishing: boolean
  publishProgress: number
  isUploadingZip: boolean
  isDraggingZip: boolean
  setIsDraggingZip: (v: boolean) => void
  fileInputRef: React.RefObject<HTMLInputElement | null>
  uploadPublishFile: (file: File) => Promise<void>
  publishSkill: () => Promise<void>
  categoryDraft: { categoryId?: number; name: string; icon: string }
  setCategoryDraft: React.Dispatch<React.SetStateAction<{ categoryId?: number; name: string; icon: string }>>
  saveCategory: () => Promise<void>
  confirmDelete: () => Promise<void>
}

const AdminSkillsContext = createContext<AdminSkillsContextValue>(null!)

export function useAdminSkills() {
  return useContext(AdminSkillsContext)
}

export function AdminSkillsProvider({ children }: { children: React.ReactNode }) {
  const t = useT()

  /* ---------- state ---------- */
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

  /* ---------- computed ---------- */
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

  const publishProgress = publishUploadId ? 100 : uploadProgress?.percentage ?? 0

  /* ---------- data loading ---------- */
  const loadSystemSkills = useCallback(async () => {
    const result = await adminSkillsApi.getSystemSkills({
      pageSize: 100,
      search: systemSearch.trim() || undefined,
      status: systemStatus === 'all' ? undefined : Number(systemStatus) as 0 | 1,
    })
    setSystemSkills(result.list.map((skill: any) => ({ ...skill, content: '' })))
  }, [systemSearch, systemStatus])

  const loadMarketSkills = useCallback(async () => {
    const result = await adminSkillsApi.getMarketSkills({
      pageSize: 100,
      search: marketSearch.trim() || undefined,
      source: marketSource === 'all' ? undefined : marketSource as SourceType,
      status: marketStatus === 'all' ? undefined : Number(marketStatus) as 0 | 1,
    })
    setMarketSkills(result.list.map((skill: any) => ({ ...skill, sourceUrl: '', remark: '', content: '' })))
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

  useEffect(() => { loadInitialData() }, [loadInitialData])

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
      setClawHubItems(result.items.map((item: any) => ({
        slug: item.slug, displayName: item.displayName, summary: item.summary,
        version: '', score: 0,
        downloads: item.stats.downloads, installs: item.stats.installsAllTime,
        stars: item.stats.stars, updatedAt: new Date(item.updatedAt).toISOString(), content: '',
      })))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setClawHubExploring(false)
    }
  }, [])

  const submitClawHubSearch = useCallback(async () => {
    const query = clawHubQuery.trim()
    if (!query) { toast.error(translate('adminSkills.emptySearchQuery')); return }
    setClawHubSelected(null)
    setClawHubSearching(true)
    setClawHubSubmittedQuery(query)
    try {
      const result = await adminSkillsApi.searchClawHub({ q: query, limit: 25 })
      setClawHubItems(result.results.map((item: any) => ({
        slug: item.slug, displayName: item.displayName, summary: item.summary,
        version: item.version, score: item.score,
        downloads: 0, installs: 0, stars: 0,
        updatedAt: new Date(item.updatedAt).toISOString(), content: '',
      })))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    } finally {
      setClawHubSearching(false)
    }
  }, [clawHubQuery])

  /* ---------- system skill actions ---------- */
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
    if (parsed.errors.length > 0) { toast.error(parsed.errors[0]); return }
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

  const toggleSystemStatus = useCallback(async (skillId: number, checked: boolean) => {
    try {
      await adminSkillsApi.updateSystemSkillStatus(skillId, checked ? 0 : 1)
      setSystemSkills((items) => items.map((item) => item.skillId === skillId ? { ...item, status: checked ? 0 : 1 } : item))
      toast.success(checked ? translate('adminSkills.globalEnabled') : translate('adminSkills.globalDisabled'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  /* ---------- market skill actions ---------- */
  const openMarketEditor = useCallback(async (skill: MarketSkill) => {
    setSelectedMarketSkillId(skill.skillId)
    try {
      const detail = await adminSkillsApi.getMarketSkillDetail(skill.skillId)
      setMarketSkills((items) => items.map((item) => item.skillId === detail.skillId ? { ...item, ...detail } : item))
      setMarketForm({
        skillId: detail.skillId, icon: detail.icon,
        categoryId: detail.categoryId, sortOrder: detail.sortOrder, remark: detail.remark,
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
        icon: marketForm.icon, categoryId: marketForm.categoryId,
        sortOrder: marketForm.sortOrder, remark: marketForm.remark,
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

  const toggleMarketStatus = useCallback(async (skillId: number, checked: boolean) => {
    try {
      await adminSkillsApi.updateMarketSkillStatus(skillId, checked ? 1 : 0)
      setMarketSkills((items) => items.map((item) => item.skillId === skillId ? { ...item, status: checked ? 1 : 0 } : item))
      toast.success(checked ? translate('adminSkills.publishedToast') : translate('adminSkills.unpublishedToast'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const openUsersDrawer = useCallback(async (skillId: number) => {
    setSelectedMarketSkillId(skillId)
    setDrawer('users')
    try {
      const skill = marketSkills.find((s) => s.skillId === skillId)
      const result = await adminSkillsApi.getMarketSkillUsers(skillId, { pageSize: 100 })
      setInstalledSkills(result.list.map((record: any, index: number) => ({
        recordId: index + 1, userId: record.userId, skillId,
        skillName: skill?.name ?? '', categoryName: skill?.categoryName ?? '',
        categoryIcon: skill?.categoryIcon ?? '', source: skill?.source ?? 'custom_upload',
        installStatus: record.installStatus, reason: '', created: record.created,
      })))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [marketSkills])

  /* ---------- delete ---------- */
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

  /* ---------- ClawHub actions ---------- */
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

  const selectClawHubItem = useCallback(async (item: ClawHubItem) => {
    setClawHubSelected(item)
    try {
      const detail = await adminSkillsApi.getClawHubSkillDetail(item.slug)
      setClawHubSelected({
        ...item,
        displayName: item.displayName || detail.name,
        summary: item.summary || detail.description,
        version: detail.version, content: detail.content,
      })
    } catch (error) {
      toast.error(error instanceof Error ? error.message : translate('common.failed'))
    }
  }, [])

  const importClawHubSkill = useCallback(async () => {
    if (!clawHubSelected) return
    if (!categories.some((item) => item.categoryId === clawHubCategoryId)) {
      toast.error(translate('adminPersonaTemplates.errCatNameRequired')); return
    }
    setImporting(true)
    try {
      await adminSkillsApi.importClawHubSkill({
        slug: clawHubSelected.slug, icon: clawHubIcon || undefined, categoryId: clawHubCategoryId,
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

  /* ---------- publish actions ---------- */
  const uploadPublishFile = useCallback(async (file: File) => {
    if (!file.name.toLowerCase().endsWith('.zip')) {
      toast.error(translate('adminSkills.uploadFirst')); return
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
    if (!publishUploadId) { toast.error(translate('adminSkills.uploadFirst')); return }
    setPublishing(true)
    try {
      await adminSkillsApi.publishMarketSkill({
        uploadId: publishUploadId, icon: publishIcon || undefined, categoryId: publishCategoryId,
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

  /* ---------- category actions ---------- */
  const saveCategory = useCallback(async () => {
    const name = categoryDraft.name.trim()
    if (!name) { toast.error(translate('adminPersonaTemplates.errCatNameRequired')); return }
    const displayWidth = [...name].reduce((w, c) => w + (c.charCodeAt(0) > 127 ? 2 : 1), 0)
    if (displayWidth > 20) { toast.error(translate('adminSkills.categoryNameTooLong')); return }
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

  /* ---------- provide ---------- */
  return (
    <AdminSkillsContext.Provider value={{
      // data
      systemSkills, marketSkills, installedSkills, categories, clawHubItems,
      // filters
      systemSearch, setSystemSearch, systemStatus, setSystemStatus,
      marketSearch, setMarketSearch, marketSource, setMarketSource, marketStatus, setMarketStatus,
      // drawer
      drawer, setDrawer, saving,
      deleteTarget, setDeleteTarget, cascadeDelete, setCascadeDelete,
      // system
      systemForm, setSystemForm, systemParsed, filteredSystemSkills,
      openSystemEditor, saveSystemSkill, toggleSystemStatus,
      // market
      marketForm, setMarketForm, filteredMarketSkills,
      selectedMarketSkill, selectedMarketSkillId, setSelectedMarketSkillId, selectedSkillUsers,
      openMarketEditor, saveMarketSkill, toggleMarketStatus, openUsersDrawer,
      // clawhub
      clawHubMode, setClawHubMode, clawHubQuery, setClawHubQuery,
      clawHubSubmittedQuery, clawHubSort, setClawHubSort,
      clawHubSearching, clawHubExploring, clawHubSelected, setClawHubSelected,
      clawHubCategoryId, setClawHubCategoryId, clawHubIcon, setClawHubIcon,
      importing, clawHubResults,
      openClawHubDrawer, changeClawHubMode, changeClawHubSort,
      submitClawHubSearch, selectClawHubItem, importClawHubSkill,
      // publish
      publishFile, publishUploadId, publishCategoryId, setPublishCategoryId,
      publishIcon, setPublishIcon, publishing, publishProgress,
      isUploadingZip, isDraggingZip, setIsDraggingZip, fileInputRef,
      uploadPublishFile, publishSkill,
      // categories
      categoryDraft, setCategoryDraft, saveCategory,
      // delete
      confirmDelete,
    }}>
      {children}
    </AdminSkillsContext.Provider>
  )
}
