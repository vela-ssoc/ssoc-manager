import { useEffect, useMemo, useRef, useState } from "react"
import {
  ArrowLeftRight,
  Boxes,
  Cable,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Clock,
  Cpu,
  Gauge,
  GitBranch,
  HardDrive,
  MemoryStick,
  Network,
  Waypoints,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime } from "@/lib/format"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"
import { formatBytes } from "@/features/broker/types"
import {
  cpuQuery,
  diskReadRateByDeviceQuery,
  diskWriteRateByDeviceQuery,
  goGoroutinesQuery,
  goHeapInuseQuery,
  goThreadsQuery,
  memTotalQuery,
  memUsedQuery,
  metricsLabel,
  netRecvRateByInterfaceQuery,
  netSentRateByInterfaceQuery,
  processOpenFdsQuery,
  rangeQuery,
  tunnelActiveQuery,
  tunnelCumulativeRateQuery,
  tunnelReceiveRateQuery,
  tunnelTransmitRateQuery,
  type VMPoint,
} from "./vm"

const QUICK_RANGES = [
  { label: "5 分钟", minutes: 5 },
  { label: "15 分钟", minutes: 15 },
  { label: "30 分钟", minutes: 30 },
  { label: "1 小时", minutes: 60 },
  { label: "3 小时", minutes: 180 },
  { label: "6 小时", minutes: 360 },
  { label: "12 小时", minutes: 720 },
  { label: "24 小时", minutes: 1440 },
  { label: "2 天", minutes: 2880 },
  { label: "7 天", minutes: 10080 },
] as const

const REFRESH_OPTIONS = [
  { label: "关闭", seconds: 0 },
  { label: "10 秒", seconds: 10 },
  { label: "30 秒", seconds: 30 },
  { label: "1 分钟", seconds: 60 },
] as const

const WEEKDAYS = ["日", "一", "二", "三", "四", "五", "六"] as const

const DEFAULT_RANGE_MINUTES = 1440
const DEFAULT_REFRESH_SECONDS = 10

const SERIES_COLORS = [
  "#55a76a",
  "#4682fa",
  "#ffaf0f",
  "#e55765",
  "#8b5cf6",
  "#0f8b8d",
  "#ee7b18",
  "#d94b8c",
]

export interface Sample {
  time: number
  value: number
}

export interface Series {
  name: string
  samples: Sample[]
  color: string
}

type TimeRange =
  | { kind: "relative"; minutes: number }
  | { kind: "absolute"; from: number; to: number }

export interface MonitorWindow {
  start: number
  end: number
  step: number
}

function stepForWindow(start: number, end: number): number {
  const seconds = Math.max(60, end - start)
  const raw = Math.ceil(seconds / 200)
  const steps = [
    10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 14400, 21600,
    43200, 86400, 172800, 604800, 2592000,
  ]
  return steps.find((item) => item >= raw) ?? Math.ceil(raw / 86400) * 86400
}

function resolveWindow(range: TimeRange): { start: number; end: number } {
  if (range.kind === "relative") {
    const end = Math.floor(Date.now() / 1000)
    return { start: end - range.minutes * 60, end }
  }
  return {
    start: Math.floor(range.from / 1000),
    end: Math.floor(range.to / 1000),
  }
}

function startOfDay(ms: number): number {
  const d = new Date(ms)
  d.setHours(0, 0, 0, 0)
  return d.getTime()
}

function startOfMonth(ms: number): number {
  const d = new Date(ms)
  d.setDate(1)
  d.setHours(0, 0, 0, 0)
  return d.getTime()
}

function withDate(baseMs: number, dayMs: number): number {
  const base = new Date(baseMs)
  const day = new Date(dayMs)
  base.setFullYear(day.getFullYear(), day.getMonth(), day.getDate())
  return base.getTime()
}

