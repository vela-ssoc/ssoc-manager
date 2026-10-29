import { useCallback, useEffect, useState } from "react"
import { Info, Plus, Power, RefreshCw, Trash2, TriangleAlert } from "lucide-react"
import { toast } from "@/lib/toast"

import { TooltipHint } from "@/components/tooltip-hint"
import { InfoTooltip } from "@/components/info-tooltip"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { useConfirmCountdown } from "@/features/broker/use-confirm-countdown"
import { formatBytes } from "@/features/broker/types"
import type { Broker, ExposeAddress } from "@/features/broker/types"

export function BasicTab({
  broker,
  slug,
  name,
  setName,
  exposes,
  onChangeExposes,
  onUpdateExpose,
}: {
  broker: Broker
  slug: string
  name: string
  setName: (value: string) => void
  exposes: ExposeAddress[]
  onChangeExposes: (rows: ExposeAddress[]) => void
  onUpdateExpose: (index: number, patch: Partial<ExposeAddress>) => void
}) {
  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">当前连接</h3>
        {broker.status ? (
          <div className="grid gap-3 sm:grid-cols-2">
            <ConnectField label="主机" value={broker.execute_stat?.hostname} />
            <ConnectField
              label="平台"
              value={[broker.execute_stat?.goos, broker.execute_stat?.goarch]
                .filter(Boolean)
                .join("/")}
            />
            <ConnectField label="版本" value={broker.execute_stat?.version} />
            <ConnectField label="网络" value={broker.tunnel_stat?.inet} />
            <ConnectField label="模块" value={broker.tunnel_stat?.module} />
            <ConnectField label="库" value={broker.tunnel_stat?.library} />
            <ConnectField
              label="连接时间"
              value={
                broker.tunnel_stat?.connected_at
                  ? formatDateTime(broker.tunnel_stat.connected_at)
                  : undefined
              }
            />
            <ConnectField label="本地地址" value={broker.tunnel_stat?.local_addr} mono />
            <ConnectField label="远端地址" value={broker.tunnel_stat?.remote_addr} mono />
            <ConnectField
              label="进程 PID"
              value={broker.execute_stat?.pid ? String(broker.execute_stat.pid) : undefined}
            />
            <ConnectField
              label="累计子流数"
              value={broker.tunnel_stat?.cumulative ? String(broker.tunnel_stat.cumulative) : undefined}
            />
            <ConnectField
              label="接收"
              value={formatBytes(broker.tunnel_stat?.receive_bytes ?? 0)}
              mono
            />
            <ConnectField
              label="发送"
              value={formatBytes(broker.tunnel_stat?.transmit_bytes ?? 0)}
              mono
            />
            <ConnectField label="工作目录" value={broker.execute_stat?.workdir} mono />
            <ConnectField label="可执行文件" value={broker.execute_stat?.executable} mono />
            <ConnectField
              label="启动参数"
              value={broker.execute_stat?.args?.length ? broker.execute_stat.args.join(" ") : undefined}
              mono
            />
          </div>
        ) : (
          <div className="text-muted-foreground text-xs">当前无连接</div>
        )}
        <div className="flex items-center justify-between">
          <StatusDot online={broker.status} />
          <span className="text-muted-foreground text-xs">
            最近心跳 {broker.heartbeat_at ? formatDateTime(broker.heartbeat_at) : "—"}
          </span>
        </div>
      </section>

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">基础信息</h3>
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label>唯一标识</Label>
            <Input value={slug} disabled />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="detail-name">名称</Label>
            <Input
              id="detail-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </div>
        </div>
        <div className="text-muted-foreground flex flex-wrap gap-x-6 gap-y-1 text-xs">
          <span>创建时间 {broker.created_at ? formatDateTime(broker.created_at) : "—"}</span>
          <span>更新时间 {broker.updated_at ? formatDateTime(broker.updated_at) : "—"}</span>
          <span>断开时间 {broker.disconnect_at ? formatDateTime(broker.disconnect_at) : "—"}</span>
        </div>
      </section>

      <ExposeSection
        exposes={exposes}
        onChange={onChangeExposes}
        onUpdate={onUpdateExpose}
      />

      <section className="grid gap-3">
        <h3 className="text-sm font-semibold">运维操作</h3>
        <div className="flex flex-wrap items-center gap-2">
          <ReconnectButton slug={slug} online={broker.status} />
          <RestartButton slug={slug} online={broker.status} />
        </div>
      </section>
    </div>
  )
}

