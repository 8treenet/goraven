import { useState, useCallback, useEffect, useMemo } from 'react'
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
  Copy,
  Check,
  Loader2,
  Circle,
  Globe,
  Terminal,
  Lightbulb,
  Plug,
} from 'lucide-react'
import { renderIcon } from '@/components/common/Icon'
import { IconPickerTrigger, DEFAULT_ICON, type IconName } from '@/components/common/IconPicker'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { adminMcpApi } from '@/api'
import type { AdminMcpItem } from '@/api'

/* ============================================
   Types
   ============================================ */

type TransportType = 'Stdio' | 'SSE' | 'StreamableHttp'
type StdioType = 'npx' | 'uvx'

interface McpFormData {
  name: string
  displayName: string
  icon: string
  description: string
  transport: TransportType
  httpUrl: string
  httpHeaders: { key: string; value: string }[]
  httpProxyUrl: string
  stdioType: StdioType
  stdioArgs: string[]
  stdioEnv: { key: string; value: string }[]
  remark: string
  status: 0 | 1
}

interface RecommendItem {
  name: string
  displayName: string
  icon: string
  description: string
  transport: TransportType
  httpUrl: string
  httpHeader: string
  stdioType: StdioType
  stdioArgs: string[]
  stdioEnv: string
  installed: boolean
  mcpId: number
  mcpStatus: 0 | 1
}

type DrawerMode = 'add' | 'edit'
type DialogMode = 'delete' | null
type PageState = 'loading' | 'data' | 'empty' | 'error'

/* ============================================
   Constants
   ============================================ */

const PAGE_SIZE = 10

const TRANSPORT_LABELS: Record<TransportType, string> = {
  Stdio: 'Stdio',
  SSE: 'SSE',
  StreamableHttp: 'HTTP',
}

const STDIO_TYPE_LABELS: Record<StdioType, string> = {
  npx: 'npx (Node.js)',
  uvx: 'uvx (Python)',
}