function shortDateTime(ms: number): string {
  const d = new Date(ms)
  const p = (n: number) => String(n).padStart(2, "0")
  return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

function formatTimeInput(ms: number): string {
  const d = new Date(ms)
  const p = (n: number) => String(n).padStart(2, "0")
  return `${p(d.getHours())}:${p(d.getMinutes())}`
}

function parseTimeInput(value: string): { hours: number; minutes: number } | null {
  const match = /^(\d{1,2}):(\d{1,2})$/.exec(value.trim())
  if (!match) return null
  const hours = Number(match[1])
  const minutes = Number(match[2])
  if (hours > 23 || minutes > 59) return null
  return { hours, minutes }
}

function normalizeTimeText(value: string): string {
  if (value.includes(":")) {
    const [hourPart = "", minutePart = ""] = value.split(":")
    const hours = hourPart.replace(/\D/g, "").slice(0, 2)
    const minutes = minutePart.replace(/\D/g, "").slice(0, 2)
    return `${hours}:${minutes}`
  }
  const digits = value.replace(/\D/g, "").slice(0, 4)
  return digits.length > 2 ? `${digits.slice(0, 2)}:${digits.slice(2)}` : digits
}

function toSeries(
  points: VMPoint[],
  keepInstance = false
): Omit<Series, "color">[] {
  return points
    .map((point) => {
      const labels = Object.entries(point.metric)
        .filter(
          ([key]) =>
            key !== "__name__" &&
            key !== "appname" &&
            (keepInstance || key !== "instance")
        )
        .map(([, value]) => value)
      const name = labels.join(" ") || "all"
      const samples = point.values
        .map(([t, v]) => ({ time: t * 1000, value: Number(v) || 0 }))
        .sort((a, b) => a.time - b.time)
      return { name, samples }
    })
    .sort((a, b) => a.name.localeCompare(b.name))
}

async function queryRangeSeries(
  q: string,
  start: number,
  end: number,
  step: string,
  keepInstance = false
): Promise<Omit<Series, "color">[]> {
  const res = await rangeQuery(q, start, end, step)
  return toSeries(res.data?.result ?? [], keepInstance)
}

function withColor(series: Omit<Series, "color">[]): Series[] {
  return series.map((s, i) => ({ ...s, color: SERIES_COLORS[i % SERIES_COLORS.length] }))
}

function mergePair(
  a: Omit<Series, "color">[],
  b: Omit<Series, "color">[],
  aSuffix: string,
  bSuffix: string
): Omit<Series, "color">[] {
  const result: Omit<Series, "color">[] = []
  for (const s of a) {
    result.push({
      name: s.name === "all" ? aSuffix : `${s.name} ${aSuffix}`,
      samples: s.samples,
    })
  }
  for (const s of b) {
    result.push({
      name: s.name === "all" ? bSuffix : `${s.name} ${bSuffix}`,
      samples: s.samples,
    })
  }
  return result
}

function timeLabel(ms: number): string {
  const d = new Date(ms)
  const p = (n: number) => String(n).padStart(2, "0")
  return `${p(d.getHours())}:${p(d.getMinutes())}`
}

function compactValue(value: number): string {
  const abs = Math.abs(value)
  if (abs >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(1)}G`
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return `${Math.round(value)}`
}

function trimNumber(value: number): string {
  return Number(value.toFixed(2)).toString()
}

function niceTicks(min: number, max: number): number[] {
  if (min === max) return [min]
  return [max, (max + min) / 2, min]
}

export function MetricsMonitoring({
  appname = "ssoc-manager",
  instance,
}: {
  appname?: string
  instance?: string
}) {
  const label = useMemo(() => metricsLabel(appname, instance), [appname, instance])
  const brokerLabel = useMemo(
    () => metricsLabel("ssoc-broker", instance),
    [instance]
  )
  const keepInstance = !instance
  const [range, setRange] = useState<TimeRange>({
    kind: "relative",
    minutes: DEFAULT_RANGE_MINUTES,
  })
  const [monitorWindow, setMonitorWindow] = useState<MonitorWindow>(() => {
    const { start, end } = resolveWindow({
      kind: "relative",
      minutes: DEFAULT_RANGE_MINUTES,
    })
    return { start, end, step: stepForWindow(start, end) }
  })
  const [refreshSeconds, setRefreshSeconds] = useState(DEFAULT_REFRESH_SECONDS)

  const [tunnelTraffic, setTunnelTraffic] = useState<Series[]>([])
  const [tunnelStreams, setTunnelStreams] = useState<Series[]>([])
  const [cpu, setCpu] = useState<Series[]>([])
  const [mem, setMem] = useState<Series[]>([])
  const [net, setNet] = useState<Series[]>([])
  const [diskIo, setDiskIo] = useState<Series[]>([])
  const [goroutines, setGoroutines] = useState<Series[]>([])
  const [heapInuse, setHeapInuse] = useState<Series[]>([])
  const [threads, setThreads] = useState<Series[]>([])
  const [openFds, setOpenFds] = useState<Series[]>([])
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setInterval> | null = null

    const load = async () => {
      const { start, end } = resolveWindow(range)
      const step = stepForWindow(start, end)
      const stepText = `${step}s`
      const fetchSeries = (query: string) =>
        queryRangeSeries(query, start, end, stepText, keepInstance)
      try {
        const [
          tunnelRxS, tunnelTxS, tunnelCumulativeS, tunnelActiveS,
          cpuS, memU, memT, netRecv, netSent, diskRead, diskWrite,
          goroutinesS, heapInuseS, threadsS, openFdsS,
        ] = await Promise.all([
          fetchSeries(tunnelReceiveRateQuery(brokerLabel)),
          fetchSeries(tunnelTransmitRateQuery(brokerLabel)),
          fetchSeries(tunnelCumulativeRateQuery(brokerLabel)),
          fetchSeries(tunnelActiveQuery(brokerLabel)),
          fetchSeries(cpuQuery(label)),
          fetchSeries(memUsedQuery(label)),
          fetchSeries(memTotalQuery(label)),
          fetchSeries(netRecvRateByInterfaceQuery(label)),
          fetchSeries(netSentRateByInterfaceQuery(label)),
          fetchSeries(diskReadRateByDeviceQuery(label)),
          fetchSeries(diskWriteRateByDeviceQuery(label)),
          fetchSeries(goGoroutinesQuery(label)),
          fetchSeries(goHeapInuseQuery(label)),
          fetchSeries(goThreadsQuery(label)),
          fetchSeries(processOpenFdsQuery(label)),
        ])

        if (cancelled) return
        setMonitorWindow({ start, end, step })
        setTunnelTraffic(
          withColor(mergePair(tunnelRxS, tunnelTxS, "接收", "发送"))
        )
        setTunnelStreams(
          withColor(
            mergePair(tunnelCumulativeS, tunnelActiveS, "累计增长率(/s)", "活跃数")
          )
        )
        setCpu(withColor(cpuS))
        setMem(withColor(mergePair(memU, memT, "已用", "总量")))
        setNet(withColor(mergePair(netRecv, netSent, "收", "发")))
        setDiskIo(withColor(mergePair(diskRead, diskWrite, "读", "写")))
        setGoroutines(withColor(goroutinesS))
        setHeapInuse(withColor(heapInuseS))
        setThreads(withColor(threadsS))
        setOpenFds(withColor(openFdsS))
        setError(false)
      } catch {
        if (!cancelled) setError(true)
      }
    }

    void load()
    if (refreshSeconds > 0) {
      timer = setInterval(() => void load(), refreshSeconds * 1000)
    }

    return () => {
      cancelled = true
      if (timer) clearInterval(timer)
    }
  }, [brokerLabel, keepInstance, label, range, refreshSeconds])

  const rangeLabel = useMemo(() => {
    if (range.kind === "relative") {
      const preset = QUICK_RANGES.find((item) => item.minutes === range.minutes)
      return `最近 ${preset?.label ?? `${range.minutes} 分钟`}`
    }
    return `${formatDateTime(new Date(range.from).toISOString())} ~ ${formatDateTime(new Date(range.to).toISOString())}`
  }, [range])

  const zoomTo = (from: number, to: number) => {
    setRange({ kind: "absolute", from, to })
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-end gap-2">
        <Select
          value={String(refreshSeconds)}
          onValueChange={(value) => setRefreshSeconds(Number(value))}
        >
          <SelectTrigger size="sm" className="w-24">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {REFRESH_OPTIONS.map((option) => (
              <SelectItem key={option.seconds} value={String(option.seconds)}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <TimeRangePicker range={range} label={rangeLabel} onChange={setRange} />
      </div>

      {error && (
        <div className="border-destructive/30 bg-destructive/5 text-destructive rounded-md border p-2.5 text-xs">
          无法获取监控指标，请确认 VictoriaMetrics 采集正常
        </div>
      )}

      <GroupHeader
        icon={Cable}
        title="虚拟通道"
        description="中继节点与中心端之间的通道流量与子流"
      />
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <MonitorChart icon={ArrowLeftRight} title="虚拟通道流量" unit="rate" series={tunnelTraffic} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={GitBranch} title="虚拟子流" unit="number" series={tunnelStreams} window={monitorWindow} onZoom={zoomTo} />
      </div>

      <GroupHeader
        icon={Gauge}
        title="系统资源"
        description="宿主机 CPU、内存、网络与磁盘的使用情况"
      />
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <MonitorChart icon={Cpu} title="CPU 使用率" unit="percent" series={cpu} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={MemoryStick} title="内存占用" unit="bytes" series={mem} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={Network} title="网络收发" unit="rate" series={net} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={HardDrive} title="磁盘 I/O" unit="rate" series={diskIo} window={monitorWindow} onZoom={zoomTo} />
      </div>

      <GroupHeader
        icon={Boxes}
        title="程序自身"
        description="当前进程的 Go 运行时与资源使用情况"
      />
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <MonitorChart icon={Waypoints} title="协程数" unit="number" series={goroutines} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={Boxes} title="堆内存" unit="bytes" series={heapInuse} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={Waypoints} title="线程数" unit="number" series={threads} window={monitorWindow} onZoom={zoomTo} />
        <MonitorChart icon={Waypoints} title="文件描述符" unit="number" series={openFds} window={monitorWindow} onZoom={zoomTo} />
      </div>
    </div>
  )
}

function TimeRangePicker({
  range,
  label,
  onChange,
}: {
  range: TimeRange
  label: string
  onChange: (range: TimeRange) => void
}) {
  const [open, setOpen] = useState(false)
  const [activeField, setActiveField] = useState<"from" | "to">("from")
  const [fromMs, setFromMs] = useState(0)
  const [toMs, setToMs] = useState(0)
  const [viewMonth, setViewMonth] = useState(() => startOfMonth(Date.now()))
  const [timeText, setTimeText] = useState("")
  const rootRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!open) return

    const onDown = (event: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false)
    }

    document.addEventListener("mousedown", onDown)
    document.addEventListener("keydown", onKey)
    return () => {
      document.removeEventListener("mousedown", onDown)
      document.removeEventListener("keydown", onKey)
    }
  }, [open])

  useEffect(() => {
    const base = activeField === "from" ? fromMs : toMs
    setTimeText(formatTimeInput(base))
  }, [activeField, fromMs, toMs])

  const toggle = () => {
    if (!open) {
      const now = Date.now()
      const from = range.kind === "relative" ? now - range.minutes * 60_000 : range.from
      const to = range.kind === "relative" ? now : range.to
      setFromMs(from)
      setToMs(to)
      setViewMonth(startOfMonth(from))
      setActiveField("from")
      setTimeText(formatTimeInput(from))
    }
    setOpen((value) => !value)
  }

  const resolveDraft = (): { from: number; to: number } | null => {
    const parsed = parseTimeInput(timeText)
    if (!parsed) return null
    const base = activeField === "from" ? fromMs : toMs
    const d = new Date(base)
    d.setHours(parsed.hours, parsed.minutes, 0, 0)
    const value = d.getTime()
    return activeField === "from"
      ? { from: value, to: toMs }
      : { from: fromMs, to: value }
  }

  const commitTime = (notify = false) => {
    const parsed = parseTimeInput(timeText)
    if (!parsed) {
      if (notify) toast.error("请输入有效的时间（HH:mm）")
      const base = activeField === "from" ? fromMs : toMs
      setTimeText(formatTimeInput(base))
      return
    }
    const d = new Date(activeField === "from" ? fromMs : toMs)
    d.setHours(parsed.hours, parsed.minutes, 0, 0)
    if (activeField === "from") setFromMs(d.getTime())
    else setToMs(d.getTime())
  }

  const pickDay = (dayMs: number) => {
    const current = resolveDraft() ?? { from: fromMs, to: toMs }
    let from = current.from
    let to = current.to
    if (activeField === "from") {
      from = withDate(from, dayMs)
      if (from >= to) to = from + 60 * 60 * 1000
      setActiveField("to")
    } else {
      const firstDay = startOfDay(from)
      const secondDay = startOfDay(dayMs)
      if (secondDay >= firstDay) {
        to = withDate(to, secondDay)
      } else {
        from = withDate(from, secondDay)
        to = withDate(to, firstDay)
      }
      if (to <= from) to = from + 60 * 60 * 1000
    }
    setFromMs(from)
    setToMs(to)
  }

  const moveMonth = (delta: number) => {
    setViewMonth((current) => {
      const d = new Date(current)
      d.setMonth(d.getMonth() + delta, 1)
      return d.getTime()
    })
  }

  const applyCustom = () => {
    const resolved = resolveDraft()
    if (!resolved) {
      toast.error("请输入有效的时间（HH:mm）")
      return
    }
    const from = Math.floor(resolved.from / 60_000) * 60_000
    const to = Math.floor(resolved.to / 60_000) * 60_000
    if (from >= to) {
      toast.error("开始时间必须早于结束时间")
      return
    }
    onChange({ kind: "absolute", from, to })
    setOpen(false)
  }

  const view = new Date(viewMonth)
  const year = view.getFullYear()
  const month = view.getMonth()
  const leading = new Date(year, month, 1).getDay()
  const gridStart = new Date(year, month, 1 - leading).getTime()
  const days = Array.from({ length: 42 }, (_, index) => {
    const d = new Date(gridStart)
    d.setDate(d.getDate() + index)
    d.setHours(0, 0, 0, 0)
    return d.getTime()
  })
  const fromDay = startOfDay(fromMs)
  const toDay = startOfDay(toMs)
  const today = startOfDay(Date.now())

  return (
    <div ref={rootRef} className="relative">
      <TooltipHint content={label}>
        <button
          type="button"
          aria-label="选择时间范围"
          className="border-input hover:bg-accent/50 flex h-7 max-w-80 items-center gap-1.5 rounded-md border bg-transparent px-2.5 text-xs transition-colors"
          onClick={toggle}
        >
          <Clock className="text-muted-foreground size-3.5 shrink-0" />
          <span className="min-w-0 truncate">{label}</span>
          <ChevronDown className="text-muted-foreground size-3.5 shrink-0" />
        </button>
      </TooltipHint>

      {open && (
        <div className="bg-popover text-popover-foreground absolute top-full right-0 z-50 mt-1 w-80 max-w-[calc(100vw-2rem)] rounded-md border p-3 shadow-lg">
          <div className="max-h-[70vh] overflow-y-auto pr-0.5">
            <div className="text-muted-foreground mb-1.5 text-[11px] font-medium">
              相对时间
            </div>
            <div className="grid grid-cols-2 gap-1">
              {QUICK_RANGES.map((item) => {
                const active =
                  range.kind === "relative" && range.minutes === item.minutes
                return (
                  <button
                    key={item.minutes}
                    type="button"
                    className={cn(
                      "rounded-md px-2 py-1.5 text-left text-xs transition-colors",
                      active
                        ? "bg-primary/10 text-foreground font-medium"
                        : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                    )}
                    onClick={() => {
                      onChange({ kind: "relative", minutes: item.minutes })
                      setOpen(false)
                    }}
                  >
                    {item.label}
                  </button>
                )
              })}
            </div>

            <div className="bg-border/70 my-2 h-px" />

            <div className="text-muted-foreground mb-2 text-[11px] font-medium">
              自定义时间范围
            </div>

            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                className={cn(
                  "min-w-0 rounded-md border px-2 py-1.5 text-left transition-colors",
                  activeField === "from"
                    ? "border-primary bg-primary/10"
                    : "hover:bg-accent/50"
                )}
                onClick={() => setActiveField("from")}
              >
                <span className="text-muted-foreground block text-[10px]">开始</span>
                <span className="mt-0.5 block truncate font-mono text-xs">
                  {shortDateTime(fromMs)}
                </span>
              </button>
              <button
                type="button"
                className={cn(
                  "min-w-0 rounded-md border px-2 py-1.5 text-left transition-colors",
                  activeField === "to"
                    ? "border-primary bg-primary/10"
                    : "hover:bg-accent/50"
                )}
                onClick={() => setActiveField("to")}
              >
                <span className="text-muted-foreground block text-[10px]">结束</span>
                <span className="mt-0.5 block truncate font-mono text-xs">
                  {shortDateTime(toMs)}
                </span>
              </button>
            </div>

            <div className="mt-2 overflow-hidden rounded-md border">
              <div className="flex items-center justify-between border-b px-1 py-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-6"
                  aria-label="上个月"
                  onClick={() => moveMonth(-1)}
                >
                  <ChevronLeft className="size-3.5" />
                </Button>
                <span className="text-xs font-medium">
                  {year} 年 {month + 1} 月
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-6"
                  aria-label="下个月"
                  onClick={() => moveMonth(1)}
                >
                  <ChevronRight className="size-3.5" />
                </Button>
              </div>

              <div className="text-muted-foreground grid grid-cols-7 pt-1.5 text-center text-[10px]">
                {WEEKDAYS.map((day) => (
                  <span key={day} className="py-0.5">
                    {day}
                  </span>
                ))}
              </div>

              <div className="grid grid-cols-7 gap-0.5 p-1.5">
                {days.map((dayMs) => {
                  const dayDate = new Date(dayMs)
                  const inMonth =
                    dayDate.getMonth() === month && dayDate.getFullYear() === year
                  const isStart = dayMs === fromDay
                  const isEnd = dayMs === toDay
                  const endpoint = isStart || isEnd
                  const inRange = dayMs >= fromDay && dayMs <= toDay
                  const isToday = dayMs === today
                  return (
                    <button
                      key={dayMs}
                      type="button"
                      className={cn(
                        "flex h-8 items-center justify-center rounded-md text-xs transition-colors",
                        !inMonth && "text-muted-foreground/35",
                        inRange && !endpoint && "bg-primary/10",
                        endpoint && "bg-primary text-primary-foreground hover:bg-primary/90",
                        !inRange && !endpoint && "hover:bg-accent/60",
                        isToday && !endpoint && "ring-primary/40 ring-1 ring-inset"
                      )}
                      onClick={() => {
                        if (!inMonth) setViewMonth(startOfMonth(dayMs))
                        pickDay(dayMs)
                      }}
                    >
                      {dayDate.getDate()}
                    </button>
                  )
                })}
              </div>
            </div>

            <div className="mt-2 flex items-center gap-2">
              <span className="text-muted-foreground shrink-0 text-[11px]">
                {activeField === "from" ? "开始时间" : "结束时间"}
              </span>
              <div className="relative">
                <Input
                  aria-label="时间（HH:mm）"
                  value={timeText}
                  placeholder="HH:mm"
                  inputMode="numeric"
                  maxLength={5}
                  className="h-7 w-20 pr-6 text-center font-mono text-xs"
                  onChange={(event) =>
                    setTimeText(normalizeTimeText(event.target.value))
                  }
                  onBlur={() => commitTime()}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      commitTime(true)
                      event.currentTarget.blur()
                    }
                  }}
                />
                <Clock className="text-muted-foreground pointer-events-none absolute top-1/2 right-1.5 size-3 -translate-y-1/2" />
              </div>
              {activeField === "to" && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 px-1.5 text-[11px]"
                  onClick={() => setToMs(Date.now())}
                >
                  现在
                </Button>
              )}
              <span className="text-muted-foreground ml-auto text-[10px]">本地时间</span>
            </div>

            <Button size="sm" className="mt-3 w-full" onClick={applyCustom}>
              应用
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function GroupHeader({
  icon: Icon,
  title,
  description,
}: {
  icon: typeof Cpu
  title: string
  description: string
}) {
  return (
    <div className="flex items-center gap-2">
      <div className="bg-primary/10 text-primary flex size-6 shrink-0 items-center justify-center rounded-md">
        <Icon className="size-3" />
      </div>
      <h3 className="shrink-0 text-sm font-semibold">{title}</h3>
      <span className="text-muted-foreground/50 shrink-0 text-[10px]">·</span>
      <p className="text-muted-foreground min-w-0 truncate text-xs">{description}</p>
    </div>
  )
}

export function MonitorChart({
  icon: Icon,
  title,
  unit,
  series,
  window: monitorWindow,
  onZoom,
  compact = false,
  mini = false,
}: {
  icon: typeof Cpu
  title: string
  unit: "percent" | "bytes" | "rate" | "number"
  series: Series[]
  window: MonitorWindow
  onZoom?: (from: number, to: number) => void
  compact?: boolean
  mini?: boolean
}) {
  const [hidden, setHidden] = useState<Set<string>>(new Set())

  const visible = series.filter((s) => !hidden.has(s.name))

  const all = visible.flatMap((s) => s.samples)
  const hasData = series.some((s) => s.samples.length > 0)
  const min = all.length ? Math.min(...all.map((p) => p.value)) : 0
  const max = all.length ? Math.max(...all.map((p) => p.value)) : 0
  const start = monitorWindow.start * 1000
  const end = monitorWindow.end * 1000

  const formatValue = (value: number): string => {
    if (unit === "percent") return `${value.toFixed(1)}%`
    if (unit === "number") return trimNumber(value)
    if (unit === "rate") return `${formatBytes(value)}/s`
    return formatBytes(value)
  }

  const ticks = niceTicks(min, max)
  const tickLabel = (value: number): string => {
    if (unit === "percent") return `${Math.round(value)}%`
    if (unit === "number") return trimNumber(value)
    return compactValue(value)
  }

  const toggleSeries = (name: string) => {
    setHidden((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }

  const showLegend = mini ? series.length > 1 : compact ? series.length > 1 : true

  const legendItems = series.map((s) => {
    const isHidden = hidden.has(s.name)
    return (
      <button
        key={s.name}
        type="button"
        className={cn(
          "flex items-center transition-opacity",
          mini
            ? "gap-1 text-[9px]"
            : compact
              ? "gap-1.5 text-[10px]"
              : "gap-1.5 text-[11px]",
          isHidden && "opacity-40"
        )}
        onClick={(event) => {
          event.stopPropagation()
          toggleSeries(s.name)
        }}
      >
        <span
          className={cn("rounded-full", mini ? "size-1.5" : "size-2")}
          style={{ background: s.color }}
        />
        <span className="text-muted-foreground">{s.name}</span>
      </button>
    )
  })

  return (
    <section
      className={cn(
        "grid content-start rounded-md border bg-card",
        mini ? "gap-1 p-1.5" : compact ? "gap-1 p-2" : "gap-2 p-3"
      )}
    >
      <div className="flex min-w-0 items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-1.5">
          <Icon
            className={cn(
              "text-muted-foreground shrink-0",
              mini ? "size-3" : compact ? "size-3" : "size-3.5"
            )}
          />
          <h3
            className={cn(
              "truncate font-semibold",
              mini ? "text-[11px]" : compact ? "text-xs" : "text-sm"
            )}
          >
            {title}
          </h3>
          {mini && (
            <span className="text-muted-foreground/60 shrink-0 text-[9px]">
              6h
            </span>
          )}
        </div>
        {mini && showLegend && (
          <div className="flex shrink-0 items-center gap-2">{legendItems}</div>
        )}
      </div>

      {!hasData ? (
        <div
          className={cn(
            "text-muted-foreground flex items-center justify-center text-xs",
            mini ? "h-12" : compact ? "h-28" : "h-44"
          )}
        >
          该时间范围内暂无数据
        </div>
      ) : (
        <Chart
          series={visible}
          min={min}
          max={max}
          start={start}
          end={end}
          stepSeconds={monitorWindow.step}
          onZoom={onZoom}
          compact={compact}
          mini={mini}
          ticks={ticks}
          tickLabel={tickLabel}
          formatValue={formatValue}
        />
      )}

      {!mini && showLegend && (
        <div
          className={cn(
            "flex flex-wrap gap-y-1",
            compact ? "gap-x-3" : "gap-x-4"
          )}
        >
          {legendItems}
        </div>
      )}
    </section>
  )
}

function Chart({
  series,
  min,
  max,
  start,
  end,
  stepSeconds,
  onZoom,
  compact = false,
  mini = false,
  ticks,
  tickLabel,
  formatValue,
}: {
  series: Series[]
  min: number
  max: number
  start: number
  end: number
  stepSeconds: number
  onZoom?: (from: number, to: number) => void
  compact?: boolean
  mini?: boolean
  ticks: number[]
  tickLabel: (value: number) => string
  formatValue: (value: number) => string
}) {
  const width = 800
  const height = mini ? 64 : compact ? 120 : 180
  const renderHeight = compact ? 96 : 160
  const span = max - min || 1
  const plotLeft = 4
  const plotRight = width - 4
  const plotTop = mini ? 6 : compact ? 8 : 12
  const plotBottom = height - plotTop
  const plotHeight = plotBottom - plotTop
  const timeSpan = Math.max(1, end - start)
  const maxGap = Math.max(stepSeconds, 1) * 1.6 * 1000
  const withDate = timeSpan > 24 * 60 * 60 * 1000

  const [hover, setHover] = useState<{
    x: number
    y: number
    time: number
    values: Record<string, number>
  } | null>(null)
  const [selection, setSelection] = useState<{
    startX: number
    endX: number
  } | null>(null)
  const dragRef = useRef<{ pointerId: number; startX: number } | null>(null)

  const xFor = (time: number) => {
    const ratio = (Math.min(end, Math.max(start, time)) - start) / timeSpan
    return plotLeft + ratio * (plotRight - plotLeft)
  }

  const coordsFor = (s: Series) => {
    if (s.samples.length === 0) return null
    return s.samples.map((p) => ({
      t: p.time,
      x: xFor(p.time),
      y: plotBottom - ((p.value - min) / span) * plotHeight,
      value: p.value,
    }))
  }

  const nearestAt = (s: Series, time: number) => {
    const coords = coordsFor(s)
    if (!coords) return null
    return coords.reduce((a, b) =>
      Math.abs(a.t - time) < Math.abs(b.t - time) ? a : b
    )
  }

  const pointerPos = (event: React.PointerEvent<HTMLDivElement>) => {
    const rect = event.currentTarget.getBoundingClientRect()
    const x = Math.max(
      plotLeft,
      Math.min(plotRight, ((event.clientX - rect.left) / rect.width) * width)
    )
    const y = Math.max(
      plotTop,
      Math.min(plotBottom, ((event.clientY - rect.top) / rect.height) * height)
    )
    return { x, y }
  }

  const updateHover = (px: number, py: number) => {
    const time = start + ((px - plotLeft) / (plotRight - plotLeft)) * timeSpan
    const values: Record<string, number> = {}
    for (const s of series) {
      const nearest = nearestAt(s, time)
      if (!nearest) continue
      if (Math.abs(nearest.t - time) > maxGap) continue
      values[s.name] = nearest.value
    }
    setHover({ x: px, y: py, time, values })
  }

  const handlePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (event.pointerType !== "mouse" || event.button !== 0) return
    if (!onZoom) return
    const { x } = pointerPos(event)
    event.currentTarget.setPointerCapture(event.pointerId)
    dragRef.current = { pointerId: event.pointerId, startX: x }
    setSelection({ startX: x, endX: x })
    setHover(null)
  }

  const handlePointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    if (event.pointerType !== "mouse") return
    const { x, y } = pointerPos(event)
    const drag = dragRef.current
    if (drag) {
      setSelection({ startX: drag.startX, endX: x })
      return
    }
    updateHover(x, y)
  }

  const handlePointerUp = (event: React.PointerEvent<HTMLDivElement>) => {
    if (event.pointerType !== "mouse") return
    if (!onZoom) return
    const drag = dragRef.current
    if (!drag) return
    const { x, y } = pointerPos(event)
    dragRef.current = null
    setSelection(null)
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }

    const fromX = Math.min(drag.startX, x)
    const toX = Math.max(drag.startX, x)
    if (toX - fromX < 8) {
      updateHover(x, y)
      return
    }

    const from = start + ((fromX - plotLeft) / (plotRight - plotLeft)) * timeSpan
    const to = start + ((toX - plotLeft) / (plotRight - plotLeft)) * timeSpan
    if (to - from < 1000) return
    onZoom(Math.floor(from), Math.ceil(to))
  }

  const handlePointerCancel = (event: React.PointerEvent<HTMLDivElement>) => {
    dragRef.current = null
    setSelection(null)
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId)
    }
  }

  const handlePointerLeave = () => {
    if (!dragRef.current) setHover(null)
  }

  const gridY = [plotBottom, (plotBottom + plotTop) / 2, plotTop]
  const selStart = selection ? Math.min(selection.startX, selection.endX) : 0
  const selEnd = selection ? Math.max(selection.startX, selection.endX) : 0

  return (
    <div
      className={cn(
        "grid gap-x-2",
        mini ? "grid-cols-1" : "grid-cols-[auto_1fr]"
      )}
    >
      {!mini && (
        <div
          className={cn(
            "text-muted-foreground flex flex-col justify-between text-right font-mono",
            compact ? "py-1 text-[9px]" : "py-[10px] text-[10px]"
          )}
          style={{ height: renderHeight }}
        >
          {ticks.map((t, i) => (
            <span key={i} className="leading-none">
              {tickLabel(t)}
            </span>
          ))}
        </div>
      )}

      <div className="min-w-0">
        <div
          className="relative select-none"
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onPointerCancel={handlePointerCancel}
          onPointerLeave={handlePointerLeave}
        >
          <svg
            viewBox={`0 0 ${width} ${height}`}
            className={cn(
              "block w-full",
              onZoom ? "cursor-crosshair" : "cursor-default",
              mini ? "h-12" : compact ? "h-24" : "h-40"
            )}
            preserveAspectRatio="none"
          >
            {gridY.map((y) => (
              <line
                key={y}
                x1={plotLeft}
                y1={y}
                x2={plotRight}
                y2={y}
                stroke="currentColor"
                strokeWidth="1"
                className="text-muted-foreground/20"
              />
            ))}
            {series.map((s) => {
              const coords = coordsFor(s)
              if (!coords) return null
              const segments = splitSegments(coords, maxGap)
              return segments.map((segment, index) => {
                if (segment.length === 1) {
                  const point = segment[0]
                  return (
                    <circle
                      key={`${s.name}-${index}`}
                      cx={point.x}
                      cy={point.y}
                      r="1.5"
                      fill={s.color}
                    />
                  )
                }
                const line = segment
                  .map((c) => `${c.x.toFixed(1)},${c.y.toFixed(1)}`)
                  .join(" ")
                return (
                  <polyline
                    key={`${s.name}-${index}`}
                    points={line}
                    fill="none"
                    stroke={s.color}
                    strokeWidth="1.5"
                    strokeLinejoin="round"
                  />
                )
              })
            })}

            {selection && selEnd - selStart > 0 && (
              <>
                <rect
                  x={selStart}
                  y={plotTop}
                  width={selEnd - selStart}
                  height={plotBottom - plotTop}
                  fill="currentColor"
                  className="text-primary/15"
                />
                <line
                  x1={selStart}
                  y1={plotTop}
                  x2={selStart}
                  y2={plotBottom}
                  stroke="currentColor"
                  strokeWidth="1"
                  className="text-primary/60"
                />
                <line
                  x1={selEnd}
                  y1={plotTop}
                  x2={selEnd}
                  y2={plotBottom}
                  stroke="currentColor"
                  strokeWidth="1"
                  className="text-primary/60"
                />
              </>
            )}

            {hover && (
              <>
                <line
                  x1={hover.x}
                  y1={plotTop}
                  x2={hover.x}
                  y2={plotBottom}
                  stroke="currentColor"
                  strokeWidth="1"
                  strokeDasharray="4 3"
                  className="text-muted-foreground/50"
                />
                <line
                  x1={plotLeft}
                  y1={hover.y}
                  x2={plotRight}
                  y2={hover.y}
                  stroke="currentColor"
                  strokeWidth="1"
                  strokeDasharray="4 3"
                  className="text-muted-foreground/50"
                />
                {series.map((s) => {
                  if (hover.values[s.name] === undefined) return null
                  const nearest = nearestAt(s, hover.time)
                  if (!nearest) return null
                  return (
                    <circle
                      key={`dot-${s.name}`}
                      cx={nearest.x}
                      cy={nearest.y}
                      r="2.5"
                      fill={s.color}
                      stroke="currentColor"
                      strokeWidth="1"
                      className="text-background"
                    />
                  )
                })}
              </>
            )}
          </svg>

          {hover && (
            <div
              className={cn(
                "bg-popover text-popover-foreground pointer-events-none absolute z-10 rounded-md border p-2 shadow-lg",
                mini
                  ? "w-40 text-[9px]"
                  : compact
                    ? "w-44 text-[10px]"
                    : "w-52 text-xs"
              )}
              style={{
                left: `clamp(0.5rem, ${(hover.x / width) * 100}%, ${
                  mini
                    ? "calc(100% - 10.5rem)"
                    : compact
                      ? "calc(100% - 11.5rem)"
                      : "calc(100% - 13.5rem)"
                })`,
                top: `${(hover.y / height) * 100}%`,
                transform:
                  hover.y < height / 2
                    ? "translateY(10px)"
                    : "translateY(calc(-100% - 10px))",
              }}
            >
              <div
                className={cn(
                  "text-muted-foreground mb-1 font-mono",
                  mini
                    ? "text-[8px]"
                    : compact
                      ? "text-[9px]"
                      : "text-[10px]"
                )}
              >
                {fullTimeLabel(hover.time, withDate)}
              </div>
              <div className="grid gap-1">
                {Object.keys(hover.values).length === 0 && (
                  <div className="text-muted-foreground text-[11px]">该时间点无数据</div>
                )}
                {series.map((s) => {
                  const value = hover.values[s.name]
                  if (value === undefined) return null
                  return (
                    <div key={s.name} className="flex items-center gap-1.5">
                      <span
                        className="size-2 shrink-0 rounded-full"
                        style={{ background: s.color }}
                      />
                      <span className="text-muted-foreground min-w-0 flex-1 truncate">
                        {s.name}
                      </span>
                      <span className="shrink-0 font-mono tabular-nums">
                        {formatValue(value)}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>

        {!mini && (
          <div
            className={cn(
              "text-muted-foreground mt-1 flex justify-between font-mono",
              compact ? "text-[9px]" : "text-[10px]"
            )}
          >
            <span>{axisTimeLabel(start, withDate)}</span>
            <span>{axisTimeLabel((start + end) / 2, withDate)}</span>
            <span>{axisTimeLabel(end, withDate)}</span>
          </div>
        )}
      </div>
    </div>
  )
}

interface Coord {
  t: number
  x: number
  y: number
  value: number
}

function splitSegments(coords: Coord[], maxGap: number): Coord[][] {
  const segments: Coord[][] = []
  let current: Coord[] = []

  for (const point of coords) {
    const prev = current[current.length - 1]
    if (prev && point.t - prev.t > maxGap) {
      if (current.length > 0) segments.push(current)
      current = []
    }
    current.push(point)
  }
  if (current.length > 0) segments.push(current)
  return segments
}

function fullTimeLabel(ms: number, withDate = false): string {
  const d = new Date(ms)
  const p = (n: number) => String(n).padStart(2, "0")
  const time = `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
  return withDate ? `${p(d.getMonth() + 1)}-${p(d.getDate())} ${time}` : time
}

function axisTimeLabel(ms: number, withDate: boolean): string {
  if (!withDate) return timeLabel(ms)
  const d = new Date(ms)
  const p = (n: number) => String(n).padStart(2, "0")
  return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}