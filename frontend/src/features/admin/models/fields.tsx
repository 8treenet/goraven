import { useState, useEffect, useRef, useMemo } from 'react'
import { useT } from '@/i18n'
import {
  X,
  AlertCircle,
  Loader2,
  Check,
  Copy,
  Eye,
  EyeOff,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { adminProvidersApi } from '@/api'
import { AVAILABLE_LOGOS } from '@/mocks/admin/models'
import type { RecommendedModel, TestState } from './shared'

/* ============================================
   Icon Picker (logo thumbnail grid)
   ============================================ */

export function IconPicker({
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
        className="flex h-8 w-full items-center gap-2 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none transition-colors hover:border-border-outer focus:border-ring focus:ring-3 focus:ring-ring/50"
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
   Searchable Combobox (model name)
   ============================================ */

export function SearchableCombobox({
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
                <div className="flex items-center gap-2 px-3 py-2 text-xs text-destructive">
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

export function ApiKeyField({
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

export function TestIndicator({ state, errorMsg }: { state: TestState; errorMsg: string }) {
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
    <div className="flex items-start gap-2 text-xs text-destructive">
      <AlertCircle className="size-3 shrink-0 mt-px" />
      <span>{errorMsg || t('adminModels.testFailed')}</span>
    </div>
  )
}
