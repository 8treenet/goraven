import { useState, useCallback } from 'react'
import { Copy, Check, ThumbsUp, ThumbsDown, ChevronDown, ChevronRight, Brain } from 'lucide-react'
import { Markdown } from '@/components/common/markdown'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { useT } from '@/i18n'

export interface ReadonlyThinkingSegment {
  type: 'reasoning' | 'tool'
  content?: string
  tool?: {
    name: string
    displayName: string
    icon: string
    action: string
  }
}

export interface ReadonlyMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinkingSegments?: ReadonlyThinkingSegment[]
  timestamp: string
}

export function ReadonlyUserMessage({ content }: { content: string }) {
  return (
    <div className="flex justify-end">
      <div className="max-w-[75%] rounded-lg bg-interactive-soft px-4 py-2.5">
        <Markdown mode="static" className="text-base leading-relaxed text-text-1 [&_p]:m-0 [&_p]:whitespace-pre-wrap">
          {content}
        </Markdown>
      </div>
    </div>
  )
}

export function ReadonlyAssistantMessage({
  content,
  thinkingSegments,
}: {
  content: string
  thinkingSegments?: ReadonlyThinkingSegment[]
}) {
  const [thinkingOpen, setThinkingOpen] = useState(true)
  const t = useT()
  const [copied, setCopied] = useState(false)
  const [liked, setLiked] = useState<'up' | 'down' | null>(null)

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 1000)
  }, [content])

  const hasThinking = !!thinkingSegments && thinkingSegments.length > 0

  return (
    <div>
      {hasThinking && (
        <div className="mb-4">
          <button
            onClick={() => setThinkingOpen(!thinkingOpen)}
            className="flex items-center gap-2 text-sm text-text-muted transition-colors hover:text-text-3 group mb-2"
          >
            <Brain className="size-[15px] shrink-0" />
            <span>{t('chat.thinkingCompleted')}</span>
            {thinkingOpen ? <ChevronDown className="size-3 shrink-0" /> : <ChevronRight className="size-3 shrink-0" />}
          </button>
          {thinkingOpen && (
            <div className="relative mt-1.5">
              <div className="absolute left-[5px] top-1.5 bottom-0 w-px bg-border-custom" />
              {thinkingSegments!.map((seg, i) => (
                <ReadonlyThinkingSegmentItem key={i} segment={seg} />
              ))}
            </div>
          )}
        </div>
      )}

      {content && (
        <Markdown mode="static" className="text-base leading-relaxed text-text-1">
          {content}
        </Markdown>
      )}

      {content && (
        <div className="mt-2 flex items-center gap-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                className="rounded p-1 text-text-muted transition-colors hover:bg-bg-layer-2 hover:text-text-2"
                onClick={handleCopy}
              >
                {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.copyTooltip')}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                className={cn(
                  'rounded p-1 transition-colors hover:bg-bg-layer-2',
                  liked === 'up' ? 'text-text-1' : 'text-text-muted hover:text-text-2',
                )}
                onClick={() => setLiked(liked === 'up' ? null : 'up')}
              >
                <ThumbsUp className="size-3.5" fill={liked === 'up' ? 'currentColor' : 'none'} />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.likeTooltip')}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                className={cn(
                  'rounded p-1 transition-colors hover:bg-bg-layer-2',
                  liked === 'down' ? 'text-text-1' : 'text-text-muted hover:text-text-2',
                )}
                onClick={() => setLiked(liked === 'down' ? null : 'down')}
              >
                <ThumbsDown className="size-3.5" fill={liked === 'down' ? 'currentColor' : 'none'} />
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('chat.dislikeTooltip')}</TooltipContent>
          </Tooltip>
        </div>
      )}
    </div>
  )
}

function ReadonlyThinkingSegmentItem({ segment }: { segment: ReadonlyThinkingSegment }) {
  if (segment.type === 'reasoning') {
    return (
      <div className="flex gap-3 pb-2.5">
        <div className="shrink-0 w-2.5" />
        <div className="flex-1 min-w-0">
          <Markdown mode="static" className="text-sm leading-relaxed text-text-3">
            {segment.content ?? ''}
          </Markdown>
        </div>
      </div>
    )
  }
  if (segment.type === 'tool' && segment.tool) {
    const tool = segment.tool
    return (
      <div className="flex gap-3 pb-2 last:pb-0">
        <div className="relative shrink-0 pt-2">
          <div className="size-2.5 rounded-full border-2 border-border-custom bg-bg-base" />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 text-sm text-text-3">
            <span className="shrink-0">{tool.icon}</span>
            <span className="font-medium text-text-2">{tool.displayName}</span>
            <span className="truncate">{tool.action}</span>
          </div>
        </div>
      </div>
    )
  }
  return null
}

export function ReadonlyMessageBlock({ message }: { message: ReadonlyMessage }) {
  if (message.role === 'user') {
    return <ReadonlyUserMessage content={message.content} />
  }
  return (
    <ReadonlyAssistantMessage
      content={message.content}
      thinkingSegments={message.thinkingSegments}
    />
  )
}
