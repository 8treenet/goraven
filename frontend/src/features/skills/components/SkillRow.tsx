import { useState } from 'react'
import { useT } from '@/i18n'
import type { TranslationKey } from '@/i18n'
import { Ellipsis, Package, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { TONE_SOFT, TONE_TEXT, toneFromKey, type Tone } from '@/lib/tone'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Icon } from '@/components/common/Icon'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import type { UserSkill, MarketSkill, ShareSkill } from '@/api/types'

type T = (key: TranslationKey) => string

const ROW_CLASS =
  'group relative flex min-h-[72px] items-center gap-3 border-b border-border px-5 py-3 transition-colors hover:bg-bg-layer-2'
const ACTION_CLASS =
  'flex items-center gap-1 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100'

/** 已安装技能来源：仅在来源会变化的列表里展示徽章 */
function sourceLabel(source: string, t: T): string | null {
  switch (source) {
    case 'clawhub': return 'ClawHub'
    case 'custom_upload': return t('skills.adminUpload')
    case 'custom': return t('skills.userCreated')
    case 'market': return t('skills.marketplaceSource')
    case 'share': return t('skills.installedShareSource')
    default: return null
  }
}

function IconTile({ icon, tone }: { icon: string; tone: Tone }) {
  return (
    <span
      className={cn(
        'flex size-8 shrink-0 items-center justify-center rounded-lg',
        TONE_SOFT[tone],
        TONE_TEXT[tone],
      )}
    >
      <Icon name={icon} className="size-4" />
    </span>
  )
}

function RowMenu({ onDelete }: { onDelete: () => void }) {
  const [open, setOpen] = useState(false)
  const t = useT()

  return (
    <div className="relative">
      <button
        type="button"
        aria-label={t('sidebar.menu')}
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        className="flex size-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-3 hover:text-text-1"
      >
        <Ellipsis className="size-4" />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full z-20 mt-1 w-32 rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
            <button
              type="button"
              className="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-destructive transition-colors hover:bg-destructive/10"
              onClick={() => {
                setOpen(false)
                onDelete()
              }}
            >
              <Trash2 className="size-3.5" />
              {t('common.delete')}
            </button>
          </div>
        </>
      )}
    </div>
  )
}

function InstallAction({
  installed,
  onInstall,
  tip,
}: {
  installed: boolean
  onInstall: () => void
  tip: string
}) {
  const t = useT()
  if (installed) {
    return <span className="text-micro whitespace-nowrap text-text-muted">{t('common.installed')}</span>
  }
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button variant="secondary" size="sm" onClick={onInstall} className="h-7 px-2.5">
          <Package className="size-3.5" />
          <span className="hidden sm:inline">{tip}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent>{tip}</TooltipContent>
    </Tooltip>
  )
}

export function InstalledSkillRow({
  skill,
  index: _index,
  total: _total,
  onOpenDrawer,
  onDelete,
}: {
  skill: UserSkill
  index: number
  total: number
  onOpenDrawer: (id: number) => void
  onDelete: (skill: UserSkill) => void
}) {
  const t = useT()
  const source = sourceLabel(skill.source, t)
  const tone = toneFromKey(skill.categoryId)

  return (
    <div className={ROW_CLASS}>
      <button
        type="button"
        className="flex min-w-0 flex-1 items-center gap-3 text-left"
        onClick={() => onOpenDrawer(skill.userSkillId)}
      >
        <IconTile icon={skill.icon} tone={tone} />

        <span className="min-w-0 flex-1">
          <span className="block truncate text-body font-medium text-text-1">
            {skill.skillName}
          </span>
          <span className="mt-0.5 flex min-w-0 items-center gap-1.5 text-meta leading-[18px] text-text-3">
            {skill.categoryName && (
              <>
                <span className="shrink-0 text-text-3">{skill.categoryName}</span>
                <span className="shrink-0 text-text-muted">·</span>
              </>
            )}
            <span className="truncate">{skill.description}</span>
          </span>
        </span>
      </button>

      <div className="flex shrink-0 items-center gap-3">
        {source && (
          <Badge variant="neutral" className="hidden sm:inline-flex">
            {source}
          </Badge>
        )}
        <span className="hidden text-micro tabular-nums text-text-muted lg:inline">
          {skill.updated.slice(0, 10)}
        </span>
        <div className={ACTION_CLASS}>
          <RowMenu onDelete={() => onDelete(skill)} />
        </div>
      </div>
    </div>
  )
}

