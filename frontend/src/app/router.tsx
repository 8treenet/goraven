import { createBrowserRouter } from 'react-router-dom'

function lazyImport<T>(importer: () => Promise<T>): () => Promise<T> {
  return () =>
    importer().then(
      (mod) => {
        sessionStorage.removeItem('goraven:chunk-reload')
        return mod
      },
      (err) => {
        if (!sessionStorage.getItem('goraven:chunk-reload')) {
          sessionStorage.setItem('goraven:chunk-reload', '1')
          window.location.reload()
          return new Promise<T>(() => {})
        }
        sessionStorage.removeItem('goraven:chunk-reload')
        throw err
      },
    )
}

export const router = createBrowserRouter([
  {
    path: '/login',
    lazy: lazyImport(() => import('@/features/auth/LoginPage')),
  },
  {
    path: '/install',
    lazy: lazyImport(() => import('@/features/auth/InstallPage')),
  },
  {
    path: '/share/:shareId',
    lazy: lazyImport(() => import('@/features/share/SharePage')),
  },
  {
    path: '/',
    lazy: lazyImport(() => import('@/features/layout/MainLayout')),
    children: [
      {
        index: true,
        lazy: lazyImport(() => import('@/features/chat/ChatPage')),
      },
      {
        path: 'chat',
        lazy: lazyImport(() => import('@/features/chat/ChatPage')),
      },
      {
        path: 'chat/:sessionId',
        lazy: lazyImport(() => import('@/features/chat/ChatPage')),
      },
      {
        path: 'dashboard',
        lazy: lazyImport(() => import('@/features/dashboard/DashboardPage')),
      },
      {
        path: 'files',
        lazy: lazyImport(() => import('@/features/files/FilesPage')),
      },
      {
        path: 'skills',
        lazy: lazyImport(() => import('@/features/skills/MySkillsPage')),
      },
      {
        path: 'personas',
        lazy: lazyImport(() => import('@/features/personas/PersonaListPage')),
      },
      {
        path: 'personas/new',
        lazy: lazyImport(() => import('@/features/personas/PersonaEditPage')),
      },
      {
        path: 'personas/:id/edit',
        lazy: lazyImport(() => import('@/features/personas/PersonaEditPage')),
      },
      {
        path: 'personas/:id',
        lazy: lazyImport(() => import('@/features/personas/PersonaDetailPage')),
      },
      {
        path: 'automation',
        lazy: lazyImport(() => import('@/features/automation/AutomationListPage')),
      },
      {
        path: 'automation/:id',
        lazy: lazyImport(() => import('@/features/automation/AutomationDetailPage')),
      },
      {
        path: 'profile',
        lazy: lazyImport(() => import('@/features/profile/ProfilePage')),
      },
      {
        path: 'admin',
        lazy: lazyImport(() => import('@/features/admin/dashboard/AdminDashboardPage')),
      },
      {
        path: 'admin/users',
        lazy: lazyImport(() => import('@/features/admin/users/AdminUsersPage')),
      },
      {
        path: 'admin/models',
        lazy: lazyImport(() => import('@/features/admin/models/AdminModelsPage')),
      },
      {
        path: 'admin/mcp',
        lazy: lazyImport(() => import('@/features/admin/mcp/AdminMcpPage')),
      },
      {
        path: 'admin/skills',
        lazy: lazyImport(() => import('@/features/admin/skills/AdminSkillsPage')),
      },
      {
        path: 'admin/persona-templates',
        lazy: lazyImport(() => import('@/features/admin/persona-templates/AdminPersonaTemplatesPage')),
      },
      {
        path: 'admin/settings',
        lazy: lazyImport(() => import('@/features/admin/settings/AdminSettingsPage')),
      },
      {
        path: 'admin/systemInfo',
        lazy: lazyImport(() => import('@/features/admin/system-info/AdminSystemInfoPage')),
      },
      {
        path: 'admin/shared-projects',
        lazy: lazyImport(() => import('@/features/admin/shared-projects/AdminSharedProjectsPage')),
      },
    ],
  },
])
