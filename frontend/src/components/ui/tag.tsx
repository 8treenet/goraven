import * as React from 'react'
import { cn } from '@/lib/utils'
import { TONE_BORDER, TONE_SOFT, TONE_TEXT, type Tone } from '@/lib/tone'

interface TagProps extends React.ComponentProps<'span'> {
  tone?: Tone
}

function Tag({ tone = 1, className, ...props }: TagProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium leading-none whitespace-nowrap',
        TONE_SOFT[tone],
        TONE_TEXT[tone],
        TONE_BORDER[tone],
        className,
      )}
      {...props}
    />
  )
}

export { Tag }
export type { TagProps }
