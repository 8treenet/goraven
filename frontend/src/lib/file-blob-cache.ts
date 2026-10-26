import { useUserStore } from '@/stores/user-store'

type CacheEntry = {
  blobUrl: string
  etag: string | null
}

const cache = new Map<string, CacheEntry>()

export async function getCachedBlobUrl(url: string): Promise<{ blobUrl: string; isCached: boolean }> {
  const entry = cache.get(url)
  const token = useUserStore.getState().token
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (entry?.etag) headers['If-None-Match'] = entry.etag

  const res = await fetch(url, { headers, cache: 'no-cache' })

  if (res.status === 304 && entry) {
    return { blobUrl: entry.blobUrl, isCached: true }
  }

  if (!res.ok) throw new Error(`fetch failed: ${res.status}`)

  const newEtag = res.headers.get('ETag')
  if (entry && (!newEtag || entry.etag === newEtag)) {
    return { blobUrl: entry.blobUrl, isCached: true }
  }

  const blob = await res.blob()
  const blobUrl = URL.createObjectURL(blob)

  cache.set(url, { blobUrl, etag: newEtag })

  return { blobUrl, isCached: false }
}

export function evictCachedBlob(url: string): void {
  cache.delete(url)
}

export function peekCachedBlobUrl(url: string): string | null {
  return cache.get(url)?.blobUrl ?? null
}

export function evictCachedBlobsByPrefix(prefix: string): void {
  for (const key of cache.keys()) {
    if (key.startsWith(prefix)) {
      cache.delete(key)
    }
  }
}
