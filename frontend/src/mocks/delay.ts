/* ============================================
   Mock delay utility — simulates network latency
   Lists: 200–600ms
   Single items: 150–300ms
   Mutations: 300–800ms
   Heavy operations: 800–1500ms
   ============================================ */

/** Generic delay */
export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** List fetch — moderate delay */
export function listDelay(): Promise<void> {
  return delay(250 + Math.random() * 350) // 250–600ms
}

/** Single item fetch — quick */
export function itemDelay(): Promise<void> {
  return delay(150 + Math.random() * 150) // 150–300ms
}

/** Mutation (create/update/delete) — slightly longer */
export function mutationDelay(): Promise<void> {
  return delay(300 + Math.random() * 500) // 300–800ms
}

/** Heavy operation (install, upload, compress) */
export function heavyDelay(): Promise<void> {
  return delay(800 + Math.random() * 700) // 800–1500ms
}

/** File upload — longer */
export function uploadDelay(): Promise<void> {
  return delay(1500 + Math.random() * 1500) // 1500–3000ms
}
