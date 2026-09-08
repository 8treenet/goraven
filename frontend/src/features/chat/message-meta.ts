const pad = (n: number) => String(n).padStart(2, '0')

/** 解析后端返回的创建时间（"2006-01-02 15:04:05"，服务器本地时区）：
 *  统一展示为 "YY-MM-DD HH:mm"（年份两位，如 2026 → 26）；
 *  mock 的 "HH:mm" 原样返回；解析失败返回 null。 */
export function formatMsgTime(created?: string): string | null {
  if (!created) return null
  if (/^\d{1,2}:\d{2}$/.test(created)) return created
  const d = new Date(created.replace(' ', 'T'))
  if (isNaN(d.getTime())) return null
  const yy = pad(d.getFullYear() % 100)
  const md = `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  return `${yy}-${md} ${hm}`
}

/** 毫秒耗时 → 展示文案，如 "800ms"、"12.4s"、"3m 20s" */
export function formatDurationLabel(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1).replace(/\.0$/, '')}s`
  const min = Math.floor(ms / 60000)
  const sec = Math.round((ms % 60000) / 1000)
  return sec > 0 ? `${min}m ${sec}s` : `${min}m`
}
