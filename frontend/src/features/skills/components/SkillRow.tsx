import { useT } from '@/i18n'
import { Eye, Trash2, Package } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Icon } from '@/components/common/Icon'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import type { UserSkill, MarketSkill, ShareSkill } from '@/api/types'

function rowPadding(total: number) {
  return total <= 5 ? 'py-3.5' : 'py-2.5'
}

export function InstalledSkillRow({
  skill,
  index,
  total,
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
  const pad = rowPadding(total)

  return (
    <div
      key={skill.userSkillId}
      className={cn(
        'group flex items-start gap-3 px-4 transition-colors',
        pad,
        index % 2 === 0 ? 'bg-interactive/10 dark:bg-bg-layer-1 hover:bg-bg-hover' : 'hover:bg-bg-hover',
      )}
    >
      <div className="shrink-0 mt-0.5">
        <Icon name={skill.icon} className="size-4 text-interactive" />
      </div>

      <div className="flex-1 min-w-0 mt-0.5">
        <span className="block text-sm font-semibold text-text-1 truncate">
          {skill.skillName}
        </span>
        <div className="mt-0.5 flex items-center gap-1.5 min-w-0">
          <span className="shrink-0 text-xs text-text-muted">{skill.skillName}</span>
          <span className="shrink-0 text-text-muted">·</span>
          <span className="truncate text-xs text-text-3">{skill.description}</span>
        </div>
      </div>

      <div className="shrink-0 w-24 flex justify-end mt-0.5">
        {skill.categoryName && (
          <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-highlight bg-highlight/10 truncate max-w-full">
            {skill.categoryName}
          </span>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-1 w-20 justify-end mt-0.5 opacity-0 transition-opacity group-hover:opacity-100">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenDrawer(skill.userSkillId)}
            >
              <Eye className="size-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('skills.details')}</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="sm" onClick={() => onDelete(skill)}>
              <Trash2 className="size-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('common.delete')}</TooltipContent>
        </Tooltip>
      </div>
    </div>
  )
}

export function MarketSkillRow({
  skill,
  index,
  total,
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
  const pad = rowPadding(total)

  return (
    <div
      key={skill.skillId}
      className={cn(
        'group flex items-start gap-3 px-4 transition-colors',
        pad,
        index % 2 === 0 ? 'bg-interactive/10 dark:bg-bg-layer-1 hover:bg-bg-hover' : 'hover:bg-bg-hover',
      )}
    >
      <div className="shrink-0 mt-0.5">
        <Icon name={skill.icon} className="size-4 text-interactive" />
      </div>

      <div className="flex-1 min-w-0 mt-0.5">
        <span className="block text-sm font-semibold text-text-1 truncate">
          {skill.name}
        </span>
        <div className="mt-0.5 flex items-center gap-1.5 min-w-0">
          <span className="truncate text-xs text-text-3">{skill.description}</span>
        </div>
      </div>

      <div className="shrink-0 w-24 flex justify-end mt-0.5">
        {skill.categoryName && (
          <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-highlight bg-highlight/10 truncate max-w-full">
            {skill.categoryName}
          </span>
        )}
      </div>

      <div className="shrink-0 w-20 text-right mt-0.5">
        <span className="text-xs text-text-muted tabular-nums">
          {t('skills.installCountSuffix').replace('{count}', String(skill.installedCount))}
        </span>
      </div>

      <div className="flex shrink-0 items-center gap-0.5 w-20 justify-end mt-0.5 opacity-0 transition-opacity group-hover:opacity-100">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenDrawer(skill.skillId)}
            >
              <Eye className="size-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('skills.details')}</TooltipContent>
        </Tooltip>
        {skill.userInstalled ? (
          <span className="px-1.5 text-xs text-text-muted whitespace-nowrap">{t('common.installed')}</span>
        ) : (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={() => onInstall(skill)}>
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

export function ShareSkillRow({
  skill,
  index,
  total,
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
  const pad = rowPadding(total)

  return (
    <div
      key={skill.shareId}
      className={cn(
        'group flex items-start gap-3 px-4 transition-colors',
        pad,
        index % 2 === 0 ? 'bg-interactive/10 dark:bg-bg-layer-1 hover:bg-bg-hover' : 'hover:bg-bg-hover',
      )}
    >
      <div className="shrink-0 mt-0.5">
        <Icon name={skill.icon} className="size-4 text-interactive" />
      </div>

      <div className="flex-1 min-w-0 mt-0.5">
        <span className="block text-sm font-semibold text-text-1 truncate">
          {skill.skillName}
        </span>
        <div className="mt-0.5 flex items-center gap-1.5 min-w-0">
          <span className="shrink-0 text-xs text-text-muted">{skill.ownerName}</span>
          <span className="shrink-0 text-text-muted">·</span>
          <span className="truncate text-xs text-text-3">{skill.description}</span>
        </div>
      </div>

      <div className="shrink-0 w-24 flex justify-end mt-0.5">
        {skill.categoryName && (
          <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs text-highlight bg-highlight/10 truncate max-w-full">
            {skill.categoryName}
          </span>
        )}
      </div>

      <div className="shrink-0 w-20 text-right mt-0.5">
        <span className="text-xs text-text-muted tabular-nums">
          {t('skills.shareInstallCount').replace('{count}', String(skill.installCount))}
        </span>
      </div>

      <div className="flex shrink-0 items-center gap-0.5 w-20 justify-end mt-0.5 opacity-0 transition-opacity group-hover:opacity-100">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenDrawer(skill.shareId)}
            >
              <Eye className="size-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('skills.details')}</TooltipContent>
        </Tooltip>
        {userInstalled ? (
          <span className="px-1.5 text-xs text-text-muted whitespace-nowrap">{t('common.installed')}</span>
        ) : (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={() => onInstall(skill)}>
                <Package className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('skills.installShareSkill')}</TooltipContent>
          </Tooltip>
        )}
        {skill.canDelete && onDelete && (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={() => onDelete(skill)}>
                <Trash2 className="size-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('skills.cancelShare')}</TooltipContent>
          </Tooltip>
        )}
      </div>
    </div>
  )
}
