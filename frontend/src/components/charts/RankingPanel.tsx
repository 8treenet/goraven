import { useT } from '@/i18n'
import { formatNumber } from '@/lib/format'
import { TONE_BG, type Tone } from '@/lib/tone'
import { cn } from '@/lib/utils'

export interface RankItem {
  name: string
  count: number
}

export function RankingPanel({
  title,
  data,
  tone = 1,
}: {
  title: string
  data: RankItem[]
  tone?: Tone
}) {
  const t = useT()
  const maxCount = Math.max(data[0]?.count ?? 0, 1)

  if (data.length === 0) {
    return (
      <div className="flex flex-1 flex-col px-5 py-5">
        <h3 className="flex shrink-0 items-center gap-1.5 text-[13px] font-semibold text-text-2">
          <span aria-hidden="true" className={cn('size-1.5 rounded-full', TONE_BG[tone])} />
          {title}
        </h3>
        <div className="flex flex-1 items-center justify-center">
          <p className="text-meta text-text-3">{t('common.noData')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col px-5 py-5">
      <h3 className="flex shrink-0 items-center gap-1.5 text-[13px] font-semibold text-text-2">
        <span aria-hidden="true" className={cn('size-1.5 rounded-full', TONE_BG[tone])} />
        {title}
      </h3>
      <div className="mt-3 flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto pr-0.5">
        {data.map((item, i) => (
          <div key={item.name} className="flex h-8 shrink-0 items-center gap-2.5">
            <span className="w-4 shrink-0 text-right text-micro text-text-muted tabular-nums">
              {i + 1}
            </span>
            <span className="w-24 shrink-0 truncate text-meta text-text-2">
              {item.name}
            </span>
            <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-bg-layer-3">
              <div
                className={cn('h-full rounded-full transition-all duration-300', TONE_BG[tone])}
                style={{ width: `${Math.max((item.count / maxCount) * 100, 2)}%` }}
              />
            </div>
            <span className="min-w-8 shrink-0 text-right text-meta text-text-3 tabular-nums">
              {formatNumber(item.count)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
