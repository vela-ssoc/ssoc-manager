import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import {
  ChevronRight,
  Copy,
  FolderTree,
  List,
  Loader2,
  MemoryStick,
  RefreshCw,
  Search,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
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
import { CollapseIcon, FolderIcon } from "@/features/zip-file/file-icon"
import { redisConfigApi } from "@/features/redis-config/api"
import { RedisValuePanel } from "@/features/redis-config/value-view"
import { redisTtl, type RedisKeyMeta } from "@/features/redis-config/types"

const SIZE_OPTIONS = [100, 200, 500, 1000]

const SEPARATOR = ":"

type ViewMode = "list" | "tree"

interface ScanOptions {
  cursor: number
  match: string
  size: number
  reset: boolean
}

interface TreeNode {
  name: string
  path: string
  children: TreeNode[]
  meta?: RedisKeyMeta
}

export function RedisKeysBrowser() {
  const [records, setRecords] = useState<RedisKeyMeta[]>([])
  const [cursor, setCursor] = useState(0)
  const [match, setMatch] = useState("*")
  const [appliedMatch, setAppliedMatch] = useState("*")
  const [size, setSize] = useState(200)
  const [loading, setLoading] = useState(false)
  const [started, setStarted] = useState(false)
  const [mode, setMode] = useState<ViewMode>("list")
  const [selected, setSelected] = useState<RedisKeyMeta | null>(null)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [now, setNow] = useState(() => Date.now())

  const requestId = useRef(0)
  const scan = useCallback(async (opts: ScanOptions) => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const data = await redisConfigApi.keys(opts.cursor, opts.match, opts.size)
      if (id !== requestId.current) return
      const incoming = data?.keys ?? []
      setCursor(data?.cursor ?? 0)
      setNow(Date.now())
      setRecords((prev) => {
        const base = opts.reset ? [] : prev
        const seen = new Set(base.map((item) => item.key))
        const merged = [...base]
        for (const item of incoming) {
          if (!seen.has(item.key)) {
            seen.add(item.key)
            merged.push(item)
          }
        }
        return merged
      })
      setStarted(true)
    } catch (err) {
      if (id !== requestId.current) return
      if (opts.reset) {
        setRecords([])
        setStarted(true)
      }
      toast.error(err instanceof Error ? err.message : "扫描 Redis Key 失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void scan({ cursor: 0, match: "*", size: 200, reset: true })
  }, [scan])

  const rescan = (nextMatch = appliedMatch) => {
    setSelected(null)
    setCursor(0)
    setRecords([])
    void scan({ cursor: 0, match: nextMatch, size, reset: true })
  }

  const applyMatch = () => {
    const next = match.trim() || "*"
    setAppliedMatch(next)
    rescan(next)
  }

  const loadMore = () => {
    if (loading || cursor === 0) return
    void scan({ cursor, match: appliedMatch, size, reset: false })
  }

  const tree = useMemo(() => buildTree(records), [records])
  const finished = started && cursor === 0

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex min-h-0 flex-1">
        <aside className="flex w-96 shrink-0 flex-col border-r">
          <div className="flex h-12 shrink-0 items-center gap-1.5 border-b px-2">
            <div className="relative min-w-0 flex-1">
              <Search className="text-muted-foreground absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2" />
              <Input
                value={match}
                placeholder="匹配模式，例如 user:*"
                className="h-7 border-transparent bg-muted/50 pl-8 font-mono text-xs focus-visible:bg-background"
                onChange={(event) => setMatch(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") applyMatch()
                }}
              />
            </div>
            <TooltipHint content="按当前模式重新扫描">
              <Button
                variant="ghost"
                size="icon"
                className="size-7 shrink-0"
                aria-label="重新扫描"
                disabled={loading}
                onClick={() => applyMatch()}
              >
                <RefreshCw className={cn("size-3.5", loading && "animate-spin")} />
              </Button>
            </TooltipHint>
          </div>

          <div className="flex h-8 shrink-0 items-center gap-2 border-b px-2">
            <ViewSwitcher mode={mode} onChange={setMode} />
            <span className="text-muted-foreground min-w-0 flex-1 truncate text-right text-[11px]">
              {loading ? "扫描中…" : `${records.length} 个${finished ? " · 全部" : ""}`}
            </span>
            <Select value={String(size)} onValueChange={(value) => setSize(Number(value))}>
              <SelectTrigger
                className="text-muted-foreground h-6 w-16 shrink-0 gap-1 border-transparent bg-transparent px-1 text-[11px] tabular-nums [&_svg]:size-3"
                aria-label="每批数量"
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SIZE_OPTIONS.map((option) => (
                  <SelectItem key={option} value={String(option)} className="text-xs tabular-nums">
                    {option}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="min-h-0 flex-1 overflow-auto py-1">
            {loading && records.length === 0 ? (
              <div className="text-muted-foreground flex items-center justify-center gap-1.5 px-3 py-8 text-xs">
                <Loader2 className="size-3.5 animate-spin" />
                正在扫描…
              </div>
            ) : records.length === 0 ? (
              <div className="text-muted-foreground flex flex-col items-center gap-2 px-4 py-10 text-center text-xs">
                <MemoryStick className="size-7 opacity-40" />
                <span>{started ? "没有匹配的 Key" : "尚未扫描"}</span>
              </div>
            ) : mode === "tree" ? (
              <TreeNodeList
                nodes={tree}
                depth={0}
                now={now}
                expanded={expanded}
                selected={selected}
                onSelect={setSelected}
                onToggle={(path) =>
                  setExpanded((prev) => {
                    const next = new Set(prev)
                    if (next.has(path)) next.delete(path)
                    else next.add(path)
                    return next
                  })
                }
              />
            ) : (
              <ul>
                {records.map((item) => (
                  <KeyRow
                    key={item.key}
                    meta={item}
                    now={now}
                    active={selected?.key === item.key}
                    onSelect={() => setSelected(item)}
                  />
                ))}
              </ul>
            )}
          </div>

          {!finished && records.length > 0 && (
            <div className="shrink-0 border-t p-2">
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                disabled={loading}
                onClick={loadMore}
              >
                {loading ? <Loader2 className="animate-spin" /> : <ChevronRight />}
                加载更多
              </Button>
            </div>
          )}
        </aside>

        <section className="flex min-w-0 flex-1 flex-col">
          {selected ? (
            <KeyDetail meta={selected} now={now} />
          ) : (
            <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2 text-sm">
              <MemoryStick className="size-9 opacity-30" />
              <span>选择左侧 Key 查看详情</span>
              <span className="text-[11px] opacity-70">
                {records.length} 个 Key · 匹配模式 {appliedMatch}
              </span>
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

function ViewSwitcher({
  mode,
  onChange,
}: {
  mode: ViewMode
  onChange: (mode: ViewMode) => void
}) {
  const options = [
    { value: "list" as const, label: "列表视图", icon: List },
    { value: "tree" as const, label: "树结构视图", icon: FolderTree },
  ]
  return (
    <div className="bg-muted flex shrink-0 items-center gap-0.5 rounded-md p-0.5">
      {options.map((option) => {
        const Icon = option.icon
        return (
          <TooltipHint key={option.value} content={option.label}>
            <button
              type="button"
              aria-label={option.label}
              aria-pressed={mode === option.value}
              className={cn(
                "flex size-5 items-center justify-center rounded transition-colors",
                mode === option.value
                  ? "bg-card text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
              onClick={() => onChange(option.value)}
            >
              <Icon className="size-3" />
            </button>
          </TooltipHint>
        )
      })}
    </div>
  )
}

function KeyRow({
  meta,
  now,
  active,
  onSelect,
}: {
  meta: RedisKeyMeta
  now: number
  active: boolean
  onSelect: () => void
}) {
  const expiry = expiryAt(meta.ttl, now)
  return (
    <li>
      <button
        type="button"
        className={cn(
          "flex h-8 w-full items-center gap-2 px-3 text-left transition-colors",
          active
            ? "bg-accent text-accent-foreground"
            : "hover:bg-accent/50"
        )}
        onClick={onSelect}
      >
        <TypeBadge type={meta.type} />
        <TooltipHint content={meta.key}>
          <span className="min-w-0 flex-1 truncate text-left font-mono text-xs">
            {meta.key}
          </span>
        </TooltipHint>
        {expiry && (
          <span className="text-muted-foreground shrink-0 text-[10px] tabular-nums">
            {expiry}
          </span>
        )}
      </button>
    </li>
  )
}

function TreeNodeList({
  nodes,
  depth,
  now,
  expanded,
  selected,
  onSelect,
  onToggle,
}: {
  nodes: TreeNode[]
  depth: number
  now: number
  expanded: Set<string>
  selected: RedisKeyMeta | null
  onSelect: (meta: RedisKeyMeta) => void
  onToggle: (path: string) => void
}) {
  return (
    <>
      {nodes.map((node) => {
        const hasChildren = node.children.length > 0
        const open = expanded.has(node.path)
        const active = selected?.key === node.path

        return (
          <div key={node.path}>
            <button
              type="button"
              aria-expanded={hasChildren ? open : undefined}
              className={cn(
                "flex h-8 w-full items-center gap-1 pr-3 transition-colors",
                active ? "bg-accent text-accent-foreground" : "hover:bg-accent/50"
              )}
              style={{ paddingLeft: depth * 12 + 4 }}
              onClick={() => {
                if (hasChildren) onToggle(node.path)
                if (node.meta) onSelect(node.meta)
              }}
            >
              {hasChildren ? (
                <CollapseIcon open={open} className="size-3.5 shrink-0 opacity-70" />
              ) : (
                <span className="size-3.5 shrink-0" />
              )}
              <span className="flex shrink-0 items-center justify-center">
                {hasChildren ? (
                  <FolderIcon className="size-4" />
                ) : node.meta ? (
                  <TypeBadge type={node.meta.type} />
                ) : null}
              </span>
              <TooltipHint content={node.meta?.key ?? node.path}>
                <span
                  className={cn(
                    "min-w-0 flex-1 truncate text-left font-mono text-xs",
                    !node.meta && "text-muted-foreground"
                  )}
                >
                  {node.name}
                </span>
              </TooltipHint>
              {node.meta && (
                <span className="text-muted-foreground w-[104px] shrink-0 text-right text-[10px] tabular-nums">
                  {expiryAt(node.meta.ttl, now)}
                </span>
              )}
            </button>
            {hasChildren && open && (
              <TreeNodeList
                nodes={node.children}
                depth={depth + 1}
                now={now}
                expanded={expanded}
                selected={selected}
                onSelect={onSelect}
                onToggle={onToggle}
              />
            )}
          </div>
        )
      })}
    </>
  )
}

function KeyDetail({ meta, now }: { meta: RedisKeyMeta; now: number }) {
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(meta.key)
      toast.success("Key 已复制")
    } catch {
      toast.error("复制失败，请手动复制")
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-12 shrink-0 items-center gap-2 border-b px-4">
        <TooltipHint content={meta.key}>
          <h3 className="min-w-0 truncate font-mono text-sm font-medium">{meta.key}</h3>
        </TooltipHint>
        <TooltipHint content="复制 Key">
          <Button
            variant="ghost"
            size="icon"
            className="size-6 shrink-0"
            aria-label="复制 Key"
            onClick={copy}
          >
            <Copy className="size-3.5" />
          </Button>
        </TooltipHint>
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-4">
        <div className="grid gap-3 sm:grid-cols-3">
          <StatCard label="类型" value={<TypeBadge type={meta.type} />} />
          <StatCard
            label="过期时间"
            value={expiryAt(meta.ttl, now) ?? "永不过期"}
          />
          <StatCard label="内存" value={formatBytes(meta.memory)} />
        </div>

        <div className="mt-3">
          <RedisValuePanel redisKey={meta.key} />
        </div>
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="rounded-md border bg-card px-3 py-2">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <div className="mt-1 text-sm">{value}</div>
    </div>
  )
}

function TypeBadge({ type }: { type: string }) {
  if (!type) {
    return (
      <Badge
        variant="outline"
        className="text-muted-foreground inline-flex h-5 w-14 shrink-0 items-center justify-center rounded px-1 py-0 font-mono text-[10px] leading-none font-medium"
      >
        UNKNOWN
      </Badge>
    )
  }
  return (
    <Badge
      variant="outline"
      className={cn(
        "inline-flex h-5 w-14 shrink-0 items-center justify-center rounded px-1 py-0 font-mono text-[10px] leading-none font-semibold tracking-wide",
        typeTone(type)
      )}
    >
      {type.toUpperCase()}
    </Badge>
  )
}

function typeTone(type: string): string {
  switch (type) {
    case "string":
      return "text-blue-600 dark:text-blue-400 border-blue-500/30 bg-blue-500/10"
    case "hash":
      return "text-purple-600 dark:text-purple-400 border-purple-500/30 bg-purple-500/10"
    case "list":
      return "text-amber-600 dark:text-amber-400 border-amber-500/30 bg-amber-500/10"
    case "set":
      return "text-teal-600 dark:text-teal-400 border-teal-500/30 bg-teal-500/10"
    case "zset":
      return "text-pink-600 dark:text-pink-400 border-pink-500/30 bg-pink-500/10"
    case "none":
      return "text-muted-foreground border-border bg-muted/40"
    default:
      return "text-green-600 dark:text-green-400 border-green-500/30 bg-green-500/10"
  }
}

function expiryAt(ttl: number, now: number): string | null {
  const { seconds, permanent, missing } = redisTtl(ttl)
  if (permanent || missing || seconds < 0) return null
  return formatDateTime(new Date(now + seconds * 1000).toISOString())
}

function buildTree(items: RedisKeyMeta[]): TreeNode[] {
  const root: TreeNode = { name: "", path: "", children: [] }

  for (const item of items) {
    const parts = item.key.split(SEPARATOR)
    let node = root
    parts.forEach((part, index) => {
      const path = parts.slice(0, index + 1).join(SEPARATOR)
      let child = node.children.find((candidate) => candidate.name === part)
      if (!child) {
        child = { name: part, path, children: [] }
        node.children.push(child)
      }
      if (index === parts.length - 1) child.meta = item
      node = child
    })
  }

  sortTree(root.children)
  return compressTree(root.children)
}

function sortTree(nodes: TreeNode[]) {
  nodes.sort((a, b) => a.name.localeCompare(b.name))
  for (const node of nodes) sortTree(node.children)
}

function compressTree(nodes: TreeNode[]): TreeNode[] {
  return nodes.map((node) => {
    let current = node
    while (current.children.length === 1 && !current.meta) {
      const child = current.children[0]
      current = {
        name: `${current.name}${SEPARATOR}${child.name}`,
        path: child.path,
        children: child.children,
        meta: child.meta,
      }
    }
    return { ...current, children: compressTree(current.children) }
  })
}