function StatusDot({ online }: { online: boolean }) {
  return (
    <div className="flex items-center gap-2">
      <TooltipHint content={online ? "在线" : "离线"}>
        <span
          role="status"
          aria-label={online ? "在线" : "离线"}
          className="inline-flex shrink-0 items-center"
        >
          <span
            className={cn(
              "size-2 rounded-full",
              online
                ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
                : "bg-muted-foreground/50"
            )}
          />
        </span>
      </TooltipHint>
      <span className="text-xs">{online ? "在线" : "离线"}</span>
    </div>
  )
}

function ExposeSection({
  exposes,
  onChange,
  onUpdate,
}: {
  exposes: ExposeAddress[]
  onChange: (rows: ExposeAddress[]) => void
  onUpdate: (index: number, patch: Partial<ExposeAddress>) => void
}) {
  return (
    <section className="grid gap-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <h3 className="text-sm font-semibold">暴露地址</h3>
          <InfoTooltip label="暴露地址说明">
            <ul>
              <li>该地址会下发到探针节点配置中，供探针节点连接并上线。</li>
              <li>探针节点会按列表顺序依次尝试连接，直到成功为止。</li>
            </ul>
          </InfoTooltip>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...exposes, { name: "", addr: "" }])}
        >
          <Plus /> 添加地址
        </Button>
      </div>
      <div className="grid gap-2">
        {exposes.map((row, index) => (
          <div key={index} className="grid grid-cols-[1fr_1.5fr_auto] gap-2">
            <Input
              aria-label={`暴露地址 ${index + 1} 名称`}
              value={row.name}
              placeholder="名称"
              onChange={(event) => onUpdate(index, { name: event.target.value })}
            />
            <Input
              aria-label={`暴露地址 ${index + 1} 地址`}
              value={row.addr}
              placeholder="例如 broker.example.com:9000"
              onChange={(event) => onUpdate(index, { addr: event.target.value })}
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 text-muted-foreground hover:text-destructive"
              aria-label="移除地址"
              onClick={() => onChange(exposes.filter((_, i) => i !== index))}
            >
              <Trash2 />
            </Button>
          </div>
        ))}
        {exposes.length === 0 && (
          <p className="text-muted-foreground text-xs">暂无暴露地址</p>
        )}
      </div>
    </section>
  )
}

const EXIT_PRESETS = [
  { label: "立即", seconds: 0 },
  { label: "10 秒后", seconds: 10 },
  { label: "30 秒后", seconds:30 },
  { label: "1 分钟后", seconds: 60 },
  { label: "1 小时后", seconds: 3600 },
]



