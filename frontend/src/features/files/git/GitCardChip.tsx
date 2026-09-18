import type { ReactNode } from 'react'
import { useT } from '@/i18n'
import { Folder, FolderGit2, GitBranch, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { GIT_CLONE_STATE } from './git-helpers'

interface GitCardChipProps {
  gitEnabled: boolean
  /** 已配置 origin 远程仓库 */
  gitHasRemote: boolean
  cloneState: number
  /** 提供则徽标可点（打开 Git 弹框）；未启用 Git 时不可点 */
  onClick?: () => void
}

const BADGE =
  'inline-flex size-8 shrink-0 items-center justify-center rounded-md border transition-colors'

export function GitCardChip({ gitEnabled, gitHasRemote, cloneState, onClick }: GitCardChipProps) {
  const t = useT()
  const interactive = gitEnabled && !!onClick

  const badge = (className: string, children: ReactNode) =>
    interactive ? (
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation()
          onClick?.()
        }}
        className={cn(
          BADGE,
          className,
          'cursor-pointer hover:opacity-80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40',
        )}
      >
        {children}
      </button>
    ) : (
      <span className={cn(BADGE, className)}>{children}</span>
    )

  if (cloneState === GIT_CLONE_STATE.RUNNING) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          {badge('border-interactive/35 bg-interactive/10 text-interactive', <Loader2 className="size-4 animate-spin" />)}
        </TooltipTrigger>
        <TooltipContent>{t('files.gitTipCloning')}</TooltipContent>
      </Tooltip>
    )
  }

  if (cloneState === GIT_CLONE_STATE.FAILED) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          {badge('border-destructive/30 text-destructive hover:bg-destructive/10', <RefreshCw className="size-4" />)}
        </TooltipTrigger>
        <TooltipContent>{t('files.gitTipFailed')}</TooltipContent>
      </Tooltip>
    )
  }

  if (gitEnabled) {
    return gitHasRemote ? (
      <Tooltip>
        <TooltipTrigger asChild>
          {badge('border-success/35 bg-success/10 text-success', <GitBranch className="size-4" />)}
        </TooltipTrigger>
        <TooltipContent>{t('files.gitTipOrigin')}</TooltipContent>
      </Tooltip>
    ) : (
      <Tooltip>
        <TooltipTrigger asChild>
          {badge('border-warning/35 bg-warning/10 text-warning', <FolderGit2 className="size-4" />)}
        </TooltipTrigger>
        <TooltipContent>{t('files.gitTipLocal')}</TooltipContent>
      </Tooltip>
    )
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        {badge('border-transparent text-text-3', <Folder className="size-4" />)}
      </TooltipTrigger>
      <TooltipContent>{t('files.gitTipNone')}</TooltipContent>
    </Tooltip>
  )
}
