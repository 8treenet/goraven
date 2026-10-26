import { useState, useCallback, useEffect, useMemo } from 'react'
import { RefreshCw, AlertCircle, Save } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { SettingRow } from '@/components/common/SettingRow'
import { SettingGroup } from '@/components/common/SettingGroup'
import { adminSystemApi } from '@/api'

/* ============================================
   Types (aligned with GET /api/admin/settings)
   ============================================ */

interface SettingItem {
  key: string
  value: string
  valueType: string
  defaultValue: string
  displayName: string
  description: string
  inputType: string
  min?: number
  max?: number
  placeholder?: string
  displayOrder: number
}

interface SettingGroupData {
  name: string
  displayName: string
  displayOrder: number
  settings: SettingItem[]
}

/* ============================================
   Control renderer
   ============================================ */

interface ControlProps {
  item: SettingItem
  value: string
  onChange: (key: string, value: string) => void
  error?: string
}

function SettingControl({ item, value, onChange, error }: ControlProps) {
  const { inputType, min, max, key } = item

  switch (inputType) {
    case 'text':
      return (
        <div className="flex flex-col items-end gap-1">
          <Input
            className="w-56"
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
            placeholder={item.placeholder || ''}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    case 'number': {
      const numMin = min ?? 0
      const numMax = max ?? 999999
      const clamp = (v: string) => {
        if (v === '') return v
        const n = Number(v)
        if (isNaN(n)) return v
        if (n < numMin) return String(numMin)
        if (n > numMax) return String(numMax)
        return v
      }
      return (
        <div className="flex flex-col items-end gap-1">
          <Input
            className="w-28"
            type="number"
            min={numMin}
            max={numMax}
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
            onBlur={() => {
              const clamped = clamp(value)
              if (clamped !== value) onChange(key, clamped)
            }}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )
    }

    case 'slider': {
      const numMin = min ?? 0
      const numMax = max ?? 100
      const numValue = Number(value) || numMin
      return (
        <div className="flex items-center gap-3">
          <Slider
            className="w-40"
            value={[numValue]}
            min={numMin}
            max={numMax}
            step={1}
            onValueChange={([v]) => {
              if (v !== undefined) onChange(key, String(v))
            }}
          />
          <span className="text-sm tabular-nums text-text-1 w-10 text-right">
            {numValue}%
          </span>
        </div>
      )
    }

    case 'switch':
      return (
        <button
          type="button"
          onClick={() =>
            onChange(key, value === 'true' ? 'false' : 'true')
          }
          className={cn(
            'flex items-center gap-1.5 rounded-md px-1.5 py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover',
            value === 'true' && 'text-text-1',
          )}
        >
          <span
            className={cn(
              'relative inline-flex h-3.5 w-8 items-center rounded-full transition-colors',
              value === 'true' ? 'bg-highlight' : 'bg-bg-hover',
            )}
          >
            <span
              className={cn(
                'inline-block h-3 w-3 rounded-full transition-transform',
                value === 'true'
                  ? 'bg-highlight-fg translate-x-5'
                  : 'bg-text-muted',
              )}
            />
          </span>
          <span>{value === 'true' ? translate('adminSettings.on') : translate('adminSettings.off')}</span>
        </button>
      )

    case 'password':
      return (
        <div className="flex flex-col items-end gap-1">
          <Input
            className="w-56"
            type="password"
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
            placeholder={item.placeholder || ''}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    case 'textarea':
    case 'jsonEditor':
      return (
        <div className="flex flex-col items-end gap-1">
          <textarea
            spellCheck={false}
            className={cn(
              'w-64 h-20 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50',
              inputType === 'jsonEditor' && 'font-mono text-xs',
            )}
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
            placeholder={item.placeholder || ''}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    case 'date':
      return (
        <div className="flex flex-col items-end gap-1">
          <Input
            className="w-44"
            type="date"
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    case 'datetime':
      return (
        <div className="flex flex-col items-end gap-1">
          <Input
            className="w-56"
            type="datetime-local"
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
          />
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    case 'select':
      return (
        <div className="flex flex-col items-end gap-1">
          <select
            className="h-8 w-44 rounded-lg border border-input bg-transparent px-2.5 text-sm transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            value={value}
            onChange={(e) => onChange(key, e.target.value)}
          >
            <option value="option1">Option 1</option>
            <option value="option2">Option 2</option>
            <option value="option3">Option 3</option>
          </select>
          {error && (
            <span className="text-xs text-destructive">{error}</span>
          )}
        </div>
      )

    default:
      return (
        <Input
          className="w-56"
          value={value}
          onChange={(e) => onChange(key, e.target.value)}
        />
      )
  }
}

/* ============================================
   Skeleton
   ============================================ */

function Skeleton() {
  return (
    <div className="flex-1 overflow-auto p-4 flex flex-col gap-3">
      {[1, 2, 3, 4, 5].map((g) => (
        <div key={g} className="rounded-lg border border-border bg-bg-layer-1">
          <div className="px-5 pt-4 pb-3">
            <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
          </div>
          {[1, 2, 3].map((r) => (
            <div
              key={r}
              className="grid grid-cols-[1fr_auto] items-center gap-4 px-5 py-3.5"
            >
              <div className="min-w-0">
                <div className="h-4 w-28 animate-pulse rounded bg-bg-layer-2" />
                <div className="mt-1 h-3 w-52 animate-pulse rounded bg-bg-layer-2" />
              </div>
              <div className="h-8 w-28 animate-pulse rounded bg-bg-layer-2" />
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

/* ============================================
   Error
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
            <p className="text-sm text-text-2">{translate('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">
          {translate('adminSettings.fetchFailed')}
        </p>
      </div>
      <button
        onClick={onRetry}
        className="inline-flex items-center gap-1 rounded-md bg-bg-layer-2 px-3 py-1.5 text-xs text-text-1 transition-colors hover:bg-bg-layer-3"
      >
        <RefreshCw className="size-3" />
        {translate('common.retry')}
      </button>
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

type PageState = 'loading' | 'data' | 'error'

export function Component() {
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [groups, setGroups] = useState<SettingGroupData[]>([])
  const [values, setValues] = useState<Record<string, string>>({})
  const [originals, setOriginals] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)

  const dirtyKeys = useMemo(() => {
    const dirty: string[] = []
    for (const key of Object.keys(values)) {
      if (values[key] !== originals[key]) {
        dirty.push(key)
      }
    }
    return dirty
  }, [values, originals])

  const loadData = useCallback(() => {
    setState('loading')
    adminSystemApi.getSettings().then((data) => {
      const v: Record<string, string> = {}
      const groupsArr = (data as { groups: SettingGroupData[] }).groups
      for (const g of groupsArr) {
        for (const s of g.settings) {
          v[s.key] = String(s.value)
        }
      }
      setGroups(groupsArr)
      setValues({ ...v })
      setOriginals({ ...v })
      setState('data')
    }).catch(() => {
      setState('error')
    })
  }, [])

  useEffect(() => {
    const cleanup = loadData()
    return cleanup
  }, [loadData])

  const handleChange = useCallback((key: string, value: string) => {
    setValues((prev) => ({ ...prev, [key]: value }))
  }, [])

  const handleSave = useCallback(() => {
    if (dirtyKeys.length === 0 || saving) return

    setSaving(true)
    const updates = dirtyKeys.map((key) => ({ key, value: values[key] }))
    adminSystemApi.updateSettings(updates).then(() => {
      setOriginals({ ...values })
      setSaving(false)
      toast.success(translate('adminSettings.settingsUpdated'))
    }).catch((err: Error) => {
      setSaving(false)
      toast.error(err.message || translate('common.saveFailed'))
    })
  }, [dirtyKeys, saving, values])

  const handleRetry = useCallback(() => {
    loadData()
  }, [loadData])

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminSettings.title')}</h1>
        <Button
          size="sm"
          disabled={dirtyKeys.length === 0 || saving}
          onClick={handleSave}
          className={cn(dirtyKeys.length === 0 || saving ? '' : 'text-highlight hover:text-highlight')}
        >
          {saving ? (
            <>
              <RefreshCw className="size-3 animate-spin" />
              {t('common.saving')}
            </>
          ) : (
            <>
              <Save className="size-3" />
              {t('common.save')}
              {dirtyKeys.length > 0 && (
                <span className="ml-0.5 text-xs opacity-70">
                  ({dirtyKeys.length})
                </span>
              )}
            </>
          )}
        </Button>
      </div>

      {/* Content */}
      {state === 'loading' && <Skeleton />}

      {state === 'error' && <ErrorState onRetry={handleRetry} />}

      {state === 'data' && groups.length > 0 && (
        <div className="flex-1 overflow-auto p-4 flex flex-col gap-3">
          {groups.map((group) => (
            <SettingGroup key={group.name} title={group.displayName}>
              {group.settings.map((item, i) => (
                <SettingRow
                  key={item.key}
                  label={item.displayName}
                  description={item.description}
                  isDirty={values[item.key] !== originals[item.key]}
                  className={i % 2 === 1 ? 'bg-interactive/10 dark:bg-bg-layer-3/40' : ''}
                >
                  <SettingControl
                    item={item}
                    value={values[item.key] ?? item.value}
                    onChange={handleChange}
                  />
                </SettingRow>
              ))}
            </SettingGroup>
          ))}
        </div>
      )}
    </div>
  )
}
