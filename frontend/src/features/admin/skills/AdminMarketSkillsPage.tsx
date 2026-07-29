import { Upload, Download, Trash2 } from 'lucide-react'
import { Icon } from '@/components/common/Icon'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import { useAdminSkills } from './context'
import {
  SearchBox, SelectField, ScopeLabel, StatusToggle, EmptyState,
  getSourceLabel,
} from './shared'

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
          <Button variant="outline" onClick={() => setDrawer('publish')} className="text-highlight hover:text-highlight">
            <Upload className="size-4" />
            {t('adminSkills.publishSkill')}
          </Button>
          <Button onClick={() => openClawHubDrawer()} className="text-highlight hover:text-highlight">
            <Download className="size-4" />
            {t('adminSkills.importFromClawhub')}
          </Button>
        </div>
      </div>

      {filteredMarketSkills.length === 0 ? (
        <EmptyState
          title={t('adminSkills.emptyMarket')}
          description={t('adminSkills.marketHint')}
          action={<div className="flex gap-2"><Button onClick={() => openClawHubDrawer()}>{t('adminSkills.importFromClawhub')}</Button><Button variant="outline" onClick={() => setDrawer('publish')}>{t('adminSkills.publishSkill')}</Button></div>}
        />
      ) : (
        <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
          <table className="w-full table-fixed text-sm">
            <colgroup>
              <col />
              <col className="w-32" />
              <col className="w-32" />
              <col className="w-20" />
              <col className="w-36" />
              <col className="w-16" />
              <col className="w-36" />
            </colgroup>
            <thead className="text-xs text-text-3">
              <tr className="border-b border-border">
                <th className="px-4 py-3 text-left font-normal">{t('common.skills')}</th>
                <th className="px-4 py-3 text-left font-normal">{translate('skills.source')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('common.category')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('adminSkills.installs')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('common.status')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('common.sort')}</th>
                <th className="pl-[26px] pr-4 py-3 text-left font-normal">{t('adminSkills.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {filteredMarketSkills.map((skill) => (
                <tr key={skill.skillId} className="border-b border-border/50 transition-colors last:border-0 hover:bg-bg-hover">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <Icon name={skill.icon} className="size-5 shrink-0 text-interactive" />
                      <div className="min-w-0">
                        <div className="font-medium text-text-1">{skill.name}</div>
                        <div className="mt-1 truncate text-xs text-text-3">{skill.description}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3"><ScopeLabel>{getSourceLabel(skill.source)}</ScopeLabel></td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 text-text-2">
                      <Icon name={skill.categoryIcon} className="size-3.5 shrink-0 text-interactive" />
                      <span className="truncate">{skill.categoryName}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => openUsersDrawer(skill.skillId)}
                      className="text-text-1 underline-offset-4 hover:underline"
                    >
                      {skill.installedCount}
                    </button>
                  </td>
                  <td className="px-4 py-3">
                    <StatusToggle
                      checked={skill.status === 1}
                      onToggle={(checked) => toggleMarketStatus(skill.skillId, checked)}
                      labelOn={t('adminSkills.published')}
                      labelOff={t('adminSkills.unpublished')}
                    />
                  </td>
                  <td className="px-4 py-3 text-text-3">{skill.sortOrder}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="sm" onClick={() => openMarketEditor(skill)}>{t('common.edit')}</Button>
                      <Button variant="ghost" size="sm" onClick={() => openUsersDrawer(skill.skillId)}>{t('common.user')}</Button>
                      <Button variant="ghost" size="icon-sm" onClick={() => setDeleteTarget({ type: 'market', id: skill.skillId, label: skill.name })}>
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
export function Component() {
  return <div>Admin Market Skills</div>
}
