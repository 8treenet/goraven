import { useEffect, useRef, useState } from 'react'

function formatElapsed(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  if (m > 0) {
    return `${m}m ${s.toString().padStart(2, '0')}s`
  }
  return `${s}s`
}

const LINES = [
  '> analyzing context structure...',
  '> retrieving relevant knowledge...',
  '> evaluating available toolchains...',
  '> parsing user intent...',
  '> loading skill definitions...',
  '> running intermediate inference...',
  '> fetching external references...',
  '> merging sub-agent results...',
  '> compressing context window...',
  '> validating response coherence...',
  '> cross-referencing sources...',
  '> pruning irrelevant context...',
  '> synthesizing final output...',
  '> checking safety constraints...',
  '> formatting structured response...',
]

const PIPELINE = [
  { label: 'context', detail: 'indexing memory' },
  { label: 'tools', detail: 'routing calls' },
  { label: 'skills', detail: 'loading runtime' },
  { label: 'verify', detail: 'checking output' },
] as const

const PIPELINE_POSITIONS = [
  'left-1/2 top-0 -translate-x-1/2 -translate-y-1/2',
  'right-0 top-1/2 translate-x-1/2 -translate-y-1/2',
  'bottom-0 left-1/2 -translate-x-1/2 translate-y-1/2',
  'left-0 top-1/2 -translate-x-1/2 -translate-y-1/2',
] as const

const TELEMETRY = [
  ['mode', 'background'],
  ['runner', 'main-agent'],
  ['stream', 'attached'],
  ['queue', 'active'],
] as const

function pickLine(prev: string[]): string {
  const recent = new Set(prev.slice(-3))
  const pool = LINES.filter((l) => !recent.has(l))
  return pool[Math.floor(Math.random() * pool.length)]
}

