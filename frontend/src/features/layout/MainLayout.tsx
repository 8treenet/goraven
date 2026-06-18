import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useSidebarStore } from '@/stores/sidebar-store'

function MobileSidebarOverlay() {
  const mobileOpen = useSidebarStore((s) => s.mobileOpen)
  const closeMobile = useSidebarStore((s) => s.closeMobile)
  if (!mobileOpen) return null
  return (
    <div
      onClick={closeMobile}
      className="fixed inset-0 z-40 bg-black/50 md:hidden"
    />
  )
}

export function Component() {
  return (
    <div className="flex h-screen bg-bg-base">
      <div className="hidden md:flex">
        <Sidebar variant="desktop" />
      </div>
      <MobileSidebarOverlay />
      <Sidebar variant="mobile" />
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
