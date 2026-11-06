import { Plus, Trash2, Archive } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useT } from '@/i18n'
import { useAdminSkills } from './context'
import {
  SearchBox, SelectField, ScopeLabel, StatusToggle,
  formatDate,
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
        <Button onClick={() => openSystemEditor()}>
          <Plus className="size-4" />
          {t('adminSkills.newGlobalSkill')}
        </Button>
      </div>

      {filteredSystemSkills.length === 0 ? (
        <EmptyState
          variant="card"
          icon={<Archive className="size-8 text-text-muted" />}
          title={t('adminSkills.noGlobalSkills')}
          description={t('adminSkills.globalSkillHint')}
          action={<Button onClick={() => openSystemEditor()}>{t('adminSkills.newGlobalSkill')}</Button>}
        />
      ) : (
        <div className="overflow-hidden rounded-lg border border-border bg-bg-layer-1">
          <Table className="table-fixed">
            <colgroup>
              <col />
              <col className="w-36" />
              <col className="w-28" />
              <col className="w-32" />
            </colgroup>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="px-4">{t('common.skills')}</TableHead>
                <TableHead className="px-4">{t('common.status')}</TableHead>
                <TableHead className="px-4">{t('common.updated')}</TableHead>
                <TableHead className="px-4">{t('adminSkills.actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSystemSkills.map((skill) => (
                <TableRow key={skill.skillId} className="group h-16">
                  <TableCell className="px-4 py-0">
                    <div className="truncate text-[13px] font-medium text-text-1">{skill.description}</div>
                    <div className="mt-1 flex items-center gap-2">
                      <ScopeLabel>{skill.name}</ScopeLabel>
                    </div>
                  </TableCell>
                  <TableCell className="px-4 py-0">
                    <StatusToggle
                      checked={skill.status === 0}
                      onToggle={(checked) => toggleSystemStatus(skill.skillId, checked)}
                      labelOn={t('common.enabled')}
                      labelOff={t('common.disabled')}
                    />
                  </TableCell>
                  <TableCell className="px-4 py-0 text-text-3">{formatDate(skill.updated)}</TableCell>
                  <TableCell className="px-4 py-0">
                    <div className="flex justify-end gap-1 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
                      <Button variant="ghost" size="sm" onClick={() => openSystemEditor(skill)}>{t('common.edit')}</Button>
                      <Button variant="ghost" size="icon-sm" aria-label={t('common.delete')} className="text-text-3 hover:text-destructive" onClick={() => setDeleteTarget({ type: 'system', id: skill.skillId, label: skill.name })}>
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
  return <div>Admin System Skills</div>
}
