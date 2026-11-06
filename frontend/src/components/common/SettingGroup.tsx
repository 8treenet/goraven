import type { ReactNode } from 'react'
import { Card } from '@/components/ui/card'

interface SettingGroupProps {
  title: string
  children: ReactNode
}

export function SettingGroup({ title, children }: SettingGroupProps) {
  return (
    <Card>
      <h2 className="px-5 pt-4 pb-3 text-section font-semibold text-text-2">{title}</h2>
      <div>{children}</div>
    </Card>
  )
}
