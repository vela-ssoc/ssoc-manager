import { useCallback, useEffect, useRef, useState } from "react"
import { Loader2, TriangleAlert } from "lucide-react"

import { Button } from "@/components/ui/button"
import { formatDuration } from "@/lib/format"
import { cn } from "@/lib/utils"
import { redisConfigApi } from "./api"

const SECTIONS = [
  "server",
  "clients",
  "memory",
  "persistence",
  "threads",
  "stats",
  "replication",
  "cpu",
  "commandstats",
  "latencystats",
  "sentinel",
  "cluster",
  "modules",
  "keyspace",
  "keysizes",
  "errorstats",
  "hotkeys",
]

const SUMMARY_FIELDS: Array<[string, string]> = [
  ["redis_version", "版本"],
  ["redis_mode", "运行模式"],
  ["role", "节点角色"],
  ["tcp_port", "端口"],
  ["connected_clients", "客户端连接"],
  ["used_memory_human", "已用内存"],
  ["maxmemory_human", "最大内存"],
]

interface InfoSection {
  name: string
  entries: Array<[string, string]>
}

export function RedisInfoPanel() {
  const [raw, setRaw] = useState("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [active, setActive] = useState("")

  const requestId = useRef(0)
  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    setError("")
    try {
      const res = await redisConfigApi.info(SECTIONS)
      if (id !== requestId.current) return
      const text = res?.data
      setRaw(typeof text === "string" ? text : "")
    } catch (err) {
      if (id !== requestId.current) return
      setRaw("")
      setError(err instanceof Error ? err.message : "读取 Redis 运行信息失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const sections = parseSections(raw)
  const info = parseInfoMap(raw)
  const current =
    sections.find((section) => section.name === active) ?? sections[0]

  if (loading) {
    return (
      <div className="text-muted-foreground flex items-center justify-center gap-1.5 px-4 py-10 text-xs">
        <Loader2 className="size-3.5 animate-spin" />
        正在读取 Redis 运行信息…
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border px-3 py-2 text-xs">
          <TriangleAlert className="size-3.5 shrink-0" />
          <span className="min-w-0 break-all">{error}</span>
        </div>
        <Button variant="outline" size="sm" className="mt-2" onClick={() => void load()}>
          重试
        </Button>
      </div>
    )
  }

  if (sections.length === 0) {
    return (
      <div className="text-muted-foreground px-4 py-10 text-center text-xs">
        暂无运行信息
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="grid shrink-0 grid-cols-2 gap-2 sm:grid-cols-4">
        {SUMMARY_FIELDS.map(([field, label]) => (
          <StatTile key={field} label={label} value={info[field]} />
        ))}
        <StatTile
          label="运行时长"
          value={
            info.uptime_in_seconds
              ? formatDuration(Number(info.uptime_in_seconds))
              : undefined
          }
        />
        <StatTile label="命中率" value={hitRate(info)} />
      </div>

      <div className="flex min-h-0 flex-1 overflow-hidden rounded-md border">
        <aside className="w-40 shrink-0 overflow-y-auto border-r p-1.5">
          {sections.map((section) => {
            const selected = section.name === current?.name
            return (
              <button
                key={section.name}
                type="button"
                aria-pressed={selected}
                className={cn(
                  "mb-0.5 flex w-full items-center justify-between gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
                  selected
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground hover:bg-accent/50"
                )}
                onClick={() => setActive(section.name)}
              >
                <span className="min-w-0 truncate">{section.name}</span>
                <span className="text-[10px] tabular-nums opacity-70">
                  {section.entries.length}
                </span>
              </button>
            )
          })}
        </aside>

        <div className="min-w-0 flex-1 overflow-auto">
          {current && (
            <div>
              {current.entries.map(([key, value], index) => (
                <div
                  key={key}
                  className={cn(
                    "grid grid-cols-[42%_58%] items-start px-3 py-1.5 text-xs",
                    index > 0 && "border-t"
                  )}
                >
                  <span className="min-w-0 truncate pr-2 font-mono">{key}</span>
                  <span className="min-w-0 font-mono break-all whitespace-pre-wrap">
                    {value || "—"}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function StatTile({ label, value }: { label: string; value?: string }) {
  return (
    <div className="rounded-md border bg-card px-2.5 py-2">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <div className="mt-0.5 truncate text-sm" title={value}>
        {value && value.trim() ? value : "—"}
      </div>
    </div>
  )
}

function parseInfoMap(text: string): Record<string, string> {
  const ret: Record<string, string> = {}
  for (const line of text.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith("#")) continue
    const index = trimmed.indexOf(":")
    if (index <= 0) continue
    ret[trimmed.slice(0, index)] = trimmed.slice(index + 1)
  }
  return ret
}

function hitRate(info: Record<string, string>): string | undefined {
  const hits = Number(info.keyspace_hits)
  const misses = Number(info.keyspace_misses)
  const total = hits + misses
  if (!Number.isFinite(total) || total <= 0) return undefined
  return `${((hits / total) * 100).toFixed(1)}%`
}

const SECTION_ORDER = new Map(
  SECTIONS.map((name, index) => [name.toLowerCase(), index] as const)
)

function orderSections(list: InfoSection[]): InfoSection[] {
  return [...list].sort((a, b) => {
    const ai = SECTION_ORDER.get(a.name.toLowerCase()) ?? Number.MAX_SAFE_INTEGER
    const bi = SECTION_ORDER.get(b.name.toLowerCase()) ?? Number.MAX_SAFE_INTEGER
    return ai - bi
  })
}

function parseSections(text: string): InfoSection[] {
  const sections: InfoSection[] = []
  let current: InfoSection | null = null

  for (const line of text.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed) continue
    if (trimmed.startsWith("#")) {
      current = { name: trimmed.replace(/^#+\s*/, ""), entries: [] }
      sections.push(current)
      continue
    }
    if (!current) continue
    const index = trimmed.indexOf(":")
    if (index <= 0) continue
    current.entries.push([trimmed.slice(0, index), trimmed.slice(index + 1)])
  }

  return orderSections(
    sections.filter((section) => section.entries.length > 0)
  )
}