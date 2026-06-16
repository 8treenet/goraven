import { Streamdown, type StreamdownProps } from 'streamdown'
import { code } from '@streamdown/code'
import { cn } from '@/lib/utils'
import { RavenFile } from './raven-file'
import { RavenChart } from './raven-chart'

interface MarkdownProps {
  children: string
  mode?: 'static' | 'streaming'
  isAnimating?: boolean
  className?: string
  components?: StreamdownProps['components']
}

const RAVEN_FILE_RE = /<raven-file\s+([^>]*?)\/?>/gi
const RAVEN_CHART_RE = /<raven-chart\s+([^>]*?)\/?>/gi
const RAVEN_UPLOAD_RE = /<raven-upload[^>]*>[\s\S]*?<\/raven-upload>/gi

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
    { regex: RAVEN_FILE_RE, type: 'file' },
    { regex: RAVEN_CHART_RE, type: 'chart' },
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
  return <Streamdown key={key} className={cn('raven-markdown', className)} {...rest} />
}

export function Markdown({
  children,
  mode = 'streaming',
  isAnimating = false,
  className,
  components,
}: MarkdownProps) {
  // Strip <raven-upload> tags — server adds these from user uploads but the frontend
  // doesn't render them (files are already shown as attachment previews in the input)
  const cleanChildren = children.replace(RAVEN_UPLOAD_RE, '').trim()

  const hasTag = RAVEN_FILE_RE.test(cleanChildren) || RAVEN_CHART_RE.test(cleanChildren)
  RAVEN_FILE_RE.lastIndex = 0
  RAVEN_CHART_RE.lastIndex = 0

  const base = {
    mode,
    isAnimating,
    plugins: { code } as StreamdownProps['plugins'],
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
    <div className={cn('raven-markdown', className)}>
      {segments.map((seg, i) => {
        if (seg.type === 'file') {
          return (
            <RavenFile
              key={i}
              kind={seg.attrs?.kind}
              path={seg.attrs?.path}
              name={seg.attrs?.name}
              description={seg.attrs?.description}
            />
          )
        }
        if (seg.type === 'chart') {
          return <RavenChart key={i} attrs={seg.attrs || {}} />
        }
        if (seg.content.trim()) {
          return renderStreamdown(i, { ...base, children: seg.content })
        }
        return null
      })}
    </div>
  )
}
