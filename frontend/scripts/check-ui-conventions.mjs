#!/usr/bin/env node
/**
 * Static UI convention checks for the frontend.
 *
 * Usage:
 *   node scripts/check-ui-conventions.mjs            # report only, exit 0
 *   node scripts/check-ui-conventions.mjs --strict   # exit 1 on violations
 *
 * Scope:
 * - `feature-local-component` covers the shared component names migrated so far.
 * - `hand-written-modal-overlay` covers full-height modal/drawer overlays
 *   (`fixed inset-0` / `fixed inset-y-0` with z-40/50).
 * - `raw-palette` covers Tailwind's numbered palette plus white/black across
 *   `src/`, except ui/ primitives and physical media chrome
 *   (`components/common/goraven-file.tsx`, `features/files/PreviewDialog.tsx`).
 * - `tiny-text` covers `features/` and `patterns/`; chart axes/legends are out
 *   of scope by design.
 * - Dropdown/popover backdrops (z-10), bottom-anchored popovers and the
 *   Phase 4 menu/select work are intentionally out of scope here; see the
 *   convergence spec §9 Phase 4. Do not read a clean run as "no hand-written
 *   popover anywhere".
 */
import { readdirSync, readFileSync } from 'node:fs'
import { join, relative, sep } from 'node:path'

const ROOT = process.cwd()
const SRC = join(ROOT, 'src')

const RULES = [
  {
    id: 'feature-local-component',
    // features/ must not redefine shared components.
    re: /(?:export\s+)?function\s+(Drawer|Skeleton|EmptyView|ErrorView|EmptyState|ErrorState|ConfirmDialog)\b/g,
    include: ['src/features'],
  },
  {
    id: 'hand-written-modal-overlay',
    // Full-height modal/drawer overlays are only allowed inside ui/ and patterns/.
    re: /fixed\s+(?:inset-0|inset-y-0)[^"'\n]*z-(?:40|50)/g,
    include: ['src'],
    exclude: ['src/components/ui', 'src/components/patterns'],
  },
  {
    id: 'raw-palette',
    // 除 ui/ 原语与物理媒体表面外，不允许 Tailwind 原生调色板（含 white/black）。
    re: /\b(?:bg|text|border|ring|from|to|via|fill|stroke|divide|outline)-(?:(?:amber|emerald|sky|rose|violet|indigo|teal|cyan|lime|pink|fuchsia|orange|yellow|green|blue|red|slate|gray|zinc|neutral|stone)-\d{2,3}|white|black)\b/g,
    include: ['src'],
    exclude: [
      'src/components/ui',
      'src/components/common/goraven-file.tsx',
      'src/features/files/PreviewDialog.tsx',
    ],
  },
  {
    id: 'tiny-text',
    // 9/10px 只允许图表/图例；features/patterns 里不允许。
    re: /text-\[(?:9|10)px\]/g,
    include: ['src/features', 'src/components/patterns'],
  },
]

function walk(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name)
    if (entry.isDirectory()) return walk(path)
    return /\.(ts|tsx)$/.test(entry.name) ? [path] : []
  })
}

const files = walk(SRC)

function collect(rule) {
  const hits = []
  for (const file of files) {
    const rel = relative(ROOT, file).split(sep).join('/')
    const included = rule.include.some((p) => rel === p || rel.startsWith(`${p}/`))
    if (!included) continue
    const excluded = (rule.exclude ?? []).some((p) => rel === p || rel.startsWith(`${p}/`))
    if (excluded) continue

    const text = readFileSync(file, 'utf8')
    const re = new RegExp(rule.re.source, rule.re.flags)
    let match
    while ((match = re.exec(text)) !== null) {
      const line = text.slice(0, match.index).split('\n').length
      hits.push({ file: rel, line, text: match[0] })
    }
  }
  return hits
}

const strict = process.argv.includes('--strict')
let total = 0

for (const rule of RULES) {
  const hits = collect(rule)
  total += hits.length
  console.log(`${rule.id}: ${hits.length}`)
  for (const hit of hits) {
    console.log(`  ${hit.file}:${hit.line}  ${hit.text}`)
  }
}

if (strict && total > 0) {
  console.error(`\nUI conventions check failed: ${total} violation(s).`)
  process.exit(1)
}
