import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import assert from 'node:assert/strict'

const backgroundThinking = readFileSync(new URL('./BackgroundThinking.tsx', import.meta.url), 'utf8')
const messageList = readFileSync(new URL('./MessageList.tsx', import.meta.url), 'utf8')

test('background task state uses a full-bleed mission control layout', () => {
  assert.match(backgroundThinking, /data-testid="background-thinking-panel"/)
  assert.match(backgroundThinking, /min-h-\[calc\(100vh-14rem\)\]/)
  assert.match(backgroundThinking, /BACKGROUND TASK/)
  assert.match(backgroundThinking, /AGENT WORKING/)
  assert.match(backgroundThinking, /agent-signal-grid/)
  assert.match(backgroundThinking, /agent-orbit/)
})

test('message list lets the background task break out of the message column', () => {
  assert.match(messageList, /data-testid="background-thinking-full-bleed"/)
  assert.match(messageList, /<div className="mx-auto max-w-3xl space-y-8">[\s\S]*isGenerating/)
  assert.match(messageList, /<div\s+data-testid="background-thinking-full-bleed"[\s\S]*<BackgroundThinking startTime=\{startTime\}/)
})

test('background task motion respects reduced motion preferences', () => {
  assert.match(backgroundThinking, /@media \(prefers-reduced-motion: reduce\)/)
  assert.match(backgroundThinking, /animation: none !important/)
})
