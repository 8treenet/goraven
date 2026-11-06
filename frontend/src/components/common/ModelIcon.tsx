import { useState } from 'react'
import { cn } from '@/lib/utils'
import { BRAND_ICON, Logo } from '@/components/brand/Logo'
import { Icon } from './Icon'

interface ModelIconProps {
  icon?: string
  /** Tailwind size classes; defaults to 16px. */
  className?: string
  /** Color classes for the fallback brain glyph. */
  fallbackClassName?: string
}

/**
 * Renders a model icon.
 *
 * The brand mark is drawn inline instead of as `<img src="/favicon.svg">`:
 * the favicon colors itself from `prefers-color-scheme` (the OS theme), which
 * turns it white-on-white when the app forces light on a dark OS. The inline
 * `Logo` follows the app's `.light`/`.dark` tokens.
 */
export function ModelIcon({
  icon,
  className = 'size-4',
  fallbackClassName = 'text-text-2',
}: ModelIconProps) {
  const [err, setErr] = useState(false)
  if (!icon || err) {
    return <Icon name="brain" className={cn(className, 'shrink-0', fallbackClassName)} />
  }
  if (icon === BRAND_ICON) {
    return <Logo className={className} />
  }
  return (
    <img
      src={icon}
      alt=""
      className={cn(className, 'shrink-0 rounded object-cover')}
      onError={() => setErr(true)}
    />
  )
}
