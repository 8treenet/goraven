import { Streamdown, type StreamdownProps } from 'streamdown'
import { code } from '@streamdown/code'
import { cn } from '@/lib/utils'
import { GoRavenFile } from './goraven-file'
import { GoRavenChart } from './goraven-chart'

const SHIKI_JS_ENGINE_SUPPORTED = (() => {
  try {
    new RegExp('(?<=a)b')
    return true
  } catch {
    return false
  }
})()

interface MarkdownProps {
  children: string
  mode?: 'static' | 'streaming'
  isAnimating?: boolean
  className?: string
  components?: StreamdownProps['components']
}

const GORAVEN_FILE_RE = /<goraven-file\s+([^>]*?)\/?>/gi
const GORAVEN_CHART_RE = /<goraven-chart\s+([^>]*?)\/?>/gi
const GORAVEN_UPLOAD_RE = /<goraven-upload[^>]*>[\s\S]*?<\/goraven-upload>/gi
const GORAVEN_REF_RE = /<goraven-ref\s+[^>]*>[\s\S]*?<\/goraven-ref>/gi

function parseAttrs(attrStr: string): Record<string, string> {
  const attrs: Record<string, string> = {}
  const re = /(\w+)="([^"]*)"/g
  let m: RegExpExecArray | null
  while ((m = re.exec(attrStr)) !== null) {
    attrs[m[1]] = m[2]
  }
  return attrs
}

interface Segment {
  type: 'md' | 'file' | 'chart'
  content: string
  attrs?: Record<string, string>
}

function splitContent(raw: string): Segment[] {
  const segments: Segment[] = []
  let last = 0

  const tags: { regex: RegExp; type: Segment['type'] }[] = [
    { regex: GORAVEN_FILE_RE, type: 'file' },
    { regex: GORAVEN_CHART_RE, type: 'chart' },
  ]

  type Match = { index: number; end: number; type: Segment['type']; attrs: Record<string, string> }
  const matches: Match[] = []

  for (const { regex, type } of tags) {
    regex.lastIndex = 0
    let m: RegExpExecArray | null
    while ((m = regex.exec(raw)) !== null) {
      matches.push({ index: m.index, end: m.index + m[0].length, type, attrs: parseAttrs(m[1]) })
    }
  }
  matches.sort((a, b) => a.index - b.index)

  for (const match of matches) {
    if (match.index >= last) {
      if (match.index > last) {
        segments.push({ type: 'md', content: raw.slice(last, match.index) })
      }
      segments.push({ type: match.type, content: raw.slice(match.index, match.end), attrs: match.attrs })
      last = match.end
    }
  }

  if (last < raw.length) {
    segments.push({ type: 'md', content: raw.slice(last) })
  }
  return segments
}

function renderStreamdown(
  key: number | string | undefined,
  props: Pick<StreamdownProps, 'mode' | 'isAnimating' | 'plugins' | 'shikiTheme' | 'controls' | 'mermaid' | 'components' | 'className'> & { children: string },
) {
  const { className, ...rest } = props
  return <Streamdown key={key} className={cn('goraven-markdown', className)} {...rest} />
}

export function Markdown({
  children,
  mode = 'streaming',
  isAnimating = false,
  className,
  components,
}: MarkdownProps) {
  // Strip <goraven-upload> tags — server adds these from user uploads but the frontend
  // doesn't render them (files are already shown as attachment previews in the input)
  const cleanChildren = children
    .replace(GORAVEN_UPLOAD_RE, '')
    .replace(GORAVEN_REF_RE, '')
    .trim()

  const hasTag = GORAVEN_FILE_RE.test(cleanChildren) || GORAVEN_CHART_RE.test(cleanChildren)
  GORAVEN_FILE_RE.lastIndex = 0
  GORAVEN_CHART_RE.lastIndex = 0

  const base = {
    mode,
    isAnimating,
    plugins: (SHIKI_JS_ENGINE_SUPPORTED ? { code } : undefined) as StreamdownProps['plugins'],
    shikiTheme: ['github-light', 'github-dark'] as StreamdownProps['shikiTheme'],
    controls: false as StreamdownProps['controls'],
    mermaid: {} as StreamdownProps['mermaid'],
    components,
    className,
  }

  if (!hasTag) {
    return renderStreamdown(undefined, { ...base, children: cleanChildren })
  }

  const segments = splitContent(cleanChildren)
  return (
    <div className={cn('goraven-markdown', className)}>
      {segments.map((seg, i) => {
        if (seg.type === 'file') {
          return (
            <GoRavenFile
              key={i}
              kind={seg.attrs?.kind}
              path={seg.attrs?.path}
              name={seg.attrs?.name}
              description={seg.attrs?.description}
            />
          )
        }
        if (seg.type === 'chart') {
          return <GoRavenChart key={i} attrs={seg.attrs || {}} />
        }
        if (seg.content.trim()) {
          return renderStreamdown(i, { ...base, children: seg.content })
        }
        return null
      })}
    </div>
  )
}
