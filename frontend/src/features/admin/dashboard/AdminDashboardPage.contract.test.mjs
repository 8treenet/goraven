import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import assert from 'node:assert/strict'

const page = readFileSync(new URL('./AdminDashboardPage.tsx', import.meta.url), 'utf8')

test('active user delta renders the backend percent rate with a percent sign', () => {
  assert.match(page, /formatPercent/)
  assert.match(page, /formatPercent\(data\.activeUsersDiff\)/)
  assert.doesNotMatch(page, /\{data\.activeUsersDiff > 0 \? '\+' : ''\}\{data\.activeUsersDiff\}/)
})
