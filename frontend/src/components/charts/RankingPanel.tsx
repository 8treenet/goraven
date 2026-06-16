import { useMemo } from 'react'
import { useT } from '@/i18n'

export interface RankItem {
  name: string
  count: number
}

export function RankingPanel({
  title,
  data,
}: {
  title: string
  data: RankItem[]
}) {
  const t = useT()
  const rankColors = useMemo(
    () => [
      'var(--highlight)',
      'var(--color-interactive)',
      'var(--chart-4)',
    ],
    [],
  )
  const maxCount = data.length > 0 ? data[0].count : 1

  if (data.length === 0) {
    return (
      <div className="flex flex-1 flex-col px-6 py-4">
        <h3 className="shrink-0 text-xs font-semibold text-text-2">{title}</h3>
        <div className="flex flex-1 items-center justify-center">
          <p className="text-xs text-text-3">{t('common.noData')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col px-6 py-4">
      <h3 className="shrink-0 text-xs font-semibold text-text-2">{title}</h3>
      <div className="mt-3 flex flex-1 flex-col justify-evenly overflow-hidden">
        {data.map((item, i) => (
          <div key={item.name} className="flex items-center gap-2">
            <span className="w-4 shrink-0 text-right text-xs text-text-3 tabular-nums">
              {i + 1}
            </span>
            <span className="w-24 shrink-0 truncate text-xs text-text-2">
              {item.name}
            </span>
            <div className="flex h-5 flex-1 overflow-hidden rounded-sm bg-bg-layer-2">
              <div
                className="h-full transition-all duration-300"
                style={{
                  width: `${Math.max((item.count / maxCount) * 100, 2)}%`,
                  backgroundColor: rankColors[i % rankColors.length],
                }}
              />
            </div>
            <span className="w-8 shrink-0 text-right text-xs text-text-1 tabular-nums">
              {item.count}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
