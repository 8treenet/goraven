import { useState } from 'react'
import {
  Check, AlertCircle, Loader2, Search, FileArchive, Trash2,
} from 'lucide-react'
import { IconPickerTrigger } from '@/components/common/IconPicker'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import { useAdminSkills } from './context'
import {
  Drawer, SelectField, SkillIcon, StatusText, EmptyState,
  getSourceLabel, getInstallStatusLabel, formatDate,
} from './shared'

/* ============================================
   System Skill Editor Drawer
   ============================================ */

export function SystemEditorDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, systemForm, setSystemForm, systemParsed,
    saving, saveSystemSkill,
  } = useAdminSkills()

  return (
    <Drawer
      open={drawer === 'system-editor'}
      onClose={() => setDrawer(null)}
      title={systemForm.skillId ? t('adminSkills.editGlobalSkill') : t('adminSkills.newGlobalSkill')}
      description={t('adminSkills.editSkillMdHint')}
      width="w-[720px]"
    >
      <div className="flex h-full">
        <div className="flex w-[260px] shrink-0 flex-col gap-3 border-r border-border p-4">
          <div className="rounded-lg bg-bg-layer-2 p-3">
            <div className="text-xs text-text-3">{t('common.name')}</div>
            <div className="mt-1 break-all text-sm font-medium text-text-1">{systemParsed.name || t('adminSkills.unresolved')}</div>
          </div>
          <div className="rounded-lg bg-bg-layer-2 p-3">
            <div className="text-xs text-text-3">{t('common.description')}</div>
            <div className="mt-1 break-words text-sm text-text-1">{systemParsed.description || t('adminSkills.unresolved')}</div>
          </div>
          <div className="rounded-lg bg-bg-layer-2 p-3">
            <div className="text-xs text-text-3">{t('adminSkills.validation')}</div>
            <div className="mt-2 space-y-1">
              {systemParsed.errors.length === 0 ? (
                <div className="flex items-center gap-1.5 text-sm text-text-1"><Check className="size-4" /> {t('adminSkills.formatValid')}</div>
              ) : systemParsed.errors.map((error) => (
                <div key={error} className="flex items-center gap-1.5 text-sm text-text-1"><AlertCircle className="size-4" /> {error}</div>
              ))}
            </div>
          </div>
          <p className="text-xs leading-5 text-text-3">{t('adminSkills.globalSkillDesc')}</p>
        </div>
        <div className="flex min-w-0 flex-1 flex-col">
          <textarea
            value={systemForm.content}
            onChange={(event) => setSystemForm((form) => ({ ...form, content: event.target.value }))}
            className="min-h-0 flex-1 resize-none break-all bg-transparent p-4 font-mono text-sm leading-6 text-text-1 outline-none placeholder:text-text-muted"
            spellCheck={false}
          />
          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <span className="text-xs text-text-3">{t('adminSkills.effectiveInNew')}</span>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{t('common.cancel')}</Button>
              <Button onClick={saveSystemSkill} disabled={saving} className="text-highlight">
                {saving && <Loader2 className="size-4 animate-spin" />}
                {t('common.save')}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Market Skill Editor Drawer
   ============================================ */

export function MarketEditorDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, marketForm, setMarketForm, selectedMarketSkill,
    categories, saving, saveMarketSkill,
  } = useAdminSkills()

  return (
    <Drawer
      open={drawer === 'market-editor'}
      onClose={() => setDrawer(null)}
      title={t('adminSkills.editMarketSkill')}
      description={t('adminSkills.marketEditHint')}
    >
      {selectedMarketSkill && marketForm && (
        <div className="flex h-full flex-col">
          <div className="flex-1 space-y-4 overflow-auto p-4">
            <div className="rounded-lg bg-bg-layer-2 p-3">
              <div className="flex items-center gap-3">
                <SkillIcon icon={marketForm.icon} />
                <div>
                  <div className="font-medium text-text-1">{selectedMarketSkill.name}</div>
                  <div className="mt-1 text-xs text-text-3">{getSourceLabel(selectedMarketSkill.source)}，{t('adminSkills.nameReadonly')}</div>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{t('common.icon')}</Label>
                <IconPickerTrigger value={marketForm.icon} onChange={(icon) => setMarketForm({ ...marketForm, icon })} />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">{t('adminPersonaTemplates.sortOrder')}</Label>
                <Input type="number" value={marketForm.sortOrder} onChange={(event) => setMarketForm({ ...marketForm, sortOrder: Number(event.target.value) })} />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-text-2">{t('common.category')}</Label>
              <SelectField value={String(marketForm.categoryId)} onChange={(value) => setMarketForm({ ...marketForm, categoryId: Number(value) })} className="w-full">
                {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
              </SelectField>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-text-2">{t('adminSkills.adminNotes')}</Label>
              <textarea
                value={marketForm.remark}
                onChange={(event) => setMarketForm({ ...marketForm, remark: event.target.value })}
                spellCheck={false}
                className="min-h-24 w-full resize-none rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm text-text-1 outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                placeholder={t('adminSkills.adminOnly')}
              />
            </div>

            <div className="rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-3">
              {translate('common.installed')} {selectedMarketSkill.installedCount} {t('adminSkills.installedCount')}
            </div>

            {selectedMarketSkill.content && (
              <div className="space-y-1.5">
                <Label className="text-xs text-text-2">SKILL.md</Label>
                <pre className="overflow-auto rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-2 whitespace-pre-wrap break-words">{selectedMarketSkill.content}</pre>
              </div>
            )}
          </div>
          <div className="flex justify-end gap-2 border-t border-border p-4">
            <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{t('common.cancel')}</Button>
            <Button onClick={saveMarketSkill} disabled={saving} className="text-highlight">{saving && <Loader2 className="size-4 animate-spin" />}{t('common.save')}</Button>
          </div>
        </div>
      )}
    </Drawer>
  )
}

