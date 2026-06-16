import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'

export function Component() {
  return (
    <div className="flex h-screen bg-bg-base">
      <Sidebar />
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
