import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "@/lib/toast"
import { ExternalLink, Network, Pencil, Plus, Search, Trash2 } from "lucide-react"

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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime } from "@/lib/format"
import { TooltipHint } from "@/components/tooltip-hint"
import { reverseProxyApi } from "@/features/reverse-proxy/api"
import { ReverseProxyDialog } from "@/features/reverse-proxy/proxy-dialog"
import type { ReverseProxy, ReverseProxyForm } from "@/features/reverse-proxy/types"

export function ReverseProxiesPage() {
  const [records, setRecords] = useState<ReverseProxy[]>([])
  const [loading, setLoading] = useState(true)
  const [query, setQuery] = useState("")
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<ReverseProxy | undefined>()
  const [submitting, setSubmitting] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ReverseProxy | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result = await reverseProxyApi.list()
      setRecords(result.records ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载反向代理失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    if (!keyword) return records
    return records.filter((item) =>
      [item.name, item.location, item.upstream].some((value) =>
        value.toLowerCase().includes(keyword)
      )
    )
  }, [query, records])

  const save = async (form: ReverseProxyForm) => {
    setSubmitting(true)
    try {
      if (editing) {
        await reverseProxyApi.update(form)
        toast.success("反向代理已更新")
      } else {
        await reverseProxyApi.create(form)
        toast.success("反向代理已创建")
      }
      setDialogOpen(false)
      setEditing(undefined)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "保存反向代理失败")
    } finally {
      setSubmitting(false)
    }
  }

  const remove = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await reverseProxyApi.remove(deleteTarget.location)
      toast.success("反向代理已删除")
      setDeleteTarget(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除反向代理失败")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={query}
            placeholder="搜索名称、路径或上游"
            className="h-7 pl-9"
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
        <Button
          size="sm"
          onClick={() => {
            setEditing(undefined)
            setDialogOpen(true)
          }}
        >
          <Plus /> 新建
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            加载中…
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2 text-sm">
            <Network className="size-9 opacity-40" />
            <span>{records.length === 0 ? "暂无反向代理" : "没有匹配项"}</span>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>路径</TableHead>
                <TableHead>上游</TableHead>
                <TableHead>Header</TableHead>
                <TableHead className="w-[18ch]">启用节点</TableHead>
                <TableHead>更新时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((item) => {
                const headers = Object.entries(item.header ?? {})
                return (
                  <TableRow key={item.location}>
                    <TableCell className="w-[20ch] max-w-[20ch] font-medium">
                      <TooltipHint content={item.name}>
                        <span className="block truncate">{item.name}</span>
                      </TooltipHint>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{item.location}</TableCell>
                    <TableCell className="w-[30ch] max-w-[30ch]">
                      <TooltipHint content={item.upstream}>
                        <span className="block truncate">{item.upstream}</span>
                      </TooltipHint>
                    </TableCell>
                    <TableCell>
                      {headers.length ? (
                        <TooltipHint content={headers.map(([key, value]) => `${key}: ${value}`).join("\n")}>
                          <div className="flex max-w-80 items-center gap-1 overflow-hidden whitespace-nowrap">
                            {headers.slice(0, 3).map(([key]) => (
                              <Badge
                                key={key}
                                variant="secondary"
                                className="max-w-28 shrink-0 truncate"
                              >
                                {key}
                              </Badge>
                            ))}
                            {headers.length > 3 && (
                              <Badge variant="outline" className="shrink-0">
                                +{headers.length - 3}
                              </Badge>
                            )}
                          </div>
                        </TooltipHint>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {item.manager && (
                          <TooltipHint content="在新标签页打开">
                            <Badge variant="outline" asChild>
                              <a
                                href={`${window.location.origin}/api/v3/reverse/${item.location}`}
                                target="_blank"
                                rel="noreferrer"
                                className="cursor-pointer hover:bg-accent"
                                onClick={(e) => e.stopPropagation()}
                              >
                                中心端
                                <ExternalLink className="size-3" />
                              </a>
                            </Badge>
                          </TooltipHint>
                        )}
                        {item.broker && <Badge variant="outline">中继节点</Badge>}
                        {!item.manager && !item.broker && (
                          <span className="text-muted-foreground">—</span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {formatDateTime(item.updated_at)}
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-1">
                        <TooltipHint content="编辑">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7"
                            aria-label="编辑"
                            onClick={() => {
                              setEditing(item)
                              setDialogOpen(true)
                            }}
                          >
                            <Pencil />
                          </Button>
                        </TooltipHint>
                        <TooltipHint content="删除">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7 text-destructive"
                            aria-label="删除"
                            onClick={() => setDeleteTarget(item)}
                          >
                            <Trash2 />
                          </Button>
                        </TooltipHint>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        )}
      </div>

      <ReverseProxyDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        initial={editing}
        submitting={submitting}
        onSubmit={save}
      />

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteTarget(null)
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除反向代理</DialogTitle>
            <DialogDescription>
              确定删除「{deleteTarget?.name}」({deleteTarget?.location})？
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
