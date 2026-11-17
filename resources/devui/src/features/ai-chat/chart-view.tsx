import { useId, useState } from "react"

export interface ChartSeries {
  name: string
  data: number[]
}

export interface ChartViewProps {
  chartType: "line" | "bar" | "pie" | "metric_card"
  title: string
  xAxis?: string[]
  series: ChartSeries[]
  summary?: string
}

interface ChartHover {
  index: number
  x: number
  y: number
  title: string
  rows: { name: string; value: number; color: string }[]
}

const PALETTE = [
  "#6366f1",
  "#10b981",
  "#f59e0b",
  "#ec4899",
  "#06b6d4",
  "#8b5cf6",
]

const WIDTH = 720
const HEIGHT = 240

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function formatNumber(value: number) {
  if (!Number.isFinite(value)) return "—"
  return new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 2 }).format(
    value
  )
}

function niceMax(value: number) {
  if (!Number.isFinite(value) || value <= 0) return 1
  const base = Math.pow(10, Math.floor(Math.log10(value)))
  const digit = value / base
  const nice = digit <= 1 ? 1 : digit <= 2 ? 2 : digit <= 5 ? 5 : 10
  return nice * base
}

function niceScale(value: number, targetTicks = 4) {
  if (!Number.isFinite(value) || value <= 0) {
    return { max: 1, ticks: 1 }
  }
  const step = niceMax(value / targetTicks)
  const ticks = Math.max(1, Math.ceil(value / step))
  return { max: step * ticks, ticks }
}

function hasData(series: ChartSeries[]) {
  return series.some((item) => item.data.some((value) => Number.isFinite(value)))
}

export function ChartView({
  chartType,
  title,
  xAxis = [],
  series,
  summary,
}: ChartViewProps) {
  const uid = useId().replace(/[^a-zA-Z0-9_-]/g, "")

  return (
    <div className="bg-card overflow-hidden rounded-md border">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-border/60 bg-muted/30 px-4 py-3">
        <div className="min-w-0">
          <div className="font-sans text-sm font-semibold tracking-tight">
            {title}
          </div>
          {summary && (
            <p className="text-muted-foreground mt-0.5 font-sans text-xs leading-relaxed">
              {summary}
            </p>
          )}
        </div>
        {chartType !== "metric_card" && series.length > 1 && (
          <Legend series={series} />
        )}
      </div>

      <div className="px-3 pt-4 pb-3">
        {hasData(series) ? (
          chartType === "metric_card" ? (
            <MetricCards series={series} />
          ) : chartType === "pie" ? (
            <PieChart uid={uid} xAxis={xAxis} series={series} />
          ) : chartType === "bar" ? (
            <BarChart uid={uid} xAxis={xAxis} series={series} />
          ) : (
            <LineChart uid={uid} xAxis={xAxis} series={series} />
          )
        ) : (
          <div className="text-muted-foreground flex h-[160px] items-center justify-center font-sans text-xs">
            暂无数据
          </div>
        )}
      </div>
    </div>
  )
}

function Legend({ series }: { series: ChartSeries[] }) {
  return (
    <div className="flex flex-wrap items-center justify-end gap-1.5">
      {series.map((item, index) => (
        <span
          key={item.name}
          className="bg-background/80 inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 font-sans text-[11px]"
        >
          <span
            className="size-2 rounded-full"
            style={{ background: PALETTE[index % PALETTE.length] }}
          />
          {item.name}
        </span>
      ))}
    </div>
  )
}

