import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import assert from 'node:assert/strict'

const modelIcon = readFileSync(new URL('./ModelIcon.tsx', import.meta.url), 'utf8')

test('brand model icon follows the in-app theme instead of the OS color scheme', () => {
  // /favicon.svg colors itself from prefers-color-scheme, which tracks the OS,
  // not the app's .light/.dark class. Rendering it in an <img> makes the
  // "Auto" icon white-on-white when the app forces light on a dark OS.
  assert.match(modelIcon, /BRAND_ICON/)
  assert.match(modelIcon, /icon === BRAND_ICON/)
  assert.match(modelIcon, /<Logo\b/)
})
