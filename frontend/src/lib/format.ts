export function formatNumber(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

export function formatBytes(bytes: number): string {
  if (bytes >= 1_073_741_824) return (bytes / 1_073_741_824).toFixed(1) + ' GB'
  if (bytes >= 1_048_576) return (bytes / 1_048_576).toFixed(1) + ' MB'
  if (bytes >= 1_024) return (bytes / 1_024).toFixed(1) + ' KB'
  return bytes + ' B'
}

export function formatPercent(n: number): string {
  return n.toFixed(1) + '%'
}

export function formatDiff(n: number): string {
  if (n > 0) return '+' + formatPercent(n * 100)
  if (n < 0) return formatPercent(n * 100)
  return '0%'
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return seconds + 's'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const parts: string[] = []
  if (days > 0) parts.push(days + 'd')
  if (hours > 0) parts.push(hours + 'h')
  if (minutes > 0) parts.push(minutes + 'm')
  return parts.join(' ')
}

export function formatRelativeTime(isoString: string): string {
  const now = Date.now()
  const then = new Date(isoString).getTime()
  const diffMs = now - then
  const diffSeconds = Math.floor(diffMs / 1000)
  if (diffSeconds < 60) return '刚刚'
  const diffMinutes = Math.floor(diffSeconds / 60)
  if (diffMinutes < 60) return diffMinutes + ' 分钟前'
  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) return diffHours + ' 小时前'
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 30) return diffDays + ' 天前'
  return new Date(isoString).toLocaleDateString('zh-CN')
}
