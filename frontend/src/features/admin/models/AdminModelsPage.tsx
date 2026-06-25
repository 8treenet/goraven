import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { toast } from 'sonner'
import { useT, t as translate } from '@/i18n'
import {
  Search,
  X,
  Plus,
  Pencil,
  Trash2,
  AlertCircle,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Check,
  MoreHorizontal,
  Eye,
  EyeOff,
  Copy,
  Loader2,
  Star,
  Zap,
  Image,

} from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { adminModelsApi, adminProvidersApi } from '@/api'
import type { AdminModelItem, ProviderItem } from '@/api'
import { AVAILABLE_LOGOS } from '@/mocks/admin/models'

/* ============================================
   Types
   ============================================ */

interface RecommendedModel {
  id: string
  object: string
  ownedBy: string
}

interface ModelFormData {
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isCompress: number
  isVisual: number
  remark: string
}

interface ModelEditFormData {
  providerDisplayName: string
  displayName: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  contextLen: number
  extraFields: string
  isDefault: number
  isCompress: number
  isVisual: number
  status: number
  remark: string
}

type DrawerMode = 'add' | 'edit'
type DialogMode = 'delete' | null

type PageState = 'loading' | 'data' | 'empty' | 'error'
type TestState = 'idle' | 'testing' | 'success' | 'fail'


/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

/* ============================================
   Icon Picker (logo thumbnail grid)
   ============================================ */

