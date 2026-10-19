import { useCallback, useEffect, useState } from "react"
import { toast } from "@/lib/toast"
import { Search, Trash2 } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { DataPagination } from "@/components/data-pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { minionApi } from "@/features/minion/api"
import { MINION_STATUSES, type Minion, type MinionStatus } from "@/features/minion/types"

const DEFAULT_PAGE_SIZE = 12

const STATUS_COLORS: Record<MinionStatus, string> = {
  1: "bg-amber-500",
  2: "bg-muted-foreground/50",
  3: "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]",
  4: "bg-destructive",
}

export function MinionsPage() {
  const [records, setRecords] = useState<Minion[]>([])
  const [current, setCurrent] = useState(1)
  const [size, setSize] = useState(DEFAULT_PAGE_SIZE)
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState("")
  const [search, setSearch] = useState("")
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<Minion | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result = await minionApi.page(current, size, search)
      setRecords(result.records ?? [])
      setTotal(result.total ?? 0)
      setCurrent(result.current ?? current)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载探针节点失败")
    } finally {
      setLoading(false)
    }
  }, [current, search, size])

  useEffect(() => {
    void load()
  }, [load])

  const remove = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await minionApi.logicalDelete([
        { key: "minion.id", operator: "eq", value: deleteTarget.id },
      ])
      toast.success("探针节点已删除")
      setDeleteTarget(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除探针节点失败")
    } finally {
      setDeleting(false)
    }
  }

  const applySearch = () => {
    setCurrent(1)
    setSearch(keyword.trim())
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={keyword}
            placeholder="搜索 IP、标签或版本"
            className="h-7 pl-9"
            onChange={(event) => setKeyword(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") applySearch()
            }}
          />
        </div>
        <Button variant="outline" size="sm" onClick={applySearch}>搜索</Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto [scrollbar-gutter:stable]">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            加载中…
          </div>
        ) : records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            暂无探针节点
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
            {records.map((minion) => (
              <MinionCard key={minion.id} minion={minion} onDelete={setDeleteTarget} />
            ))}
          </div>
        )}
      </div>

      <DataPagination
        current={current}
        size={size}
        total={total}
        loading={loading}
        onPageChange={setCurrent}
        onPageSizeChange={(nextSize) => {
          setSize(nextSize)
          setCurrent(1)
        }}
      />

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteTarget(null)
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除探针节点</DialogTitle>
            <DialogDescription>
              将逻辑删除探针「{deleteTarget?.inet}」。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" disabled={deleting} onClick={() => setDeleteTarget(null)}>
              取消
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={remove}>
              {deleting ? "删除中…" : "删除"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function MinionCard({
  minion,
  onDelete,
}: {
  minion: Minion
  onDelete: (minion: Minion) => void
}) {
  const status = MINION_STATUSES[minion.status] ?? "未知"
  const tags = minion.tags?.filter(
    (tag) => tag !== minion.inet && tag !== minion.goos
  ) ?? []
  const displayedTags = tags.slice(0, 3)
  const remainingTags = Math.max(0, tags.length - displayedTags.length)

  return (
    <article className="group flex min-w-0 flex-col gap-3 rounded-md border bg-card p-3 transition-colors hover:bg-accent/20">
      <div className="flex min-w-0 items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <TooltipHint content={status}>
            <span
              role="status"
              aria-label={status}
              className="flex size-5 shrink-0 cursor-help items-center justify-center"
            >
              <span
                className={cn("size-2.5 rounded-full", STATUS_COLORS[minion.status])}
              />
            </span>
          </TooltipHint>
          <TooltipHint content={minion.inet}>
            <h2 className="min-w-0 truncate font-mono text-sm font-semibold">
              {minion.inet}
            </h2>
          </TooltipHint>
        </div>
        <div className="flex shrink-0 items-center gap-1 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
          <TooltipHint content="删除">
            <Button
              variant="ghost"
              size="icon"
              className="size-7 text-destructive"
              aria-label="删除探针节点"
              onClick={() => onDelete(minion)}
            >
              <Trash2 />
            </Button>
          </TooltipHint>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-x-4 gap-y-2">
        <MinionField label="操作系统" value={minion.goos || "—"} />
        <MinionField label="Agent 版本" value={minion.edition || "—"} />
        <MinionField label="代理节点" value={minion.broker_name || "—"} />
        <MinionField label="运维部门" value={minion.ibu || "—"} />
        <MinionField label="机房" value={minion.idc || "—"} />
        {minion.cpu_core > 0 && <MinionField label="CPU 核心" value={String(minion.cpu_core)} />}
      </div>

      {displayedTags.length > 0 && (
        <div className="flex flex-wrap items-center gap-1">
          {displayedTags.map((tag) => <Badge key={tag} variant="secondary">{tag}</Badge>)}
          {remainingTags > 0 && <Badge variant="outline">+{remainingTags}</Badge>}
        </div>
      )}

      <div className="text-muted-foreground flex items-center justify-between border-t pt-2 text-[11px]">
        <span>{minion.unload ? "静默" : "运行中"}{minion.unstable ? " · 不稳定版本" : ""}</span>
        <span>{minion.heartbeat_at ? formatDateTime(minion.heartbeat_at) : "—"}</span>
      </div>
    </article>
  )
}

function MinionField({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={value}>
        <div className="truncate text-xs">{value}</div>
      </TooltipHint>
    </div>
  )
}
