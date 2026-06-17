import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { toast } from 'sonner'
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
  Maximize2,
} from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { Icon, renderIcon } from '@/components/common/Icon'
import { IconPickerTrigger } from '@/components/common/IconPicker'
import type { IconName } from '@/components/common/icon-registry'
import {
  getPersonaTemplates,
  getAllPersonaCategories,
  createPersonaTemplate,
  updatePersonaTemplate,
  deletePersonaTemplate,
  createPersonaCategory,
  updatePersonaCategory,
  deletePersonaCategory,
} from '@/api/admin/personas'
import type { AdminPersonaTemplateItem, AdminPersonaCategoryItem } from '@/api/types'
import { useT, t as translate } from '@/i18n'

/* ============================================
   Types
   ============================================ */

interface TemplateFormData {
  name: string
  icon: string
  description: string
  roleInfo: string
  categoryId: number
  sortOrder: number
}

interface CategoryFormData {
  name: string
  icon: string
}

type PageState = 'loading' | 'data' | 'empty' | 'error'
type DrawerMode = 'add' | 'edit'
type CategoryDialogMode = 'add' | 'edit'



/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str
  return str.slice(0, maxLen) + '...'
}

/* ============================================
   RoleInfo Popover
   ============================================ */

function RoleInfoPopover({ content }: { content: string }) {
  const [visible, setVisible] = useState(false)
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>()

  const show = () => {
    clearTimeout(timeoutRef.current)
    setVisible(true)
  }

  const hide = () => {
    timeoutRef.current = setTimeout(() => setVisible(false), 150)
  }

  useEffect(() => {
    return () => clearTimeout(timeoutRef.current)
  }, [])

  return (
    <div className="relative" onMouseEnter={show} onMouseLeave={hide}>
      <span className="cursor-default text-xs text-text-2 font-mono">
        {truncate(content, 50)}
      </span>
      {visible && (
        <div
          className="absolute bottom-full left-0 z-50 mb-1 w-80 rounded-md border border-border bg-bg-layer-1 p-3 shadow-pop"
          onMouseEnter={show}
          onMouseLeave={hide}
        >
          <p className="max-h-60 overflow-auto text-xs leading-relaxed text-text-2 font-mono whitespace-pre-wrap">
            {content}
          </p>
        </div>
      )}
    </div>
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
  footer,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
  footer?: React.ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <div className={cn('fixed inset-0 z-50', open ? 'visible' : 'invisible')}>
        <div
          className={cn(
            'absolute inset-0 bg-black/60 transition-opacity duration-200',
            open ? 'opacity-100' : 'opacity-0',
          )}
          onClick={onClose}
        />
        <div
          className={cn(
            'absolute right-0 top-0 z-50 flex h-full w-[480px] flex-col border-l border-border bg-bg-layer-1 shadow-pop transition-transform duration-200 ease-out',
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
          {footer && (
            <div className="shrink-0 border-t border-border px-4 py-3">
              {footer}
            </div>
          )}
        </div>
      </div>
    </Dialog>
  )
}

/* ============================================
   Template Drawer (Create / Edit)
   ============================================ */

function TemplateDrawer({
  open,
  mode,
  onClose,
  onSave,
  template,
  categories,
}: {
  open: boolean
  mode: DrawerMode
  onClose: () => void
  onSave: (data: TemplateFormData) => void
  template: AdminPersonaTemplateItem | null
  categories: AdminPersonaCategoryItem[]
}) {
  const [form, setForm] = useState<TemplateFormData>({
    name: '',
    icon: 'bot',
    description: '',
    roleInfo: '',
    categoryId: categories[0]?.categoryId ?? 1,
    sortOrder: 0,
  })
  const [fullscreenOpen, setFullscreenOpen] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const t = useT()

  useEffect(() => {
    if (open) {
      if (mode === 'edit' && template) {
        setForm({
          name: template.name,
          icon: template.icon,
          description: template.description,
          roleInfo: template.roleInfo,
          categoryId: template.categoryId,
          sortOrder: template.sortOrder,
        })
      } else {
        setForm({
          name: '',
          icon: 'bot',
          description: '',
          roleInfo: '',
          categoryId: categories[0]?.categoryId ?? 1,
          sortOrder: 0,
        })
      }
      setErrors({})
    }
  }, [open, mode, template, categories])

  const validate = (): boolean => {
    const e: Record<string, string> = {}
    if (!form.name.trim()) e.name = translate('adminPersonaTemplates.errNameRequired')
    if (!form.roleInfo.trim()) e.roleInfo = translate('adminPersonaTemplates.errPromptRequired')
    else if (form.roleInfo.length > 500) e.roleInfo = translate('adminPersonaTemplates.errPromptTooLong')
    if (!form.categoryId) e.categoryId = translate('adminPersonaTemplates.errCategoryRequired')
    setErrors(e)
    return Object.keys(e).length === 0
  }

  const handleSave = () => {
    if (!validate()) return
    onSave(form)
  }

  const roleInfoCharCount = form.roleInfo.length

  return (
    <>
      <Drawer
        open={open}
        onClose={onClose}
        title={mode === 'add' ? t('adminPersonaTemplates.createTemplate') : t('adminPersonaTemplates.editTemplate')}
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" size="default" onClick={handleSave}>
              {t('common.save')}
            </Button>
          </div>
        }
      >
        <div className="flex flex-col gap-5">
          {/* Basic Info */}
          <div>
            <h3 className="mb-3 text-xs font-semibold text-text-3 uppercase tracking-wide">{t('adminPersonaTemplates.basicInfo')}</h3>
            <div className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">
                  {t('adminPersonaTemplates.templateName')} <span className="text-red-500">*</span>
                </Label>
                <Input
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder={t('adminPersonaTemplates.templateName')}
                  className="h-8 text-sm"
                  maxLength={64}
                />
                {errors.name && <p className="text-xs text-red-500">{errors.name}</p>}
              </div>

              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('common.icon')}</Label>
                <IconPickerTrigger
                  value={form.icon}
                  onChange={(name: IconName) => setForm((f) => ({ ...f, icon: name }))}
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('adminPersonaTemplates.templateDesc')}</Label>
                <Input
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  placeholder={t('adminPersonaTemplates.templateDescHint')}
                  className="h-8 text-sm"
                  maxLength={128}
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">
                  {t('common.category')} <span className="text-red-500">*</span>
                </Label>
                <select
                  value={form.categoryId}
                  onChange={(e) => setForm((f) => ({ ...f, categoryId: Number(e.target.value) }))}
                  className="h-8 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none focus:border-interactive focus:ring-3 focus:ring-interactive/50"
                >
                  {categories.map((c) => (
                    <option key={c.categoryId} value={c.categoryId}>
                      {c.name}
                    </option>
                  ))}
                </select>
                {errors.categoryId && <p className="text-xs text-red-500">{errors.categoryId}</p>}
              </div>
            </div>
          </div>

          {/* Core Content */}
          <div>
            <h3 className="mb-3 text-xs font-semibold text-text-3 uppercase tracking-wide">{t('adminPersonaTemplates.coreContent')}</h3>
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between">
                <Label className="text-xs text-text-2">
                  {t('adminPersonaTemplates.systemPrompt')} <span className="text-red-500">*</span>
                </Label>
                <button
                  type="button"
                  onClick={() => setFullscreenOpen(true)}
                  className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
                >
                  <Maximize2 className="size-3" />
                  {t('adminPersonaTemplates.fullscreenEdit')}
                </button>
              </div>
              <textarea
                value={form.roleInfo}
                onChange={(e) => setForm((f) => ({ ...f, roleInfo: e.target.value }))}
                placeholder={t('adminPersonaTemplates.promptPlaceholder')}
                maxLength={500}
                className="min-h-[200px] resize-y rounded-lg border border-border-strong bg-transparent px-2.5 py-2 text-sm text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-3 focus:ring-interactive/50 font-mono"
              />
              <div className="flex items-center justify-between">
                {errors.roleInfo ? (
                  <p className="text-xs text-red-500">{errors.roleInfo}</p>
                ) : (
                  <span />
                )}
                <span className="text-xs text-text-muted tabular-nums">{t('adminPersonaTemplates.charCount').replace('{count}', String(roleInfoCharCount))}</span>
              </div>
            </div>
          </div>

          {/* Other Settings */}
          <div>
            <h3 className="mb-3 text-xs font-semibold text-text-3 uppercase tracking-wide">{t('adminPersonaTemplates.otherSettings')}</h3>
            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-text-2">{t('adminPersonaTemplates.sortOrder')}</Label>
              <Input
                type="number"
                value={form.sortOrder}
                onChange={(e) => setForm((f) => ({ ...f, sortOrder: Number(e.target.value) || 0 }))}
                className="h-8 w-24 text-sm"
              />
              <p className="text-xs text-text-muted">{t('adminPersonaTemplates.sortHint')}</p>
            </div>
          </div>
        </div>
      </Drawer>

      {/* Fullscreen roleInfo editor */}
      <Dialog open={fullscreenOpen} onOpenChange={setFullscreenOpen}>
        <DialogContent className="max-w-3xl h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{t('adminPersonaTemplates.fullscreenTitle')}</DialogTitle>
            <DialogDescription>{t('adminPersonaTemplates.fullscreenHint')}</DialogDescription>
          </DialogHeader>
          <textarea
            value={form.roleInfo}
            onChange={(e) => setForm((f) => ({ ...f, roleInfo: e.target.value }))}
            placeholder={t('adminPersonaTemplates.promptPlaceholder')}
            maxLength={500}
            className="flex-1 resize-none rounded-lg border border-border-strong bg-transparent px-3 py-2.5 text-sm text-text-1 outline-none placeholder:text-text-muted focus:border-interactive focus:ring-3 focus:ring-interactive/50 font-mono"
          />
          <div className="flex items-center justify-between pt-2">
            <span className="text-xs text-text-muted tabular-nums">{t('adminPersonaTemplates.charCount').replace('{count}', String(roleInfoCharCount))}</span>
            <Button variant="default" size="default" onClick={() => setFullscreenOpen(false)}>
              {t('adminPersonaTemplates.finishEdit')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}

/* ============================================
   Category Management Dialog
   ============================================ */

function CategoryManagementDialog({
  open,
  onClose,
  categories,
  onAdd,
  onEdit,
  onDelete,
}: {
  open: boolean
  onClose: () => void
  categories: AdminPersonaCategoryItem[]
  onAdd: (data: CategoryFormData) => void
  onEdit: (categoryId: number, data: CategoryFormData) => void
  onDelete: (categoryId: number) => void
}) {
  const [mode, setMode] = useState<CategoryDialogMode | null>(null)
  const [editTarget, setEditTarget] = useState<AdminPersonaCategoryItem | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AdminPersonaCategoryItem | null>(null)
  const [form, setForm] = useState<CategoryFormData>({ name: '', icon: 'tag' })
  const [formError, setFormError] = useState('')
  const t = useT()

  useEffect(() => {
    if (open) {
      setMode(null)
      setEditTarget(null)
      setDeleteTarget(null)
      setForm({ name: '', icon: 'tag' })
      setFormError('')
    }
  }, [open])

  const handleAddSave = () => {
    if (!form.name.trim()) {
      setFormError(translate('adminPersonaTemplates.errCatNameRequired'))
      return
    }
    onAdd(form)
    setMode(null)
    setForm({ name: '', icon: 'tag' })
    setFormError('')
  }

  const handleEditSave = () => {
    if (!form.name.trim()) {
      setFormError(translate('adminPersonaTemplates.errCatNameRequired'))
      return
    }
    if (editTarget) {
      onEdit(editTarget.categoryId, form)
    }
    setMode(null)
    setEditTarget(null)
    setFormError('')
  }

  return (
    <>
      <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('adminPersonaTemplates.catManageTitle')}</DialogTitle>
                <DialogDescription>{t('adminPersonaTemplates.catManageDesc')}</DialogDescription>
          </DialogHeader>

          {mode === null && (
            <>
              <div className="max-h-80 overflow-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-border text-left text-xs text-text-3">
                      <th className="pb-2 pl-1 pr-2 font-normal">{t('common.icon')}</th>
                      <th className="pb-2 pr-4 font-normal">{t('common.name')}</th>
                      <th className="pb-2 pr-4 font-normal text-right tabular-nums">{t('adminPersonaTemplates.templateCount')}</th>
                      <th className="pb-2 pr-1 font-normal" />
                    </tr>
                  </thead>
                  <tbody>
                    {categories.map((cat) => (
                      <tr key={cat.categoryId} className="border-b border-border transition-colors hover:bg-bg-hover">
                        <td className="py-2 pl-1 pr-2 text-text-2">
                          {renderIcon(cat.icon, 'size-3.5')}
                        </td>
                        <td className="py-2 pr-4">
                          <div className="flex items-center gap-1.5">
                            <span className="text-sm text-text-1">{cat.name}</span>
                            {cat.isDefault === 1 && (
                              <span className="inline-flex items-center rounded px-1 py-0 text-[10px] text-text-3 bg-bg-layer-3">
                                {t('common.default')}
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="py-2 pr-4 text-right text-sm tabular-nums text-text-2">
                          {cat.templateCount}
                        </td>
                        <td className="py-2 pr-1">
                          <div className="flex items-center gap-0.5">
                            <button
                              onClick={() => {
                                setEditTarget(cat)
                                setForm({ name: cat.name, icon: cat.icon })
                                setFormError('')
                                setMode('edit')
                              }}
                              disabled={cat.isDefault === 1}
                               className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-text-3"
                               title={cat.isDefault === 1 ? translate('adminPersonaTemplates.defaultUneditable') : translate('common.edit')}
                             >
                               <Pencil className="size-3.5" />
                             </button>
                             <button
                               onClick={() => setDeleteTarget(cat)}
                               disabled={cat.isDefault === 1}
                               className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-red-500 disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-text-3"
                               title={cat.isDefault === 1 ? translate('adminPersonaTemplates.defaultUndeletable') : translate('common.delete')}
                             >
                              <Trash2 className="size-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="flex justify-between pt-2">
                <Button
                  variant="outline"
                  size="default"
                  onClick={() => {
                    setForm({ name: '', icon: 'tag' })
                    setFormError('')
                    setMode('add')
                  }}
                >
                  <Plus className="size-3.5" />
                  {t('adminPersonaTemplates.addCategory')}
                </Button>
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.close')}
                </Button>
              </div>
            </>
          )}

          {(mode === 'add' || mode === 'edit') && (
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('adminPersonaTemplates.categoryName')}</Label>
                <Input
                  value={form.name}
                  onChange={(e) => {
                    setForm((f) => ({ ...f, name: e.target.value }))
                    setFormError('')
                  }}
                  placeholder={t('adminPersonaTemplates.catNamePlaceholder')}
                  className="h-8 text-sm"
                  maxLength={32}
                />
                {formError && <p className="text-xs text-red-500">{formError}</p>}
              </div>

              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-text-2">{t('common.icon')}</Label>
                <IconPickerTrigger
                  value={form.icon}
                  onChange={(name: IconName) => setForm((f) => ({ ...f, icon: name }))}
                />
              </div>

              <div className="flex justify-end gap-2">
                <Button
                  variant="ghost"
                  size="default"
                  onClick={() => {
                    setMode(null)
                    setEditTarget(null)
                    setFormError('')
                  }}
                >
                  {t('common.cancel')}
                </Button>
                <Button
                  variant="default"
                  size="default"
                  onClick={mode === 'add' ? handleAddSave : handleEditSave}
                >
                  {t('common.save')}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete category confirmation */}
      <Dialog open={deleteTarget !== null} onOpenChange={(v) => !v && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('adminPersonaTemplates.deleteCategory')}</DialogTitle>
            <DialogDescription>
              {translate('adminPersonaTemplates.deleteCatConfirm')}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={() => setDeleteTarget(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="destructive"
              size="default"
              onClick={() => {
                if (deleteTarget) {
                  onDelete(deleteTarget.categoryId)
                  setDeleteTarget(null)
                }
              }}
            >
              {t('common.delete')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}

/* ============================================
   Delete Template Dialog
   ============================================ */

function DeleteTemplateDialog({
  open,
  onClose,
  onConfirm,
  templateName,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  templateName: string
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{translate('adminPersonaTemplates.deleteTemplate')}</DialogTitle>
          <DialogDescription>
            {translate('adminPersonaTemplates.deleteTemplateDesc')} {translate('adminPersonaTemplates.deleteTemplateConfirm').replace('{name}', templateName)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose}>
            {translate('common.cancel')}
          </Button>
          <Button variant="destructive" size="default" onClick={onConfirm}>
            {translate('common.delete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Table Row
   ============================================ */

function TemplateRow({
  template,
  isEven,
  onEdit,
  onDelete,
}: {
  template: AdminPersonaTemplateItem
  isEven: boolean
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <tr className={cn(isEven && 'bg-bg-layer-1', 'transition-colors hover:bg-bg-hover')}>
      <td className="py-2.5 pl-4 pr-2 text-text-2">
        {renderIcon(template.icon, 'size-4')}
      </td>
      <td className="py-2.5 pr-4">
        <span className="text-sm font-medium text-text-1">{template.name}</span>
      </td>
      <td className="py-2.5 pr-4">
        <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs bg-bg-layer-3 text-text-2">
          {renderIcon(template.categoryIcon, 'size-3')}
          {template.categoryName}
        </span>
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-2 max-w-48">
        <span className="line-clamp-1">{template.description || '—'}</span>
      </td>
      <td className="py-2.5 pr-4">
        <RoleInfoPopover content={template.roleInfo} />
      </td>
      <td className="py-2.5 pr-4 text-sm tabular-nums text-text-2">
        {template.usageCount.toLocaleString()}
      </td>
      <td className="py-2.5 pr-4 text-sm tabular-nums text-text-3">
        {template.sortOrder}
      </td>
      <td className="py-2.5 pr-4 text-sm text-text-3 whitespace-nowrap">
        {formatDate(template.updated)}
      </td>
      <td className="py-2.5 pr-4">
        <div className="flex items-center gap-0.5">
          <button
            onClick={onEdit}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
            title={translate('common.edit')}
          >
            <Pencil className="size-3.5" />
          </button>
          <button
            onClick={onDelete}
            className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-red-500"
            title={translate('common.delete')}
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
  return (
    <div className="flex-1 overflow-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border text-left text-xs text-text-3">
            <th className="pb-2 pl-4 pr-2 font-normal" />
            <th className="pb-2 pr-4 font-normal">{translate('common.name')}</th>
            <th className="pb-2 pr-4 font-normal">{translate('common.category')}</th>
            <th className="pb-2 pr-4 font-normal">{translate('common.description')}</th>
            <th className="pb-2 pr-4 font-normal">{translate('adminPersonaTemplates.systemPrompt')}</th>
                  <th className="pb-2 pr-4 font-normal">{translate('adminPersonaTemplates.usageCount')}</th>
            <th className="pb-2 pr-4 font-normal">{translate('common.sort')}</th>
            <th className="pb-2 pr-4 font-normal">{translate('common.updated')}</th>
            <th className="pb-2 pr-4 font-normal" />
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: 8 }).map((_, i) => (
            <tr key={i} className={cn(i % 2 === 0 && 'bg-bg-layer-1')}>
              <td className="py-2.5 pl-4 pr-2">
                <div className="size-4 animate-pulse rounded-sm bg-bg-layer-3" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-24 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-32 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-40 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-10 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </td>
              <td className="py-2.5 pr-4">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
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
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3">
      <Icon name="bot" className="size-10 text-text-muted" />
      <div className="text-center">
        <p className="text-sm text-text-2">
          {hasFilter ? translate('adminPersonaTemplates.noMatch') : translate('adminPersonaTemplates.noTemplates')}
        </p>
      </div>
      {hasFilter ? (
        <button
          onClick={onClearFilter}
          className="text-xs text-interactive transition-colors hover:text-interactive-hover"
        >
          {translate('common.clear')}
        </button>
      ) : (
        <p className="text-xs text-text-3">{translate('adminPersonaTemplates.createFirst')}</p>
      )}
    </div>
  )
}

/* ============================================
   Error State
   ============================================ */

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <AlertCircle className="size-8 text-text-3" />
      <div className="text-center">
        <p className="text-sm text-text-2">{translate('common.loadFailed')}</p>
        <p className="mt-1 text-xs text-text-3">{translate('adminPersonaTemplates.fetchFailed')}</p>
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
      <span className="text-xs text-text-3">{translate('common.total')} {totalCount} {translate('adminPersonaTemplates.totalTemplates')}</span>
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
  const [templates, setTemplates] = useState<AdminPersonaTemplateItem[]>([])
  const [categories, setCategories] = useState<AdminPersonaCategoryItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<number | 'all'>('all')
  const [page, setPage] = useState(1)

  const [drawerMode, setDrawerMode] = useState<DrawerMode | null>(null)
  const [editTarget, setEditTarget] = useState<AdminPersonaTemplateItem | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AdminPersonaTemplateItem | null>(null)
  const [categoryDialogOpen, setCategoryDialogOpen] = useState(false)

  const loadData = useCallback(() => {
    setState('loading')
    const catPromise = categories.length === 0
      ? getAllPersonaCategories().then((res) => { setCategories(res.list); return res.list })
      : Promise.resolve(categories)

    Promise.all([
      getPersonaTemplates({
        page,
        pageSize: 15,
        search: search || undefined,
        categoryId: categoryFilter === 'all' ? undefined : categoryFilter,
      }),
      catPromise,
    ]).then(([res]) => {
      setTemplates(res.list)
      setTotalCount(res.totalCount)
      setTotalPages(res.totalPage)
      setState(res.list.length > 0 ? 'data' : 'empty')
    }).catch(() => {
      setState('error')
    })
  }, [page, search, categoryFilter, categories.length])

  useEffect(() => {
    loadData()
  }, [loadData])

  const safePage = Math.min(page, totalPages)

  // Reset page on filter/search change
  useEffect(() => {
    setPage(1)
  }, [search, categoryFilter])

  const handleSearch = useCallback((val: string) => {
    setSearch(val)
  }, [])

  const handleClearFilter = useCallback(() => {
    setSearch('')
    setCategoryFilter('all')
  }, [])

  const hasFilter = search.trim().length > 0 || categoryFilter !== 'all'

  // Category sorted by name
  const sortedCategories = useMemo(
    () => [...categories].sort((a, b) => a.name.localeCompare(b.name, 'zh')),
    [categories],
  )

  // Refresh categories after changes
  const refreshCategories = useCallback(() => {
    getAllPersonaCategories().then((res) => setCategories(res.list)).catch(() => {})
  }, [])

  // Create template
  const handleCreateTemplate = useCallback(
    (data: TemplateFormData) => {
      createPersonaTemplate(data).then(() => {
        setDrawerMode(null)
        toast.success(translate('adminPersonaTemplates.templateCreated'))
        loadData()
        refreshCategories()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [loadData, refreshCategories],
  )

  // Edit template
  const handleEditTemplate = useCallback(
    (data: TemplateFormData) => {
      if (!editTarget) return
      updatePersonaTemplate(editTarget.templateId, data).then(() => {
        setDrawerMode(null)
        setEditTarget(null)
        toast.success(translate('adminPersonaTemplates.templateUpdated'))
        loadData()
        refreshCategories()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [editTarget, loadData, refreshCategories],
  )

  // Delete template
  const handleDeleteTemplate = useCallback(() => {
    if (!deleteTarget) return
    deletePersonaTemplate(deleteTarget.templateId).then(() => {
      setDeleteTarget(null)
      toast.success(translate('adminPersonaTemplates.templateDeleted'))
      loadData()
      refreshCategories()
    }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
  }, [deleteTarget, loadData, refreshCategories])

  // Category CRUD
  const handleAddCategory = useCallback(
    (data: CategoryFormData) => {
      createPersonaCategory({ name: data.name, icon: data.icon }).then(() => {
        toast.success(translate('adminPersonaTemplates.categoryAdded'))
        refreshCategories()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [refreshCategories],
  )

  const handleEditCategory = useCallback(
    (categoryId: number, data: CategoryFormData) => {
      updatePersonaCategory(categoryId, { name: data.name, icon: data.icon }).then(() => {
        toast.success(translate('adminPersonaTemplates.categoryUpdated'))
        refreshCategories()
        loadData()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [refreshCategories, loadData],
  )

  const handleDeleteCategory = useCallback(
    (categoryId: number) => {
      deletePersonaCategory(categoryId).then(() => {
        toast.success(translate('adminPersonaTemplates.categoryDeleted'))
        refreshCategories()
        loadData()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [refreshCategories, loadData],
  )

  return (
    <div className="flex h-full flex-col bg-bg-base">
      {/* Toolbar */}
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-border px-4">
        <h1 className="text-base font-semibold text-text-1">{t('adminPersonaTemplates.title')}</h1>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-3" />
            <input
              type="text"
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              placeholder={t('adminPersonaTemplates.searchPlaceholder')}
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
            value={categoryFilter}
            onChange={(e) =>
              setCategoryFilter(e.target.value === 'all' ? 'all' : Number(e.target.value))
            }
            className="h-7 rounded-lg border border-border-strong bg-transparent px-2 text-xs text-text-1 outline-none focus:border-interactive focus:ring-2 focus:ring-interactive/30"
          >
            <option value="all">{t('adminPersonaTemplates.allCategories')}</option>
            {sortedCategories.map((c) => (
              <option key={c.categoryId} value={c.categoryId}>
                {c.name}
              </option>
            ))}
          </select>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCategoryDialogOpen(true)}
          >
            <Icon name="folder-open" className="size-3.5" />
            {t('adminPersonaTemplates.categoryManagement')}
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={() => setDrawerMode('add')}
          >
            <Plus className="size-3.5" />
            {t('adminPersonaTemplates.createTemplate')}
          </Button>
        </div>
      </div>

      {/* Content */}
      {state === 'loading' && <TableSkeleton />}

      {state === 'error' && <ErrorState onRetry={loadData} />}

      {state === 'empty' && <EmptyState hasFilter={false} onClearFilter={handleClearFilter} />}

      {state === 'data' && templates.length === 0 && (
        <EmptyState hasFilter={hasFilter} onClearFilter={handleClearFilter} />
      )}

      {state === 'data' && templates.length > 0 && (
        <>
          <div className="flex-1 overflow-auto">
            <table className="w-full">
              <thead>
                <tr className="sticky top-0 z-10 border-b border-border bg-bg-base text-left text-xs text-text-3">
                  <th className="pb-2 pl-4 pr-2 font-normal" />
                  <th className="pb-2 pr-4 font-normal">{t('common.name')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('common.category')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('common.description')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminPersonaTemplates.systemPrompt')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('adminPersonaTemplates.usageCount')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('common.sort')}</th>
                  <th className="pb-2 pr-4 font-normal">{t('common.updated')}</th>
                  <th className="pb-2 pr-4 font-normal" />
                </tr>
              </thead>
              <tbody>
                {templates.map((t, i) => (
                  <TemplateRow
                    key={t.templateId}
                    template={t}
                    isEven={i % 2 === 0}
                    onEdit={() => {
                      setEditTarget(t)
                      setDrawerMode('edit')
                    }}
                    onDelete={() => setDeleteTarget(t)}
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

      {/* Drawer */}
      <TemplateDrawer
        open={drawerMode !== null}
        mode={drawerMode ?? 'add'}
        onClose={() => {
          setDrawerMode(null)
          setEditTarget(null)
        }}
        onSave={drawerMode === 'add' ? handleCreateTemplate : handleEditTemplate}
        template={editTarget}
        categories={sortedCategories}
      />

      {/* Delete Dialog */}
      <DeleteTemplateDialog
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDeleteTemplate}
        templateName={deleteTarget?.name ?? ''}
      />

      {/* Category Management Dialog */}
      <CategoryManagementDialog
        open={categoryDialogOpen}
        onClose={() => setCategoryDialogOpen(false)}
        categories={sortedCategories}
        onAdd={handleAddCategory}
        onEdit={handleEditCategory}
        onDelete={handleDeleteCategory}
      />
    </div>
  )
}
