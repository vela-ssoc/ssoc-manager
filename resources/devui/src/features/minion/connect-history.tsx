import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "@/lib/toast"

import { DataPagination } from "@/components/data-pagination"
import { TooltipHint } from "@/components/tooltip-hint"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { formatDateTime, formatDuration } from "@/lib/format"
import { cn } from "@/lib/utils"
import { formatBytes } from "@/features/broker/types"
import { minionApi } from "@/features/minion/api"
import type { AgentBrokerEndpoint, AgentConnectHistory } from "@/features/minion/types"

const DEFAULT_PAGE_SIZE = 24

function brokerLabel(broker?: AgentBrokerEndpoint) {
  const name = broker?.name?.trim()
  const slug = broker?.slug?.trim()
  if (name && slug) return `${name}（${slug}）`
  return name || slug || "—"
}

export function AgentConnectHistorySection({ agentId }: { agentId: string }) {
  const [records, setRecords] = useState<AgentConnectHistory[]>([])
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(DEFAULT_PAGE_SIZE)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<AgentConnectHistory | null>(null)
  const requestId = useRef(0)

  const load = useCallback(
    async (nextPage: number, nextSize: number) => {
      const id = ++requestId.current
      setLoading(true)
      try {
        const ret = await minionApi.connectHistories(agentId, nextPage, nextSize)
        if (id !== requestId.current) return
        setRecords(ret.records ?? [])
        setTotal(ret.total ?? 0)
      } catch (err) {
        if (id !== requestId.current) return
        setRecords([])
        setTotal(0)
        toast.error(err instanceof Error ? err.message : "加载连接历史失败")
      } finally {
        if (id === requestId.current) setLoading(false)
      }
    },
    [agentId]
  )

  useEffect(() => {
    void load(page, size)
  }, [load, page, size])

  return (
    <div className="grid gap-3">
      <div className="overflow-hidden rounded-md border">
        <div className="overflow-x-auto">
          <table className="w-full table-fixed text-xs">
            <thead>
              <tr className="bg-muted/40 text-muted-foreground border-b">
                <th className="w-40 px-2 py-1.5 text-left font-medium">中继节点</th>
                <th className="w-24 px-2 py-1.5 text-left font-medium">版本</th>
                <th className="w-20 px-2 py-1.5 text-left font-medium">隧道类型</th>
                <th className="w-16 px-2 py-1.5 text-right font-medium">子流数</th>
                <th className="w-32 px-2 py-1.5 text-left font-medium">出口 IP</th>
                <th className="w-40 px-2 py-1.5 text-left font-medium">本地地址</th>
                <th className="w-40 px-2 py-1.5 text-left font-medium">远端地址</th>
                <th className="w-20 px-2 py-1.5 text-right font-medium">接收</th>
                <th className="w-20 px-2 py-1.5 text-right font-medium">发送</th>
                <th className="w-36 px-2 py-1.5 text-left font-medium">连接时间</th>
                <th className="w-36 px-2 py-1.5 text-left font-medium">断开时间</th>
                <th className="w-24 px-2 py-1.5 text-right font-medium">连接时长</th>
              </tr>
            </thead>
            <tbody>
              {records.map((item) => (
                <tr
                  key={item.id}
                  className="hover:bg-accent/40 cursor-pointer border-b last:border-b-0"
                  onClick={() => setDetail(item)}
                >
                  <td className="px-2 py-1.5">
                    <TooltipHint content={brokerLabel(item.broker)}>
                      <div className="truncate">{brokerLabel(item.broker)}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.execute_stat?.version || "—"}>
                      <div className="truncate">
                        {item.execute_stat?.version || "—"}
                      </div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.library || "—"}>
                      <div className="truncate">
                        {item.tunnel_stat?.library || "—"}
                      </div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {item.tunnel_stat?.cumulative ?? 0}
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.inet || "—"}>
                      <div className="truncate font-mono">
                        {item.tunnel_stat?.inet || "—"}
                      </div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.local_addr || "—"}>
                      <div className="truncate font-mono">
                        {item.tunnel_stat?.local_addr || "—"}
                      </div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.remote_addr || "—"}>
                      <div className="truncate font-mono">
                        {item.tunnel_stat?.remote_addr || "—"}
                      </div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {formatBytes(item.tunnel_stat?.receive_bytes ?? 0)}
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {formatBytes(item.tunnel_stat?.transmit_bytes ?? 0)}
                  </td>
                  <td className="px-2 py-1.5">
                    <div className="truncate">
                      {item.tunnel_stat?.connected_at
                        ? formatDateTime(item.tunnel_stat.connected_at)
                        : "—"}
                    </div>
                  </td>
                  <td className="px-2 py-1.5">
                    <div className="truncate">
                      {item.disconnected_at
                        ? formatDateTime(item.disconnected_at)
                        : "—"}
                    </div>
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {formatDuration(item.seconds ?? 0)}
                  </td>
                </tr>
              ))}
              {records.length === 0 && (
                <tr>
                  <td
                    colSpan={12}
                    className="text-muted-foreground px-2 py-4 text-center"
                  >
                    {loading ? "加载中…" : "暂无连接历史"}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

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

      <AgentHistoryDetailDialog detail={detail} onClose={() => setDetail(null)} />
    </div>
  )
}

