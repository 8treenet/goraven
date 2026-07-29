import { useState } from 'react'
import { Settings2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useT } from '@/i18n'
import { useAdminSkills, AdminSkillsProvider } from './context'
import { AdminSystemSkillsPage } from './AdminSystemSkillsPage'
import { AdminMarketSkillsPage } from './AdminMarketSkillsPage'
import {
  SystemEditorDrawer, MarketEditorDrawer, ClawHubDrawer,
  PublishDrawer, CategoriesDrawer, UsersDrawer, DeleteConfirmDialog,
} from './drawers'
import type { TabKey } from './shared'

function AdminSkillsInner() {
  const t = useT()
  const { systemSkills, marketSkills, setDrawer } = useAdminSkills()
  const [activeTab, setActiveTab] = useState<TabKey>('global')

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <div className="border-b border-border px-6 py-4">
        <div className="flex items-start justify-between gap-6">
          <div>
            <h1 className="text-xl font-semibold text-text-1">{t('adminSkills.title')}</h1>
            <p className="mt-1 text-sm text-text-3">{t('adminSkills.description')}</p>
          </div>
          <Button variant="outline" onClick={() => setDrawer('categories')} className="text-highlight hover:text-highlight">
            <Settings2 className="size-4" />
            {t('adminSkills.categoryManagement')}
          </Button>
        </div>

        <div className="mt-5 inline-flex rounded-lg bg-bg-layer-2 p-1">
          {([
            ['global', t('adminSkills.global'), systemSkills.length],
            ['market', t('adminSkills.market'), marketSkills.length],
          ] as const).map(([key, label, count]) => (
            <button
              key={key}
              onClick={() => setActiveTab(key)}
              className={cn(
                'flex h-8 items-center gap-2 rounded-md px-3 text-sm transition-colors',
                activeTab === key ? 'bg-highlight text-highlight-fg' : 'text-text-3 hover:text-text-1',
              )}
            >
              {label}
              <span className={cn('text-xs', activeTab === key ? 'text-highlight-fg/70' : 'text-text-muted')}>{count}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
        {activeTab === 'global' && <AdminSystemSkillsPage />}
        {activeTab === 'market' && <AdminMarketSkillsPage />}
      </div>

      <SystemEditorDrawer />
      <MarketEditorDrawer />
      <ClawHubDrawer />
      <PublishDrawer />
      <CategoriesDrawer />
      <UsersDrawer />
      <DeleteConfirmDialog />
    </div>
  )
}

export function Component() {
  return (
    <AdminSkillsProvider>
      <AdminSkillsInner />
    </AdminSkillsProvider>
  )
}