/* ============================================
   ClawHub Import Drawer
   ============================================ */

export function ClawHubDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, categories,
    clawHubMode, clawHubQuery, setClawHubQuery, clawHubSort,
    clawHubSearching, clawHubExploring, clawHubSelected, setClawHubSelected,
    clawHubCategoryId, setClawHubCategoryId, clawHubIcon, setClawHubIcon,
    importing, clawHubResults,
    changeClawHubMode, changeClawHubSort, submitClawHubSearch,
    selectClawHubItem, importClawHubSkill,
  } = useAdminSkills()

  const [clawHubTokenConfigured] = useState(true)

  return (
    <Drawer
      open={drawer === 'clawhub'}
      onClose={() => setDrawer(null)}
      title={t('adminSkills.importTitle')}
      description={t('adminSkills.importDesc')}
      width="w-[760px]"
    >
      {!clawHubTokenConfigured ? (
        <div className="flex h-full items-center justify-center p-6">
          <div className="max-w-sm text-center">
            <AlertCircle className="mx-auto mb-3 size-8 text-text-muted" />
            <h3 className="text-sm font-semibold text-text-1">{t('adminSkills.clawhubNoToken')}</h3>
            <p className="mt-1 text-sm text-text-3">{t('adminSkills.clawhubTokenHint')}</p>
          </div>
        </div>
      ) : (
        <div className="flex h-full">
          <div className="flex w-[48%] flex-col border-r border-border">
            <div className="space-y-3 border-b border-border p-4">
              <div className="inline-flex rounded-lg bg-bg-layer-2 p-1">
                <button onClick={() => changeClawHubMode('search')} className={cn('h-7 rounded-md px-3 text-sm', clawHubMode === 'search' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3')}>{t('common.search')}</button>
                <button onClick={() => changeClawHubMode('explore')} className={cn('h-7 rounded-md px-3 text-sm', clawHubMode === 'explore' ? 'bg-bg-layer-3 text-text-1' : 'text-text-3')}>{t('adminSkills.browse')}</button>
              </div>
              {clawHubMode === 'search' && (
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <div className="relative min-w-0 flex-1">
                      <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-text-3" />
                      <Input
                        value={clawHubQuery}
                        onChange={(event) => setClawHubQuery(event.target.value)}
                        onKeyDown={(event) => { if (event.key === 'Enter') submitClawHubSearch() }}
                        placeholder={t('adminSkills.searchHint')}
                        className="pl-8 text-sm"
                        disabled={clawHubSearching}
                      />
                    </div>
                    <Button onClick={submitClawHubSearch} disabled={clawHubSearching || !clawHubQuery.trim()}>
                      {clawHubSearching && <Loader2 className="size-4 animate-spin" />}
                      {t('common.search')}
                    </Button>
                  </div>
                  <p className="text-xs leading-5 text-text-3">{t('adminSkills.searchSlow')}</p>
                </div>
              )}
              {clawHubMode === 'explore' && (
                <div className="grid grid-cols-3 gap-1 rounded-lg bg-bg-layer-2 p-1">
                  {([
                    ['trending', t('adminSkills.trending')],
                    ['newest', t('adminSkills.newest')],
                    ['updated', t('common.updated')],
                    ['downloads', t('adminSkills.downloads')],
                    ['stars', t('adminSkills.stars')],
                    ['installs', t('skills.installCount')],
                  ] as const).map(([sort, label]) => (
                    <button
                      key={sort}
                      onClick={() => changeClawHubSort(sort)}
                      className={cn(
                        'h-7 rounded-md px-2 text-xs transition-colors',
                        clawHubSort === sort ? 'bg-bg-layer-3 text-text-1' : 'text-text-3 hover:text-text-1',
                      )}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="flex-1 overflow-auto p-2">
              {clawHubSearching || clawHubExploring ? (
                <div className="space-y-2 p-2">
                  {[0, 1, 2].map((item) => (
                    <div key={item} className="rounded-lg bg-bg-layer-2 p-3">
                      <div className="h-4 w-2/3 rounded bg-bg-layer-3" />
                      <div className="mt-3 h-3 w-full rounded bg-bg-layer-3" />
                      <div className="mt-2 h-3 w-1/2 rounded bg-bg-layer-3" />
                    </div>
                  ))}
                </div>
              ) : clawHubResults.length === 0 ? (
                <div className="flex h-full items-center justify-center px-6 text-center">
                  <div>
                    <Search className="mx-auto mb-3 size-8 text-text-muted" />
                    <div className="text-sm font-medium text-text-1">
                      {clawHubMode === 'search' && clawHubQuery ? translate('adminPersonaTemplates.noMatch') : translate('adminSkills.searchHint')}
                    </div>
                    <p className="mt-1 text-xs leading-5 text-text-3">
                      {clawHubMode === 'search'
                        ? translate('adminSkills.clawhubSearchTip')
                        : translate('adminSkills.clawhubRankingWait')}
                    </p>
                  </div>
                </div>
              ) : clawHubResults.map((item) => (
                <button
                  key={item.slug}
                  onClick={() => selectClawHubItem(item)}
                  className={cn(
                    'mb-1 w-full rounded-lg px-3 py-2 text-left transition-colors hover:bg-bg-hover',
                    clawHubSelected?.slug === item.slug && 'bg-bg-layer-3',
                  )}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="font-medium text-text-1">{item.displayName}</div>
                    <span className="text-xs text-text-3">v{item.version}</span>
                  </div>
                  <div className="mt-1 line-clamp-2 text-xs leading-5 text-text-3">{item.summary}</div>
                  <div className="mt-2 flex gap-3 text-xs text-text-muted">
                    <span>score {item.score}</span>
                    <span>{item.downloads} {t('adminSkills.downloads')}</span>
                    <span>{item.installs} {translate('skills.installCount')}</span>
                    <span>{item.stars} {t('adminSkills.stars')}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>
          <div className="flex min-w-0 flex-1 flex-col">
            {clawHubSelected ? (
              <>
                <div className="flex-1 overflow-auto p-4">
                  <div className="mb-4">
                    <h3 className="text-base font-semibold text-text-1">{clawHubSelected.displayName}</h3>
                    <p className="mt-1 text-sm text-text-3">{clawHubSelected.summary}</p>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-1.5">
                      <Label className="text-xs text-text-2">{t('common.category')}</Label>
                      <SelectField value={String(clawHubCategoryId)} onChange={(value) => setClawHubCategoryId(Number(value))} className="w-full">
                        {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
                      </SelectField>
                    </div>
                    <div className="space-y-1.5">
                      <Label className="text-xs text-text-2">{translate('common.icon')}</Label>
                      <IconPickerTrigger value={clawHubIcon} onChange={setClawHubIcon} />
                    </div>
                  </div>
                  <pre className="mt-4 max-h-[360px] overflow-auto rounded-lg bg-bg-layer-2 p-3 text-xs leading-5 text-text-2">{clawHubSelected.content}</pre>
                </div>
                <div className="border-t border-border p-4">
                  {importing && <div className="mb-3 flex items-center gap-2 text-xs text-text-3"><Loader2 className="size-4 animate-spin" /> {t('adminSkills.downloadingMd')}</div>}
                  <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={() => setClawHubSelected(null)} disabled={importing} className="text-highlight hover:text-highlight/80">{t('common.back')}</Button>
                    <Button onClick={importClawHubSkill} disabled={importing} className="text-highlight">{importing && <Loader2 className="size-4 animate-spin" />}{t('adminSkills.confirmImport')}</Button>
                  </div>
                </div>
              </>
            ) : (
              <div className="flex h-full items-center justify-center p-6 text-center text-sm text-text-3">
                {t('adminSkills.selectToPreview')}
              </div>
            )}
          </div>
        </div>
      )}
    </Drawer>
  )
}

/* ============================================
   Publish Drawer
   ============================================ */

export function PublishDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, categories,
    publishFile, publishUploadId, publishCategoryId, setPublishCategoryId,
    publishIcon, setPublishIcon, publishing, publishProgress,
    isUploadingZip, isDraggingZip, setIsDraggingZip, fileInputRef,
    uploadPublishFile, publishSkill,
  } = useAdminSkills()

  return (
    <Drawer open={drawer === 'publish'} onClose={() => setDrawer(null)} title={t('adminSkills.publishTitle')} description={t('adminSkills.publishDesc')}>
      <div className="flex h-full flex-col">
        <div className="flex-1 space-y-4 overflow-auto p-4">
          <input
            ref={fileInputRef as React.RefObject<HTMLInputElement>}
            type="file"
            accept=".zip,application/zip,application/x-zip-compressed"
            className="hidden"
            onChange={(event) => {
              const file = event.target.files?.[0]
              if (file) uploadPublishFile(file)
              event.target.value = ''
            }}
          />
          <div
            role="button"
            tabIndex={0}
            onClick={() => { if (!isUploadingZip && !publishing) fileInputRef.current?.click() }}
            onKeyDown={(event) => {
              if ((event.key === 'Enter' || event.key === ' ') && !isUploadingZip && !publishing) {
                event.preventDefault()
                fileInputRef.current?.click()
              }
            }}
            onDragOver={(event) => {
              event.preventDefault()
              event.stopPropagation()
              if (isUploadingZip || publishing) return
              event.dataTransfer.dropEffect = 'copy'
              if (!isDraggingZip) setIsDraggingZip(true)
            }}
            onDragEnter={(event) => {
              event.preventDefault()
              event.stopPropagation()
              if (isUploadingZip || publishing) return
              setIsDraggingZip(true)
            }}
            onDragLeave={(event) => {
              event.preventDefault()
              event.stopPropagation()
              if (event.currentTarget.contains(event.relatedTarget as Node)) return
              setIsDraggingZip(false)
            }}
            onDrop={(event) => {
              event.preventDefault()
              event.stopPropagation()
              setIsDraggingZip(false)
              if (isUploadingZip || publishing) return
              const file = event.dataTransfer.files?.[0]
              if (file) uploadPublishFile(file)
            }}
            aria-disabled={isUploadingZip || publishing}
            className={cn(
              'flex h-36 w-full flex-col items-center justify-center rounded-lg border border-dashed text-center transition-colors',
              isUploadingZip || publishing
                ? 'cursor-not-allowed border-border bg-bg-layer-2 opacity-60'
                : 'cursor-pointer hover:bg-bg-hover',
              isDraggingZip
                ? 'border-interactive bg-bg-hover'
                : 'border-border bg-bg-layer-2',
            )}
          >
            {isUploadingZip ? <Loader2 className="mb-2 size-8 animate-spin text-text-3" /> : <FileArchive className="mb-2 size-8 text-text-3" />}
            <span className="text-sm font-medium text-text-1">{publishFile || t('adminSkills.zipDropHint')}</span>
            <span className="mt-1 text-xs text-text-3">{t('adminSkills.zipHint')}</span>
          </div>
          {publishFile && (
            <div className="rounded-lg bg-bg-layer-2 p-3">
              <div className="flex justify-between text-xs text-text-3"><span>{publishUploadId || publishFile}</span><span>{publishProgress}%</span></div>
              <div className="mt-2 h-1.5 overflow-hidden rounded bg-bg-layer-3">
                <div className="h-full bg-text-1 transition-all duration-200" style={{ width: `${publishProgress}%` }} />
              </div>
            </div>
          )}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label className="text-xs text-text-2">{translate('common.category')}</Label>
              <SelectField value={String(publishCategoryId)} onChange={(value) => setPublishCategoryId(Number(value))} className="w-full">
                {categories.map((category) => <option key={category.categoryId} value={category.categoryId}>{category.name}</option>)}
              </SelectField>
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs text-text-2">{translate('common.icon')}</Label>
              <IconPickerTrigger value={publishIcon} onChange={setPublishIcon} />
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 border-t border-border p-4">
          <Button variant="outline" onClick={() => setDrawer(null)} className="text-highlight hover:text-highlight/80">{translate('common.cancel')}</Button>
          <Button onClick={publishSkill} disabled={publishing || isUploadingZip} className="text-highlight">{publishing && <Loader2 className="size-4 animate-spin" />}{translate('adminSkills.publishSkill')}</Button>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Categories Drawer
   ============================================ */

export function CategoriesDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, categories,
    categoryDraft, setCategoryDraft, saveCategory, setDeleteTarget,
  } = useAdminSkills()

  return (
    <Drawer open={drawer === 'categories'} onClose={() => setDrawer(null)} title={t('adminSkills.categoryManagement')} description={t('adminSkills.catManageHint')}>
      <div className="flex h-full flex-col">
        <div className="border-b border-border p-4">
          <div className="grid grid-cols-[1fr_180px_auto] gap-2">
            <Input placeholder={t('adminPersonaTemplates.categoryName')} maxLength={20} value={categoryDraft.name} onChange={(event) => setCategoryDraft({ ...categoryDraft, name: event.target.value })} />
            <IconPickerTrigger value={categoryDraft.icon} onChange={(icon) => setCategoryDraft({ ...categoryDraft, icon })} />
            <Button onClick={saveCategory}>{categoryDraft.categoryId ? t('common.save') : t('common.create')}</Button>
          </div>
        </div>
        <div className="flex-1 overflow-auto p-2">
          {categories.map((category) => (
            <div key={category.categoryId} className="mb-1 flex items-center justify-between rounded-lg px-3 py-2 transition-colors hover:bg-bg-hover">
              <div className="flex items-center gap-3">
                <SkillIcon icon={category.icon} />
                <div>
                  <div className="flex items-center gap-2 text-sm font-medium text-text-1">
                    {category.name}
                    {category.isDefault === 1 && <span className="rounded bg-interactive/10 px-1.5 py-0.5 text-xs text-interactive">{t('common.default')}</span>}
                  </div>
                  <div className="mt-1 text-xs text-text-3">{category.skillCount} {t('adminSkills.skillNumUpdated')} {formatDate(category.updated)}</div>
                </div>
              </div>
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={category.isDefault === 1}
                  onClick={() => setCategoryDraft({ categoryId: category.categoryId, name: category.name, icon: category.icon })}
                >{t('common.edit')}</Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  disabled={category.isDefault === 1}
                  onClick={() => setDeleteTarget({ type: 'category', id: category.categoryId, label: category.name })}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Users Drawer
   ============================================ */

export function UsersDrawer() {
  const t = useT()
  const {
    drawer, setDrawer, selectedMarketSkill, selectedSkillUsers,
  } = useAdminSkills()

  return (
    <Drawer
      open={drawer === 'users'}
      onClose={() => setDrawer(null)}
      title={t('adminSkills.installedUsers')}
      description={selectedMarketSkill ? selectedMarketSkill.name : undefined}
    >
      {selectedSkillUsers.length === 0 ? (
        <EmptyState title={t('adminSkills.noInstalledUsers')} description={t('adminSkills.installedUsersHint')} />
      ) : (
        <div className="p-4">
          <div className="overflow-hidden rounded-lg bg-bg-layer-2">
            <table className="w-full text-left text-sm">
              <thead className="text-xs text-text-3">
                <tr>
                  <th className="px-3 py-2 font-normal">{t('common.user')}</th>
                  <th className="px-3 py-2 font-normal">{t('common.status')}</th>
                  <th className="px-3 py-2 font-normal">{translate('skills.installTime')}</th>
                </tr>
              </thead>
              <tbody>
                {selectedSkillUsers.map((record) => (
                  <tr key={record.recordId}>
                    <td className="px-3 py-2 text-text-1">{record.userId}</td>
                    <td className="px-3 py-2"><StatusText tone={record.installStatus === 3 ? 'warn' : record.installStatus === 2 ? 'good' : 'neutral'}>{getInstallStatusLabel(record.installStatus)}</StatusText></td>
                    <td className="px-3 py-2 text-text-3">{formatDate(record.created)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </Drawer>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

export function DeleteConfirmDialog() {
  const {
    deleteTarget, setDeleteTarget, cascadeDelete, setCascadeDelete, confirmDelete,
  } = useAdminSkills()
  const t = useT()

  return (
    <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {deleteTarget?.type === 'market' ? t('adminSkills.deleteMarketSkill') : deleteTarget?.type === 'category' ? t('adminSkills.deleteCategory') : t('adminSkills.deleteGlobalSkill')}
          </DialogTitle>
          <DialogDescription>
            {deleteTarget?.type === 'market'
              ? t('adminSkills.deleteMarketSkillDesc')
              : deleteTarget?.type === 'category'
                ? t('adminSkills.deleteCatDesc')
                : t('adminSkills.deleteGlobalSkillDesc')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="rounded-lg bg-bg-layer-2 px-3 py-2 text-sm text-text-1">{deleteTarget?.label}</div>
          {deleteTarget?.type === 'market' && (
            <label className="flex items-center gap-2 text-sm text-text-2">
              <input
                type="checkbox"
                checked={cascadeDelete}
                onChange={(event) => setCascadeDelete(event.target.checked)}
                className="size-4 rounded border-border bg-transparent"
              />
              {translate('adminSkills.alsoDeleteRecords')}
            </label>
          )}
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>{translate('common.cancel')}</Button>
            <Button variant="destructive" onClick={confirmDelete}>
              {deleteTarget?.type === 'market' && cascadeDelete ? translate('adminSkills.deleteSkillAndRecords') : translate('common.delete')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
