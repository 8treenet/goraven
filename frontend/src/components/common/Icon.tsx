import { getIconComponent, DEFAULT_ICON, type IconName } from './icon-registry'

export interface IconProps {
  /** Icon name from the registry, e.g. 'code', 'bot'. Falls back to Bot. */
  name: string
  className?: string
}

/**
 * Universal icon renderer.
 *
 * <Icon name="code" className="size-4" />
 * <Icon name="unknown" className="size-4" />  → renders Bot as fallback
 */
export function Icon({ name, className }: IconProps) {
  const Component = getIconComponent(name)
  return <Component className={className} />
}

/**
 * Convenience: render an icon inline with a default size.
 * Returns a React element, not a component, so you can use it in expressions:
 *
 *   {renderIcon('code')}
 *   {renderIcon('bot', 'size-3.5 text-text-3')}
 */
export function renderIcon(name: string, className = 'size-4') {
  const Component = getIconComponent(name)
  return <Component className={className} />
}

export { type IconName, DEFAULT_ICON }
