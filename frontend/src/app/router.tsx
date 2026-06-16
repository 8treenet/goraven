import { createBrowserRouter } from 'react-router-dom'

export const router = createBrowserRouter([
  {
    path: '/login',
    lazy: () => import('@/features/auth/LoginPage'),
  },
  {
    path: '/install',
    lazy: () => import('@/features/auth/InstallPage'),
  },
  {
    path: '/share/:shareId',
    lazy: () => import('@/features/share/SharePage'),
  },
  {
    path: '/',
    lazy: () => import('@/features/layout/MainLayout'),
    children: [
      {
        index: true,
        lazy: () => import('@/features/chat/ChatPage'),
      },
      {
        path: 'chat',
        lazy: () => import('@/features/chat/ChatPage'),
      },
      {
        path: 'chat/:sessionId',
        lazy: () => import('@/features/chat/ChatPage'),
      },
      {
        path: 'dashboard',
        lazy: () => import('@/features/dashboard/DashboardPage'),
      },
      {
        path: 'files',
        lazy: () => import('@/features/files/FilesPage'),
      },
      {
        path: 'skills',
        lazy: () => import('@/features/skills/MySkillsPage'),
      },
      {
        path: 'personas',
        lazy: () => import('@/features/personas/PersonaListPage'),
      },
      {
        path: 'personas/new',
        lazy: () => import('@/features/personas/PersonaEditPage'),
      },
      {
        path: 'personas/:id/edit',
        lazy: () => import('@/features/personas/PersonaEditPage'),
      },
      {
        path: 'personas/:id',
        lazy: () => import('@/features/personas/PersonaDetailPage'),
      },
      {
        path: 'profile',
        lazy: () => import('@/features/profile/ProfilePage'),
      },
      {
        path: 'admin',
        lazy: () => import('@/features/admin/dashboard/AdminDashboardPage'),
      },
      {
        path: 'admin/users',
        lazy: () => import('@/features/admin/users/AdminUsersPage'),
      },
      {
        path: 'admin/models',
        lazy: () => import('@/features/admin/models/AdminModelsPage'),
      },
      {
        path: 'admin/mcp',
        lazy: () => import('@/features/admin/mcp/AdminMcpPage'),
      },
      {
        path: 'admin/skills',
        lazy: () => import('@/features/admin/skills/AdminSkillsPage'),
      },
      {
        path: 'admin/persona-templates',
        lazy: () => import('@/features/admin/persona-templates/AdminPersonaTemplatesPage'),
      },
      {
        path: 'admin/settings',
        lazy: () => import('@/features/admin/settings/AdminSettingsPage'),
      },
      {
        path: 'admin/systemInfo',
        lazy: () => import('@/features/admin/system-info/AdminSystemInfoPage'),
      },
    ],
  },
])
