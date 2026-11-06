import { Upload, Download, Trash2, Archive } from 'lucide-react'
import { Icon } from '@/components/common/Icon'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { useAdminSkills } from './context'
import {
  SearchBox, SelectField, ScopeLabel, StatusToggle,
  getSourceLabel,
} from './shared'
import { EmptyState } from '@/components/patterns/EmptyState'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'

export function AdminMarketSkillsPage() {
  const t = useT()
  const {
    marketSearch, setMarketSearch, marketSource, setMarketSource, marketStatus, setMarketStatus,
    filteredMarketSkills, openMarketEditor, setDeleteTarget, toggleMarketStatus,
    setDrawer, openClawHubDrawer, openUsersDrawer,
  } = useAdminSkills()

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <SearchBox value={marketSearch} onChange={setMarketSearch} placeholder={t('adminSkills.searchSkillName')} />
          <SelectField value={marketSource} onChange={setMarketSource}>
            <option value="all">{t('adminSkills.allSources')}</option>
            <option value="clawhub">ClawHub</option>
            <option value="custom_upload">{t('adminSkills.customUpload')}</option>
          </SelectField>
          <SelectField value={marketStatus} onChange={setMarketStatus}>
            <option value="all">{t('adminSkills.allStatuses')}</option>
            <option value="1">{t('adminSkills.published')}</option>
            <option value="0">{t('adminSkills.unpublished')}</option>
          </SelectField>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => setDrawer('publish')} className="text-accent hover:text-accent-hover">
            <Upload className="size-4" />
            {t('adminSkills.publishSkill')}
          </Button>
          <Button onClick={() => openClawHubDrawer()}>
            <Download className="size-4" />
            {t('adminSkills.importFromClawhub')}
          </Button>
        </div>
      </div>

      {filteredMarketSkills.length === 0 ? (
        <EmptyState
          variant="card"
          icon={<Archive className="size-8 text-text-muted" />}
          title={t('adminSkills.emptyMarket')}
          description={t('adminSkills.marketHint')}
          action={<div className="flex gap-2"><Button onClick={() => openClawHubDrawer()}>{t('adminSkills.importFromClawhub')}</Button><Button variant="outline" onClick={() => setDrawer('publish')}>{t('adminSkills.publishSkill')}</Button></div>}
        />
      ) : (
        <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
          <Table className="table-fixed">
            <colgroup>
              <col />
              <col className="w-32" />
              <col className="w-32" />
              <col className="w-20" />
              <col className="w-36" />
              <col className="w-16" />
              <col className="w-36" />
            </colgroup>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="px-4">{t('common.skills')}</TableHead>
                <TableHead className="px-4">{translate('skills.source')}</TableHead>
                <TableHead className="px-4">{t('common.category')}</TableHead>
                <TableHead className="px-4 text-right">{t('adminSkills.installs')}</TableHead>
                <TableHead className="px-4">{t('common.status')}</TableHead>
                <TableHead className="px-4 text-right">{t('common.sort')}</TableHead>
                <TableHead className="px-4">{t('adminSkills.actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMarketSkills.map((skill) => (
                <TableRow key={skill.skillId} className="group h-16">
                  <TableCell className="px-4 py-0">
                    <div className="flex items-center gap-3">
                      <Icon name={skill.icon} className="size-5 shrink-0 text-accent" />
                      <div className="min-w-0">
                        <div className="truncate text-[13px] font-medium text-text-1">{skill.name}</div>
                        <div className="mt-1 truncate text-xs text-text-3">{skill.description}</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="px-4 py-0"><ScopeLabel>{getSourceLabel(skill.source)}</ScopeLabel></TableCell>
                  <TableCell className="px-4 py-0">
                    <div className="flex items-center gap-2 text-text-2">
                      <Icon name={skill.categoryIcon} className="size-3.5 shrink-0 text-interactive" />
                      <span className="truncate">{skill.categoryName}</span>
                    </div>
                  </TableCell>
                  <TableCell className="px-4 py-0 text-right tabular-nums">
                    <button
                      onClick={() => openUsersDrawer(skill.skillId)}
                      className="text-text-1 underline-offset-4 hover:underline"
                    >
                      {skill.installedCount}
                    </button>
                  </TableCell>
                  <TableCell className="px-4 py-0">
                    <StatusToggle
                      checked={skill.status === 1}
                      onToggle={(checked) => toggleMarketStatus(skill.skillId, checked)}
                      labelOn={t('adminSkills.published')}
                      labelOff={t('adminSkills.unpublished')}
                    />
                  </TableCell>
                  <TableCell className="px-4 py-0 text-right tabular-nums text-text-3">{skill.sortOrder}</TableCell>
                  <TableCell className="px-4 py-0">
                    <div className="flex justify-end gap-1 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
                      <Button variant="ghost" size="sm" onClick={() => openMarketEditor(skill)}>{t('common.edit')}</Button>
                      <Button variant="ghost" size="sm" onClick={() => openUsersDrawer(skill.skillId)}>{t('common.user')}</Button>
                      <Button variant="ghost" size="icon-sm" aria-label={t('common.delete')} className="text-text-3 hover:text-destructive" onClick={() => setDeleteTarget({ type: 'market', id: skill.skillId, label: skill.name })}>
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </section>
  )
}
export function Component() {
  return <div>Admin Market Skills</div>
}