function ChartTooltip({
  hover,
  width,
  height,
}: {
  hover: ChartHover
  width: number
  height: number
}) {
  const margin = Math.min(40, (70 / width) * 100)
  const left = clamp((hover.x / width) * 100, margin, 100 - margin)
  const below = hover.y < height * 0.35

  return (
    <div
      className="bg-popover text-popover-foreground pointer-events-none absolute z-10 min-w-28 rounded-md border px-2.5 py-2 shadow-md"
      style={{
        left: `${left}%`,
        top: `${(hover.y / height) * 100}%`,
        transform: below
          ? "translate(-50%, 10px)"
          : "translate(-50%, calc(-100% - 10px))",
      }}
    >
      <div className="mb-1 font-sans text-[11px] font-medium">
        {hover.title}
      </div>
      <div className="grid gap-1">
        {hover.rows.map((row) => (
          <div
            key={row.name}
            className="flex items-center gap-2 font-sans text-[11px]"
          >
            <span
              className="size-2 shrink-0 rounded-full"
              style={{ background: row.color }}
            />
            <span className="text-muted-foreground">{row.name}</span>
            <span className="ml-auto font-semibold tabular-nums">
              {formatNumber(row.value)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

function LineChart({
  uid,
  xAxis,
  series,
}: {
  uid: string
  xAxis: string[]
  series: ChartSeries[]
}) {
  const [hover, setHover] = useState<ChartHover | null>(null)
  const padding = { left: 38, right: 10, top: 10, bottom: 26 }
  const innerW = WIDTH - padding.left - padding.right
  const innerH = HEIGHT - padding.top - padding.bottom
  const count = Math.max(...series.map((item) => item.data.length), 1)
  const { max, ticks } = niceScale(
    Math.max(...series.flatMap((item) => item.data.filter(Number.isFinite)))
  )
  const xAt = (index: number) =>
    count <= 1
      ? padding.left + innerW / 2
      : padding.left + (index / (count - 1)) * innerW
  const yAt = (value: number) =>
    padding.top + innerH - (Math.max(0, value) / max) * innerH
  const labelStep = Math.ceil(xAxis.length / 7)

  const handleMove = (
    clientX: number,
    rect: DOMRect
  ) => {
    const px = ((clientX - rect.left) / rect.width) * WIDTH
    const ratio = count <= 1 ? 0 : (px - padding.left) / innerW
    const index = clamp(Math.round(ratio * (count - 1)), 0, count - 1)
    setHover((prev) => {
      if (prev?.index === index) return prev
      const rows = series.map((item, itemIndex) => ({
        name: item.name,
        value: item.data[index] ?? 0,
        color: PALETTE[itemIndex % PALETTE.length],
      }))
      const top = Math.min(
        ...series.map((item) => yAt(item.data[index] ?? 0))
      )
      return {
        index,
        x: xAt(index),
        y: top,
        title: xAxis[index] ?? `#${index + 1}`,
        rows,
      }
    })
  }

  return (
    <div className="relative">
      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        className="h-auto w-full cursor-crosshair font-sans"
        role="img"
        aria-label={series.map((item) => item.name).join("、")}
        onMouseMove={(event) =>
          handleMove(
            event.clientX,
            event.currentTarget.getBoundingClientRect()
          )
        }
        onMouseLeave={() => setHover(null)}
      >
        <defs>
          {series.map((_, index) => (
            <linearGradient
              key={`${uid}-area-${index}`}
              id={`${uid}-area-${index}`}
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop
                offset="0%"
                stopColor={PALETTE[index % PALETTE.length]}
                stopOpacity={0.24}
              />
              <stop
                offset="100%"
                stopColor={PALETTE[index % PALETTE.length]}
                stopOpacity={0}
              />
            </linearGradient>
          ))}
        </defs>

        {Array.from({ length: ticks + 1 }, (_, index) => {
          const value = (max / ticks) * index
          const y = yAt(value)
          return (
            <g key={index}>
              <line
                x1={padding.left}
                x2={WIDTH - padding.right}
                y1={y}
                y2={y}
                className="stroke-border"
                strokeDasharray={index === 0 ? undefined : "3 5"}
              />
              <text
                x={padding.left - 8}
                y={y + 3}
                textAnchor="end"
                className="fill-muted-foreground text-[10px]"
              >
                {formatNumber(value)}
              </text>
            </g>
          )
        })}

        {xAxis.map((label, index) => {
          if (index % labelStep !== 0) return null
          return (
            <text
              key={`${label}-${index}`}
              x={xAt(index)}
              y={HEIGHT - 8}
              textAnchor="middle"
              className="fill-muted-foreground text-[10px]"
            >
              {label}
            </text>
          )
        })}

        {hover && (
          <line
            x1={hover.x}
            x2={hover.x}
            y1={padding.top}
            y2={padding.top + innerH}
            className="stroke-border"
            strokeDasharray="3 4"
            pointerEvents="none"
          />
        )}

        {series.map((item, index) => {
          const color = PALETTE[index % PALETTE.length]
          const points = item.data.map((value, i) => ({
            x: xAt(i),
            y: yAt(value),
          }))
          const path = smoothPath(points)
          if (!path) return null
          const first = points[0]
          const last = points[points.length - 1]
          const area =
            series.length === 1 && points.length > 1
              ? `${path} L ${last.x} ${padding.top + innerH} L ${first.x} ${padding.top + innerH} Z`
              : ""

          return (
            <g key={item.name}>
              {area && (
                <path
                  d={area}
                  fill={`url(#${uid}-area-${index})`}
                  stroke="none"
                />
              )}
              <path
                d={path}
                fill="none"
                stroke={color}
                strokeWidth={2.5}
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              {points.length <= 16 &&
                points.map((point, pointIndex) => (
                  <circle
                    key={pointIndex}
                    cx={point.x}
                    cy={point.y}
                    r={3}
                    fill={color}
                    stroke="var(--card)"
                    strokeWidth={2}
                  />
                ))}
            </g>
          )
        })}

        {hover &&
          series.map((item, index) => (
            <circle
              key={item.name}
              cx={xAt(hover.index)}
              cy={yAt(item.data[hover.index] ?? 0)}
              r={4.5}
              fill={PALETTE[index % PALETTE.length]}
              stroke="var(--card)"
              strokeWidth={2}
              pointerEvents="none"
            />
          ))}
      </svg>
      {hover && <ChartTooltip hover={hover} width={WIDTH} height={HEIGHT} />}
    </div>
  )
}

function smoothPath(points: { x: number; y: number }[]) {
  if (points.length === 0) return ""
  if (points.length === 1) {
    return `M ${points[0].x} ${points[0].y}`
  }

  let path = `M ${points[0].x} ${points[0].y}`
  for (let index = 0; index < points.length - 1; index++) {
    const p0 = points[index - 1] ?? points[index]
    const p1 = points[index]
    const p2 = points[index + 1]
    const p3 = points[index + 2] ?? p2
    const tension = 0.18
    const c1x = p1.x + (p2.x - p0.x) * tension
    const c1y = p1.y + (p2.y - p0.y) * tension
    const c2x = p2.x - (p3.x - p1.x) * tension
    const c2y = p2.y - (p3.y - p1.y) * tension
    path += ` C ${c1x} ${c1y}, ${c2x} ${c2y}, ${p2.x} ${p2.y}`
  }

  return path
}

function BarChart({
  uid,
  xAxis,
  series,
}: {
  uid: string
  xAxis: string[]
  series: ChartSeries[]
}) {
  const [hover, setHover] = useState<ChartHover | null>(null)
  const padding = { left: 38, right: 10, top: 10, bottom: 26 }
  const innerW = WIDTH - padding.left - padding.right
  const innerH = HEIGHT - padding.top - padding.bottom
  const count = Math.max(
    xAxis.length,
    ...series.map((item) => item.data.length),
    1
  )
  const { max, ticks } = niceScale(
    Math.max(...series.flatMap((item) => item.data.filter(Number.isFinite)))
  )
  const groupWidth = innerW / count
  const barWidth = Math.max(4, (groupWidth * 0.62) / Math.max(series.length, 1))
  const yAt = (value: number) =>
    padding.top + innerH - (Math.max(0, value) / max) * innerH
  const labelStep = Math.ceil(xAxis.length / 8)

  const showGroup = (index: number) => {
    const rows = series.map((item, itemIndex) => ({
      name: item.name,
      value: item.data[index] ?? 0,
      color: PALETTE[itemIndex % PALETTE.length],
    }))
    const top = Math.min(
      ...series.map((item) => yAt(item.data[index] ?? 0))
    )
    setHover({
      index,
      x: padding.left + groupWidth * (index + 0.5),
      y: top,
      title: xAxis[index] ?? `#${index + 1}`,
      rows,
    })
  }

  return (
    <div className="relative">
      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        className="h-auto w-full font-sans"
        role="img"
        aria-label={series.map((item) => item.name).join("、")}
      >
        <defs>
          {series.map((_, index) => (
            <linearGradient
              key={`${uid}-bar-${index}`}
              id={`${uid}-bar-${index}`}
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop
                offset="0%"
                stopColor={PALETTE[index % PALETTE.length]}
                stopOpacity={0.95}
              />
              <stop
                offset="100%"
                stopColor={PALETTE[index % PALETTE.length]}
                stopOpacity={0.55}
              />
            </linearGradient>
          ))}
        </defs>

        {Array.from({ length: ticks + 1 }, (_, index) => {
          const value = (max / ticks) * index
          const y = yAt(value)
          return (
            <g key={index}>
              <line
                x1={padding.left}
                x2={WIDTH - padding.right}
                y1={y}
                y2={y}
                className="stroke-border"
                strokeDasharray={index === 0 ? undefined : "3 5"}
              />
              <text
                x={padding.left - 8}
                y={y + 3}
                textAnchor="end"
                className="fill-muted-foreground text-[10px]"
              >
                {formatNumber(value)}
              </text>
            </g>
          )
        })}

        {xAxis.map((label, index) => {
          if (index % labelStep !== 0) return null
          return (
            <text
              key={`${label}-${index}`}
              x={padding.left + groupWidth * (index + 0.5)}
              y={HEIGHT - 8}
              textAnchor="middle"
              className="fill-muted-foreground text-[10px]"
            >
              {label}
            </text>
          )
        })}

        {hover && (
          <rect
            x={padding.left + groupWidth * hover.index + 2}
            y={padding.top - 4}
            width={Math.max(0, groupWidth - 4)}
            height={innerH + 4}
            rx={4}
            style={{ fill: "var(--muted)" }}
            pointerEvents="none"
          />
        )}

        {series.map((item, seriesIndex) =>
          item.data.map((value, index) => {
            const height = Math.max(3, innerH - (yAt(value) - padding.top))
            const offset =
              (groupWidth - barWidth * series.length) / 2 +
              barWidth * seriesIndex
            return (
              <rect
                key={`${item.name}-${index}`}
                x={padding.left + groupWidth * index + offset}
                y={yAt(value)}
                width={barWidth}
                height={height}
                fill={`url(#${uid}-bar-${seriesIndex})`}
                className="transition-opacity"
                opacity={
                  hover === null || hover.index === index ? 1 : 0.45
                }
              />
            )
          })
        )}

        {Array.from({ length: count }, (_, index) => (
          <rect
            key={index}
            x={padding.left + groupWidth * index}
            y={padding.top}
            width={groupWidth}
            height={innerH}
            fill="transparent"
            className="cursor-pointer"
            onMouseEnter={() => showGroup(index)}
            onMouseLeave={() => setHover(null)}
          />
        ))}
      </svg>
      {hover && <ChartTooltip hover={hover} width={WIDTH} height={HEIGHT} />}
    </div>
  )
}

function PieChart({
  uid,
  xAxis,
  series,
}: {
  uid: string
  xAxis: string[]
  series: ChartSeries[]
}) {
  const [hover, setHover] = useState<ChartHover | null>(null)
  const first = series[0]
  const values = first.data.map((value) => Math.max(0, value))
  const total = values.reduce((sum, value) => sum + value, 0)
  const cx = 110
  const cy = 110
  const outer = 86
  const inner = 56

  const startAngle = (index: number) => {
    if (total <= 0) return -Math.PI / 2
    const ratio =
      values.slice(0, index).reduce((sum, value) => sum + value, 0) / total
    return -Math.PI / 2 + ratio * Math.PI * 2
  }

  const arcs = values.map((value, index) => {
    const start = startAngle(index)
    const sweep = total > 0 ? (value / total) * Math.PI * 2 : 0
    const end = start + sweep
    if (sweep >= Math.PI * 2 - 1e-6) {
      return { index, full: true, path: "" }
    }
    const large = sweep > Math.PI ? 1 : 0
    const x1 = cx + outer * Math.cos(start)
    const y1 = cy + outer * Math.sin(start)
    const x2 = cx + outer * Math.cos(end)
    const y2 = cy + outer * Math.sin(end)
    const x3 = cx + inner * Math.cos(end)
    const y3 = cy + inner * Math.sin(end)
    const x4 = cx + inner * Math.cos(start)
    const y4 = cy + inner * Math.sin(start)
    const path =
      `M ${x1} ${y1} A ${outer} ${outer} 0 ${large} 1 ${x2} ${y2} ` +
      `L ${x3} ${y3} A ${inner} ${inner} 0 ${large} 0 ${x4} ${y4} Z`
    return { index, full: false, path }
  })

  const showSlice = (index: number) => {
    const value = values[index] ?? 0
    const percent = total > 0 ? (value / total) * 100 : 0
    const start = startAngle(index)
    const sweep = total > 0 ? (value / total) * Math.PI * 2 : 0
    const middle = start + sweep / 2
    const radius = (outer + inner) / 2
    setHover({
      index,
      x: cx + radius * Math.cos(middle),
      y: cy + radius * Math.sin(middle),
      title: xAxis[index] ?? `${index + 1}`,
      rows: [
        {
          name: `${first.name} · ${percent.toFixed(1)}%`,
          value,
          color: PALETTE[index % PALETTE.length],
        },
      ],
    })
  }

  return (
    <div className="flex flex-col items-center gap-5 font-sans sm:flex-row sm:gap-6">
      <div className="relative shrink-0">
        <svg
          viewBox="0 0 220 220"
          className="h-[176px] w-[176px]"
          role="img"
          aria-label={first.name}
        >
          <defs>
            <filter
              id={`${uid}-pie-shadow`}
              x="-20%"
              y="-20%"
              width="140%"
              height="140%"
            >
              <feDropShadow
                dx="0"
                dy="2"
                stdDeviation="3"
                floodColor="#000"
                floodOpacity="0.12"
              />
            </filter>
          </defs>
          <g filter={`url(#${uid}-pie-shadow)`}>
            {arcs.map((arc) => {
              const color = PALETTE[arc.index % PALETTE.length]
              const active = hover === null || hover.index === arc.index
              if (arc.full) {
                return (
                  <circle
                    key={arc.index}
                    cx={cx}
                    cy={cy}
                    r={(outer + inner) / 2}
                    fill="none"
                    stroke={color}
                    strokeWidth={outer - inner}
                    className="cursor-pointer transition-opacity"
                    opacity={active ? 1 : 0.45}
                    onMouseEnter={() => showSlice(arc.index)}
                    onMouseLeave={() => setHover(null)}
                  />
                )
              }
              return (
                <path
                  key={arc.index}
                  d={arc.path}
                  fill={color}
                  stroke="var(--card)"
                  strokeWidth={2.5}
                  strokeLinejoin="round"
                  className="cursor-pointer transition-opacity"
                  opacity={active ? 1 : 0.45}
                  onMouseEnter={() => showSlice(arc.index)}
                  onMouseLeave={() => setHover(null)}
                />
              )
            })}
          </g>
          <text
            x={cx}
            y={cy - 2}
            textAnchor="middle"
            className="fill-foreground text-[22px] font-semibold tabular-nums"
            pointerEvents="none"
          >
            {formatNumber(total)}
          </text>
          <text
            x={cx}
            y={cy + 16}
            textAnchor="middle"
            className="fill-muted-foreground text-[11px]"
            pointerEvents="none"
          >
            总计
          </text>
        </svg>
        {hover && (
          <ChartTooltip hover={hover} width={220} height={220} />
        )}
      </div>

      <div className="grid w-full gap-2.5">
        {values.map((value, index) => {
          const color = PALETTE[index % PALETTE.length]
          const percent = total > 0 ? (value / total) * 100 : 0
          const active = hover === null || hover.index === index
          return (
            <div
              key={`${xAxis[index] ?? index}`}
              className="grid cursor-pointer gap-1"
              onMouseEnter={() => showSlice(index)}
              onMouseLeave={() => setHover(null)}
            >
              <div className="flex items-center gap-2 text-xs">
                <span
                  className="size-2 shrink-0 rounded-full transition-opacity"
                  style={{ background: color, opacity: active ? 1 : 0.45 }}
                />
                <span
                  className="min-w-0 flex-1 truncate transition-opacity"
                  style={{ opacity: active ? 1 : 0.6 }}
                >
                  {xAxis[index] ?? `${index + 1}`}
                </span>
                <span className="text-muted-foreground tabular-nums">
                  {formatNumber(value)}
                </span>
                <span className="text-muted-foreground w-11 text-right text-[11px] tabular-nums">
                  {percent.toFixed(1)}%
                </span>
              </div>
              <div className="bg-muted ml-4 h-1 overflow-hidden rounded-full">
                <div
                  className="h-full rounded-full transition-[width,opacity]"
                  style={{
                    width: `${percent}%`,
                    background: color,
                    opacity: active ? 1 : 0.45,
                  }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function MetricCards({ series }: { series: ChartSeries[] }) {
  return (
    <div
      className={
        series.length > 1
          ? "grid gap-2.5 font-sans sm:grid-cols-2"
          : "grid gap-2.5 font-sans"
      }
    >
      {series.map((item, index) => {
        const color = PALETTE[index % PALETTE.length]
        const value = item.data[item.data.length - 1]
        return (
          <div
            key={item.name}
            className="rounded-xl border p-3.5"
            style={{
              background: `linear-gradient(135deg, ${color}14, transparent 70%)`,
            }}
          >
            <div className="text-muted-foreground flex items-center gap-1.5 text-[11px]">
              <span
                className="size-2 rounded-full"
                style={{ background: color }}
              />
              {item.name}
            </div>
            <div className="mt-1.5 text-2xl font-semibold tracking-tight tabular-nums">
              {Number.isFinite(value) ? formatNumber(value) : "—"}
            </div>
          </div>
        )
      })}
    </div>
  )
}