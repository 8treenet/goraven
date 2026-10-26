import { useMemo } from 'react'
import { cn } from '@/lib/utils'

function lineClass(line: string): string {
  if (line.startsWith('+++') || line.startsWith('---')) return 'text-text-3'
  if (line.startsWith('+')) return 'bg-success/10 text-success'
  if (line.startsWith('-')) return 'bg-destructive/10 text-destructive'
  if (line.startsWith('@@')) return 'text-info'
  return 'text-text-2'
}

interface GitDiffViewProps {
  diff: string
  emptyHint: string
}

export function GitDiffView({ diff, emptyHint }: GitDiffViewProps) {
  const lines = useMemo(() => diff.split('\n'), [diff])
  if (!diff.trim()) {
    return <p className="rounded-md border border-border bg-bg-layer-2 px-3 py-4 text-center text-xs text-text-3">{emptyHint}</p>
  }
  return (
    <div className="max-h-72 overflow-auto rounded-md border border-border bg-bg-layer-2">
      <pre className="w-max min-w-full font-mono text-xs leading-5">
        {lines.map((line, i) => (
          <div key={i} className={cn('px-2', lineClass(line))}>{line || ' '}</div>
        ))}
      </pre>
    </div>
  )
}
