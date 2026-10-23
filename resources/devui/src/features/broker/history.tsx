import { useState } from "react"

import { TooltipHint } from "@/components/tooltip-hint"
import { DataPagination } from "@/components/data-pagination"
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
import type { BrokerConnectHistory } from "@/features/broker/types"

export function HistorySection({
  history,
  historyPage,
  historySize,
  historyTotal,
  historyLoading,
  onHistoryPageChange,
  onHistorySizeChange,
}: {
  history: BrokerConnectHistory[]
  historyPage: number
  historySize: number
  historyTotal: number
  historyLoading: boolean
  onHistoryPageChange: (page: number) => void
  onHistorySizeChange: (size: number) => void
}) {
  const [detail, setDetail] = useState<BrokerConnectHistory | null>(null)

  return (
    <div className="grid gap-3">
      <div className="overflow-x-auto overflow-hidden rounded-md border">
        <table className="w-full table-fixed text-xs">
          <thead>
            <tr className="bg-muted/40 text-muted-foreground border-b">
              <th className="w-20 px-2 py-1.5 text-left font-medium">主机名</th>
              <th className="w-14 px-1.5 py-1.5 text-left font-medium">隧道类型</th>
              <th className="w-[72px] px-1.5 py-1.5 text-right font-medium">子流数</th>
              <th className="w-24 px-2 py-1.5 text-left font-medium">版本</th>
              <th className="w-36 px-2 py-1.5 text-left font-medium">出口 IP</th>
              <th className="w-36 px-2 py-1.5 text-left font-medium">本地地址</th>
              <th className="w-36 px-2 py-1.5 text-left font-medium">远端地址</th>
              <th className="w-20 px-2 py-1.5 text-right font-medium">接收</th>
              <th className="w-20 px-2 py-1.5 text-right font-medium">发送</th>
              <th className="w-40 px-2 py-1.5 text-left font-medium">连接时间</th>
              <th className="w-40 px-2 py-1.5 text-left font-medium">断开时间</th>
              <th className="w-24 px-2 py-1.5 text-right font-medium">连接时长</th>
            </tr>
          </thead>
          <tbody>
            {history.map((item) => {
              return (
                <tr
                  key={item.id}
                  className="hover:bg-accent/40 cursor-pointer border-b last:border-b-0"
                  onClick={() => setDetail(item)}
                >
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.execute_stat?.hostname || "—"}>
                      <div className="truncate">{item.execute_stat?.hostname || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.library || "—"}>
                      <div className="truncate">{item.tunnel_stat?.library || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {item.tunnel_stat?.cumulative ?? 0}
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.execute_stat?.version || "—"}>
                      <div className="truncate">{item.execute_stat?.version || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.inet || "—"}>
                      <div className="truncate font-mono">{item.tunnel_stat?.inet || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.local_addr || "—"}>
                      <div className="truncate font-mono">{item.tunnel_stat?.local_addr || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.tunnel_stat?.remote_addr || "—"}>
                      <div className="truncate font-mono">{item.tunnel_stat?.remote_addr || "—"}</div>
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
                      {item.disconnected_at ? formatDateTime(item.disconnected_at) : "—"}
                    </div>
                  </td>
                  <td className="px-2 py-1.5 text-right font-mono">
                    {formatDuration(item.seconds ?? 0)}
                  </td>
                </tr>
              )
            })}
            {history.length === 0 && (
              <tr>
                <td colSpan={12} className="text-muted-foreground px-2 py-4 text-center">
                  {historyLoading ? "加载中…" : "暂无连接历史"}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <DataPagination
        current={historyPage}
        size={historySize}
        total={historyTotal}
        loading={historyLoading}
        onPageChange={onHistoryPageChange}
        onPageSizeChange={onHistorySizeChange}
      />

      <HistoryDetailDialog detail={detail} onClose={() => setDetail(null)} />
    </div>
  )
}

function HistoryDetailDialog({
  detail,
  onClose,
}: {
  detail: BrokerConnectHistory | null
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
              <h4 className="text-muted-foreground text-xs font-medium">运行信息</h4>
              <div className="grid gap-3 sm:grid-cols-2">
                <ConnectField label="主机" value={detail.execute_stat?.hostname} />
                <ConnectField
                  label="平台"
                  value={[detail.execute_stat?.goos, detail.execute_stat?.goarch]
                    .filter(Boolean)
                    .join("/")}
                />
                <ConnectField label="版本" value={detail.execute_stat?.version} />
                <ConnectField label="进程 PID" value={detail.execute_stat?.pid ? String(detail.execute_stat.pid) : undefined} />
                <ConnectField label="工作目录" value={detail.execute_stat?.workdir} mono />
                <ConnectField
                  label="可执行文件"
                  value={detail.execute_stat?.executable}
                  mono
                />
                <ConnectField
                  label="启动参数"
                  value={detail.execute_stat?.args?.length ? detail.execute_stat.args.join(" ") : undefined}
                  mono
                />
              </div>
            </section>

            <section className="grid gap-3">
              <h4 className="text-muted-foreground text-xs font-medium">隧道信息</h4>
              <div className="grid gap-3 sm:grid-cols-2">
                <ConnectField label="库" value={detail.tunnel_stat?.library} />
                <ConnectField label="模块" value={detail.tunnel_stat?.module} />
                <ConnectField label="网络" value={detail.tunnel_stat?.inet} />
                <ConnectField
                  label="连接时间"
                  value={
                    detail.tunnel_stat?.connected_at
                      ? formatDateTime(detail.tunnel_stat.connected_at)
                      : undefined
                  }
                />
                <ConnectField
                  label="累计子流数"
                  value={detail.tunnel_stat?.cumulative ? String(detail.tunnel_stat.cumulative) : undefined}
                />
                <ConnectField label="本地地址" value={detail.tunnel_stat?.local_addr} mono />
                <ConnectField label="远端地址" value={detail.tunnel_stat?.remote_addr} mono />
                <ConnectField
                  label="接收"
                  value={formatBytes(detail.tunnel_stat?.receive_bytes ?? 0)}
                  mono
                />
                <ConnectField
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

function ConnectField({
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