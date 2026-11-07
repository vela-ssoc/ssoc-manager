import { useCallback, useEffect, useRef, useState } from "react"
import { Loader2, RefreshCw } from "lucide-react"

import { Button } from "@/components/ui/button"
import { toast } from "@/lib/toast"
import { redisConfigApi } from "./api"
import { RedisStringViewer } from "./string-viewer"

const PAGE_SIZE = 100

const SCAN_TYPES = new Set(["hash", "set", "zset"])

interface StreamEntry {
  ID: string
  Values: Record<string, unknown>
}

type View =
  | { kind: "empty" }
  | { kind: "string"; text: string }
  | { kind: "hash"; rows: Array<[string, string]> }
  | { kind: "list"; items: string[] }
  | { kind: "set"; items: string[] }
  | { kind: "zset"; rows: Array<[string, string]> }
  | { kind: "stream"; rows: StreamEntry[] }

export function RedisValuePanel({ redisKey }: { redisKey: string }) {
  const [view, setView] = useState<View | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [cursor, setCursor] = useState(0)
  const [hasMore, setHasMore] = useState(false)
  const [type, setType] = useState("")

  const requestId = useRef(0)
  const load = useCallback(
    async (from: number, reset: boolean) => {
      const id = ++requestId.current
      if (reset) setLoading(true)
      else setLoadingMore(true)
      try {
        const data = await redisConfigApi.value(redisKey, from, "*", PAGE_SIZE)
        if (id !== requestId.current) return
        const nextType = data?.type ?? ""
        const page = toView(nextType, data?.data)
        setType(nextType)
        setView((prev) => (reset || !prev ? page : mergeView(prev, page)))
        setCursor(data?.cursor ?? 0)
        setHasMore(moreOf(nextType, data?.cursor ?? 0, page))
      } catch (err) {
        if (id !== requestId.current) return
        if (reset) {
          setView(null)
          setHasMore(false)
        }
        toast.error(err instanceof Error ? err.message : "读取 Key 内容失败")
      } finally {
        if (id === requestId.current) {
          setLoading(false)
          setLoadingMore(false)
        }
      }
    },
    [redisKey]
  )

  useEffect(() => {
    setView(null)
    setCursor(0)
    setHasMore(false)
    void load(0, true)
  }, [load])

  const reload = () => {
    setCursor(0)
    void load(0, true)
  }

  const loadMore = () => {
    if (!hasMore || loading || loadingMore) return
    void load(cursor, false)
  }

  return (
    <div className="flex min-h-0 flex-col gap-2">
      <div className="flex h-8 items-center gap-2">
        <span className="text-muted-foreground text-[11px]">
          {type ? `${type.toUpperCase()} 内容` : "内容"}
        </span>
        <span className="text-muted-foreground text-[11px]">
          {SCAN_TYPES.has(type) ? "每批 100 条" : ""}
        </span>
        <Button
          variant="ghost"
          size="sm"
          className="ml-auto h-7"
          disabled={loading || loadingMore}
          onClick={reload}
        >
          <RefreshCw className={loading ? "animate-spin" : undefined} /> 刷新
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto rounded-md border">
        {loading ? (
          <div className="text-muted-foreground flex items-center justify-center gap-1.5 px-3 py-10 text-xs">
            <Loader2 className="size-3.5 animate-spin" />
            正在读取…
          </div>
        ) : !view ? (
          <div className="text-muted-foreground px-3 py-10 text-center text-xs">
            暂无内容
          </div>
        ) : (
          <ViewBody view={view} />
        )}
      </div>

      {hasMore && (
        <div className="flex justify-center">
          <Button variant="outline" size="sm" disabled={loadingMore} onClick={loadMore}>
            {loadingMore && <Loader2 className="animate-spin" />}
            加载更多
          </Button>
        </div>
      )}
    </div>
  )
}

function ViewBody({ view }: { view: View }) {
  switch (view.kind) {
    case "string":
      return <RedisStringViewer text={view.text} />
    case "hash":
      return <PairTable rows={view.rows} left="字段" right="值" empty="（空 Hash）" />
    case "zset":
      return <PairTable rows={view.rows} left="成员" right="分值" empty="（空 ZSet）" />
    case "list":
      return (
        <PairTable
          rows={view.items.map((item, index) => [String(index), item])}
          left="下标"
          right="值"
          empty="（空 List）"
        />
      )
    case "set":
      return <PairTable rows={view.items.map((item) => [item, ""])} left="成员" right="" empty="（空 Set）" />
    case "stream":
      return <StreamTable rows={view.rows} />
    default:
      return (
        <div className="text-muted-foreground px-3 py-10 text-center text-xs">
          该类型的值暂不支持展示
        </div>
      )
  }
}

