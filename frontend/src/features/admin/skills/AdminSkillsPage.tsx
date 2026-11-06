import { useState } from 'react'
import { Settings2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { PageTabs } from '@/components/patterns/PageTabs'
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
      <PageHeader
        title={t('adminSkills.title')}
        description={t('adminSkills.description')}
        actions={
          <Button variant="outline" onClick={() => setDrawer('categories')} className="text-accent hover:text-accent-hover">
            <Settings2 className="size-4" />
            {t('adminSkills.categoryManagement')}
          </Button>
        }
      />

      <div className="px-6 pt-4">
        <PageTabs
          value={activeTab}
          onChange={setActiveTab}
          items={[
            { value: 'global', label: `${t('adminSkills.global')} ${systemSkills.length}` },
            { value: 'market', label: `${t('adminSkills.market')} ${marketSkills.length}` },
          ]}
        />
      </div>

      <PageContainer className="flex flex-1 flex-col overflow-auto">
        {activeTab === 'global' && <AdminSystemSkillsPage />}
        {activeTab === 'market' && <AdminMarketSkillsPage />}
      </PageContainer>

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