function IconPicker({
  value,
  onChange,
}: {
  value: string
  onChange: (path: string) => void
}) {
  const t = useT()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex h-8 w-full items-center gap-2 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none transition-colors hover:border-border-outer focus:border-interactive focus:ring-3 focus:ring-interactive/50"
      >
        {value ? (
          <>
            <span className="inline-flex size-5 shrink-0 items-center justify-center rounded-sm bg-white/20">
              <img
                src={value}
                alt=""
                className="size-4 rounded object-contain"
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = 'none'
                }}
              />
            </span>
            <span className="truncate text-xs text-text-3">
              {value.split('/').pop()}
            </span>
          </>
        ) : (
          <span className="text-text-muted">{t('adminModels.selectIcon')}</span>
        )}
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-[320px] rounded-lg border border-border bg-bg-layer-1 p-2 shadow-pop">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-xs text-text-3">{t('adminModels.availableIcons')}</span>
            {value && (
              <button
                onClick={() => {
                  onChange('')
                  setOpen(false)
                }}
                className="text-xs text-text-muted transition-colors hover:text-text-2"
              >
                {t('common.clear')}
              </button>
            )}
          </div>
          <div className="grid max-h-48 grid-cols-5 gap-2 overflow-auto">
            {AVAILABLE_LOGOS.map((path) => {
              const selected = path === value
              return (
                <button
                  key={path}
                  onClick={() => {
                    onChange(path)
                    setOpen(false)
                  }}
                  className={cn(
                    'flex size-12 items-center justify-center rounded-md border bg-bg-base transition-colors hover:bg-bg-layer-2',
                    selected
                      ? 'border-interactive ring-2 ring-interactive/30'
                      : 'border-border-strong',
                  )}
                >
                  <span className="inline-flex size-8 items-center justify-center rounded-sm bg-white/20">
                    <img
                      src={path}
                      alt=""
                      className="size-6 rounded object-contain"
                    />
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

/* ============================================
   Status Toggle (reused from users page)
   ============================================ */

function StatusToggle({
  checked,
  onChange,
}: {
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label className="relative inline-flex cursor-pointer items-center">
      <input
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <div
        className={cn(
          'h-4 w-8 rounded-full border transition-colors',
          'peer-checked:border-highlight peer-checked:bg-highlight',
          'border-border-strong bg-bg-layer-2',
          'peer-hover:border-text-2',
        )}
      />
      <div
        className={cn(
          'absolute left-0.5 top-0.5 size-3 rounded-full bg-white shadow-sm transition-transform',
          'peer-checked:translate-x-4',
        )}
      />
    </label>
  )
}

/* ============================================
   Drawer
   ============================================ */

function Drawer({
  open,
  onClose,
  title,
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <div
        className={cn(
          'fixed inset-0 z-50',
          open ? 'visible' : 'invisible',
        )}
      >
        <div
          className={cn(
            'absolute inset-0 bg-black/60 transition-opacity duration-200',
            open ? 'opacity-100' : 'opacity-0',
          )}
          onClick={onClose}
        />
        <div
          className={cn(
            'absolute right-0 top-0 z-50 flex h-full w-[400px] flex-col border-l border-border bg-bg-layer-1 shadow-pop transition-transform duration-200 ease-out',
            open ? 'translate-x-0' : 'translate-x-full',
          )}
        >
          <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
            <h2 className="text-sm font-semibold text-text-1">{title}</h2>
            <button
              onClick={onClose}
              className="rounded-sm p-0.5 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
            >
              <X className="size-4" />
            </button>
          </div>
          <div className="flex-1 overflow-auto p-4">
            {children}
          </div>
        </div>
      </div>
    </Dialog>
  )
}

/* ============================================
   Searchable Combobox (model name)
   ============================================ */

function SearchableCombobox({
  value,
  onChange,
  providerId,
  canFetch,
  apiKey,
  baseUrl,
}: {
  value: string
  onChange: (v: string) => void
  providerId: string
  canFetch: boolean
  apiKey: string
  baseUrl: string
}) {
  const t = useT()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [recommends, setRecommends] = useState<RecommendedModel[]>([])
  const [fetching, setFetching] = useState(false)
  const [fetchError, setFetchError] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setRecommends([])
    setFetchError(false)
  }, [providerId])

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const handleFetch = () => {
    if (!canFetch || fetching) return
    setFetching(true)
    setFetchError(false)
    adminProvidersApi.getRecommendModels({ providerId, apiKey, baseUrl })
      .then((r) => {
        setRecommends(r.list)
        setFetching(false)
        setOpen(r.list.length > 0)
      })
      .catch(() => {
        setFetchError(true)
        setFetching(false)
      })
  }

  const filtered = useMemo(() => {
    if (!search.trim()) return recommends
    const q = search.toLowerCase()
    return recommends.filter((m) => m.id.toLowerCase().includes(q))
  }, [recommends, search])

  const showDropdown = open && (fetchError || recommends.length > 0)

  return (
    <div ref={containerRef} className="relative">
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Input
            value={value}
            onChange={(e) => {
              onChange(e.target.value)
              setSearch(e.target.value)
            }}
            onFocus={() => {
              if (recommends.length > 0) setOpen(true)
            }}
            placeholder="deepseek-v4-pro"
            className="h-8 pr-8 text-sm"
          />
          {value && (
            <button
              onClick={() => {
                onChange('')
                setSearch('')
              }}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-sm p-0.5 text-text-3 hover:text-text-1"
            >
              <X className="size-3" />
            </button>
          )}
          {showDropdown && (
            <div className="absolute left-0 top-full z-20 mt-1 max-h-48 w-full overflow-auto rounded-md border border-border bg-bg-layer-1 shadow-pop">
              {fetchError ? (
                <div className="flex items-center gap-2 px-3 py-2 text-xs text-red-500">
                  <AlertCircle className="size-3 shrink-0" />
                  {t('adminModels.fetchFailed')}
                </div>
              ) : filtered.length === 0 ? (
                <div className="px-3 py-2 text-xs text-text-muted">{t('adminModels.noMatchModel')}</div>
              ) : (
                filtered.map((m) => (
                  <button
                    key={m.id}
                    onClick={() => {
                      onChange(m.id)
                      setSearch('')
                      setOpen(false)
                    }}
                    className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
                  >
                    <span className="truncate">{m.id}</span>
                    <span className="shrink-0 text-xs text-text-3">{m.ownedBy}</span>
                  </button>
                ))
              )}
            </div>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleFetch}
          disabled={!canFetch || fetching}
        >
          {fetching ? (
            <>
              <Loader2 className="size-3 animate-spin" />
              {t('adminModels.fetching')}
            </>
          ) : (
            t('adminModels.getRecommended')
          )}
        </Button>
      </div>
    </div>
  )
}

/* ============================================
   API Key Field (single key with toggle + copy)
   ============================================ */

function ApiKeyField({
  value,
  onChange,
  disabled,
}: {
  value: string
  onChange: (v: string) => void
  disabled: boolean
}) {
  const t = useT()
  const [visible, setVisible] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="relative">
      <Input
        type={visible ? 'text' : 'password'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="sk-..."
        className="h-8 pr-16 font-mono text-xs"
        disabled={disabled}
      />
      <div className="absolute right-1 top-1/2 flex -translate-y-1/2 items-center gap-0.5">
        <button
          type="button"
          onClick={() => setVisible(!visible)}
          className="rounded p-0.5 text-text-3 transition-colors hover:text-text-2"
        >
          {visible ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
        </button>
        {value && (
          <button
            type="button"
            onClick={handleCopy}
            className={cn(
              'rounded p-0.5 transition-colors',
              copied ? 'text-highlight' : 'text-text-3 hover:text-text-2',
            )}
            title={t('common.copy')}
          >
            {copied ? <Check className="size-3" /> : <Copy className="size-3" />}
          </button>
        )}
      </div>
    </div>
  )
}

/* ============================================
   Connectivity Test Indicator
   ============================================ */

function TestIndicator({ state, errorMsg }: { state: TestState; errorMsg: string }) {
  const t = useT()
  if (state === 'idle') return null

  if (state === 'testing') {
    return (
      <div className="flex items-center gap-2 text-xs text-text-3">
        <Loader2 className="size-3 animate-spin" />
        {t('adminModels.testingConnection')}
      </div>
    )
  }

  if (state === 'success') {
    return (
      <div className="flex items-center gap-2 text-xs text-highlight">
        <Check className="size-3" />
        {t('adminModels.testPassed')}
      </div>
    )
  }

  return (
    <div className="flex items-start gap-2 text-xs text-red-500">
      <AlertCircle className="size-3 shrink-0 mt-px" />
      <span>{errorMsg || t('adminModels.testFailed')}</span>
    </div>
  )
}

/* ============================================
   Add Model Drawer
   ============================================ */

function AddModelDrawer({
  open,
  onClose,
  onSave,
  providers,
  duplicateModel,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: ModelFormData) => Promise<void>
  providers: ProviderItem[]
  duplicateModel: AdminModelItem | null
}) {
  const t = useT()
  const [form, setForm] = useState<ModelFormData>({
    providerDisplayName: '',
    displayName: '',
    providerId: '',
    modelName: '',
    icon: '',
    apiKey: '',
    baseUrl: '',
    proxyUrl: '',
    contextLen: 200,
    extraFields: '',
    isDefault: 0,
    isCompress: 0,
    isVisual: 0,
    remark: '',
  })
  const [testState, setTestState] = useState<TestState>('idle')
  const [testError, setTestError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    if (!open) {
      setForm({
        providerDisplayName: '',
        displayName: '',
        providerId: '',
        modelName: '',
        icon: '',
        apiKey: '',
        baseUrl: '',
        proxyUrl: '',
        contextLen: 200,
        extraFields: '',
        isDefault: 0,
        isCompress: 0,
        isVisual: 0,
        remark: '',
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(false)
      return
    }

    if (duplicateModel) {
      setForm({
        providerDisplayName: duplicateModel.providerDisplayName,
        displayName: duplicateModel.displayName + '-duplicate',
        providerId: duplicateModel.providerId,
        modelName: duplicateModel.modelName,
        icon: duplicateModel.icon,
        apiKey: '',
        baseUrl: duplicateModel.baseUrl,
        proxyUrl: duplicateModel.proxyUrl,
        contextLen: duplicateModel.contextLen,
        extraFields: '',
        isDefault: 0,
        isCompress: duplicateModel.isCompress,
        isVisual: duplicateModel.isVisual,
        remark: duplicateModel.remark,
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(true)
      adminModelsApi
        .getModelDetail(duplicateModel.aiModelId)
        .then((detail) => {
          setForm((f) => ({
            ...f,
            apiKey: detail.apiKey || '',
            extraFields: detail.extraFields || '',
          }))
        })
        .catch(() => {
          setTestError(t('adminModels.fetchFailed'))
          setTestState('fail')
        })
        .finally(() => {
          setDetailLoading(false)
        })
    } else {
      setForm({
        providerDisplayName: '',
        displayName: '',
        providerId: '',
        modelName: '',
        icon: '',
        apiKey: '',
        baseUrl: '',
        proxyUrl: '',
        contextLen: 200,
        extraFields: '',
        isDefault: 0,
        isCompress: 0,
        isVisual: 0,
        remark: '',
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(false)
    }
  }, [open, duplicateModel, t])

  const selectedProvider = providers.find((p) => p.providerId === form.providerId)
  const showApiKey = !selectedProvider || selectedProvider.requireApiKey
  const showBaseUrl = selectedProvider?.requireBaseUrl || false

  const canFetchModels =
    form.providerId.length > 0 &&
    (!showApiKey || form.apiKey.trim().length > 0) &&
    (!showBaseUrl || form.baseUrl.trim().length > 0)

  const canSave =
    form.providerDisplayName.trim().length > 0 &&
    form.providerId.length > 0 &&
    form.modelName.trim().length > 0 &&
    (!showBaseUrl || form.baseUrl.trim().length > 0) &&
    (!showApiKey || form.apiKey.trim().length > 0) &&
    !submitting &&
    !detailLoading

  const handleSubmit = () => {
    setSubmitting(true)
    setTestState('testing')
    onSave(form)
      .then(() => {
        setTestState('success')
        setTimeout(() => {
          setSubmitting(false)
        }, 600)
      })
      .catch((err) => {
        setTestState('fail')
        setTestError(err?.message || translate('adminModels.testFailed'))
        setSubmitting(false)
      })
  }

  return (
    <Drawer open={open} onClose={onClose} title={t('adminModels.addModel')}>
      <div className="flex flex-col gap-4">
        {/* provider select */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.provider')}</Label>
          <select
            value={form.providerId}
            onChange={(e) => {
              const pid = e.target.value
              const p = providers.find((pr) => pr.providerId === pid)
              setForm((f) => ({
                ...f,
                providerId: pid,
                modelName: '',
                providerDisplayName: p ? p.providerDisplayNameZh : f.providerDisplayName,
                icon: p && p.icon ? p.icon : f.icon,
                baseUrl: p ? p.defaultBaseUrl : '',
              }))
            }}
            className="h-8 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none focus:border-interactive focus:ring-3 focus:ring-interactive/50"
          >
            <option value="">{t('adminModels.selectProvider')}</option>
            {providers.map((p) => (
              <option key={p.providerId} value={p.providerId}>
                {p.providerDisplayNameZh}
              </option>
            ))}
          </select>
        </div>

        {/* display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.provider')}
            <span className="ml-0.5 text-red-500">*</span>
          </Label>
          <Input
            value={form.providerDisplayName}
            onChange={(e) => setForm((f) => ({ ...f, providerDisplayName: e.target.value }))}
            placeholder="DeepSeek"
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.nameHint')}</p>
        </div>

        {/* model display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.modelDisplayName')}
          </Label>
          <Input
            value={form.displayName}
            onChange={(e) => setForm((f) => ({ ...f, displayName: e.target.value }))}
            placeholder="deepseek-v4-pro"
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.displayNameHint')}</p>
        </div>

        {/* icon */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('common.icon')}</Label>
          <IconPicker
            value={form.icon}
            onChange={(path) => setForm((f) => ({ ...f, icon: path }))}
          />
        </div>

        {/* api keys */}
        {showApiKey && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">
              API Key
              {selectedProvider && <span className="ml-0.5 text-red-500">*</span>}
            </Label>
            <ApiKeyField
              value={form.apiKey}
              onChange={(key) => setForm((f) => ({ ...f, apiKey: key }))}
              disabled={false}
            />
          </div>
        )}

        {/* base url */}
        {showBaseUrl && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">
              Base URL
              <span className="ml-0.5 text-red-500">*</span>
            </Label>
            <Input
              value={form.baseUrl}
              onChange={(e) => setForm((f) => ({ ...f, baseUrl: e.target.value }))}
              placeholder="https://api.deepseek.com"
              className="h-8 text-sm font-mono"
            />
          </div>
        )}

        {/* model name searchable combobox */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('common.model')}
            <span className="ml-0.5 text-red-500">*</span>
          </Label>
          <SearchableCombobox
            value={form.modelName}
            onChange={(v) => setForm((f) => ({ ...f, modelName: v }))}
            providerId={form.providerId}
            canFetch={canFetchModels}
            apiKey={form.apiKey}
            baseUrl={form.baseUrl}
          />
        </div>

        {/* proxy url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.networkProxy')}</Label>
          <Input
            value={form.proxyUrl}
            onChange={(e) => setForm((f) => ({ ...f, proxyUrl: e.target.value }))}
            placeholder={`http://127.0.0.1:7890（${t('common.optional')}）`}
            className="h-8 text-sm font-mono"
          />
        </div>

        {/* context len */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.contextKB')}</Label>
          <Input
            type="number"
            value={String(form.contextLen)}
            onChange={(e) => setForm((f) => ({ ...f, contextLen: Number(e.target.value) || 0 }))}
            className="h-8 text-sm"
          />
        </div>

        {/* flags */}
        <div className="flex flex-col gap-2">
          <Label className="text-xs text-text-2">{t('adminModels.modelLabels')}</Label>
          <div className="flex flex-col gap-2">
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.defaultModel')}</span>
              <StatusToggle
                checked={form.isDefault === 1}
                onChange={(v) => setForm((f) => ({ ...f, isDefault: v ? 1 : 0 }))}
              />
            </label>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.compressionModel')}</span>
              <StatusToggle
                checked={form.isCompress === 1}
                onChange={(v) => setForm((f) => ({ ...f, isCompress: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.compressHint')}</p>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.multimodalModel')}</span>
              <StatusToggle
                checked={form.isVisual === 1}
                onChange={(v) => setForm((f) => ({ ...f, isVisual: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.multimodalHint')}</p>
          </div>
        </div>

        {/* extra fields - only for openai_compatible / claude_compatible */}
        {(form.providerId === 'openai_compatible' || form.providerId === 'claude_compatible') && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">extra_body</Label>
            <textarea
              value={form.extraFields}
              onChange={(e) => setForm((f) => ({ ...f, extraFields: e.target.value }))}
              placeholder='{"thinking":{"type":"enabled"}}'
              rows={3}
              className="min-h-[60px] w-full rounded-lg border border-border-strong bg-transparent px-2.5 py-1.5 font-mono text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-3 focus:ring-interactive/50"
            />
          </div>
        )}

        {/* remark */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.remark')}</Label>
          <Input
            value={form.remark}
            onChange={(e) => setForm((f) => ({ ...f, remark: e.target.value }))}
            placeholder={t('common.optional')}
            className="h-8 text-sm"
          />
        </div>

        {/* test indicator */}
        <TestIndicator state={testState} errorMsg={testError} />

        {/* buttons */}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={submitting}>
            {t('common.cancel')}
          </Button>
          <Button variant="default" size="default" disabled={!canSave} onClick={handleSubmit}>
            {submitting ? (
              <>
                <Loader2 className="size-3.5 animate-spin" />
                {t('common.testing')}
              </>
            ) : (
              t('adminModels.saveAndTest')
            )}
          </Button>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Edit Model Drawer
   ============================================ */

function EditModelDrawer({
  open,
  onClose,
  onSave,
  model,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: ModelEditFormData) => Promise<void>
  model: AdminModelItem | null
}) {
  const t = useT()
  const [form, setForm] = useState<ModelEditFormData>({
    providerDisplayName: '',
    displayName: '',
    modelName: '',
    icon: '',
    apiKey: '',
    baseUrl: '',
    proxyUrl: '',
    contextLen: 200,
    extraFields: '',
    isDefault: 0,
    isCompress: 0,
    isVisual: 0,
    status: 1,
    remark: '',
  })
  const [testState, setTestState] = useState<TestState>('idle')
  const [testError, setTestError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [connectionChanged, setConnectionChanged] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    if (model && open) {
      setForm({
        providerDisplayName: model.providerDisplayName,
        displayName: model.displayName,
        modelName: model.modelName,
        icon: model.icon,
        apiKey: '',
        baseUrl: model.baseUrl,
        proxyUrl: model.proxyUrl,
        contextLen: model.contextLen,
        extraFields: '',
        isDefault: model.isDefault,
        isCompress: model.isCompress,
        isVisual: model.isVisual,
        status: model.status,
        remark: model.remark,
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setConnectionChanged(false)
      setDetailLoading(true)
      adminModelsApi
        .getModelDetail(model.aiModelId)
        .then((detail) => {
          setForm((f) => ({
            ...f,
            apiKey: detail.apiKey || '',
            extraFields: detail.extraFields || '',
          }))
        })
        .catch(() => {
          setTestError(t('adminModels.fetchFailed'))
          setTestState('fail')
        })
        .finally(() => {
          setDetailLoading(false)
        })
    }
  }, [model, open, t])

  if (!model) return null

  const canSave =
    form.providerDisplayName.trim().length > 0 &&
    form.modelName.trim().length > 0 &&
    !submitting &&
    !detailLoading

  const handleSubmit = () => {
    setSubmitting(true)
    if (connectionChanged) {
      setTestState('testing')
    }
    const submitData: any = { ...form }
    if (!connectionChanged) {
      delete submitData.apiKey
    }
    onSave(submitData)
      .then(() => {
        if (connectionChanged) {
          setTestState('success')
        }
        setTimeout(() => {
          setSubmitting(false)
        }, 600)
      })
      .catch((err) => {
        if (connectionChanged) {
          setTestState('fail')
          setTestError(err?.message || translate('adminModels.testFailed'))
        }
        setSubmitting(false)
      })
  }

  const markConnectionChanged = () => setConnectionChanged(true)

  return (
    <Drawer open={open} onClose={onClose} title={t('adminModels.editModel')}>
      <div className="flex flex-col gap-4">
        {/* provider id (disabled) */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.provider')}</Label>
          <Input value={model.providerId} disabled className="h-8 text-sm opacity-50" />
          <p className="text-xs text-text-muted">{t('adminModels.providerReadonly')}</p>
        </div>

        {/* display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.provider')}
            <span className="ml-0.5 text-red-500">*</span>
          </Label>
          <Input
            value={form.providerDisplayName}
            onChange={(e) => setForm((f) => ({ ...f, providerDisplayName: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* model display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.modelDisplayName')}
          </Label>
          <Input
            value={form.displayName}
            onChange={(e) => setForm((f) => ({ ...f, displayName: e.target.value }))}
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.displayNameHint')}</p>
        </div>

        {/* icon */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('common.icon')}</Label>
          <IconPicker
            value={form.icon}
            onChange={(path) => setForm((f) => ({ ...f, icon: path }))}
          />
        </div>

        {/* api keys (with copy) */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">API Key</Label>
          <ApiKeyField
            value={form.apiKey}
            onChange={(key) => {
              setForm((f) => ({ ...f, apiKey: key }))
              markConnectionChanged()
            }}
            disabled={detailLoading}
          />
          {detailLoading && (
            <div className="flex items-center gap-2 text-xs text-text-3">
              <Loader2 className="size-3 animate-spin" />
              {t('adminModels.testingConnection')}
            </div>
          )}
        </div>

        {/* base url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">Base URL</Label>
          <Input
            value={form.baseUrl}
            onChange={(e) => {
              setForm((f) => ({ ...f, baseUrl: e.target.value }))
              markConnectionChanged()
            }}
            className="h-8 text-sm font-mono"
            placeholder="https://api.deepseek.com"
          />
        </div>

        {/* model name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('common.model')}
            <span className="ml-0.5 text-red-500">*</span>
          </Label>
          <Input
            value={form.modelName}
            onChange={(e) => setForm((f) => ({ ...f, modelName: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* proxy url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.networkProxy')}</Label>
          <Input
            value={form.proxyUrl}
            onChange={(e) => {
              setForm((f) => ({ ...f, proxyUrl: e.target.value }))
              markConnectionChanged()
            }}
            className="h-8 text-sm font-mono"
          />
        </div>

        {/* context len */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.contextKB')}</Label>
          <Input
            type="number"
            value={String(form.contextLen)}
            onChange={(e) => setForm((f) => ({ ...f, contextLen: Number(e.target.value) || 0 }))}
            className="h-8 text-sm"
          />
        </div>

        {/* flags */}
        <div className="flex flex-col gap-2">
          <Label className="text-xs text-text-2">{t('adminModels.modelLabels')}</Label>
          <div className="flex flex-col gap-2">
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.defaultModel')}</span>
              <StatusToggle
                checked={form.isDefault === 1}
                onChange={(v) => setForm((f) => ({ ...f, isDefault: v ? 1 : 0 }))}
              />
            </label>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.compressionModel')}</span>
              <StatusToggle
                checked={form.isCompress === 1}
                onChange={(v) => setForm((f) => ({ ...f, isCompress: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.compressHint')}</p>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.multimodalModel')}</span>
              <StatusToggle
                checked={form.isVisual === 1}
                onChange={(v) => setForm((f) => ({ ...f, isVisual: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.multimodalHint')}</p>
          </div>
        </div>

        {/* extra fields - only for openai_compatible / claude_compatible */}
        {(model.providerId === 'openai_compatible' || model.providerId === 'claude_compatible') && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">extra_body</Label>
            <textarea
              value={form.extraFields}
              onChange={(e) => setForm((f) => ({ ...f, extraFields: e.target.value }))}
              rows={3}
              className="min-h-[60px] w-full rounded-lg border border-border-strong bg-transparent px-2.5 py-1.5 font-mono text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-3 focus:ring-interactive/50"
            />
          </div>
        )}

        {/* remark */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.remark')}</Label>
          <Input
            value={form.remark}
            onChange={(e) => setForm((f) => ({ ...f, remark: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* test indicator */}
        <TestIndicator state={testState} errorMsg={testError} />

        {/* buttons */}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={submitting}>
            {t('common.cancel')}
          </Button>
          <Button variant="default" size="default" disabled={!canSave} onClick={handleSubmit}>
            {submitting ? (
              <>
                <Loader2 className="size-3.5 animate-spin" />
                {t('common.saving')}
              </>
            ) : (
              t('common.save')
            )}
          </Button>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

function DeleteModelDialog({
  open,
  onClose,
  onConfirm,
  modelName,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  modelName: string
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminModels.deleteModel')}</DialogTitle>
          <DialogDescription>
            {t('adminModels.deleteConfirm').replace('{name}', modelName)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose}>{t('common.cancel')}</Button>
          <Button variant="destructive" size="default" onClick={onConfirm}>
            {t('common.delete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Row Actions Dropdown
   ============================================ */

function RowActionsMenu({
  onEdit,
  onDelete,
  onDuplicate,
  onSetDefault,
  onSetCompress,
  onSetVisual,
  isDefault,
  isCompress,
  isVisual,
}: {
  onEdit: () => void
  onDelete: () => void
  onDuplicate: () => void
  onSetDefault: () => void
  onSetCompress: () => void
  onSetVisual: () => void
  isDefault: boolean
  isCompress: boolean
  isVisual: boolean
}) {
  const t = useT()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
      >
        <MoreHorizontal className="size-3.5" />
      </button>
      {open && (
        <div
          className={cn(
            'absolute right-0 top-full z-30 mt-1 w-36 overflow-hidden rounded-md border border-border bg-bg-layer-1 shadow-pop',
          )}
        >
          <button
            onClick={() => { onEdit(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Pencil className="size-3.5 text-text-3" />
            {t('common.edit')}
          </button>
          <button
            onClick={() => { onDuplicate(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
          >
            <Copy className="size-3.5 text-text-3" />
            {t('common.duplicate')}
          </button>
          {!isDefault && (
            <button
              onClick={() => { onSetDefault(); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
            >
              <Star className="size-3.5 text-text-3" />

              {t('common.default')}
            </button>
          )}
          {!isCompress && (
            <button
              onClick={() => { onSetCompress(); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
            >
              <Zap className="size-3.5 text-text-3" />

              {t('adminModels.compress')}
            </button>
          )}
          {!isVisual && (
            <button
              onClick={() => { onSetVisual(); setOpen(false) }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-text-1 transition-colors hover:bg-bg-hover"
            >
              <Image className="size-3.5 text-text-3" />

              {t('adminModels.multimodal')}
            </button>
          )}
          <div className="border-t border-border" />
          <button
            onClick={() => { onDelete(); setOpen(false) }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-red-500 transition-colors hover:bg-bg-hover"
          >
            <Trash2 className="size-3.5" />
            {t('common.delete')}
          </button>
        </div>
      )}
    </div>
  )
}

/* ============================================
   Table Row
   ============================================ */

function ModelIcon({ icon, name }: { icon: string; name: string }) {
  const [error, setError] = useState(false)

  if (!icon || error) {
    return (
      <div className="flex size-7 shrink-0 items-center justify-center rounded-sm bg-bg-layer-3 text-xs font-medium text-interactive">
        {name.charAt(0).toUpperCase()}
      </div>
    )
  }

  return (
    <span className="inline-flex size-7 shrink-0 items-center justify-center rounded-sm bg-white/20">
      <img
        src={icon}
        alt={name}
        className="size-5 rounded-sm object-contain"
        onError={() => setError(true)}
      />
    </span>
  )
}

function ModelRow({
  model,
  isEven,
  onEdit,
  onDelete,
  onDuplicate,
  onSetDefault,
  onSetCompress,
  onSetVisual,
}: {
  model: AdminModelItem
  isEven: boolean
  onEdit: () => void
  onDelete: () => void
  onDuplicate: () => void
  onSetDefault: () => void
  onSetCompress: () => void
  onSetVisual: () => void
}) {
  const t = useT()
  return (
    <tr className={cn(isEven && 'bg-bg-layer-1', 'transition-colors hover:bg-bg-hover')}>
      <td className="py-2.5 pl-4 pr-2">
        <div className="flex items-center gap-2">
          <ModelIcon icon={model.icon} name={model.providerDisplayName} />
          <span className="text-sm font-medium text-text-1">{model.providerDisplayName}</span>
        </div>
      </td>
      <td className="py-2.5 pr-2 text-sm font-medium text-text-1">{model.displayName}</td>
      <td className="py-2.5 pr-2 text-sm font-mono text-text-1">{model.modelName}</td>
      <td className="py-2.5 pr-2">
        <div className="flex items-center gap-1">
          {model.isDefault === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-highlight/15 text-highlight">
              <Star className="size-2.5" />
              {t('common.default')}
            </span>
          )}
          {model.isCompress === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-bg-layer-3 text-text-2">
              <Zap className="size-2.5" />
              {t('adminModels.compress')}
            </span>
          )}
          {model.isVisual === 1 && (
            <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs bg-interactive/15 text-interactive">
              <Image className="size-2.5" />
              {t('adminModels.multimodal')}
            </span>
          )}
          {model.isDefault === 0 && model.isCompress === 0 && model.isVisual === 0 && (
            <span className="text-xs text-text-muted">—</span>
          )}
        </div>
      </td>
      <td className="py-2.5 pr-2 text-sm tabular-nums text-text-2">{model.contextLen} KB</td>
      <td className="py-2.5 pr-2 text-sm text-text-3">{formatDate(model.updated)}</td>
      <td className="py-2.5 pr-4">
        <RowActionsMenu
          onEdit={onEdit}
          onDelete={onDelete}
          onDuplicate={onDuplicate}
          onSetDefault={onSetDefault}
          onSetCompress={onSetCompress}
          onSetVisual={onSetVisual}
          isDefault={model.isDefault === 1}
          isCompress={model.isCompress === 1}
          isVisual={model.isVisual === 1}
        />
      </td>
    </tr>
  )
}

/* ============================================
   Skeleton
   ============================================ */

function TableSkeleton() {
  const t = useT()
  return (
    <div className="flex-1 overflow-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border text-left text-xs text-text-3">
            <th className="pb-2 pl-4 pr-2 font-normal">{t('adminModels.provider')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.modelDisplayName')}</th>
            <th className="pb-2 pr-2 font-normal">{t('common.model')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.labels')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.context')}</th>
            <th className="pb-2 pr-2 font-normal">{t('adminModels.updatedAt')}</th>
            <th className="pb-2 pr-4 font-normal" />
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: 8 }).map((_, i) => (
            <tr key={i} className={cn(i % 2 === 0 && 'bg-bg-layer-1')}>
              <td className="py-2.5 pl-4 pr-2">
                <div className="flex items-center gap-2">
                  <div className="size-7 animate-pulse rounded-sm bg-bg-layer-3" />
                  <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
                </div>
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-2">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/* ============================================
   Empty State
   ============================================ */

function EmptyState({ hasFilter, onClearFilter, onAdd }: {
  hasFilter: boolean
  onClearFilter: () => void
  onAdd: () => void
}) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3">
      <div className="flex size-10 items-center justify-center rounded-full bg-bg-layer-2">
        <Zap className="size-5 text-text-muted" />
      </div>
      <div className="text-center">
        <p className="text-sm text-text-2">
          {hasFilter ? t('adminModels.noMatch') : t('adminModels.noModels')}
        </p>
      </div>
      {hasFilter ? (
        <button
          onClick={onClearFilter}
          className="text-xs text-interactive transition-colors hover:text-interactive-hover"
        >
          {t('adminUsers.clearFilter')}
        </button>
      ) : (
        <Button variant="default" size="sm" onClick={onAdd}>
          <Plus className="size-3.5" />
          {t('adminModels.addFirst')}
        </Button>
      )}
    </div>
  )
}

/* ============================================
   Error State
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{t('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">{t('adminModels.fetchFailedList')}</p>
      </div>
      <button
        onClick={onRetry}
        className="inline-flex items-center gap-1 rounded-md bg-bg-layer-2 px-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-layer-3"
      >
        <RefreshCw className="size-3" />
        {t('common.retry')}
      </button>
    </div>
  )
}

/* ============================================
   Pagination
   ============================================ */

function Pagination({
  page,
  totalPages,
  totalCount,
  onPageChange,
}: {
  page: number
  totalPages: number
  totalCount: number
  onPageChange: (p: number) => void
}) {
  const t = useT()
  const pages = useMemo(() => {
    const result: (number | '...')[] = []
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) result.push(i)
    } else {
      result.push(1)
      if (page > 3) result.push('...')
      for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
        result.push(i)
      }
      if (page < totalPages - 2) result.push('...')
      result.push(totalPages)
    }
    return result
  }, [page, totalPages])

  if (totalPages <= 1) return null

  return (
    <div className="flex items-center justify-between border-t border-border px-4 py-2">
      <span className="text-xs text-text-3">{t('common.total')} {totalCount} {t('adminModels.totalModels')}</span>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={page === 1}
          className="inline-flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronLeft className="size-3.5" />
        </button>
        {pages.map((p, i) =>
          p === '...' ? (
            <span key={`dots-${i}`} className="px-1 text-xs text-text-3">
              ...
            </span>
          ) : (
            <button
              key={p}
              onClick={() => onPageChange(p)}
              className={cn(
                'inline-flex size-7 items-center justify-center rounded-md text-xs tabular-nums transition-colors',
                p === page
                  ? 'bg-bg-layer-3 text-text-1 font-medium'
                  : 'text-text-3 hover:bg-bg-layer-2 hover:text-text-1',
              )}
            >
              {p}
            </button>
          ),
        )}
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={page === totalPages}
          className="inline-flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronRight className="size-3.5" />
        </button>
      </div>
    </div>
  )
}

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
    (modelId: number) => {
      adminModelsApi.setDefaultModel(modelId).then(() => {
        setModels((prev) =>
          prev.map((m) => ({
            ...m,
            isDefault: (m.aiModelId === modelId ? 1 : 0) as number,
          })),
        )
        toast.success(translate('adminModels.setDefault'))
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [],
  )

  const handleSetCompress = useCallback(
    (modelId: number) => {
      adminModelsApi.setCompressModel(modelId).then(() => {
        setModels((prev) =>
          prev.map((m) => ({
            ...m,
            isCompress: (m.aiModelId === modelId ? 1 : 0) as number,
          })),
        )
        toast.success(translate('adminModels.setCompress'))
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
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-3" />
            <input
              type="text"
              name="search"
              autoComplete="off"
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              placeholder={t('adminModels.searchPlaceholder')}
              className="h-7 w-44 rounded-lg border border-border-strong bg-transparent pl-7 pr-2 text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-2 focus:ring-interactive/30"
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
            className="h-7 rounded-lg border border-border-strong bg-transparent px-2 text-xs text-text-1 outline-none focus:border-interactive focus:ring-2 focus:ring-interactive/30"
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
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.contextLength')}</th>
                  <th className="pb-2 pr-2 font-normal">{t('adminModels.updatedAt')}</th>
                  <th className="pb-2 pr-4 font-normal" />
                </tr>
              </thead>
              <tbody>
                {models.map((model, i) => (
                  <ModelRow
                    key={model.aiModelId}
                    model={model}
                    isEven={i % 2 === 0}
                    onEdit={() => {
                      setEditTarget(model)
                      setDrawerMode('edit')
                    }}
                    onDuplicate={() => {
                      setDuplicateTarget(model)
                      setDrawerMode('add')
                    }}
                    onDelete={() => {
                      setDialogTarget(model)
                      setDialogMode('delete')
                    }}
                    onSetDefault={() => handleSetDefault(model.aiModelId)}
                    onSetCompress={() => handleSetCompress(model.aiModelId)}
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
    </div>
  )
}
