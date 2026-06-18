import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SidebarState {
  collapsed: boolean
  expandedGroups: string[]
  mobileOpen: boolean

  toggleCollapsed: () => void
  toggleGroup: (key: string) => void
  openMobile: () => void
  closeMobile: () => void
  toggleMobile: () => void
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      collapsed: false,
      expandedGroups: ['workspace', 'admin-overview', 'admin-system', 'admin-config'],
      mobileOpen: false,

      toggleCollapsed: () => set((s) => ({ collapsed: !s.collapsed })),
      toggleGroup: (key) =>
        set((s) => ({
          expandedGroups: s.expandedGroups.includes(key)
            ? s.expandedGroups.filter((k) => k !== key)
            : [...s.expandedGroups, key],
        })),
      openMobile: () => set({ mobileOpen: true }),
      closeMobile: () => set({ mobileOpen: false }),
      toggleMobile: () => set((s) => ({ mobileOpen: !s.mobileOpen })),
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