export function MarketSkillRow({
  skill,
  index: _index,
  total: _total,
  onOpenDrawer,
  onInstall,
}: {
  skill: MarketSkill
  index: number
  total: number
  onOpenDrawer: (id: number) => void
  onInstall: (skill: MarketSkill) => void
}) {
  const t = useT()
  const tone = toneFromKey(skill.categoryId)

  return (
    <div className={ROW_CLASS}>
      <button
        type="button"
        className="flex min-w-0 flex-1 items-center gap-3 text-left"
        onClick={() => onOpenDrawer(skill.skillId)}
      >
        <IconTile icon={skill.icon} tone={tone} />

        <span className="min-w-0 flex-1">
          <span className="block truncate text-body font-medium text-text-1">
            {skill.name}
          </span>
          <span className="mt-0.5 flex min-w-0 items-center gap-1.5 text-meta leading-[18px] text-text-3">
            {skill.categoryName && (
              <>
                <span className="shrink-0 text-text-3">{skill.categoryName}</span>
                <span className="shrink-0 text-text-muted">·</span>
              </>
            )}
            <span className="truncate">{skill.description}</span>
          </span>
        </span>
      </button>

      <div className="flex shrink-0 items-center gap-3">
        <span className="hidden w-24 whitespace-nowrap text-right text-micro tabular-nums text-text-muted sm:inline">
          {t('skills.installCountSuffix').replace('{count}', String(skill.installedCount))}
        </span>
        <div className={cn(ACTION_CLASS, 'w-28 justify-end')}>
          <InstallAction
            installed={skill.userInstalled}
            onInstall={() => onInstall(skill)}
            tip={t('skills.install')}
          />
        </div>
      </div>
    </div>
  )
}

export function ShareSkillRow({
  skill,
  index: _index,
  total: _total,
  userInstalled,
  onOpenDrawer,
  onInstall,
  onDelete,
}: {
  skill: ShareSkill
  index: number
  total: number
  userInstalled: boolean
  onOpenDrawer: (id: number) => void
  onInstall: (skill: ShareSkill) => void
  onDelete?: (skill: ShareSkill) => void
}) {
  const t = useT()
  const tone = toneFromKey(skill.categoryId)

  return (
    <div className={ROW_CLASS}>
      <button
        type="button"
        className="flex min-w-0 flex-1 items-center gap-3 text-left"
        onClick={() => onOpenDrawer(skill.shareId)}
      >
        <IconTile icon={skill.icon} tone={tone} />

        <span className="min-w-0 flex-1">
          <span className="block truncate text-body font-medium text-text-1">
            {skill.skillName}
          </span>
          <span className="mt-0.5 flex min-w-0 items-center gap-1.5 text-meta leading-[18px] text-text-3">
            <span className="shrink-0 text-text-3">{skill.ownerName}</span>
            {skill.categoryName && (
              <>
                <span className="shrink-0 text-text-muted">·</span>
                <span className="shrink-0 text-text-3">{skill.categoryName}</span>
              </>
            )}
            <span className="shrink-0 text-text-muted">·</span>
            <span className="truncate">{skill.description}</span>
          </span>
        </span>
      </button>

      <div className="flex shrink-0 items-center gap-3">
        <span className="hidden w-24 whitespace-nowrap text-right text-micro tabular-nums text-text-muted sm:inline">
          {t('skills.shareInstallCount').replace('{count}', String(skill.installCount))}
        </span>
        <div className={cn(ACTION_CLASS, 'w-28 justify-end')}>
          <InstallAction
            installed={userInstalled}
            onInstall={() => onInstall(skill)}
            tip={t('skills.installShareSkill')}
          />
          {skill.canDelete && onDelete && (
            <RowMenu onDelete={() => onDelete(skill)} />
          )}
        </div>
      </div>
    </div>
  )
}
