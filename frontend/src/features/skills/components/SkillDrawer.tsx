import { type ReactNode } from 'react'
import { useT } from '@/i18n'
import { X, ChevronRight, ChevronDown, Trash2, Loader2, Share2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
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

export function SkillDrawer({
  name,
  dirName,
  icon,
  description,
  sourceLabel,
  categoryName,
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

  return (
    <>
      <div
        className="fixed inset-0 z-40 bg-black/60 animate-in fade-in duration-150"
        onClick={onClose}
      />
      <div className="fixed inset-y-0 right-0 z-50 w-[400px] animate-in slide-in-from-right duration-200 border-l border-border bg-bg-layer-1 shadow-pop">
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-border px-4 py-3">
            <h2 className="text-base font-semibold text-text-1">{t('skills.details')}</h2>
            <button
              onClick={onClose}
              className="text-text-3 transition-colors hover:text-text-1"
            >
              <X className="size-4" />
            </button>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-auto p-4 space-y-4">
            {/* Icon + Name */}
            <div className="flex items-start gap-3">
              {editableIcon ? (
                <div ref={iconPickerRef} className="relative shrink-0 mt-0.5">
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        onClick={() => setIconPickerOpen(!iconPickerOpen)}
                        className={cn(
                          'inline-flex size-8 items-center justify-center rounded-md transition-colors',
                          iconPickerOpen
                            ? 'bg-bg-layer-3 text-text-1'
                            : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                        )}
                      >
                        <Icon name={icon} className="size-5" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>{t('skills.editIcon')}</TooltipContent>
                  </Tooltip>
                  {iconPickerOpen && onIconChange && (
                    <div className="absolute left-0 top-full z-50 mt-1 w-72 rounded-md border border-border bg-bg-layer-1 p-3 shadow-pop">
                      <IconPicker value={icon} onChange={onIconChange} />
                    </div>
                  )}
                </div>
              ) : (
                <Icon name={icon} className="size-5 shrink-0 mt-0.5 text-text-2" />
              )}
              <div className="min-w-0">
                <h3 className="text-base font-semibold text-text-1">{name}</h3>
                <p className="text-xs text-text-muted">{dirName}</p>
              </div>
            </div>

            {/* Description */}
            <p className="text-sm text-text-2 leading-relaxed">{description}</p>

            <div className="border-t border-border" />

            {/* Meta */}
            <div className="space-y-2.5">
              <div className="flex items-center gap-2">
                <span className="shrink-0 text-xs text-text-muted w-16">
                  {t('skills.source')}
                </span>
                <span className="text-xs text-text-2">{sourceLabel}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="shrink-0 text-xs text-text-muted w-16">
                  {t('common.category')}
                </span>
                {editableCategory && categories ? (
                  <div ref={categoryPickerRef} className="relative">
                    <button
                      onClick={() => setCategoryPickerOpen(!categoryPickerOpen)}
                      className={cn(
                        'flex items-center gap-1 rounded py-0.5 text-xs transition-colors',
                        categoryPickerOpen
                          ? 'bg-bg-layer-3 text-text-1'
                          : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                      )}
                    >
                      {categoryName || t('skills.noCategory')}
                      <ChevronDown className="size-3" />
                    </button>
                    {categoryPickerOpen && onCategoryChange && (
                      <div className="absolute top-full left-0 z-30 mt-1 min-w-[120px] rounded-md border border-border bg-bg-layer-2 py-1 shadow-pop">
                        {categories.map((cat) => (
                          <button
                            key={cat.categoryId}
                            onClick={() => onCategoryChange(cat.categoryId)}
                            className={cn(
                              'w-full whitespace-nowrap px-3 py-1.5 text-left text-xs transition-colors',
                              categoryName === cat.name
                                ? 'bg-bg-layer-3 text-text-1'
                                : 'text-text-3 hover:bg-bg-hover hover:text-text-2',
                            )}
                          >
                            {cat.name}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                ) : (
                  <span className="text-xs text-text-2">
                    {categoryName || t('skills.noCategory')}
                  </span>
                )}
              </div>
              {metaFields.map((field) => (
                <div key={field.label} className="flex items-center gap-2">
                  <span className="shrink-0 text-xs text-text-muted w-16">
                    {field.label}
                  </span>
                  <span className="text-xs text-text-2 tabular-nums">{field.value}</span>
                </div>
              ))}
              {showAlwaysOn && (
                <div className="flex items-center gap-2">
                  <span className="shrink-0 text-xs text-text-muted w-16">
                    {t('skills.alwaysOn')}
                  </span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        onClick={onToggleAlwaysOn}
                        className="flex items-center gap-1.5 rounded-md py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover"
                      >
                        <span
                          className={cn(
                            'relative inline-flex h-3.5 w-8 items-center rounded-full transition-colors',
                            alwaysOn === 1 ? 'bg-highlight' : 'bg-bg-hover',
                          )}
                        >
                          <span
                            className={cn(
                              'inline-block h-3 w-3 rounded-full transition-transform',
                              alwaysOn === 1
                                ? 'bg-highlight-fg translate-x-5'
                                : 'bg-text-muted',
                            )}
                          />
                        </span>
                        <span>
                          {alwaysOn === 1 ? t('common.enabled') : t('common.disabled')}
                        </span>
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>{t('skills.alwaysOnTip')}</TooltipContent>
                  </Tooltip>
                </div>
              )}
            </div>

            <div className="border-t border-border" />

            {/* SKILL.md collapsible */}
            <div>
              <button
                onClick={onToggleSkillMd}
                className="flex items-center gap-1.5 text-xs text-text-3 transition-colors hover:text-text-2"
              >
                <ChevronRight
                  className={cn('size-4 transition-transform', showSkillMd && 'rotate-90')}
                />
                SKILL.md {t('skills.skillMdContent')}
              </button>
              {showSkillMd && (
                <div className="mt-2 rounded-md border border-border bg-bg-layer-2 p-3">
                  {skillMdLoading ? (
                    <div className="flex items-center gap-2 py-2">
                      <Loader2 className="size-4 animate-spin text-text-3" />
                      <span className="text-xs text-text-3">{t('common.loading')}</span>
                    </div>
                  ) : skillMdContent ? (
                    <pre className="text-xs text-text-3 whitespace-pre-wrap leading-relaxed">
                      {skillMdContent}
                    </pre>
                  ) : (
                    <span className="text-xs text-text-3">{name}</span>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Bottom actions */}
          {(showShareButton || showDeleteButton) && (
            <div className="border-t border-border px-4 py-3 space-y-2">
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
          )}
        </div>
      </div>
    </>
  )
}