/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatRelative(iso: string | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return translate('files.justNow')
  if (mins < 60) return `${mins} ${translate('files.minutesAgo')}`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} ${translate('files.hoursAgo')}`
  return `${Math.floor(hours / 24)} ${translate('files.daysAgo')}`
}

function parseJsonArray(s: string): string[] {
  if (!s) return []
  try {
    const parsed = JSON.parse(s)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function parseJsonMap(s: string): Record<string, string> {
  if (!s) return {}
  try {
    const parsed = JSON.parse(s)
    return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

function parseEnvArray(s: string): Record<string, string> {
  if (!s) return {}
  try {
    const parsed = JSON.parse(s)
    if (!Array.isArray(parsed)) return {}
    const result: Record<string, string> = {}
    for (const item of parsed) {
      if (typeof item !== 'string') continue
      const idx = item.indexOf('=')
      if (idx > 0) {
        result[item.slice(0, idx)] = item.slice(idx + 1)
      }
    }
    return result
  } catch {
    return {}
  }
}

/* ============================================
   JSON Preview Builder
   ============================================ */

function buildJsonPreview(form: McpFormData): string {
  const cleanHeaders = form.httpHeaders.filter((h) => h.key.trim())
  const cleanEnv = form.stdioEnv.filter((e) => e.key.trim())

  const baseFields = (): Record<string, unknown> => ({
    name: form.name || '',
    description: form.description || '',
  })

  if (form.transport === 'Stdio') {
    const obj: Record<string, unknown> = {
      ...baseFields(),
      type: 'stdio',
      command: form.stdioType,
      args: form.stdioType === 'npx' ? ['-y', ...form.stdioArgs.filter(Boolean)] : form.stdioArgs.filter(Boolean),
    }
    if (cleanEnv.length > 0) {
      const envObj: Record<string, string> = {}
      cleanEnv.forEach((e) => { envObj[e.key] = e.value })
      obj.env = envObj
    }
    return JSON.stringify(obj, null, 2)
  }

  if (form.transport === 'SSE') {
    const obj: Record<string, unknown> = {
      ...baseFields(),
      type: 'sse',
      url: form.httpUrl || '',
    }
    if (cleanHeaders.length > 0) {
      const headersObj: Record<string, string> = {}
      cleanHeaders.forEach((h) => { headersObj[h.key] = h.value })
      obj.headers = headersObj
    }
    return JSON.stringify(obj, null, 2)
  }

  // StreamableHttp
  const obj: Record<string, unknown> = {
    ...baseFields(),
    type: 'streamable-http',
    url: form.httpUrl || '',
  }
  if (cleanHeaders.length > 0) {
    const headersObj: Record<string, string> = {}
    cleanHeaders.forEach((h) => { headersObj[h.key] = h.value })
    obj.headers = headersObj
  }
  return JSON.stringify(obj, null, 2)
}

function getEmptyForm(transport: TransportType = 'Stdio'): McpFormData {
  const isHttp = transport === 'SSE' || transport === 'StreamableHttp'
  return {
    name: '',
    displayName: '',
    icon: DEFAULT_ICON,
    description: '',
    transport,
    httpUrl: '',
    httpHeaders: isHttp ? [{ key: 'Authorization', value: 'Bearer {API_KEY}' }] : [],
    httpProxyUrl: '',
    stdioType: 'npx',
    stdioArgs: [''],
    stdioEnv: [],
    remark: '',
    status: 1,
  }
}

function itemToForm(item: AdminMcpItem): McpFormData {
  const headers = parseJsonMap(item.httpHeader)
  const envMap = parseEnvArray(item.stdioEnv)
  return {
    name: item.name,
    displayName: item.displayName,
    icon: item.icon,
    description: item.description,
    transport: item.transport as TransportType,
    httpUrl: item.httpUrl,
    httpHeaders: Object.entries(headers).map(([key, value]) => ({ key, value })),
    httpProxyUrl: item.httpProxyUrl,
    stdioType: (item.stdioType || 'npx') as StdioType,
    stdioArgs: parseJsonArray(item.stdioArgs),
    stdioEnv: Object.entries(envMap).map(([key, value]) => ({ key, value })),
    remark: item.remark,
    status: item.status as 0 | 1,
  }
}

function recommendToForm(r: RecommendItem): McpFormData {
  return {
    name: r.name,
    displayName: r.displayName,
    icon: r.icon,
    description: r.description,
    transport: r.transport,
    httpUrl: r.httpUrl,
    httpHeaders: Object.entries(parseJsonMap(r.httpHeader)).map(([key, value]) => ({ key, value })),
    httpProxyUrl: '',
    stdioType: r.stdioType || 'npx',
    stdioArgs: r.stdioArgs.length > 0 ? r.stdioArgs : [''],
    stdioEnv: Object.entries(parseEnvArray(r.stdioEnv)).map(([key, value]) => ({ key, value })),
    remark: '',
    status: 1,
  }
}

/* ============================================
   Status Toggle
   ============================================ */

function StatusToggle({
  checked,
  loading,
  onChange,
}: {
  checked: boolean
  loading: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label
      className={cn(
        'relative inline-flex cursor-pointer items-center',
        loading && 'cursor-not-allowed opacity-60',
      )}
    >
      <input
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        disabled={loading}
        onChange={(e) => onChange(e.target.checked)}
      />
      {loading && (
        <Loader2 className="absolute left-1/2 top-1/2 z-10 size-3 -translate-x-1/2 -translate-y-1/2 animate-spin text-text-2" />
      )}
      <div
        className={cn(
          'h-4 w-8 rounded-full border transition-colors',
          'peer-checked:border-highlight peer-checked:bg-highlight',
          'border-border-strong bg-bg-layer-2',
          !loading && 'peer-hover:border-text-2',
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
   Icon Renderer
   ============================================ */

function McpIcon({ icon }: { icon: string }) {
  return (
    <div className="inline-flex size-8 shrink-0 items-center justify-center rounded bg-bg-layer-3 text-interactive">
      {renderIcon(icon || DEFAULT_ICON, 'size-4')}
    </div>
  )
}

/* ============================================
   Drawer (wider, for left-right split)
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
        className={cn('fixed inset-0 z-50', open ? 'visible' : 'invisible')}
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
            'absolute right-0 top-0 z-50 flex h-full w-[680px] flex-col border-l border-border bg-bg-layer-1 shadow-pop transition-transform duration-200 ease-out',
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
          <div className="flex-1 overflow-hidden">
            {children}
          </div>
        </div>
      </div>
    </Dialog>
  )
}

/* ============================================
   MCP Form (shared by add / edit / recommend)
   ============================================ */

function McpForm({
  form,
  onChange,
  onSave,
  onCancel,
  saving,
  error,
  mode,
  nameReadonly,
}: {
  form: McpFormData
  onChange: (f: McpFormData) => void
  onSave: () => void
  onCancel: () => void
  saving: boolean
  error: string | null
  mode: 'add' | 'edit'
  nameReadonly: boolean
}) {
  const t = useT()
  const jsonPreview = useMemo(() => buildJsonPreview(form), [form])
  const [jsonCopied, setJsonCopied] = useState(false)

  const handleCopyJson = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(jsonPreview)
      setJsonCopied(true)
      setTimeout(() => setJsonCopied(false), 2000)
    } catch {
      // fallback
    }
  }, [jsonPreview])

  const updateField = useCallback(
    <K extends keyof McpFormData>(key: K, value: McpFormData[K]) => {
      onChange({ ...form, [key]: value })
    },
    [form, onChange],
  )

  const handleTransportChange = useCallback(
    (t: TransportType) => {
      const isHttp = t === 'SSE' || t === 'StreamableHttp'
      onChange({
        ...form,
        transport: t,
        httpUrl: '',
        httpHeaders: isHttp ? [{ key: 'Authorization', value: 'Bearer {API_KEY}' }] : [],
        httpProxyUrl: '',
        stdioArgs: [''],
        stdioEnv: [],
      })
    },
    [form, onChange],
  )

  const canSave =
    form.displayName.trim().length > 0 &&
    (mode === 'edit' || (form.name.trim().length > 0 && /^[A-Za-z][A-Za-z0-9-]{1,63}$/.test(form.name.trim()))) &&
    (form.transport === 'Stdio'
      ? form.stdioArgs.some((a) => a.trim().length > 0)
      : form.httpUrl.trim().length > 0)

  return (
    <div className="flex h-full">
      {/* Left: Form */}
      <div className="flex w-[55%] flex-col border-r border-border">
        <div className="flex-1 space-y-4 overflow-auto p-4">
          {/* Basic info */}
          <div className="flex flex-col gap-1.5">
               <Label className="text-xs text-text-2">{t('adminMcp.identifierName')}</Label>
            <Input
              value={form.name}
              onChange={(e) => {
                const v = e.target.value.replace(/[^A-Za-z0-9-]/g, '')
                updateField('name', v)
              }}
              placeholder={t('adminMcp.identifierHint')}
              className={cn('h-8 text-sm', nameReadonly && 'opacity-50')}
              disabled={nameReadonly}
              maxLength={64}
            />
            {form.name.length > 0 && !/^[A-Za-z][A-Za-z0-9-]{1,63}$/.test(form.name) && (
              <p className="text-xs text-red-500">{t('adminMcp.identifierValidation')}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
               <Label className="text-xs text-text-2">{t('adminMcp.displayName')}</Label>
            <Input
              value={form.displayName}
              onChange={(e) => updateField('displayName', e.target.value)}
              placeholder={t('adminMcp.displayNamePlaceholder')}
              className="h-8 text-sm"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">{t('common.icon')}</Label>
            <IconPickerTrigger
              value={form.icon}
              onChange={(name: IconName) => updateField('icon', name)}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">{t('common.description')}</Label>
            <textarea
              value={form.description}
              onChange={(e) => updateField('description', e.target.value)}
              placeholder={t('adminMcp.descriptionPlaceholder')}
              rows={3}
              className="w-full resize-none rounded-lg border border-border-strong bg-transparent px-2.5 py-2 text-sm text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-2 focus:ring-interactive/30"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">{t('adminMcp.transportType')}</Label>
            <select
              value={form.transport}
              onChange={(e) => handleTransportChange(e.target.value as TransportType)}
              disabled={mode === 'edit'}
              className={cn(
                'h-8 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none focus:border-interactive focus:ring-3 focus:ring-interactive/50',
                mode === 'edit' && 'opacity-50',
              )}
            >
              <option value="Stdio">Stdio</option>
              <option value="SSE">SSE</option>
              <option value="StreamableHttp">StreamableHttp</option>
            </select>
            {mode === 'edit' && (
              <p className="text-xs text-text-muted">{t('adminMcp.transportReadonly')}</p>
            )}
          </div>

          {/* Transport-specific fields */}
          {form.transport === 'Stdio' && (
            <>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('adminMcp.runnerType')}</Label>
                <select
                  value={form.stdioType}
                  onChange={(e) => updateField('stdioType', e.target.value as StdioType)}
                  className="h-8 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none focus:border-interactive focus:ring-3 focus:ring-interactive/50"
                >
                  <option value="npx">{STDIO_TYPE_LABELS.npx}</option>
                  <option value="uvx">{STDIO_TYPE_LABELS.uvx}</option>
                </select>
              </div>

              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('adminMcp.startupArgs')}</Label>
                <div className="space-y-1.5">
                  {form.stdioArgs.map((arg, i) => (
                    <div key={i} className="flex items-center gap-1">
                      <Input
                        value={arg}
                        onChange={(e) => {
                          const next = [...form.stdioArgs]
                          next[i] = e.target.value
                          updateField('stdioArgs', next)
                        }}
                        placeholder={`arg ${i + 1}`}
                        className="h-8 flex-1 text-sm"
                      />
                      <button
                        onClick={() => {
                          const next = form.stdioArgs.filter((_, j) => j !== i)
                          updateField('stdioArgs', next.length === 0 ? [''] : next)
                        }}
                        className="rounded p-1 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
                      >
                        <X className="size-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
                <button
                  onClick={() => updateField('stdioArgs', [...form.stdioArgs, ''])}
                  className="inline-flex items-center gap-1 self-start text-xs text-interactive transition-colors hover:text-interactive-hover"
                >
                  <Plus className="size-3" />
                  {t('adminMcp.addArg')}
                </button>
              </div>

              <div className="flex flex-col gap-1.5">
                 <Label className="text-xs text-text-2">{t('adminMcp.envVars')}</Label>
                 <div className="space-y-1.5">
                   {form.stdioEnv.map((env, i) => (
                     <div key={i} className="flex items-center gap-1">
                       <Input
                         value={env.key}
                         onChange={(e) => {
                           const next = [...form.stdioEnv]
                           next[i] = { ...next[i], key: e.target.value }
                           updateField('stdioEnv', next)
                         }}
                         placeholder="KEY"
                         className="h-8 w-[45%] text-sm font-mono text-xs"
                       />
                       <Input
                         value={env.value}
                         onChange={(e) => {
                           const next = [...form.stdioEnv]
                           next[i] = { ...next[i], value: e.target.value }
                           updateField('stdioEnv', next)
                         }}
                         placeholder="VALUE"
                         className="h-8 flex-1 text-sm"
                       />
                       <button
                         onClick={() => {
                           const next = form.stdioEnv.filter((_, j) => j !== i)
                           updateField('stdioEnv', next)
                         }}
                         className="rounded p-1 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
                       >
                         <X className="size-3.5" />
                       </button>
                     </div>
                   ))}
                   {form.stdioEnv.length === 0 && (
                     <p className="text-xs text-text-muted">{t('adminMcp.noEnvVars')}</p>
                   )}
                 </div>
                 <button
                   onClick={() => updateField('stdioEnv', [...form.stdioEnv, { key: '', value: '' }])}
                   className="inline-flex items-center gap-1 self-start text-xs text-interactive transition-colors hover:text-interactive-hover"
                 >
                   <Plus className="size-3" />
                   {t('adminMcp.addEnvVar')}
                 </button>
               </div>
            </>
          )}

          {(form.transport === 'SSE' || form.transport === 'StreamableHttp') && (
            <>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('adminMcp.serviceAddr')}</Label>
                <Input
                  value={form.httpUrl}
                  onChange={(e) => updateField('httpUrl', e.target.value)}
                  placeholder="https://mcp.example.com/sse"
                  className="h-8 text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                 <Label className="text-xs text-text-2">{t('adminMcp.requestHeaders')}</Label>
                 <div className="space-y-1.5">
                   {form.httpHeaders.map((header, i) => (
                     <div key={i} className="flex items-center gap-1">
                       <Input
                         value={header.key}
                         onChange={(e) => {
                           const next = [...form.httpHeaders]
                           next[i] = { ...next[i], key: e.target.value }
                           updateField('httpHeaders', next)
                         }}
                         placeholder="Header"
                         className="h-8 w-[45%] text-sm"
                       />
                       <Input
                         value={header.value}
                         onChange={(e) => {
                           const next = [...form.httpHeaders]
                           next[i] = { ...next[i], value: e.target.value }
                           updateField('httpHeaders', next)
                         }}
                         placeholder="Value"
                         className="h-8 flex-1 text-sm"
                       />
                       <button
                         onClick={() => {
                           const next = form.httpHeaders.filter((_, j) => j !== i)
                           updateField('httpHeaders', next)
                         }}
                         className="rounded p-1 text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
                       >
                         <X className="size-3.5" />
                       </button>
                     </div>
                   ))}
                   {form.httpHeaders.length === 0 && (
                     <p className="text-xs text-text-muted">{t('adminMcp.noHeaders')}</p>
                   )}
                 </div>
                 <button
                   onClick={() => updateField('httpHeaders', [...form.httpHeaders, { key: '', value: '' }])}
                   className="inline-flex items-center gap-1 self-start text-xs text-interactive transition-colors hover:text-interactive-hover"
                 >
                   <Plus className="size-3" />
                   {t('adminMcp.addHeader')}
                 </button>
               </div>

               <div className="flex flex-col gap-1.5">
                 <Label className="text-xs text-text-2">{t('adminModels.networkProxy')}</Label>
                 <Input
                   value={form.httpProxyUrl}
                   onChange={(e) => updateField('httpProxyUrl', e.target.value)}
                   placeholder="http://127.0.0.1:7890"
                   className="h-8 text-sm"
                 />
               </div>
            </>
          )}

          {/* Remark */}
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">{t('adminMcp.adminNotes')}</Label>
            <textarea
              value={form.remark}
              onChange={(e) => updateField('remark', e.target.value)}
              placeholder={t('common.optional')}
              rows={2}
              className="w-full resize-none rounded-lg border border-border-strong bg-transparent px-2.5 py-2 text-sm text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-2 focus:ring-interactive/30"
            />
          </div>

          {error && (
            <div className="flex items-start gap-2 rounded-md bg-red-500/10 p-3">
              <AlertCircle className="mt-0.5 size-3.5 shrink-0 text-red-500" />
              <p className="text-xs text-red-500">{error}</p>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 border-t border-border px-4 py-3">
          <Button variant="ghost" size="default" onClick={onCancel} className="text-highlight hover:text-highlight/80">
            {t('common.cancel')}
          </Button>
          <Button
            variant="default"
            size="default"
            disabled={!canSave || saving}
            onClick={onSave}
            className="text-highlight"
          >
            {saving ? (
              <>
                <Loader2 className="size-3.5 animate-spin" />
                {t('common.testing')}
              </>
            ) : (
              t('common.save')
            )}
          </Button>
        </div>
      </div>

      {/* Right: JSON Preview */}
      <div className="flex w-[45%] flex-col">
        <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
          <span className="text-xs text-text-3">{t('adminMcp.jsonPreview')}</span>
          <button
            onClick={handleCopyJson}
            className="inline-flex items-center gap-1 rounded-sm px-1.5 py-0.5 text-xs text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1"
          >
            {jsonCopied ? (
              <>
                <Check className="size-3" />
                {t('common.copied')}
              </>
            ) : (
              <>
                <Copy className="size-3" />
                {t('common.copy')}
              </>
            )}
          </button>
        </div>
        <div className="flex-1 overflow-auto p-4">
          <pre className="select-text whitespace-pre font-mono text-xs leading-relaxed text-text-2">
            {jsonPreview}
          </pre>
        </div>
      </div>
    </div>
  )
}

/* ============================================
   Add / Edit / Recommend Drawer
   ============================================ */

function McpDrawer({
  open,
  mode,
  initialForm,
  onClose,
  onSave,
  saving,
  error,
  nameReadonly,
}: {
  open: boolean
  mode: 'add' | 'edit'
  initialForm: McpFormData
  onClose: () => void
  onSave: (form: McpFormData) => void
  saving: boolean
  error: string | null
  nameReadonly: boolean
}) {
  const t = useT()
  const [form, setForm] = useState<McpFormData>(initialForm)

  useEffect(() => {
    if (open) {
      setForm(initialForm)
    }
  }, [open, initialForm])

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={mode === 'add' ? t('adminMcp.addMcp') : t('adminMcp.editMcp')}
    >
      <McpForm
        form={form}
        onChange={setForm}
        onSave={() => onSave(form)}
        onCancel={onClose}
        saving={saving}
        error={error}
        mode={mode}
        nameReadonly={nameReadonly}
      />
    </Drawer>
  )
}

/* ============================================
   Recommend Templates Dialog
   ============================================ */

function RecommendDialog({
  open,
  onClose,
  recommendItems,
  onInstall,
}: {
  open: boolean
  onClose: () => void
  recommendItems: RecommendItem[]
  onInstall: (item: RecommendItem) => void
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="flex max-h-[85vh] w-full max-w-4xl flex-col gap-0 p-0">
        <div className="shrink-0 border-b border-border px-6 pb-4 pt-6">
          <DialogTitle>{t('adminMcp.recommendedTemplates')}</DialogTitle>
          <DialogDescription className="mt-1">
            {t('adminMcp.templateHint')}
          </DialogDescription>
        </div>
        <div className="flex-1 overflow-auto px-6 py-4">
          <div className="grid grid-cols-3 gap-3">
            {recommendItems.map((item) => (
              <button
                key={item.name}
                disabled={item.installed}
                onClick={() => onInstall(item)}
                className={cn(
                  'group flex flex-col gap-2.5 rounded-lg border border-border bg-bg-base p-4 text-left transition-colors',
                  !item.installed && 'hover:border-border-strong hover:bg-bg-layer-1',
                  item.installed && 'cursor-default opacity-60',
                )}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2.5">
                    <McpIcon icon={item.icon} />
                    <div>
                      <h4 className="text-sm font-semibold text-text-1">
                        {item.displayName}
                      </h4>
                      <span className="text-xs text-text-3">{item.name}</span>
                    </div>
                  </div>
                  {!item.installed ? (
                    <span className="shrink-0 rounded-md bg-bg-layer-2 px-2.5 py-1 text-xs font-medium text-text-2 opacity-0 transition-opacity group-hover:opacity-100">
                      {t('common.install')}
                    </span>
                  ) : (
                    <span
                      className={cn(
                        'shrink-0 inline-flex items-center gap-1.5 rounded px-1.5 py-0.5 text-xs',
                        item.mcpStatus === 1
                          ? 'bg-highlight/10 text-highlight'
                          : 'bg-bg-layer-3 text-text-3',
                      )}
                    >
                      <Circle
                        className={cn(
                          'size-1.5',
                          item.mcpStatus === 1 ? 'fill-highlight' : 'fill-text-3',
                        )}
                      />
                      {item.mcpStatus === 1 ? t('common.installed') : t('common.disabled')}
                    </span>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <span className="inline-flex w-fit items-center gap-1 rounded bg-bg-layer-3 px-1.5 py-0.5 text-xs text-text-2">
                    {item.transport === 'Stdio' ? (
                      <Terminal className="size-2.5" />
                    ) : (
                      <Globe className="size-2.5" />
                    )}
          {TRANSPORT_LABELS[item.transport as TransportType]}
                  </span>
                  <p className="line-clamp-2 min-h-[2.5em] text-xs leading-relaxed text-text-3">
                    {item.description}
                  </p>
                </div>
              </button>
            ))}
          </div>
        </div>
        <div className="flex shrink-0 items-center justify-end gap-2 border-t border-border px-6 py-3">
          <span className="mr-auto text-xs text-text-3">
            {t('common.total')} {recommendItems.length}
          </span>
          <Button variant="ghost" size="default" onClick={onClose}>
            {t('common.close')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

function DeleteMcpDialog({
  open,
  onClose,
  onConfirm,
  displayName,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  displayName: string
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminMcp.deleteEndpoint')}</DialogTitle>
          <DialogDescription>
            {t('adminMcp.deleteConfirm').replace('{name}', displayName)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button variant="destructive" size="default" onClick={onConfirm}>
            {t('common.delete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Table Row
   ============================================ */

function McpRow({
  item,
  isEven,
  onEdit,
  onDelete,
  onToggleStatus,
  toggling,
  onToggleAlwaysOn,
  togglingAlwaysOn,
}: {
  item: AdminMcpItem
  isEven: boolean
  onEdit: () => void
  onDelete: () => void
  onToggleStatus: (v: boolean) => void
  toggling: boolean
  onToggleAlwaysOn: (v: boolean) => void
  togglingAlwaysOn: boolean
}) {
  const t = useT()
  return (
    <tr className={cn(isEven && 'bg-bg-layer-1', 'transition-colors hover:bg-bg-hover')}>
      <td className="py-2.5 pl-4 pr-2">
        <McpIcon icon={item.icon} />
      </td>
      <td className="py-2.5 pr-4">
        <div className="flex flex-col gap-0.5">
          <span className="text-sm font-medium text-text-1">{item.displayName}</span>
          <span className="text-xs text-text-3">{item.name}</span>
        </div>
      </td>
      <td className="py-2.5 pr-4">
        <span className="inline-flex items-center gap-1 rounded bg-bg-layer-3 px-1.5 py-0.5 text-xs text-text-2">
          {item.transport === 'Stdio' ? (
            <Terminal className="size-3" />
          ) : (
            <Globe className="size-3" />
          )}
          {TRANSPORT_LABELS[item.transport as TransportType]}
        </span>
      </td>
      <td className="py-2.5 pr-4">
        {item.healthLatency > 0 ? (
          <span
            className={cn(
              'text-sm tabular-nums',
              item.healthLatency < 50
                ? 'text-highlight'
                : item.healthLatency < 200
                  ? 'text-text-2'
                  : 'text-text-3',
            )}
          >
            {item.healthLatency}ms
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 text-sm text-text-3">
            <Circle className="size-1.5 fill-current" />
            {t('common.offline')}
          </span>
        )}
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-3">
        {formatRelative(item.healthCheckedAt)}
      </td>
      <td className="py-2.5 pr-4">
        <StatusToggle
          checked={item.status === 1}
          loading={toggling}
          onChange={onToggleStatus}
        />
      </td>
      <td className="py-2.5 pr-4">
        <StatusToggle
          checked={item.alwaysOn === 1}
          loading={togglingAlwaysOn}
          onChange={onToggleAlwaysOn}
        />
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-3">
        {formatDate(item.updated)}
      </td>
      <td className="py-2.5 pr-4">
        <div className="flex items-center gap-0.5">
          <button
            onClick={onEdit}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
            title={t('common.edit')}
          >
            <Pencil className="size-3.5" />
          </button>
          <button
            onClick={onDelete}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-red-500"
            title={t('common.delete')}
          >
            <Trash2 className="size-3.5" />
          </button>
        </div>
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
            <th className="pb-2 pl-4 pr-2 font-normal" />
            <th className="pb-2 pr-4 font-normal">{t('common.name')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminMcp.transport')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminMcp.latency')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminMcp.checkTime')}</th>
            <th className="pb-2 pr-4 font-normal">{t('common.status')}</th>
            <th className="pb-2 pr-4 font-normal">{t('adminMcp.alwaysOn')}</th>
            <th className="pb-2 pr-4 font-normal">{t('common.updated')}</th>
            <th className="pb-2 pr-4 font-normal" />
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: 8 }).map((_, i) => (
            <tr key={i} className={cn(i % 2 === 0 && 'bg-bg-layer-1')}>
              <td className="py-2.5 pl-4 pr-2">
                <div className="size-8 animate-pulse rounded bg-bg-layer-3" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
                <div className="mt-1 h-2.5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-12 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-4 w-8 animate-pulse rounded-full bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-4 w-8 animate-pulse rounded-full bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-12 animate-pulse rounded bg-bg-layer-2" />
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

function EmptyState({ hasFilter, onClearFilter }: { hasFilter: boolean; onClearFilter: () => void }) {
  const t = useT()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3">
      <Plug className="size-10 text-text-muted" />
      <div className="text-center">
        <p className="text-sm text-text-2">
          {hasFilter ? t('adminMcp.noMatch') : t('adminMcp.noEndpoints')}
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
        <p className="text-xs text-text-3">{t('adminMcp.startHint')}</p>
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
        <p className="mt-1 text-xs text-text-3">{t('adminMcp.fetchFailed')}</p>
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
      <span className="text-xs text-text-3">{t('common.total')} {totalCount} {t('adminMcp.totalEndpoints')}</span>
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
  const [mcps, setMcps] = useState<AdminMcpItem[]>([])
  const [search, setSearch] = useState('')
  const [transportFilter, setTransportFilter] = useState<'all' | TransportType>('all')
  const [page, setPage] = useState(1)

  const [drawerMode, setDrawerMode] = useState<DrawerMode | null>(null)
  const [drawerInitialForm, setDrawerInitialForm] = useState<McpFormData>(getEmptyForm())
  const [drawerNameReadonly, setDrawerNameReadonly] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  const [dialogMode, setDialogMode] = useState<DialogMode>(null)
  const [dialogTarget, setDialogTarget] = useState<AdminMcpItem | null>(null)

  const [recommendOpen, setRecommendOpen] = useState(false)
  const [recommendItems, setRecommendItems] = useState<RecommendItem[]>([])

  const [togglingIds, setTogglingIds] = useState<Set<number>>(new Set())
  const [togglingAlwaysOnIds, setTogglingAlwaysOnIds] = useState<Set<number>>(new Set())

  const loadData = useCallback(async () => {
    setState('loading')
    try {
      const result = await adminMcpApi.getMCPs({ pageSize: 1000 })
      setMcps(result.list)
      setState(result.list.length > 0 ? 'data' : 'empty')
    } catch {
      setState('error')
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const filteredMcps = useMemo(() => {
    let result = mcps
    if (search.trim()) {
      const q = search.trim().toLowerCase()
      result = result.filter(
        (m) => m.name.toLowerCase().includes(q) || m.displayName.toLowerCase().includes(q),
      )
    }
    if (transportFilter !== 'all') {
      result = result.filter((m) => m.transport === transportFilter)
    }
    return result
  }, [mcps, search, transportFilter])

  const totalPages = Math.max(1, Math.ceil(mcps.length / 12) + (mcps.length === 12 ? 1 : 0))
  const safePage = Math.min(page, totalPages)

  useEffect(() => {
    if (safePage !== page) setPage(safePage)
  }, [safePage, page])

  const pagedMcps = useMemo(() => {
    const start = (safePage - 1) * PAGE_SIZE
    return filteredMcps.slice(start, start + PAGE_SIZE)
  }, [filteredMcps, safePage])

  useEffect(() => {
    setPage(1)
  }, [search, transportFilter])

  const handleClearFilter = useCallback(() => {
    setSearch('')
    setTransportFilter('all')
  }, [])

  // Open add drawer
  const handleOpenAdd = useCallback(() => {
    setDrawerMode('add')
    setDrawerInitialForm(getEmptyForm())
    setDrawerNameReadonly(false)
    setSaveError(null)
  }, [])

  // Open edit drawer
  const handleEdit = useCallback((item: AdminMcpItem) => {
    setDrawerMode('edit')
    setDrawerInitialForm(itemToForm(item))
    setDrawerNameReadonly(true)
    setSaveError(null)

    adminMcpApi.getMCPDetail(item.mcpId).then((detail) => {
      setDrawerInitialForm(itemToForm(detail))
    }).catch(() => {})
  }, [])

  // Save (add or edit)
  const handleSave = useCallback(
    async (formData: McpFormData) => {
      setSaving(true)
      setSaveError(null)

      try {
        if (drawerMode === 'add') {
          await adminMcpApi.createMCP({
            name: formData.name.trim(),
            displayName: formData.displayName.trim() || formData.name.trim(),
            icon: formData.icon.trim(),
            description: formData.description.trim(),
            transport: formData.transport,
            httpUrl: formData.httpUrl,
            httpHeader: formData.httpHeaders.length > 0
              ? Object.fromEntries(formData.httpHeaders.filter((h) => h.key).map((h) => [h.key, h.value]))
              : undefined,
            httpProxyUrl: formData.httpProxyUrl,
            stdioType: formData.stdioType,
            stdioEnv: formData.stdioEnv.length > 0
              ? Object.fromEntries(formData.stdioEnv.filter((e) => e.key).map((e) => [e.key, e.value]))
              : undefined,
            stdioArgs: formData.stdioArgs.filter(Boolean),
            remark: formData.remark.trim(),
          })
          setDrawerMode(null)
          toast.success(translate('adminMcp.created'))
        } else if (drawerMode === 'edit') {
          const target = mcps.find(
            (m) => m.name === drawerInitialForm.name && m.transport === drawerInitialForm.transport,
          )
          if (target) {
            await adminMcpApi.updateMCP(target.mcpId, {
              displayName: formData.displayName.trim() || target.name,
              icon: formData.icon.trim(),
              description: formData.description.trim(),
              httpUrl: formData.httpUrl,
              httpHeader: formData.httpHeaders.length > 0
                ? Object.fromEntries(formData.httpHeaders.filter((h) => h.key).map((h) => [h.key, h.value]))
                : undefined,
              httpProxyUrl: formData.httpProxyUrl,
              stdioType: formData.stdioType,
              stdioEnv: formData.stdioEnv.length > 0
                ? Object.fromEntries(formData.stdioEnv.filter((e) => e.key).map((e) => [e.key, e.value]))
                : undefined,
              stdioArgs: formData.stdioArgs.filter(Boolean),
              remark: formData.remark.trim(),
              status: formData.status,
            })
          }
          setDrawerMode(null)
          toast.success(translate('adminMcp.updated'))
        }
        loadData()
      } catch (err) {
        setSaving(false)
        setSaveError(err instanceof Error ? err.message : translate('common.saveFailed'))
        return
      }

      setSaving(false)
    },
    [drawerMode, drawerInitialForm, mcps, loadData],
  )

  // Delete
  const handleDeleteConfirm = useCallback(async () => {
    if (!dialogTarget) return
    try {
      await adminMcpApi.deleteMCP(dialogTarget.mcpId)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('common.deleteFailed'))
      return
    }
    setDialogMode(null)
    setDialogTarget(null)
    toast.success(translate('adminMcp.deleted'))
    loadData()
  }, [dialogTarget, loadData])

  // Toggle status
  const handleToggleStatus = useCallback(
    async (item: AdminMcpItem, enabled: boolean) => {
      setTogglingIds((prev) => new Set(prev).add(item.mcpId))
      try {
        await adminMcpApi.updateMCPStatus(item.mcpId, enabled ? 1 : 0)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('adminMcp.statusToggleFailed'))
      }
      setTogglingIds((prev) => {
        const next = new Set(prev)
        next.delete(item.mcpId)
        return next
      })
      loadData()
    },
    [loadData],
  )

  // Toggle always-on
  const handleToggleAlwaysOn = useCallback(
    async (item: AdminMcpItem, enabled: boolean) => {
      setTogglingAlwaysOnIds((prev) => new Set(prev).add(item.mcpId))
      try {
        await adminMcpApi.toggleMCPAlwaysOn(item.mcpId, enabled ? 1 : 0)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('adminMcp.alwaysOnToggleFailed'))
      }
      setTogglingAlwaysOnIds((prev) => {
        const next = new Set(prev)
        next.delete(item.mcpId)
        return next
      })
      loadData()
    },
    [loadData],
  )

  // Recommend
  const handleOpenRecommend = useCallback(async () => {
    try {
      const result = await adminMcpApi.getRecommendMCPs()
      setRecommendItems(
        result.list.map((item) => ({
          ...item,
          transport: item.transport as TransportType,
          stdioType: (item.stdioType || 'npx') as StdioType,
          stdioArgs: parseJsonArray(item.stdioArgs),
          stdioEnv: item.stdioEnv ?? '',
          httpHeader: item.httpHeader ?? '',
          mcpId: item.mcpId ?? 0,
          mcpStatus: (item.mcpStatus ?? 0) as 0 | 1,
        })),
      )
    } catch (err) {
      toast.error(err instanceof Error ? err.message : translate('adminMcp.fetchTemplatesFailed'))
    }
    setRecommendOpen(true)
  }, [])

  const handleInstallRecommend = useCallback((item: RecommendItem) => {
    setRecommendOpen(false)
    setDrawerMode('add')
    setDrawerInitialForm(recommendToForm(item))
    setDrawerNameReadonly(true)
    setSaveError(null)
  }, [])

  const handleCloseDrawer = useCallback(() => {
    setDrawerMode(null)
    setSaveError(null)
  }, [])

  const hasFilter = search.trim().length > 0 || transportFilter !== 'all'

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminMcp.title')}</h1>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-3" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('adminMcp.searchPlaceholder')}
              className="h-7 w-48 rounded-lg border border-border-strong bg-transparent pl-7 pr-2 text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-2 focus:ring-interactive/30"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded-sm p-0.5 text-text-3 hover:text-text-1"
              >
                <X className="size-3" />
              </button>
            )}
          </div>
          <select
            value={transportFilter}
            onChange={(e) => setTransportFilter(e.target.value as 'all' | TransportType)}
            className="h-7 rounded-lg border border-border-strong bg-transparent px-2 text-xs text-text-1 outline-none focus:border-interactive focus:ring-2 focus:ring-interactive/30"
          >
            <option value="all">{t('adminMcp.allTypes')}</option>
            <option value="Stdio">Stdio</option>
            <option value="SSE">SSE</option>
            <option value="StreamableHttp">HTTP</option>
          </select>
          <Button
            variant="outline"
            size="sm"
            onClick={handleOpenRecommend}
            className="text-highlight hover:text-highlight"
          >
            <Lightbulb className="size-3.5" />
            {t('adminMcp.recommendedTemplates')}
          </Button>
          <Button variant="default" size="sm" onClick={handleOpenAdd} className="text-highlight hover:text-highlight">
            <Plus className="size-3.5" />
            {t('adminMcp.addMcp')}
          </Button>
        </div>
      </div>

      {/* Content */}
      {state === 'loading' && <TableSkeleton />}

      {state === 'error' && <ErrorState onRetry={loadData} />}

      {state === 'empty' && <EmptyState hasFilter={false} onClearFilter={handleClearFilter} />}

      {state === 'data' && filteredMcps.length === 0 && (
        <EmptyState hasFilter={hasFilter} onClearFilter={handleClearFilter} />
      )}

      {state === 'data' && filteredMcps.length > 0 && (
        <>
          <div className="flex-1 overflow-auto">
            <table className="w-full">
              <thead>
                <tr className="sticky top-0 z-10 border-b border-border bg-bg-base text-left text-xs text-text-3">
                  <th className="pb-2 pl-4 pr-2 font-normal" />
                  <th className="pb-2 pr-4 font-normal">{t('common.name')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminMcp.transport')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminMcp.latency')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminMcp.checkTime')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('common.status')}</th>
                  <th className="pb-2 pr-4 font-normal">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="cursor-help border-b border-dashed border-text-muted">
                          {t('adminMcp.alwaysOn')}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent>{t('adminMcp.alwaysOnTip')}</TooltipContent>
                    </Tooltip>
                  </th>
                  <th className="pb-2 pr-4 font-normal">{t('common.updated')}</th>
                  <th className="pb-2 pr-4 font-normal" />
                </tr>
              </thead>
              <tbody>
                {pagedMcps.map((item, i) => (
                  <McpRow
                    key={item.mcpId}
                    item={item}
                    isEven={i % 2 === 0}
                    onEdit={() => handleEdit(item)}
                    onDelete={() => {
                      setDialogTarget(item)
                      setDialogMode('delete')
                    }}
                    onToggleStatus={(v) => handleToggleStatus(item, v)}
                    toggling={togglingIds.has(item.mcpId)}
                    onToggleAlwaysOn={(v) => handleToggleAlwaysOn(item, v)}
                    togglingAlwaysOn={togglingAlwaysOnIds.has(item.mcpId)}
                  />
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            page={safePage}
            totalPages={totalPages}
            totalCount={filteredMcps.length}
            onPageChange={setPage}
          />
        </>
      )}

      {/* Drawer */}
      {drawerMode && (
        <McpDrawer
          open
          mode={drawerMode}
          initialForm={drawerInitialForm}
          onClose={handleCloseDrawer}
          onSave={handleSave}
          saving={saving}
          error={saveError}
          nameReadonly={drawerNameReadonly}
        />
      )}

      {/* Dialogs */}
      <DeleteMcpDialog
        open={dialogMode === 'delete'}
        onClose={() => {
          setDialogMode(null)
          setDialogTarget(null)
        }}
        onConfirm={handleDeleteConfirm}
        displayName={dialogTarget?.displayName ?? ''}
      />

      <RecommendDialog
        open={recommendOpen}
        onClose={() => setRecommendOpen(false)}
        recommendItems={recommendItems}
        onInstall={handleInstallRecommend}
      />
    </div>
  )
}
