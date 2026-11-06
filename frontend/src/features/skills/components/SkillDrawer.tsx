import { type ReactNode } from 'react'
import { useT } from '@/i18n'
import { ChevronRight, ChevronDown, Trash2, Loader2, Share2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { TONE_BG, TONE_SOFT, TONE_TEXT, toneFromKey } from '@/lib/tone'
import { Button } from '@/components/ui/button'
import { AppDrawer } from '@/components/patterns/AppDrawer'
import { Switch } from '@/components/ui/switch'
import { Icon } from '@/components/common/Icon'
import { IconPicker } from '@/components/common/IconPicker'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import type { SkillCategory } from '@/api/types'

export interface SkillDrawerMetaField {
  label: string
  value: string | ReactNode
}

export interface SkillDrawerProps {
  name: string
  dirName: string
  icon: string
  description: string
  sourceLabel: string
  categoryName: string | null
  categoryId?: number

  metaFields: SkillDrawerMetaField[]

  editableIcon?: boolean
  editableCategory?: boolean
  categories?: SkillCategory[]
  onIconChange?: (name: string) => void
  onCategoryChange?: (categoryId: number) => void

  showAlwaysOn?: boolean
  alwaysOn?: number
  onToggleAlwaysOn?: () => void

  showDeleteButton?: boolean
  onDelete?: () => void

  showShareButton?: boolean
  isShared?: boolean
  onShare?: () => void

  skillMdContent: string | null
  skillMdLoading: boolean

  iconPickerOpen: boolean
  setIconPickerOpen: (v: boolean) => void
  categoryPickerOpen: boolean
  setCategoryPickerOpen: (v: boolean) => void
  iconPickerRef: React.RefObject<HTMLDivElement>
  categoryPickerRef: React.RefObject<HTMLDivElement>

  showSkillMd: boolean
  onToggleSkillMd: () => void

  onClose: () => void
}

function MetaRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-start gap-3">
      <dt className="w-16 shrink-0 pt-0.5 text-meta text-text-3">{label}</dt>
      <dd className="min-w-0 flex-1 text-[13px] leading-[1.65] text-text-1">{children}</dd>
    </div>
  )
}

