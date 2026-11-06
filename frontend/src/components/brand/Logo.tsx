import { useId } from 'react'
import { cn } from '@/lib/utils'

/** URL of the standalone brand mark used by the virtual "Auto" model. */
export const BRAND_ICON = '/favicon.svg'

/**
 * Inline GoRaven mark. Keep the geometry in sync with `public/favicon.svg`:
 * the favicon must remain a standalone file, while this inline copy can use
 * the app's theme tokens instead of the operating system color scheme.
 */
export function Logo({ size = 28, className }: { size?: number; className?: string }) {
  const eyeMaskId = `raven-eye-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      className={cn('shrink-0 text-text-1', className)}
      aria-hidden="true"
    >
      <defs>
        <mask id={eyeMaskId} maskUnits="userSpaceOnUse" x="0" y="0" width="64" height="64">
          <rect width="64" height="64" fill="white" />
          <circle cx="48" cy="19" r="3" fill="black" />
        </mask>
      </defs>
      <g fill="currentColor" mask={`url(#${eyeMaskId})`}>
        <ellipse cx="32" cy="38" rx="17" ry="14" />
        <circle cx="45" cy="23" r="11.5" />
        <path d="M20 36 2 48l7 7 15-8z" />
      </g>
      <path d="M55 19.5 64 24 55 28.5z" fill="var(--amber)" />
      <path d="M27 50v5M37 50v5" fill="none" stroke="currentColor" strokeWidth="3.4" strokeLinecap="round" />
    </svg>
  )
}
