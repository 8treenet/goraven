import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SidebarState {
  collapsed: boolean
  expandedGroups: string[]

  toggleCollapsed: () => void
  toggleGroup: (key: string) => void
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      collapsed: false,
      expandedGroups: ['workspace', 'admin-overview', 'admin-system', 'admin-config'],

      toggleCollapsed: () => set((s) => ({ collapsed: !s.collapsed })),
      toggleGroup: (key) =>
        set((s) => ({
          expandedGroups: s.expandedGroups.includes(key)
            ? s.expandedGroups.filter((k) => k !== key)
            : [...s.expandedGroups, key],
        })),
    }),
    {
      name: 'sidebar-storage',
      partialize: (state) => ({
        collapsed: state.collapsed,
        expandedGroups: state.expandedGroups,
      }),
    },
  ),
)
