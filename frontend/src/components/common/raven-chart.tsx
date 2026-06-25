import { useMemo } from 'react'
import { useT } from '@/i18n'
import {
  LineChart, Line, BarChart, Bar, PieChart, Pie, Cell,
  AreaChart, Area, ScatterChart, Scatter,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts'

const COLORS = [
  'var(--highlight)',
  'var(--interactive)',
  'oklch(0.55 0.03 250)',
  'oklch(0.60 0.02 200)',
  'oklch(0.65 0.04 60)',
]

const CHART_HEIGHT = 280

interface Point {
  name: string
  [key: string]: string | number
}

function parseArray(raw: string | undefined): string[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.map(String)
  } catch {
    return raw.replace(/^\[|\]$/g, '').split(',').map((s) => s.trim().replace(/^['"]|['"]$/g, ''))
  }
  return []
}

function parseNumbers(raw: string | undefined): number[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.map(Number)
  } catch { /* fall through */ }
  return []
}

function buildData(xLabels: string[], series: { key: string; values: number[] }[]): Point[] {
  return xLabels.map((name, i) => {
    const point: Point = { name }
    for (const s of series) {
      point[s.key] = s.values[i] ?? 0
    }
    return point
  })
}

function buildPieData(labels: string[], values: number[]) {
  return labels.map((name, i) => ({ name, value: values[i] ?? 0 }))
}

interface ChartAttrs {
  type: string
  title?: string
  height?: number
  x?: string
  labels?: string
  y1?: string
  y1name?: string
  y2?: string
  y2name?: string
  y3?: string
  y3name?: string
}

function ChartRenderer({ attrs }: { attrs: ChartAttrs }) {
  const t = useT()
  const { type, title, height = CHART_HEIGHT } = attrs

  const data = useMemo(() => {
    if (type === 'pie') {
      const labels = parseArray(attrs.labels ?? attrs.x)
      const values = parseNumbers(attrs.y1)
      return buildPieData(labels, values)
    }
    const labels = parseArray(attrs.x ?? attrs.labels)
    const series: { key: string; values: number[] }[] = []
    if (attrs.y1) series.push({ key: attrs.y1name || t('chart.series1'), values: parseNumbers(attrs.y1) })
    if (attrs.y2) series.push({ key: attrs.y2name || t('chart.series2'), values: parseNumbers(attrs.y2) })
    if (attrs.y3) series.push({ key: attrs.y3name || t('chart.series3'), values: parseNumbers(attrs.y3) })
    return buildData(labels, series)
  }, [attrs, type])

  const seriesKeys = useMemo(() => {
    if (type === 'pie') return []
    return Object.keys(data[0] || {}).filter((k) => k !== 'name')
  }, [data, type])

  const chart = useMemo(() => {
    const isPie = type === 'pie'
    const shared = { data, margin: { top: 8, right: 8, left: 0, bottom: 0 } }

    if (isPie) {
      const pieData = data as unknown as { name: string; value: number }[]
      return (
        <PieChart {...shared}>
          <Pie
            data={pieData}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            innerRadius={0}
            outerRadius="75%"
            label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
            labelLine={{ stroke: 'var(--text-muted)', strokeWidth: 1 }}
          >
            {pieData.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              background: 'var(--bg-layer-2)',
              border: '1px solid var(--border)',
              borderRadius: '6px',
              fontSize: '12px',
              color: 'var(--text-1)',
            }}
          />
        </PieChart>
      )
    }

    const grid = <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" opacity={0.4} />
    const xAxis = <XAxis dataKey="name" tick={{ fontSize: 11, fill: 'var(--text-muted)' }} axisLine={{ stroke: 'var(--border)' }} tickLine={false} />
    const yAxis = <YAxis tick={{ fontSize: 11, fill: 'var(--text-muted)' }} axisLine={false} tickLine={false} width={40} />
    const tooltip = (
      <Tooltip
        contentStyle={{
          background: 'var(--bg-layer-2)',
          border: '1px solid var(--border)',
          borderRadius: '6px',
          fontSize: '12px',
          color: 'var(--text-1)',
        }}
      />
    )

    const seriesElements = seriesKeys.map((key, i) => {
      const color = COLORS[i % COLORS.length]
      const baseProps = { key, dataKey: key, fill: color, stroke: color, isAnimationActive: false }

      switch (type) {
        case 'line':
          return <Line {...baseProps} />
        case 'area':
          return <Area {...baseProps} type="monotone" fillOpacity={0.15} />
        case 'bar':
          return <Bar {...baseProps} radius={[4, 4, 0, 0] as [number, number, number, number]} />
        case 'scatter':
          return <Scatter {...baseProps} />
        default:
          return <Bar {...baseProps} radius={[4, 4, 0, 0] as [number, number, number, number]} />
      }
    })

    switch (type) {
      case 'line':
        return (
          <LineChart {...shared}>
            {grid}{xAxis}{yAxis}{tooltip}{seriesElements}
          </LineChart>
        )
      case 'area':
        return (
          <AreaChart {...shared}>
            {grid}{xAxis}{yAxis}{tooltip}{seriesElements}
          </AreaChart>
        )
      case 'bar':
        return (
          <BarChart {...shared}>
            {grid}{xAxis}{yAxis}{tooltip}{seriesElements}
          </BarChart>
        )
      case 'scatter':
        return (
          <ScatterChart {...shared}>
            {grid}<XAxis dataKey="name" tick={false} />{yAxis}{tooltip}{seriesElements}
          </ScatterChart>
        )
      default:
        return (
          <BarChart {...shared}>
            {grid}{xAxis}{yAxis}{tooltip}{seriesElements}
          </BarChart>
        )
    }
  }, [type, data, seriesKeys])

  return (
    <figure className="my-3">
      {title && <p className="mb-2 text-xs text-text-2">{title}</p>}
      <div className="overflow-hidden rounded-md border border-border bg-bg-layer-1 p-3">
        <div style={{ width: '100%', height }}>
          <ResponsiveContainer>
            {chart}
          </ResponsiveContainer>
        </div>
      </div>
    </figure>
  )
}

export function RavenChart({ attrs }: { attrs: Record<string, string> }) {
  const type = (attrs.type || 'bar').toLowerCase()
  if (!['bar', 'line', 'pie', 'area', 'scatter'].includes(type)) return null

  return (
    <ChartRenderer
      attrs={{
        type,
        title: attrs.title,
        height: attrs.height ? Number(attrs.height) : undefined,
        x: attrs.x,
        labels: attrs.labels,
        y1: attrs.y1,
        y1name: attrs.y1name,
        y2: attrs.y2,
        y2name: attrs.y2name,
        y3: attrs.y3,
        y3name: attrs.y3name,
      }}
    />
  )
}