function AgentHistoryDetailDialog({
  detail,
  onClose,
}: {
  detail: AgentConnectHistory | null
  onClose: () => void
}) {
  return (
    <Dialog open={detail !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>连接详情</DialogTitle>
          <DialogDescription>
            {detail?.disconnected_at ? formatDateTime(detail.disconnected_at) : ""}
          </DialogDescription>
        </DialogHeader>
        {detail && (
          <div className="grid gap-4">
            <section className="grid gap-3">
              <h4 className="text-muted-foreground text-xs font-medium">连接信息</h4>
              <div className="grid gap-3 sm:grid-cols-2">
                <HistoryField
                  label="中继节点"
                  value={brokerLabel(detail.broker)}
                />
                <HistoryField
                  label="连接时长"
                  value={formatDuration(detail.seconds ?? 0)}
                />
                <HistoryField
                  label="连接时间"
                  value={
                    detail.tunnel_stat?.connected_at
                      ? formatDateTime(detail.tunnel_stat.connected_at)
                      : undefined
                  }
                />
                <HistoryField
                  label="断开时间"
                  value={
                    detail.disconnected_at
                      ? formatDateTime(detail.disconnected_at)
                      : undefined
                  }
                />
              </div>
            </section>

            <section className="grid gap-3">
              <h4 className="text-muted-foreground text-xs font-medium">运行信息</h4>
              <div className="grid gap-3 sm:grid-cols-2">
                <HistoryField label="主机" value={detail.execute_stat?.hostname} />
                <HistoryField
                  label="平台"
                  value={[detail.execute_stat?.goos, detail.execute_stat?.goarch]
                    .filter(Boolean)
                    .join("/")}
                />
                <HistoryField
                  label="版本"
                  value={
                    detail.execute_stat?.version
                      ? `${detail.execute_stat.version}${detail.execute_stat.unstable ? " · 内测" : ""}`
                      : undefined
                  }
                />
                <HistoryField
                  label="定制版"
                  value={detail.execute_stat?.customized}
                />
                <HistoryField
                  label="进程 PID"
                  value={
                    detail.execute_stat?.pid
                      ? String(detail.execute_stat.pid)
                      : undefined
                  }
                />
                <HistoryField
                  label="工作目录"
                  value={detail.execute_stat?.workdir}
                  mono
                />
                <HistoryField
                  label="可执行文件"
                  value={detail.execute_stat?.executable}
                  mono
                />
                <HistoryField
                  label="启动参数"
                  value={
                    detail.execute_stat?.args?.length
                      ? detail.execute_stat.args.join(" ")
                      : undefined
                  }
                  mono
                />
              </div>
            </section>

            <section className="grid gap-3">
              <h4 className="text-muted-foreground text-xs font-medium">隧道信息</h4>
              <div className="grid gap-3 sm:grid-cols-2">
                <HistoryField label="库" value={detail.tunnel_stat?.library} />
                <HistoryField label="模块" value={detail.tunnel_stat?.module} />
                <HistoryField label="出口 IP" value={detail.tunnel_stat?.inet} mono />
                <HistoryField
                  label="累计子流数"
                  value={
                    detail.tunnel_stat?.cumulative
                      ? String(detail.tunnel_stat.cumulative)
                      : undefined
                  }
                />
                <HistoryField
                  label="本地地址"
                  value={detail.tunnel_stat?.local_addr}
                  mono
                />
                <HistoryField
                  label="远端地址"
                  value={detail.tunnel_stat?.remote_addr}
                  mono
                />
                <HistoryField
                  label="接收"
                  value={formatBytes(detail.tunnel_stat?.receive_bytes ?? 0)}
                  mono
                />
                <HistoryField
                  label="发送"
                  value={formatBytes(detail.tunnel_stat?.transmit_bytes ?? 0)}
                  mono
                />
              </div>
            </section>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

function HistoryField({
  label,
  value,
  mono = false,
}: {
  label: string
  value?: string
  mono?: boolean
}) {
  const text = value && value.trim() ? value : "—"
  return (
    <div className="grid content-start gap-0.5">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={text}>
        <div className={cn("truncate text-xs", mono && "font-mono")}>{text}</div>
      </TooltipHint>
    </div>
  )
}