function ReconnectButton({ slug, online }: { slug: string; online: boolean }) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const remaining = useConfirmCountdown(open)

  const exec = async () => {
    setLoading(true)
    try {
      await brokerApi.closeStream(slug)
      toast.success("已断开主连接，中继节点将自动重连")
      setOpen(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "重连失败")
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Button
        size="sm"
        variant="outline"
        disabled={!online}
        onClick={() => setOpen(true)}
      >
        <RefreshCw /> 重连
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>重连</DialogTitle>
          </DialogHeader>
          <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
            <TriangleAlert className="size-4 shrink-0" />
            重连操作会断开中继节点与中心端的网络连接，可能导致正在进行中的业务异常，请谨慎操作。
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" disabled={loading} onClick={() => setOpen(false)}>
              取消
            </Button>
            <Button disabled={loading || remaining > 0} onClick={() => void exec()}>
              {loading ? "处理中…" : remaining > 0 ? `确定 (${remaining})` : "确定"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}

function RestartButton({ slug, online }: { slug: string; online: boolean }) {
  const [open, setOpen] = useState(false)
  const [deadline, setDeadline] = useState("")
  const [seconds, setSeconds] = useState(10)
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState("")
  const remaining = useConfirmCountdown(open)

  const loadDeadline = useCallback(async () => {
    if (!online) {
      setDeadline("")
      return
    }
    try {
      const data = await brokerApi.softExitDeadline(slug)
      const value = typeof data === "string" ? data : data?.data
      setDeadline(typeof value === "string" && value ? value : "")
    } catch {
      setDeadline("")
    }
  }, [slug, online])

  useEffect(() => {
    void loadDeadline()
  }, [loadDeadline])

  useEffect(() => {
    if (!deadline) {
      setCountdown("")
      return
    }
    const update = () => {
      const diff = new Date(deadline).getTime() - Date.now()
      if (diff <= 0) {
        setCountdown("即将执行关机")
        return
      }
      const total = Math.floor(diff / 1000)
      const h = Math.floor(total / 3600)
      const m = Math.floor((total % 3600) / 60)
      const s = total % 60
      setCountdown(
        [h > 0 ? `${h} 时` : "", m > 0 || h > 0 ? `${m} 分` : "", `${s} 秒`]
          .filter(Boolean)
          .join(" ")
      )
    }
    update()
    const timer = setInterval(update, 1000)
    return () => clearInterval(timer)
  }, [deadline])

  const exec = async () => {
    setLoading(true)
    try {
      await brokerApi.softExit(slug, seconds)
      toast.success("已提交关机，程序退出后由服务管理器拉起")
      setOpen(false)
      await loadDeadline()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "执行关机失败")
    } finally {
      setLoading(false)
    }
  }

  const cancel = async () => {
    setLoading(true)
    try {
      await brokerApi.softExitCancel(slug)
      toast.success("已取消关机")
      setDeadline("")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "取消关机失败")
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      {deadline ? (
        <Button
          size="sm"
          variant="destructive"
          disabled={loading}
          onClick={() => void cancel()}
        >
          <Power /> 取消重启{countdown ? ` (${countdown})` : ""}
        </Button>
      ) : (
        <Button
          size="sm"
          variant="outline"
          disabled={!online}
          onClick={() => setOpen(true)}
        >
          <Power /> 重启
        </Button>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>重启中继节点</DialogTitle>
          </DialogHeader>
          <div className="text-muted-foreground flex items-start gap-2 rounded-md border p-2.5 text-xs leading-relaxed">
            <Info className="mt-px size-4 shrink-0" />
            <span>
              此处重启的是<strong className="text-primary font-semibold">中继节点程序本体</strong>，并非操作系统，程序退出后需要操作系统的服务管理器拉起（需要先将中继节点程序注册为系统服务）。
            </span>
          </div>

          {deadline ? (
            <div className="grid gap-3">
              <div className="border-destructive/30 bg-destructive/5 text-destructive flex items-center gap-2 rounded-md border p-2.5 text-xs">
                <TriangleAlert className="size-4 shrink-0" />
                存在定时关机任务，倒计时 {countdown}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" disabled={loading} onClick={() => void cancel()}>
                  取消关机
                </Button>
              </div>
            </div>
          ) : (
            <div className="grid gap-3">
              <div className="flex flex-wrap justify-center gap-1.5">
                {EXIT_PRESETS.map((preset) => (
                  <button
                    key={preset.seconds}
                    type="button"
                    className={cn(
                      "rounded-md border px-3 py-1.5 text-xs transition-colors",
                      seconds === preset.seconds
                        ? "border-primary bg-primary/10 text-foreground font-medium"
                        : "text-muted-foreground hover:bg-accent/50"
                    )}
                    onClick={() => setSeconds(preset.seconds)}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  disabled={loading}
                  onClick={() => setOpen(false)}
                >
                  取消
                </Button>
                <Button disabled={loading || remaining > 0} onClick={() => void exec()}>
                  {loading ? "处理中…" : remaining > 0 ? `确定 (${remaining})` : "确定"}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
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