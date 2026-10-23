import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { toast } from "@/lib/toast"
import { ArrowDown, ArrowUp, ChevronsUpDown, RefreshCw, TriangleAlert, XCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime, formatDuration } from "@/lib/format"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { useConfirmCountdown } from "@/features/broker/use-confirm-countdown"
import { formatBytes } from "@/features/broker/types"
import type { BrokerSubstream } from "@/features/broker/types"

type SortKey = "id" | "rx" | "tx" | "established_at" | "duration"

interface SortState {
  key: SortKey
  desc: boolean
}

export function SubstreamsSection({ slug }: { slug: string }) {
  const [records, setRecords] = useState<BrokerSubstream[]>([])
  const [loading, setLoading] = useState(false)
  const [killTarget, setKillTarget] = useState<BrokerSubstream | null>(null)
  const [killing, setKilling] = useState(false)
  const [now, setNow] = useState(() => Date.now())
  const [sort, setSort] = useState<SortState>({ key: "id", desc: false })
  const remaining = useConfirmCountdown(killTarget !== null)
  const requestId = useRef(0)

  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const ret = await brokerApi.substreams(slug)
      if (id !== requestId.current) return
      setRecords(ret.records ?? [])
    } catch (err) {
      if (id === requestId.current) {
        toast.error(err instanceof Error ? err.message : "加载活跃子流失败")
      }
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [slug])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  const sorted = useMemo(() => sortRecords(records, sort), [records, sort])

  const toggleSort = (key: SortKey) => {
    setSort((prev) =>
      prev.key === key ? { key, desc: !prev.desc } : { key, desc: false }
    )
  }

  const handleKill = async () => {
    if (!killTarget) return
    setKilling(true)
    try {
      await brokerApi.killSubstream(slug, killTarget.id)
      toast.success("子流已结束")
      setKillTarget(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "结束子流失败")
    } finally {
      setKilling(false)
    }
  }

  return (
    <div className="grid gap-3">
      <div className="flex items-center justify-between">
        <div className="text-muted-foreground text-xs">
          共 {records.length} 条活跃子流
        </div>
        <TooltipHint content="刷新">
          <Button
            variant="outline"
            size="sm"
            className="h-7"
            disabled={loading}
            onClick={() => void load()}
          >
            <RefreshCw className={cn(loading && "animate-spin")} /> 刷新
          </Button>
        </TooltipHint>
      </div>

      <div className="overflow-hidden rounded-md border">
        <table className="w-full table-fixed text-xs">
          <thead>
            <tr className="bg-muted/40 text-muted-foreground border-b">
              <SortableTh
                className="w-48 text-left"
                active={sort.key === "id"}
                desc={sort.desc}
                onClick={() => toggleSort("id")}
              >
                子流 ID
              </SortableTh>
              <SortableTh
                className="w-28 text-right"
                align="right"
                active={sort.key === "rx"}
                desc={sort.desc}
                onClick={() => toggleSort("rx")}
              >
                接收
              </SortableTh>
              <SortableTh
                className="w-28 text-right"
                align="right"
                active={sort.key === "tx"}
                desc={sort.desc}
                onClick={() => toggleSort("tx")}
              >
                发送
              </SortableTh>
              <SortableTh
                className="text-left"
                active={sort.key === "established_at"}
                desc={sort.desc}
                onClick={() => toggleSort("established_at")}
              >
                建立时间
              </SortableTh>
              <SortableTh
                className="w-28 text-right"
                align="right"
                active={sort.key === "duration"}
                desc={sort.desc}
                onClick={() => toggleSort("duration")}
              >
                持续时间
              </SortableTh>
              <th className="w-16 px-3 py-1.5" />
            </tr>
          </thead>
          <tbody>
            {sorted.map((item) => (
              <tr key={item.id} className="border-b last:border-b-0">
                <td className="px-3 py-1.5">
                  <TooltipHint content={item.id}>
                    <div className="truncate font-mono">{item.id}</div>
                  </TooltipHint>
                </td>
                <td className="px-3 py-1.5 text-right font-mono">
                  {formatBytes(item.rx ?? 0)}
                </td>
                <td className="px-3 py-1.5 text-right font-mono">
                  {formatBytes(item.tx ?? 0)}
                </td>
                <td className="px-3 py-1.5">
                  {item.established_at ? formatDateTime(item.established_at) : "—"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono">
                  {duration(item.established_at, now)}
                </td>
                <td className="px-3 py-1.5 text-right">
                  <TooltipHint content="结束子流">
                    <button
                      type="button"
                      className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive ml-auto flex size-6 items-center justify-center rounded transition-colors"
                      aria-label="结束子流"
                      onClick={() => setKillTarget(item)}
                    >
                      <XCircle className="size-3.5" />
                    </button>
                  </TooltipHint>
                </td>
              </tr>
            ))}
            {records.length === 0 && (
              <tr>
                <td colSpan={6} className="text-muted-foreground px-3 py-6 text-center">
                  {loading ? "加载中…" : "暂无活跃子流"}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Dialog open={killTarget !== null} onOpenChange={(open) => !open && setKillTarget(null)}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>结束子流</DialogTitle>
          </DialogHeader>
          <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
            <TriangleAlert className="size-4 shrink-0" />
            强制结束子流可能导致正在进行中的业务异常，请谨慎操作。
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setKillTarget(null)}>
              取消
            </Button>
            <Button
              variant="destructive"
              disabled={killing || remaining > 0}
              onClick={() => void handleKill()}
            >
              {killing ? "处理中…" : remaining > 0 ? `确定 (${remaining})` : "确定"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function SortableTh({
  className,
  align = "left",
  active,
  desc,
  onClick,
  children,
}: {
  className?: string
  align?: "left" | "right"
  active: boolean
  desc: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  const Icon = active ? (desc ? ArrowDown : ArrowUp) : ChevronsUpDown
  return (
    <th className={cn("px-3 py-1.5 font-medium", className)}>
      <button
        type="button"
        className={cn(
          "inline-flex items-center gap-1 rounded transition-colors hover:text-foreground",
          align === "right" ? "ml-auto" : "-ml-1 px-1"
        )}
        onClick={onClick}
      >
        {children}
        <Icon className={cn("size-3", !active && "opacity-40")} />
      </button>
    </th>
  )
}

function sortRecords(records: BrokerSubstream[], sort: SortState): BrokerSubstream[] {
  const sorted = [...records].sort((a, b) => compare(a, b, sort.key))
  return sort.desc ? sorted.reverse() : sorted
}

function compare(a: BrokerSubstream, b: BrokerSubstream, key: SortKey): number {
  switch (key) {
    case "id":
      return compareBigInt(a.id, b.id)
    case "rx":
      return (a.rx ?? 0) - (b.rx ?? 0)
    case "tx":
      return (a.tx ?? 0) - (b.tx ?? 0)
    case "established_at":
      return timeValue(a.established_at) - timeValue(b.established_at)
    case "duration":
      return timeValue(b.established_at) - timeValue(a.established_at)
    default:
      return 0
  }
}

function compareBigInt(a: string, b: string): number {
  try {
    const av = BigInt(a)
    const bv = BigInt(b)
    if (av < bv) return -1
    if (av > bv) return 1
    return 0
  } catch {
    return a.localeCompare(b)
  }
}

function timeValue(value: string): number {
  if (!value) return 0
  const t = new Date(value).getTime()
  return Number.isNaN(t) ? 0 : t
}

function duration(establishedAt: string, now: number): string {
  if (!establishedAt) return "—"
  const started = new Date(establishedAt).getTime()
  if (Number.isNaN(started)) return "—"
  return formatDuration((now - started) / 1000)
}