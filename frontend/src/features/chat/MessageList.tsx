import { useState, useCallback } from 'react'
import { Brain, ChevronRight, ChevronDown, Copy, Check, ThumbsUp, ThumbsDown, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useChatStore, type Message, type RetryInfo, type ThinkingSegment } from '@/stores/chat-store'
import { Markdown } from '@/components/common/markdown'
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'
import { useT } from '@/i18n'
import { BackgroundThinking } from './BackgroundThinking'

function MessageList({ messages, isGenerating, isBackground, startTime }: { messages: Message[]; isGenerating: boolean; isBackground: boolean; startTime?: string }) {
  const streamingContent = useChatStore((s) => s.streamingContent)
  const streamingThinkingSegments = useChatStore((s) => s.streamingThinkingSegments)
  const streamingRetry = useChatStore((s) => s.streamingRetry)

  return (
    <div className="px-4 py-6">
      <div className="mx-auto max-w-3xl space-y-8">
        {messages.map((msg) => (
          <MessageBlock key={msg.id} message={msg} />
        ))}
        {isGenerating && (
          <AssistantMessage
            content={streamingContent}
            thinkingSegments={streamingThinkingSegments}
            retry={streamingRetry}
            isStreaming
          />
        )}
      </div>
      {isBackground && (
        <div
          data-testid="background-thinking-full-bleed"
          className="mt-8"
        >
          <BackgroundThinking startTime={startTime} />
        </div>
      )}
    </div>
  )
}

function MessageBlock({ message }: { message: Message }) {
  if (message.role === 'user') {
    return <UserMessage content={message.content} />
  }
  if (message.role === 'assistant') {
    return (
      <AssistantMessage
        content={message.content}
        thinkingSegments={message.thinkingSegments}
      />
    )
  }
  return null
}

function UserMessage({ content }: { content: string }) {
  return (
    <div className="flex justify-end">
      <div className="max-w-[75%] rounded-lg bg-bg-layer-2 px-4 py-2.5">
        <Markdown mode="static" className="text-base leading-relaxed text-text-1 [&_p]:m-0 [&_p]:whitespace-pre-wrap">
          {content}
        </Markdown>
      </div>
    </div>
  )
}

function AssistantMessage({
  content,
  thinkingSegments,
  retry,
  isStreaming,
}: {
  content: string
  thinkingSegments?: ThinkingSegment[]
  retry?: RetryInfo | null
  isStreaming?: boolean
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

  const hasThinking = isStreaming || (thinkingSegments && thinkingSegments.length > 0) || !!retry

  const renderSegment = (seg: ThinkingSegment, i: number) => {
    switch (seg.type) {
      case 'reasoning':
        return (
          <div key={i} className="flex gap-3 pb-2.5">
            <div className="shrink-0 w-2.5" />
            <div className="flex-1 min-w-0">
              <Markdown mode="static" className="text-sm leading-relaxed text-text-3">
                {seg.content}
              </Markdown>
            </div>
          </div>
        )
      case 'tool':
        return (
          <div key={i} className="flex gap-3 pb-2 last:pb-0">
            <div className="relative shrink-0 pt-2">
              <div className="size-2.5 rounded-full border-2 border-border-custom bg-bg-base" />
            </div>
            <div className="flex-1 min-w-0">
              <ToolCallItem toolCall={seg} />
            </div>
          </div>
        )
      case 'retry':
        return (
          <div key={i} className="flex gap-3 pb-1">
            <div className="relative shrink-0 pt-1.5">
              <div className="size-2.5 rounded-full border-2 border-border-custom bg-bg-base" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5 text-xs text-text-muted">
                <Loader2 className="size-3 shrink-0 animate-spin" />
                <span>{t('chat.retrying')}</span>
                <span className="text-text-muted">
                  ({seg.attempt}/{seg.maxRetries}) · {seg.error}
                </span>
              </div>
            </div>
          </div>
        )
      default:
        return null
    }
  }

  const hasExtraRetry = retry && !thinkingSegments?.some(seg => seg.type === 'retry')

  return (
    <div>
      {hasThinking && (
        <div className="mb-4">
          <button
            onClick={() => setThinkingOpen(!thinkingOpen)}
            className="flex items-center gap-2 text-sm text-text-muted transition-colors hover:text-text-3 group mb-2"
          >
            {isStreaming ? (
              <Loader2 className="size-[15px] shrink-0 animate-spin" />
            ) : (
              <Brain className="size-[15px] shrink-0" />
            )}
            <span className={cn(isStreaming && 'animate-pulse')}>
              {isStreaming ? t('chat.thinkingInProgress') : t('chat.thinkingCompleted')}
            </span>
            {thinkingOpen ? <ChevronDown className="size-3 shrink-0" /> : <ChevronRight className="size-3 shrink-0" />}
          </button>
          {thinkingOpen && (
            <div className="relative mt-1.5">
              <div className="absolute left-[5px] top-1.5 bottom-0 w-px bg-border-custom" />
              {(isStreaming ? thinkingSegments?.slice(-20) : thinkingSegments)?.map(renderSegment)}
              {hasExtraRetry && (
                <div className="flex gap-3 pb-1">
                  <div className="relative shrink-0 pt-1.5">
                    <div className="size-2.5 rounded-full border-2 border-border-custom bg-bg-base" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5 text-xs text-text-muted">
                      <Loader2 className="size-3 shrink-0 animate-spin" />
                      <span>{t('chat.retrying')}</span>
                      <span className="text-text-muted">
                        ({retry!.attempt}/{retry!.maxRetries}) · {retry!.error}
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {content && (
        <Markdown
          mode="static"
          className="text-base leading-relaxed text-text-1"
        >
          {content}
        </Markdown>
      )}

      {!isStreaming && content && (
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

function ToolCallItem({ toolCall }: { toolCall: { icon: string, displayName: string, action: string, duration?: number, success?: boolean } }) {
  const { icon, displayName, action, duration, success } = toolCall
  return (
    <div className="flex items-center gap-1.5 text-sm text-text-3">
      <span className="shrink-0">{icon}</span>
      <span className="font-medium text-text-2">{displayName}</span>
      <span className="truncate">{action}</span>
      {duration !== undefined && (
        <span className="ml-auto shrink-0 text-xs text-text-muted tabular-nums">{duration}ms</span>
      )}
      {success !== undefined && (
        <span className={cn('shrink-0 text-xs', success ? 'text-text-3' : 'text-text-muted')}>
          {success ? '✓' : '✗'}
        </span>
      )}
    </div>
  )
}

export { MessageList }