function BackgroundThinking({ startTime }: { startTime?: string }) {
  const [elapsed, setElapsed] = useState('0s')
  const [history, setHistory] = useState<string[]>([])
  const [current, setCurrent] = useState('')
  const typingRef = useRef<ReturnType<typeof setInterval>>()
  const nextRef = useRef<ReturnType<typeof setTimeout>>()
  const recentRef = useRef<string[]>([])

  useEffect(() => {
    if (!startTime) return
    const start = new Date(startTime).getTime()
    if (Number.isNaN(start)) return

    const compute = () => {
      const diff = Math.max(0, Math.floor((Date.now() - start) / 1000))
      setElapsed(formatElapsed(diff))
    }

    compute()
    const t = setInterval(compute, 1000)
    return () => clearInterval(t)
  }, [startTime])

  useEffect(() => {
    let cancelled = false

    const typeNext = () => {
      if (cancelled) return

      const line = pickLine(recentRef.current)
      recentRef.current = [...recentRef.current.slice(-2), line]
      let i = 0
      setCurrent('')

      typingRef.current = setInterval(() => {
        if (cancelled) {
          clearInterval(typingRef.current)
          return
        }

        i += 1
        setCurrent(line.slice(0, i))

        if (i >= line.length) {
          clearInterval(typingRef.current)
          typingRef.current = undefined
          setHistory((prev) => [...prev.slice(-6), line])
          setCurrent('')
          nextRef.current = setTimeout(typeNext, 880)
        }
      }, 32)
    }

    typeNext()

    return () => {
      cancelled = true
      clearInterval(typingRef.current)
      clearTimeout(nextRef.current)
    }
  }, [])

  return (
    <div className="my-4 sm:my-8">
      <style>{`
        @keyframes terminal-cursor {
          0%, 100% { opacity: 1; }
          50% { opacity: 0; }
        }

        @keyframes agent-scan {
          0% { transform: translateY(-18%); opacity: 0; }
          18% { opacity: 0.62; }
          100% { transform: translateY(118%); opacity: 0; }
        }

        @keyframes agent-orbit {
          to { transform: rotate(360deg); }
        }

        @keyframes agent-orbit-reverse {
          to { transform: rotate(-360deg); }
        }

        @keyframes agent-pulse {
          0%, 100% { opacity: 0.44; transform: scale(0.96); }
          50% { opacity: 1; transform: scale(1.04); }
        }

        @keyframes agent-flow {
          to { background-position: 128px 0; }
        }

        .agent-signal-grid {
          background-image:
            linear-gradient(var(--border-custom) 1px, transparent 1px),
            linear-gradient(90deg, var(--border-custom) 1px, transparent 1px);
          background-size: 44px 44px;
          mask-image: radial-gradient(circle at 50% 42%, black 0%, transparent 72%);
        }

        .agent-scan-line {
          animation: agent-scan 4.8s cubic-bezier(0.16, 1, 0.3, 1) infinite;
        }

        .agent-orbit {
          animation: agent-orbit 18s linear infinite;
        }

        .agent-orbit-reverse {
          animation: agent-orbit-reverse 28s linear infinite;
        }

        .agent-pulse {
          animation: agent-pulse 2.4s cubic-bezier(0.16, 1, 0.3, 1) infinite;
        }

        .agent-flow {
          background-image: repeating-linear-gradient(90deg, transparent 0 18px, var(--highlight) 18px 26px, transparent 26px 52px);
          background-size: 128px 1px;
          animation: agent-flow 2.8s linear infinite;
        }

        .term-cursor {
          animation: terminal-cursor 1.06s step-end infinite;
        }

        @media (prefers-reduced-motion: reduce) {
          .agent-scan-line,
          .agent-orbit,
          .agent-orbit-reverse,
          .agent-pulse,
          .agent-flow,
          .term-cursor { animation: none !important; }
        }
      `}</style>

      <section
        data-testid="background-thinking-panel"
        className="relative mx-auto flex min-h-[calc(100vh-14rem)] w-full max-w-7xl overflow-hidden rounded-xl border border-border-custom bg-bg-layer-1 text-text-1 shadow-pop"
      >
        <div className="agent-signal-grid pointer-events-none absolute inset-0 opacity-50" />
        <div className="agent-scan-line pointer-events-none absolute inset-x-0 top-0 h-24 bg-highlight/10" />
        <div className="pointer-events-none absolute inset-x-8 top-0 h-px bg-highlight/70" />

        <div className="relative z-10 flex min-h-full w-full flex-col p-4 sm:p-6 lg:p-8">
          <header className="flex flex-col gap-4 border-b border-border-custom pb-5 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div className="mb-3 flex flex-wrap items-center gap-2 font-mono text-[11px] uppercase tracking-[0.28em] text-text-muted">
                <span>BACKGROUND TASK</span>
                <span className="h-px w-8 bg-border-strong" />
                <span>AGENT WORKING</span>
              </div>
              <h2 className="text-2xl font-semibold tracking-[-0.04em] text-text-1 sm:text-3xl">
                Running outside the foreground loop
              </h2>
            </div>

            <div className="grid grid-cols-2 gap-2 font-mono text-xs sm:grid-cols-4 lg:min-w-[420px]">
              {TELEMETRY.map(([label, value]) => (
                <div key={label} className="rounded-lg bg-bg-layer-2 px-3 py-2">
                  <div className="mb-1 text-[10px] uppercase tracking-[0.2em] text-text-muted">{label}</div>
                  <div className="truncate text-text-2">{value}</div>
                </div>
              ))}
            </div>
          </header>

          <div className="grid flex-1 gap-6 py-6 lg:grid-cols-12 lg:py-8">
            <div className="relative min-h-[360px] overflow-hidden rounded-lg bg-bg-base/70 lg:col-span-7">
              <div className="absolute inset-0 border border-border-custom" />
              <div className="absolute left-5 top-5 font-mono text-[11px] uppercase tracking-[0.24em] text-text-muted">
                runtime map
              </div>
              <div className="absolute right-5 top-5 rounded-full bg-highlight px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.18em] text-highlight-fg">
                active run
              </div>

              <div className="absolute left-1/2 top-1/2 size-56 -translate-x-1/2 -translate-y-1/2 rounded-full border border-border-custom sm:size-72">
                <div className="agent-orbit absolute inset-0 rounded-full border border-dashed border-highlight/45" />
                <div className="agent-orbit-reverse absolute -inset-10 rounded-full border border-border-custom" />
                <div className="agent-pulse absolute left-1/2 top-1/2 size-24 -translate-x-1/2 -translate-y-1/2 rounded-lg border border-highlight/60 bg-bg-layer-2 shadow-pop sm:size-28">
                  <div className="flex h-full flex-col items-center justify-center font-mono">
                    <span className="text-[10px] uppercase tracking-[0.28em] text-text-muted">agent</span>
                    <span className="mt-1 text-lg text-highlight">RUN</span>
                  </div>
                </div>

                {PIPELINE.map((step, index) => {
                  return (
                    <div
                      key={step.label}
                      className={`${PIPELINE_POSITIONS[index]} absolute w-28 rounded-lg border border-border-custom bg-bg-layer-1 px-3 py-2 font-mono shadow-soft`}
                    >
                      <div className="mb-1 flex items-center gap-2 text-[10px] uppercase tracking-[0.2em] text-text-muted">
                        <span className="size-1.5 rounded-full bg-highlight" />
                        {step.label}
                      </div>
                      <div className="truncate text-[11px] text-text-3">{step.detail}</div>
                    </div>
                  )
                })}
              </div>

              <div className="absolute inset-x-6 bottom-8 space-y-3">
                {PIPELINE.map((step, index) => (
                  <div key={step.label} className="grid grid-cols-[72px_1fr] items-center gap-3 font-mono text-[11px] text-text-muted">
                    <span className="uppercase tracking-[0.18em]">{step.label}</span>
                    <span className="agent-flow h-px bg-border-custom opacity-80" style={{ animationDelay: `${index * 180}ms` }} />
                  </div>
                ))}
              </div>
            </div>

            <aside className="flex min-h-[360px] flex-col rounded-lg border border-border-custom bg-bg-base/80 lg:col-span-5">
              <div className="flex items-center justify-between border-b border-border-custom px-4 py-3">
                <div className="font-mono text-[11px] uppercase tracking-[0.24em] text-text-muted">live console</div>
                <div className="font-mono text-xs tabular-nums text-highlight">running {elapsed}</div>
              </div>

              <div className="flex-1 px-4 py-4 font-mono text-[13px] leading-relaxed text-text-3">
                {history.map((line, i) => (
                  <div key={`${line}-${i}`} className="text-highlight/65">
                    {line}
                  </div>
                ))}
                {current && <div className="text-highlight">{current}</div>}
                <div className="mt-1 flex items-center gap-1.5">
                  <span className="text-highlight">$</span>
                  <span className="term-cursor inline-block h-[15px] w-2 bg-highlight align-middle" />
                </div>
              </div>

              <div className="grid grid-cols-3 border-t border-border-custom font-mono text-[11px] text-text-muted">
                <div className="px-4 py-3">
                  <div className="mb-1 uppercase tracking-[0.18em]">state</div>
                  <div className="text-text-2">processing</div>
                </div>
                <div className="border-x border-border-custom px-4 py-3">
                  <div className="mb-1 uppercase tracking-[0.18em]">tools</div>
                  <div className="text-text-2">armed</div>
                </div>
                <div className="px-4 py-3">
                  <div className="mb-1 uppercase tracking-[0.18em]">output</div>
                  <div className="text-text-2">pending</div>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </section>
    </div>
  )
}

export { BackgroundThinking }
