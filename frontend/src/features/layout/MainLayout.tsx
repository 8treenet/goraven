import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'

export function Component() {
  return (
    <div className="flex h-dvh bg-bg-base">
      <div className="hidden md:flex">
        <Sidebar variant="desktop" />
      </div>
      <Sidebar variant="mobile" />
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
