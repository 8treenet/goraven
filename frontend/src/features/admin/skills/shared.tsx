import { X, Archive, Search } from 'lucide-react'
import { renderIcon } from '@/components/common/Icon'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Dialog } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'
import { t as translate } from '@/i18n'

/* ============================================
   Types
   ============================================ */

export type TabKey = 'global' | 'market'
export type SourceType = 'clawhub' | 'custom_upload'
export type InstallStatus = 0 | 1 | 2 | 3
export type ClawHubSort = 'newest' | 'updated' | 'downloads' | 'stars' | 'installs' | 'trending'
export type DrawerKey = 'system-editor' | 'market-editor' | 'clawhub' | 'publish' | 'categories' | 'users' | null

export interface SystemSkill {
  skillId: number
  name: string
  description: string
  content: string
  status: 0 | 1
  updated: string
}

export interface SkillCategory {
  categoryId: number
  name: string
  icon: string
  isDefault: 0 | 1
  skillCount: number
  updated: string
}

export interface MarketSkill {
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

export interface InstalledSkill {
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

export interface ClawHubItem {
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

export interface DeleteTarget {
  type: 'system' | 'market' | 'category'
  id: number
  label: string
}

export interface SystemForm {
  skillId?: number
  content: string
}

export interface MarketForm {
  skillId: number
  icon: string
  categoryId: number
  sortOrder: number
  remark: string
}

/* ============================================
   Helpers
   ============================================ */

export function formatDate(iso: string): string {
  const date = new Date(iso)
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

export function getSourceLabel(source: SourceType): string {
  return source === 'clawhub' ? 'ClawHub' : translate('adminSkills.customUpload')
}

export function getInstallStatusLabel(status: InstallStatus): string {
  if (status === 0) return translate('common.notInstalled')
  if (status === 1) return translate('common.installing')
  if (status === 2) return translate('common.installed')
  return translate('common.failed')
}

export function parseSkillContent(content: string): { name: string; description: string; errors: string[] } {
  const errors: string[] = []
  const match = content.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/)

  if (!match) {
    return { name: '', description: '', errors: [translate('adminSkills.frontmatterError')] }
  }

  const frontmatter = match[1]
  const name = frontmatter.match(/^name:[ \t]+(.+)$/m)?.[1]?.trim().replace(/^["']|["']$/g, '') ?? ''
  const description = frontmatter.match(/^description:[ \t]+(.+)$/m)?.[1]?.trim().replace(/^["']|["']$/g, '') ?? ''

  if (!name) errors.push(translate('adminSkills.errMissingName'))
  if (!description) errors.push(translate('adminSkills.errMissingDescription'))
  if (name && !name.startsWith('goraven-')) errors.push(translate('adminSkills.errNamePrefix'))
  if (name && !/^goraven-[a-z0-9][a-z0-9-]*$/.test(name)) errors.push(translate('adminSkills.errNameFormat'))

  return { name, description, errors }
}

export const GLOBAL_TEMPLATE = `---
name: goraven-
description: fill skill description here
---

`

/* ============================================
   Small components
   ============================================ */

export function Drawer({
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

export function SelectField({ value, onChange, children, className }: {
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

export function SearchBox({ value, onChange, placeholder }: { value: string; onChange: (value: string) => void; placeholder: string }) {
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

export function SkillIcon({ icon }: { icon: string }) {
  return (
    <div className="inline-flex size-8 shrink-0 items-center justify-center rounded bg-bg-layer-3 text-interactive">
      {renderIcon(icon || 'bot', 'size-4')}
    </div>
  )
}

export function ScopeLabel({ children }: { children: React.ReactNode }) {
  return <span className="rounded bg-interactive/10 px-1.5 py-0.5 text-xs text-interactive">{children}</span>
}

export function StatusToggle({
  checked,
  onToggle,
  labelOn,
  labelOff,
}: {
  checked: boolean
  onToggle: (checked: boolean) => void
  labelOn: string
  labelOff: string
}) {
  return (
    <div className="flex items-center gap-1.5">
      <Switch size="sm" checked={checked} onCheckedChange={onToggle} />
      <span className="whitespace-nowrap text-xs text-text-2">
        {checked ? labelOn : labelOff}
      </span>
    </div>
  )
}

export function StatusText({ children, tone = 'neutral' }: { children: React.ReactNode; tone?: 'neutral' | 'good' | 'warn' }) {
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

export function EmptyState({ title, description, action }: { title: string; description: string; action?: React.ReactNode }) {
  return (
    <div className="flex min-h-[280px] flex-col items-center justify-center rounded-lg border border-border bg-bg-layer-1 px-6 text-center">
      <Archive className="mb-3 size-8 text-text-muted" />
      <h3 className="text-sm font-semibold text-text-1">{title}</h3>
      <p className="mt-1 max-w-sm text-sm text-text-3">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
