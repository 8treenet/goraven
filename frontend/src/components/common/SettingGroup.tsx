import type { ReactNode } from 'react'

interface SettingGroupProps {
  title: string
  children: ReactNode
}

export function SettingGroup({ title, children }: SettingGroupProps) {
  return (
    <div className="rounded-lg border border-border bg-bg-layer-1">
      <h2 className="text-xs font-semibold text-text-2 px-5 pt-4 pb-3">{title}</h2>
      <div>
        {children}
      </div>
    </div>
  )
}