export function SkillDrawer({
  name,
  dirName,
  icon,
  description,
  sourceLabel,
  categoryName,
  categoryId,
  metaFields,
  editableIcon,
  editableCategory,
  categories,
  onIconChange,
  onCategoryChange,
  showAlwaysOn,
  alwaysOn,
  onToggleAlwaysOn,
  showDeleteButton,
  onDelete,
  showShareButton,
  isShared,
  onShare,
  skillMdContent,
  skillMdLoading,
  iconPickerOpen,
  setIconPickerOpen,
  categoryPickerOpen,
  setCategoryPickerOpen,
  iconPickerRef,
  categoryPickerRef,
  showSkillMd,
  onToggleSkillMd,
  onClose,
}: SkillDrawerProps) {
  const t = useT()

  const footer =
    showShareButton || showDeleteButton ? (
      <div className="space-y-2">
        {showShareButton && onShare && (
          <Button
            variant="default"
            size="default"
            className="w-full"
            disabled={isShared}
            onClick={isShared ? undefined : onShare}
          >
            <Share2 className="size-4" />
            {isShared ? t('skills.alreadyShared') : t('skills.shareToTeam')}
          </Button>
        )}
        {showDeleteButton && onDelete && (
          <Button
            variant="destructive"
            size="default"
            className="w-full"
            onClick={onDelete}
          >
            <Trash2 className="size-4" />
            {t('skills.deleteSkill')}
          </Button>
        )}
      </div>
    ) : undefined

  return (
    <AppDrawer
      open
      onClose={onClose}
      title={name}
      description={dirName !== name ? dirName : undefined}
      width="sm"
      footer={footer}
    >
      <div className="space-y-5">
        {/* Summary */}
        <div className="flex items-start gap-3">
          {editableIcon ? (
            <div ref={iconPickerRef} className="relative shrink-0">
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    onClick={() => setIconPickerOpen(!iconPickerOpen)}
                    className={cn(
                      'flex size-10 items-center justify-center rounded-lg transition-colors hover:bg-bg-layer-3 hover:text-text-1',
                      categoryId
                        ? `${TONE_SOFT[toneFromKey(categoryId)]} ${TONE_TEXT[toneFromKey(categoryId)]}`
                        : 'bg-bg-layer-2 text-text-2',
                      iconPickerOpen && 'bg-bg-layer-3 text-text-1',
                    )}
                  >
                    <Icon name={icon} className="size-5" />
                  </button>
                </TooltipTrigger>
                <TooltipContent>{t('skills.editIcon')}</TooltipContent>
              </Tooltip>
              {iconPickerOpen && onIconChange && (
                <div className="absolute left-0 top-full z-50 mt-1 w-72 rounded-lg border border-border bg-bg-layer-1 p-3 shadow-pop">
                  <IconPicker value={icon} onChange={onIconChange} />
                </div>
              )}
            </div>
          ) : (
            <span
              className={cn(
                'flex size-10 shrink-0 items-center justify-center rounded-lg',
                categoryId
                  ? `${TONE_SOFT[toneFromKey(categoryId)]} ${TONE_TEXT[toneFromKey(categoryId)]}`
                  : 'bg-bg-layer-2 text-text-2',
              )}
            >
              <Icon name={icon} className="size-5" />
            </span>
          )}
          <div className="min-w-0 pt-0.5">
            <p className="text-[13px] leading-[1.65] text-text-2">
              {description}
            </p>
          </div>
        </div>

        <div className="border-t border-border" />

        {/* Meta */}
        <dl className="space-y-2.5">
          <MetaRow label={t('skills.source')}>{sourceLabel}</MetaRow>
          <MetaRow label={t('common.category')}>
            {editableCategory && categories ? (
              <div ref={categoryPickerRef} className="relative">
                <button
                  onClick={() => setCategoryPickerOpen(!categoryPickerOpen)}
                  className={cn(
                    'flex items-center gap-1 rounded-md py-0.5 transition-colors',
                    categoryPickerOpen
                      ? 'text-accent'
                      : 'text-text-1 hover:text-accent',
                  )}
                >
                  {categoryName || t('skills.noCategory')}
                  <ChevronDown className="size-3.5" />
                </button>
                {categoryPickerOpen && onCategoryChange && (
                  <div className="absolute top-full left-0 z-30 mt-1 min-w-[160px] rounded-lg border border-border bg-bg-layer-1 py-1 shadow-pop">
                    {categories.map((cat) => (
                      <button
                        key={cat.categoryId}
                        onClick={() => onCategoryChange(cat.categoryId)}
                        className={cn(
                          'flex w-full items-center gap-1.5 whitespace-nowrap px-3 py-1.5 text-left text-meta transition-colors',
                          categoryName === cat.name
                            ? 'bg-accent-soft text-accent'
                            : 'text-text-3 hover:bg-bg-layer-2 hover:text-text-1',
                        )}
                      >
                        <span
                          aria-hidden="true"
                          className={cn('size-1.5 shrink-0 rounded-full', TONE_BG[toneFromKey(cat.categoryId)])}
                        />
                        {cat.name}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <span>{categoryName || t('skills.noCategory')}</span>
            )}
          </MetaRow>
          {metaFields.map((field) => (
            <MetaRow key={field.label} label={field.label}>
              <span className="tabular-nums">{field.value}</span>
            </MetaRow>
          ))}
          {showAlwaysOn && (
            <MetaRow label={t('skills.alwaysOn')}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="flex items-center gap-2">
                    <Switch
                      size="sm"
                      checked={alwaysOn === 1}
                      onCheckedChange={() => onToggleAlwaysOn?.()}
                      aria-label={t('skills.alwaysOn')}
                    />
                    <span>
                      {alwaysOn === 1 ? t('common.enabled') : t('common.disabled')}
                    </span>
                  </span>
                </TooltipTrigger>
                <TooltipContent>{t('skills.alwaysOnTip')}</TooltipContent>
              </Tooltip>
            </MetaRow>
          )}
        </dl>

        <div className="border-t border-border" />

        {/* SKILL.md collapsible */}
        <div>
          <button
            onClick={onToggleSkillMd}
            className="flex items-center gap-1.5 text-[13px] leading-[1.65] text-text-3 transition-colors hover:text-text-1"
          >
            <ChevronRight
              className={cn('size-4 transition-transform', showSkillMd && 'rotate-90')}
            />
            SKILL.md {t('skills.skillMdContent')}
          </button>
          {showSkillMd && (
            <div className="mt-2 rounded-lg border border-border bg-bg-layer-2 p-3">
              {skillMdLoading ? (
                <div className="flex items-center gap-2 py-2">
                  <Loader2 className="size-4 animate-spin text-text-3" />
                  <span className="text-meta text-text-3">{t('common.loading')}</span>
                </div>
              ) : skillMdContent ? (
                <pre className="text-[13px] leading-[1.65] whitespace-pre-wrap text-text-3">
                  {skillMdContent}
                </pre>
              ) : (
                <span className="text-meta text-text-3">{name}</span>
              )}
            </div>
          )}
        </div>
      </div>
    </AppDrawer>
  )
}
