import { useCallback, useEffect, useMemo, useState } from "react"
import { CalendarClock, Loader2, RefreshCw, Trash2 } from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { TooltipHint } from "@/components/tooltip-hint"
import { formatDateTime, formatTimeout } from "@/lib/format"
import { centralApi } from "./api"
import type { CronJobInfo } from "./types"

export function CentralCrontab() {
  const [records, setRecords] = useState<CronJobInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<CronJobInfo | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result = await centralApi.crontabJobs()
      setRecords(result.records ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载定时任务失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const sorted = useMemo(() => {
    return [...records].sort((a, b) => {
      const at = a.next ? new Date(a.next).getTime() : Number.POSITIVE_INFINITY
      const bt = b.next ? new Date(b.next).getTime() : Number.POSITIVE_INFINITY
      if (Number.isNaN(at)) return Number.isNaN(bt) ? 0 : 1
      if (Number.isNaN(bt)) return -1
      return at - bt
    })
  }, [records])

  const remove = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await centralApi.crontabRemove(deleteTarget.id)
      toast.success("定时任务已删除")
      setDeleteTarget(null)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除定时任务失败")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">定时任务</h3>
        <TooltipHint content="刷新">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="刷新"
            disabled={loading}
            onClick={() => void load()}
          >
            <RefreshCw className="size-3.5" />
          </Button>
        </TooltipHint>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex items-center gap-2 text-xs">
            <Loader2 className="size-3.5 animate-spin" />
            正在加载定时任务…
          </div>
        ) : records.length === 0 ? (
          <div className="text-muted-foreground flex flex-col items-center justify-center gap-2 py-10 text-xs">
            <CalendarClock className="size-8 opacity-40" />
            <span>暂无定时任务</span>
          </div>
        ) : (
          <div className="overflow-hidden rounded-md border">
            <div className="bg-muted/40 text-muted-foreground grid grid-cols-[minmax(0,1.3fr)_minmax(0,1.4fr)_minmax(0,1.1fr)_minmax(0,0.9fr)_minmax(0,1.2fr)_minmax(0,1fr)_2.5rem] border-b px-2 py-1 text-[11px]">
              <span>任务名</span>
              <span className="px-2">任务 ID</span>
              <span className="px-2">Cron 表达式</span>
              <span className="px-2">方法</span>
              <span className="px-2">下次执行</span>
              <span className="px-2">超时</span>
              <span />
            </div>
            {sorted.map((job) => (
              <div
                key={job.id}
                className="grid grid-cols-[minmax(0,1.3fr)_minmax(0,1.4fr)_minmax(0,1.1fr)_minmax(0,0.9fr)_minmax(0,1.2fr)_minmax(0,1fr)_2.5rem] items-center border-b px-2 py-1.5 text-xs last:border-b-0"
              >
                <TooltipHint content={job.name}>
                  <span className="block truncate font-medium">{job.name}</span>
                </TooltipHint>
                <TooltipHint content={job.id}>
                  <span className="block truncate px-2 font-mono">{job.id}</span>
                </TooltipHint>
                <TooltipHint content={job.expr}>
                  <span className="block truncate px-2 font-mono">{job.expr}</span>
                </TooltipHint>
                <span className="px-2">{job.immediate ? "立即执行" : "按计划执行"}</span>
                <span className="px-2 font-mono">{formatDateTime(job.next)}</span>
                <span className="text-muted-foreground px-2">{formatTimeout(job.timeout)}</span>
                <div className="flex justify-end">
                  <TooltipHint content="删除任务">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="text-muted-foreground hover:text-destructive size-7"
                      aria-label="删除任务"
                      onClick={() => setDeleteTarget(job)}
                    >
                      <Trash2 />
                    </Button>
                  </TooltipHint>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteTarget(null)
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>删除定时任务</DialogTitle>
            <DialogDescription>
              确定删除定时任务「{deleteTarget?.name || deleteTarget?.id}」？
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
            <Button variant="destructive" disabled={deleting} onClick={() => void remove()}>
              {deleting ? "删除中…" : "删除"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}