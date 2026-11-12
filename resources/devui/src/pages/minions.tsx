import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "@/lib/toast"
import {
  ArrowRight,
  Building2,
  ExternalLink,
  LayoutDashboard,
  Loader2,
  Radar,
  Search,
  Trash2,
  type LucideIcon,
} from "lucide-react"

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
import { cmdbApi } from "@/features/cmdb/api"
import { DepartmentTree } from "@/features/cmdb/department-tree"
import {
  countDepartmentNodes,
  type CMDBDepartmentNode,
} from "@/features/cmdb/types"
import { minionApi } from "@/features/minion/api"
import {
  MINION_STATUSES,
  MINION_STATUS_COLORS,
  type MinionNode,
} from "@/features/minion/types"

const DEFAULT_PAGE_SIZE = 12

export function MinionsPage() {
  const [departments, setDepartments] = useState<CMDBDepartmentNode[]>([])
  const [treeLoading, setTreeLoading] = useState(true)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [selected, setSelected] = useState("")

  const [records, setRecords] = useState<MinionNode[]>([])
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(DEFAULT_PAGE_SIZE)
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState("")
  const [query, setQuery] = useState("")
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<MinionNode | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [modeTarget, setModeTarget] = useState<string | null>(null)

  const navigate = useNavigate()

  const toggleMode = (minion: MinionNode) => {
    setModeTarget((prev) => (prev === minion.id ? null : minion.id))
  }

  const closeMode = useCallback(() => setModeTarget(null), [])

  const confirmDelete = (minion: MinionNode) => {
    setModeTarget(null)
    setDeleteTarget(minion)
  }

  const chooseMode = (minion: MinionNode, mode: "legacy" | "devui") => {
    setModeTarget(null)
    if (mode === "legacy") {
      navigate(`/legacy?path=${encodeURIComponent(`/#/node/${minion.id}`)}`)
      return
    }
    navigate(`/minions/${minion.id}`)
  }

  useEffect(() => {
    let cancelled = false
    cmdbApi
      .departments()
      .then((data) => {
        if (cancelled) return
        const roots = (data.records ?? []).filter(
          (node) => node.name || node.department
        )
        setDepartments(roots)
      })
      .catch((err) => {
        if (cancelled) return
        toast.error(err instanceof Error ? err.message : "加载 CMDB 部门失败")
      })
      .finally(() => {
        if (!cancelled) setTreeLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const requestId = useRef(0)
  const load = useCallback(async () => {
    const id = ++requestId.current
    setLoading(true)
    try {
      const result = await minionApi.pageNodes(page, size, query, selected)
      if (id !== requestId.current) return
      setRecords(result.records ?? [])
      setTotal(result.total ?? 0)
    } catch (err) {
      if (id !== requestId.current) return
      setRecords([])
      setTotal(0)
      toast.error(err instanceof Error ? err.message : "加载探针节点失败")
    } finally {
      if (id === requestId.current) setLoading(false)
    }
  }, [page, query, selected, size])

  useEffect(() => {
    void load()
  }, [load])

  const departmentCount = useMemo(
    () => countDepartmentNodes(departments),
    [departments]
  )

  const selectDepartment = (department: string) => {
    setSelected(department)
    setPage(1)
  }

  const toggleDepartment = (department: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(department)) next.delete(department)
      else next.add(department)
      return next
    })
  }

  const applySearch = () => {
    const next = keyword.trim()
    if (page === 1 && next === query) {
      void load()
      return
    }
    setPage(1)
    setQuery(next)
  }

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

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-12 items-center gap-2 border-b px-4">
        <Radar className="text-muted-foreground size-4 shrink-0" />
        <h2 className="min-w-0 truncate text-sm font-semibold">探针节点</h2>
        <TooltipHint content={selected || "全部部门"}>
          <span className="text-muted-foreground max-w-64 truncate text-xs">
            {selected || "全部部门"}
          </span>
        </TooltipHint>
      </div>

      <div className="flex min-h-0 flex-1">
        <aside className="flex w-60 shrink-0 flex-col border-r">
          <div className="flex h-12 items-center justify-between gap-2 border-b px-3">
            <div className="flex items-center gap-1.5">
              <Building2 className="text-muted-foreground size-3.5" />
              <span className="text-xs font-semibold">部门</span>
            </div>
            <span className="text-muted-foreground text-[10px]">
              {treeLoading ? "加载中…" : `${departmentCount} 个`}
            </span>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            <button
              type="button"
              className={cn(
                "mb-1 flex w-full items-center gap-1 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
                selected === ""
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/60"
              )}
              onClick={() => selectDepartment("")}
            >
              全部节点
            </button>

            {treeLoading ? (
              <div className="text-muted-foreground flex items-center gap-1.5 px-2 py-3 text-xs">
                <Loader2 className="size-3.5 animate-spin" />
                正在加载部门…
              </div>
            ) : departments.length === 0 ? (
              <p className="text-muted-foreground px-2 py-3 text-xs">
                暂无部门数据
              </p>
            ) : (
              <DepartmentTree
                nodes={departments}
                depth={0}
                expanded={expanded}
                selected={selected}
                onSelect={selectDepartment}
                onToggle={toggleDepartment}
              />
            )}
          </div>
        </aside>

        <section className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-12 flex-wrap items-center gap-2 border-b px-4">
            <div className="relative min-w-0 flex-1 sm:max-w-xs">
              <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
              <Input
                value={keyword}
                placeholder="搜索 IP、机器码或中继节点"
                className="h-7 pl-9"
                onChange={(event) => setKeyword(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") applySearch()
                }}
              />
            </div>
            <Button variant="outline" size="sm" onClick={applySearch}>
              搜索
            </Button>
            <span className="text-muted-foreground ml-auto shrink-0 text-xs">
              {loading ? "加载中…" : `共 ${total} 个节点`}
            </span>
          </div>

          <div className="min-h-0 flex-1 overflow-auto p-4">
            {loading && records.length === 0 ? (
              <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
                加载中…
              </div>
            ) : records.length === 0 ? (
              <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
                暂无探针节点
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 2xl:grid-cols-3">
                {records.map((minion) => (
                  <MinionCard
                    key={minion.id}
                    minion={minion}
                    choosing={modeTarget === minion.id}
                    onToggle={toggleMode}
                    onCancel={closeMode}
                    onChoose={chooseMode}
                    onDelete={confirmDelete}
                  />
                ))}
              </div>
            )}
          </div>

          <div className="border-t px-4 py-2">
            <DataPagination
              current={page}
              size={size}
              total={total}
              loading={loading}
              onPageChange={setPage}
              onPageSizeChange={(nextSize) => {
                setSize(nextSize)
                setPage(1)
              }}
            />
          </div>
        </section>
      </div>

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteTarget(null)
        }}
      >
        <DialogContent showCloseButton={false} className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除探针节点</DialogTitle>
            <DialogDescription>
              将逻辑删除探针「{deleteTarget?.inet}」。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              disabled={deleting}
              onClick={() => setDeleteTarget(null)}
            >
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
  choosing,
  onToggle,
  onCancel,
  onChoose,
  onDelete,
}: {
  minion: MinionNode
  choosing: boolean
  onToggle: (minion: MinionNode) => void
  onCancel: () => void
  onChoose: (minion: MinionNode, mode: "legacy" | "devui") => void
  onDelete: (minion: MinionNode) => void
}) {
  const cardRef = useRef<HTMLElement>(null)
  const status = MINION_STATUSES[minion.status] ?? "未知"
  const statusColor =
    MINION_STATUS_COLORS[minion.status] ?? "bg-muted-foreground/50"

  useEffect(() => {
    if (!choosing) return
    const handlePointerDown = (event: PointerEvent) => {
      if (cardRef.current?.contains(event.target as Node)) return
      onCancel()
    }
    document.addEventListener("pointerdown", handlePointerDown)
    return () => document.removeEventListener("pointerdown", handlePointerDown)
  }, [choosing, onCancel])

  return (
    <article
      ref={cardRef}
      role="button"
      tabIndex={0}
      aria-expanded={choosing}
      aria-label={`打开探针 ${minion.inet}`}
      onClick={() => onToggle(minion)}
      onKeyDown={(event) => {
        if (event.key === "Escape" && choosing) {
          event.preventDefault()
          onCancel()
          return
        }
        if (event.target !== event.currentTarget) return
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault()
          onToggle(minion)
        }
      }}
      className={cn(
        "group flex min-w-0 cursor-pointer flex-col gap-3 rounded-md border bg-card p-3 transition-colors hover:bg-accent/20 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
        choosing && "border-primary/50 bg-accent/10 hover:bg-accent/10"
      )}
    >
      <div className="flex min-w-0 items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <TooltipHint content={status}>
            <span
              role="status"
              aria-label={status}
              className="flex size-5 shrink-0 cursor-help items-center justify-center"
            >
              <span className={cn("size-2.5 rounded-full", statusColor)} />
            </span>
          </TooltipHint>
          <TooltipHint content={minion.inet}>
            <h2 className="min-w-0 truncate font-mono text-sm font-semibold">
              {minion.inet}
            </h2>
          </TooltipHint>
          {minion.unstable && <Badge variant="outline">内测</Badge>}
          {minion.unload && <Badge variant="secondary">静默</Badge>}
        </div>
        <div className="flex shrink-0 items-center gap-1 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
          <TooltipHint content="删除">
            <Button
              variant="ghost"
              size="icon"
              className="text-destructive size-7"
              aria-label="删除探针节点"
              onClick={(event) => {
                event.stopPropagation()
                onDelete(minion)
              }}
            >
              <Trash2 />
            </Button>
          </TooltipHint>
        </div>
      </div>

      {choosing ? (
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground text-[11px]">
              选择打开方式
            </span>
            <button
              type="button"
              className="text-muted-foreground hover:text-foreground text-[11px] transition-colors"
              onClick={(event) => {
                event.stopPropagation()
                onCancel()
              }}
            >
              取消
            </button>
          </div>
          <ModeOption
            icon={ExternalLink}
            title="老 UI"
            description="旧版界面查看详情"
            autoFocus
            onClick={() => onChoose(minion, "legacy")}
          />
          <ModeOption
            icon={LayoutDashboard}
            title="开发 UI"
            description="新版详情页查看"
            onClick={() => onChoose(minion, "devui")}
          />
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-x-4 gap-y-2">
          <MinionField label="机器码" value={minion.machine_id} mono />
          <MinionField label="Agent 版本" value={minion.edition} />
          <MinionField label="操作系统" value={minion.goos} />
          <MinionField label="中继节点" value={minion.broker_name} />
          <MinionField label="部门" value={minion.org_path || minion.ibu} />
          <MinionField label="机房" value={minion.idc} />
        </div>
      )}

      <div className="text-muted-foreground flex items-center justify-between border-t pt-2 text-[11px]">
        <span className="truncate">{minion.comment || "—"}</span>
        <span className="shrink-0">
          {minion.heartbeat_at ? formatDateTime(minion.heartbeat_at) : "—"}
        </span>
      </div>
    </article>
  )
}

function ModeOption({
  icon: Icon,
  title,
  description,
  autoFocus,
  onClick,
}: {
  icon: LucideIcon
  title: string
  description: string
  autoFocus?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      autoFocus={autoFocus}
      onClick={(event) => {
        event.stopPropagation()
        onClick()
      }}
      className="group/opt flex w-full items-center gap-2 rounded-md border bg-background/60 px-2 py-1.5 text-left transition-colors hover:border-primary/40 hover:bg-accent/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <span className="bg-muted text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded transition-colors group-hover/opt:bg-primary/10 group-hover/opt:text-primary">
        <Icon className="size-3.5" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-xs leading-tight font-medium">{title}</span>
        <span className="text-muted-foreground block truncate text-[10px] leading-tight">
          {description}
        </span>
      </span>
      <ArrowRight className="text-muted-foreground size-3.5 shrink-0 opacity-0 transition-opacity group-hover/opt:opacity-100" />
    </button>
  )
}

function MinionField({
  label,
  value,
  mono,
}: {
  label: string
  value?: string
  mono?: boolean
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={text}>
        <div className={cn("truncate text-xs", mono && "font-mono")}>{text}</div>
      </TooltipHint>
    </div>
  )
}