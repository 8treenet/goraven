import { Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'
import { useAdminSkills } from './context'
import {
  SearchBox, SelectField, ScopeLabel, StatusToggle, EmptyState,
  formatDate,
} from './shared'

export function AdminSystemSkillsPage() {
  const t = useT()
  const {
    systemSearch, setSystemSearch, systemStatus, setSystemStatus,
    filteredSystemSkills, openSystemEditor, setDeleteTarget, toggleSystemStatus,
  } = useAdminSkills()

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <SearchBox value={systemSearch} onChange={setSystemSearch} placeholder={t('adminSkills.searchPlaceholder')} />
          <SelectField value={systemStatus} onChange={setSystemStatus}>
            <option value="all">{t('adminSkills.allStatuses')}</option>
            <option value="0">{t('common.enabled')}</option>
            <option value="1">{t('common.disabled')}</option>
          </SelectField>
        </div>
        <Button onClick={() => openSystemEditor()} className="text-highlight hover:text-highlight">
          <Plus className="size-4" />
          {t('adminSkills.newGlobalSkill')}
        </Button>
      </div>

      {filteredSystemSkills.length === 0 ? (
        <EmptyState
          title={t('adminSkills.noGlobalSkills')}
          description={t('adminSkills.globalSkillHint')}
          action={<Button onClick={() => openSystemEditor()}>{t('adminSkills.newGlobalSkill')}</Button>}
        />
      ) : (
        <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
          <table className="w-full table-fixed text-sm">
            <colgroup>
              <col />
              <col className="w-36" />
              <col className="w-28" />
              <col className="w-32" />
            </colgroup>
            <thead className="text-xs text-text-3">
              <tr className="border-b border-border">
                <th className="px-4 py-3 text-left font-normal">{t('common.skills')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('common.status')}</th>
                <th className="px-4 py-3 text-left font-normal">{t('common.updated')}</th>
                <th className="pl-[26px] pr-4 py-3 text-left font-normal">{t('adminSkills.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {filteredSystemSkills.map((skill) => (
                <tr key={skill.skillId} className="border-b border-border/50 transition-colors last:border-0 hover:bg-bg-hover">
                  <td className="px-4 py-3">
                    <div className="font-medium text-text-1">{skill.description}</div>
                    <div className="mt-1 flex items-center gap-2">
                      <ScopeLabel>{skill.name}</ScopeLabel>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <StatusToggle
                      checked={skill.status === 0}
                      onToggle={(checked) => toggleSystemStatus(skill.skillId, checked)}
                      labelOn={t('common.enabled')}
                      labelOff={t('common.disabled')}
                    />
                  </td>
                  <td className="px-4 py-3 text-text-3">{formatDate(skill.updated)}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="sm" onClick={() => openSystemEditor(skill)}>{t('common.edit')}</Button>
                      <Button variant="ghost" size="icon-sm" onClick={() => setDeleteTarget({ type: 'system', id: skill.skillId, label: skill.name })}>
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
  return <div>Admin System Skills</div>
}