function PairTable({
  rows,
  left,
  right,
  empty,
}: {
  rows: Array<[string, string]>
  left: string
  right: string
  empty: string
}) {
  if (rows.length === 0) {
    return <div className="text-muted-foreground px-3 py-8 text-center text-xs">{empty}</div>
  }
  return (
    <div>
      <div className="bg-muted/40 text-muted-foreground sticky top-0 z-10 grid grid-cols-[38%_62%] border-b px-3 py-1.5 text-[11px] font-medium">
        <span className="truncate">{left}</span>
        {right && <span className="truncate">{right}</span>}
      </div>
      {rows.map(([key, value], index) => (
        <div
          key={`${key}-${index}`}
          className="grid grid-cols-[38%_62%] items-start border-b px-3 py-1.5 text-xs last:border-b-0"
        >
          <span className="min-w-0 truncate pr-2 font-mono">{key}</span>
          {right && (
            <span className="min-w-0 font-mono break-all whitespace-pre-wrap">
              {value || "—"}
            </span>
          )}
        </div>
      ))}
    </div>
  )
}

function StreamTable({ rows }: { rows: StreamEntry[] }) {
  if (rows.length === 0) {
    return <div className="text-muted-foreground px-3 py-8 text-center text-xs">（空 Stream）</div>
  }
  return (
    <div>
      <div className="bg-muted/40 text-muted-foreground sticky top-0 z-10 grid grid-cols-[32%_68%] border-b px-3 py-1.5 text-[11px] font-medium">
        <span>消息 ID</span>
        <span>字段</span>
      </div>
      {rows.map((entry) => (
        <div
          key={entry.ID}
          className="grid grid-cols-[32%_68%] items-start border-b px-3 py-1.5 text-xs last:border-b-0"
        >
          <span className="min-w-0 truncate pr-2 font-mono">{entry.ID}</span>
          <span className="min-w-0 font-mono break-all whitespace-pre-wrap">
            {formatStreamValues(entry.Values)}
          </span>
        </div>
      ))}
    </div>
  )
}

function toView(type: string, data: unknown): View {
  switch (type) {
    case "string":
      return { kind: "string", text: typeof data === "string" ? data : stringify(data) }
    case "hash":
      return { kind: "hash", rows: asRecord(data) }
    case "list":
      return { kind: "list", items: asArray(data) }
    case "set":
      return { kind: "set", items: asArray(data) }
    case "zset":
      return { kind: "zset", rows: asPairs(data) }
    case "stream":
      return { kind: "stream", rows: asStream(data) }
    default:
      return { kind: "empty" }
  }
}

function mergeView(prev: View, next: View): View {
  if (prev.kind === "hash" && next.kind === "hash") {
    const map = new Map(prev.rows)
    for (const [field, value] of next.rows) map.set(field, value)
    return { kind: "hash", rows: [...map.entries()] }
  }
  if (prev.kind === "zset" && next.kind === "zset") {
    const map = new Map(prev.rows)
    for (const [member, score] of next.rows) map.set(member, score)
    return { kind: "zset", rows: [...map.entries()] }
  }
  if (prev.kind === "set" && next.kind === "set") {
    return { kind: "set", items: [...new Set([...prev.items, ...next.items])] }
  }
  if (prev.kind === "list" && next.kind === "list") {
    return { kind: "list", items: [...prev.items, ...next.items] }
  }
  if (prev.kind === "stream" && next.kind === "stream") {
    return { kind: "stream", rows: [...prev.rows, ...next.rows] }
  }
  return next
}

function moreOf(type: string, cursor: number, page: View): boolean {
  if (type === "list") return page.kind === "list" && page.items.length >= PAGE_SIZE
  if (SCAN_TYPES.has(type)) return cursor !== 0
  return false
}

function asRecord(data: unknown): Array<[string, string]> {
  if (!data || typeof data !== "object") return []
  return Object.entries(data as Record<string, unknown>).map(([key, value]) => [
    key,
    stringify(value),
  ])
}

function asArray(data: unknown): string[] {
  if (!Array.isArray(data)) return []
  return data.map((item) => (typeof item === "string" ? item : stringify(item)))
}

function asPairs(data: unknown): Array<[string, string]> {
  if (!Array.isArray(data)) return []
  const rows: Array<[string, string]> = []
  for (let i = 0; i + 1 < data.length; i += 2) {
    rows.push([String(data[i]), String(data[i + 1])])
  }
  return rows
}

function asStream(data: unknown): StreamEntry[] {
  if (!Array.isArray(data)) return []
  return data.map((item) => {
    const entry = (item ?? {}) as { ID?: string; Values?: Record<string, unknown> }
    return { ID: entry.ID ?? "", Values: entry.Values ?? {} }
  })
}

function formatStreamValues(values: Record<string, unknown>): string {
  const pairs = Object.entries(values)
  if (pairs.length === 0) return "—"
  return pairs.map(([key, value]) => `${key}=${stringify(value)}`).join("  ")
}

function stringify(value: unknown): string {
  if (value === null || value === undefined) return ""
  if (typeof value === "string") return value
  if (typeof value === "number" || typeof value === "boolean") return String(value)
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}