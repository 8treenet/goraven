export type Tone = 1 | 2 | 3 | 4 | 5

export const TONE_TEXT: Record<Tone, string> = {
  1: 'text-cat-1',
  2: 'text-cat-2',
  3: 'text-cat-3',
  4: 'text-cat-4',
  5: 'text-cat-5',
}

export const TONE_SOFT: Record<Tone, string> = {
  1: 'bg-cat-1-soft',
  2: 'bg-cat-2-soft',
  3: 'bg-cat-3-soft',
  4: 'bg-cat-4-soft',
  5: 'bg-cat-5-soft',
}

export const TONE_BG: Record<Tone, string> = {
  1: 'bg-cat-1',
  2: 'bg-cat-2',
  3: 'bg-cat-3',
  4: 'bg-cat-4',
  5: 'bg-cat-5',
}

export const TONE_BORDER: Record<Tone, string> = {
  1: 'border-cat-1/30',
  2: 'border-cat-2/30',
  3: 'border-cat-3/30',
  4: 'border-cat-4/30',
  5: 'border-cat-5/30',
}

/**
 * Stable tone for a category / provider / file key.
 * Do not use names or array indexes: the same object must keep its color.
 */
export function toneFromKey(key: string | number): Tone {
  const value = String(key)
  let hash = 0
  for (let i = 0; i < value.length; i++) {
    hash = (hash * 31 + value.charCodeAt(i)) >>> 0
  }
  return ((hash % 5) + 1) as Tone
}
