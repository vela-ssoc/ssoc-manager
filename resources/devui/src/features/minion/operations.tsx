import { useState } from "react"
import { RefreshCw, TriangleAlert } from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useConfirmCountdown } from "@/features/broker/use-confirm-countdown"
import { minionApi } from "@/features/minion/api"
import { MINION_STATUS_ONLINE } from "@/features/minion/types"
import type { MinionDetail } from "@/features/minion/types"

export function MinionOperations({ detail }: { detail: MinionDetail }) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const remaining = useConfirmCountdown(open)
  const online = detail.status === MINION_STATUS_ONLINE
  const brokerSlug = detail.broker_slug

  const reconnect = async () => {
    if (!brokerSlug) return
    setLoading(true)
    try {
      await minionApi.reconnect(brokerSlug, detail.id)
      toast.success("已断开主连接，探针节点将自动重连")
      setOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "重连失败")
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">运维操作</h3>
        <div className="flex flex-wrap items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={!online || !brokerSlug}
            onClick={() => setOpen(true)}
          >
            <RefreshCw /> 重连
          </Button>
        </div>
      </section>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>重连探针节点</DialogTitle>
          </DialogHeader>
          <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
            <TriangleAlert className="size-4 shrink-0" />
            重连操作会断开探针节点与中继节点的网络连接，可能导致正在进行中的业务异常，请谨慎操作。
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              disabled={loading}
              onClick={() => setOpen(false)}
            >
              取消
            </Button>
            <Button disabled={loading || remaining > 0} onClick={() => void reconnect()}>
              {loading ? "处理中…" : remaining > 0 ? `确定 (${remaining})` : "确定"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}