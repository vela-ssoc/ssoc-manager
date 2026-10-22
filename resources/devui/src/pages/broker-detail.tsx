import { useCallback, useEffect, useMemo, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { toast } from "@/lib/toast"
import Editor from "@monaco-editor/react"
import { ArrowLeft, Braces, Loader2, Plus, Play, Power, Save, Trash2, TriangleAlert } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { InfoTooltip } from "@/components/info-tooltip"
import { Button } from "@/components/ui/button"
import { DataPagination } from "@/components/data-pagination"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { rawRequest, type RawResponse } from "@/lib/http"
import { useIsDark } from "@/features/zip-file/file-icon"
import { getMonacoFontFamily } from "@/lib/fonts"
import { useTheme } from "@/providers/theme-provider"
import { brokerApi } from "@/features/broker/api"
import { ConfigEditor } from "@/features/broker/config-editor"
import { ConfigDiff } from "@/features/broker/config-diff"
import { emptyConfig, formatBytes } from "@/features/broker/types"
import type {
  Broker,
  BrokerConfig,
  BrokerConnectHistory,
  BrokerForm,
  BrokerRouteInfo,
  ExposeAddress,
  ManagerConfig,
} from "@/features/broker/types"

type Tab = "basic" | "config" | "history" | "brokerRoutes" | "agentRoutes" | "exposeRoutes"

const TABS: { key: Tab; label: string }[] = [
  { key: "basic", label: "基本信息" },
  { key: "config", label: "配置信息" },
  { key: "history", label: "连接历史" },
  { key: "brokerRoutes", label: "中心端侧接口" },
  { key: "agentRoutes", label: "探针侧接口" },
  { key: "exposeRoutes", label: "用户侧接口" },
]

export function BrokerDetailPage() {
  const { slug = "" } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [broker, setBroker] = useState<Broker | null>(null)
  const [managerConfig, setManagerConfig] = useState<ManagerConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [tab, setTab] = useState<Tab>("basic")

  const [name, setName] = useState("")
  const [exposes, setExposes] = useState<ExposeAddress[]>([])
  const [config, setConfig] = useState<BrokerConfig>(emptyConfig)

  const [history, setHistory] = useState<BrokerConnectHistory[]>([])
  const [historyTotal, setHistoryTotal] = useState(0)
  const [historyPage, setHistoryPage] = useState(1)
  const [historySize, setHistorySize] = useState(12)
  const [historyLoading, setHistoryLoading] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await brokerApi.info(slug)
      setBroker(data)
      setName(data.name ?? "")
      setExposes(data.exposes?.length ? data.exposes.map((e) => ({ ...e })) : [])
      setConfig(data.config ?? emptyConfig)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载中继节点失败")
    } finally {
      setLoading(false)
    }
  }, [slug])

  const loadHistory = useCallback(
    async (page: number, size: number) => {
      if (!broker) return
      setHistoryLoading(true)
      try {
        const data = await brokerApi.connectHistories(broker.slug, page, size)
        setHistory(data.records ?? [])
        setHistoryTotal(data.total ?? 0)
        setHistoryPage(data.page ?? page)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "加载连接历史失败")
      } finally {
        setHistoryLoading(false)
      }
    },
    [broker]
  )

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    void brokerApi
      .config()
      .then(setManagerConfig)
      .catch(() => setManagerConfig(null))
  }, [])

  useEffect(() => {
    if (broker) void loadHistory(historyPage, historySize)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [broker])

  const changeHistoryPage = (page: number) => {
    setHistoryPage(page)
    void loadHistory(page, historySize)
  }

  const changeHistorySize = (size: number) => {
    setHistorySize(size)
    void loadHistory(1, size)
  }

  const updateExpose = (index: number, patch: Partial<ExposeAddress>) =>
    setExposes((rows) => rows.map((row, i) => (i === index ? { ...row, ...patch } : row)))

  const save = async () => {
    if (!broker) return
    const validExposes = exposes.filter(
      (item) => item.name.trim() && item.addr.trim()
    )
    if (!name.trim() || validExposes.length === 0) {
      toast.error("名称和暴露地址不能为空")
      return
    }
    const form: BrokerForm = { slug, name, exposes: validExposes, config }
    setSaving(true)
    try {
      await brokerApi.update(form)
      toast.success("中继节点已更新")
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "更新中继节点失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 px-4 py-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft />
        </Button>
        <h2 className="min-w-0 truncate text-sm font-semibold">
          {broker?.name || slug}
        </h2>
        {broker && (
          <TooltipHint content={broker.status ? "在线" : "离线"}>
            <span
              role="status"
              aria-label={broker.status ? "在线" : "离线"}
              className="inline-flex shrink-0 items-center"
            >
              <span
                className={cn(
                  "size-2 rounded-full",
                  broker.status
                    ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
                    : "bg-muted-foreground/50"
                )}
              />
            </span>
          </TooltipHint>
        )}
        <span className="min-w-0 flex-1" />
        {(tab === "basic" || tab === "config") && (
          <Button size="sm" disabled={saving || loading} onClick={() => void save()}>
            <Save /> {saving ? "保存中…" : "保存"}
          </Button>
        )}
      </div>

      <div className="flex items-center gap-1 border-b px-4">
        {TABS.map((item) => (
          <button
            key={item.key}
            type="button"
            className={cn(
              "border-b-2 px-3 py-2 text-sm transition-colors",
              tab === item.key
                ? "border-primary text-foreground font-medium"
                : "text-muted-foreground border-transparent hover:text-foreground"
            )}
            onClick={() => setTab(item.key)}
          >
            {item.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="text-muted-foreground size-5 animate-spin" />
        </div>
      ) : !broker ? (
        <div className="text-muted-foreground flex flex-1 items-center justify-center text-sm">
          未找到该中继节点
        </div>
      ) : (
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          {tab === "basic" && (
            <BasicTab
              broker={broker}
              slug={slug}
              name={name}
              setName={setName}
              exposes={exposes}
              onChangeExposes={setExposes}
              onUpdateExpose={updateExpose}
            />
          )}

          {tab === "config" && (
            <div className="mx-auto flex max-w-3xl flex-col gap-6">
              <ConfigEditor
                config={config}
                managerConfig={managerConfig}
                onChange={setConfig}
              />
              <ConfigDiff
                slug={slug}
                config={config}
                managerConfig={managerConfig}
                online={broker.status}
              />
            </div>
          )}

          {tab === "history" && (
            <HistorySection
              history={history}
              historyPage={historyPage}
              historySize={historySize}
              historyTotal={historyTotal}
              historyLoading={historyLoading}
              onHistoryPageChange={changeHistoryPage}
              onHistorySizeChange={changeHistorySize}
            />
          )}

          {tab === "brokerRoutes" && (
            <RouteApiSection slug={slug} online={broker.status} kind="brokers" />
          )}

          {tab === "agentRoutes" && (
            <RouteApiSection slug={slug} online={broker.status} kind="agents" />
          )}

          {tab === "exposeRoutes" && (
            <RouteApiSection slug={slug} online={broker.status} kind="exposes" />
          )}
        </div>
      )}
    </div>
  )
}

const EXIT_PRESETS = [
  { label: "立即", seconds: 0 },
  { label: "10 秒后", seconds: 10 },
  { label: "30 秒后", seconds: 30 },
  { label: "1 分钟后", seconds: 60 },
  { label: "1 小时后", seconds: 3600 },
]

function RestartButton({ slug, online }: { slug: string; online: boolean }) {
  const [open, setOpen] = useState(false)
  const [deadline, setDeadline] = useState("")
  const [seconds, setSeconds] = useState(10)
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState("")

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
                <Button disabled={loading} onClick={() => void exec()}>
                  {loading ? "处理中…" : "确定"}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}

function BasicTab({
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
    <div className="mx-auto flex max-w-3xl flex-col gap-6">
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
          <div className="flex items-center gap-2">
            <TooltipHint content={broker.status ? "在线" : "离线"}>
              <span
                role="status"
                aria-label={broker.status ? "在线" : "离线"}
                className="inline-flex shrink-0 items-center"
              >
                <span
                  className={cn(
                    "size-2 rounded-full",
                    broker.status
                      ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
                      : "bg-muted-foreground/50"
                  )}
                />
              </span>
            </TooltipHint>
            <span className="text-xs">{broker.status ? "在线" : "离线"}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground text-xs">
              最近心跳 {broker.heartbeat_at ? formatDateTime(broker.heartbeat_at) : "—"}
            </span>
            <RestartButton slug={slug} online={broker.status} />
          </div>
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

function HistorySection({
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
              <th className="w-40 px-2 py-1.5 text-left font-medium">主机</th>
              <th className="w-28 px-2 py-1.5 text-left font-medium">平台</th>
              <th className="w-24 px-2 py-1.5 text-left font-medium">版本</th>
              <th className="w-16 px-2 py-1.5 text-left font-medium">网络</th>
              <th className="w-36 px-2 py-1.5 text-left font-medium">本地地址</th>
              <th className="w-36 px-2 py-1.5 text-left font-medium">远端地址</th>
              <th className="w-20 px-2 py-1.5 text-right font-medium">接收</th>
              <th className="w-20 px-2 py-1.5 text-right font-medium">发送</th>
              <th className="w-40 px-2 py-1.5 text-left font-medium">断开时间</th>
            </tr>
          </thead>
          <tbody>
            {history.map((item) => {
              const plat = [item.execute_stat?.goos, item.execute_stat?.goarch]
                .filter(Boolean)
                .join("/")
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
                    <div className="truncate">{plat || "—"}</div>
                  </td>
                  <td className="px-2 py-1.5">
                    <TooltipHint content={item.execute_stat?.version || "—"}>
                      <div className="truncate">{item.execute_stat?.version || "—"}</div>
                    </TooltipHint>
                  </td>
                  <td className="px-2 py-1.5">
                    <div className="truncate">{item.tunnel_stat?.inet || "—"}</div>
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
                      {item.disconnected_at ? formatDateTime(item.disconnected_at) : "—"}
                    </div>
                  </td>
                </tr>
              )
            })}
            {history.length === 0 && (
              <tr>
                <td colSpan={9} className="text-muted-foreground px-2 py-4 text-center">
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

type RouteKind = "brokers" | "agents" | "exposes"

function RouteApiSection({
  slug,
  online,
  kind,
}: {
  slug: string
  online: boolean
  kind: RouteKind
}) {
  const [routes, setRoutes] = useState<BrokerRouteInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!online) {
      setRoutes([])
      setError(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(false)
    brokerApi
      .routes(slug, kind)
      .then((data) => {
        if (!cancelled) setRoutes(data.records ?? [])
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [slug, online, kind])

  if (!online) return null

  if (loading && routes.length === 0) {
    return (
      <div className="text-muted-foreground flex items-center gap-2 text-xs">
        <Loader2 className="size-3.5 animate-spin" />
        正在加载接口列表…
      </div>
    )
  }

  if (error && routes.length === 0) {
    return (
      <div className="text-muted-foreground text-xs">
        无法获取接口列表，请确认节点在线且连接正常
      </div>
    )
  }

  if (kind === "agents" || kind === "exposes") {
    return <AgentRoutesTable routes={routes} />
  }

  return <RoutesExplorer slug={slug} routes={routes} />
}

function AgentRoutesTable({ routes }: { routes: BrokerRouteInfo[] }) {
  if (routes.length === 0) {
    return <div className="text-muted-foreground text-xs">暂无探针路由</div>
  }

  return (
    <div className="overflow-hidden rounded-md border">
      <table className="w-full table-fixed text-xs">
        <thead>
          <tr className="bg-muted/40 text-muted-foreground border-b">
            <th className="w-28 px-3 py-1.5 text-left font-medium">方法</th>
            <th className="px-3 py-1.5 text-left font-medium">路径</th>
          </tr>
        </thead>
        <tbody>
          {routes.map((route, index) => (
            <tr key={index} className="border-b last:border-b-0">
              <td className="px-3 py-1.5">
                <span className={cn("font-mono text-xs font-semibold", methodClass(route.method))}>
                  {route.method}
                </span>
              </td>
              <td className="px-3 py-1.5">
                <TooltipHint content={route.path}>
                  <div className="truncate font-mono">{route.path}</div>
                </TooltipHint>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function fillPath(template: string, values: Record<string, string>): string {
  const segments = template.split("/")
  const filled = segments.map((segment) => {
    if (segment.startsWith(":")) {
      return values[segment.slice(1)] ?? segment
    }
    if (segment.startsWith("*")) {
      return values[segment.slice(1)] ?? ""
    }

    return segment
  })

  return filled.join("/")
}

type RequestPane = "params" | "headers" | "body"

const METHOD_ORDER = ["GET", "POST", "PUT", "PATCH", "DELETE"]

function RoutesExplorer({
  slug,
  routes,
}: {
  slug: string
  routes: BrokerRouteInfo[]
}) {
  const methods = useMemo(() => {
    const set = new Set<string>()
    routes.forEach((route) => set.add(route.method))
    const rest = Array.from(set).filter((m) => !METHOD_ORDER.includes(m))
    return [...METHOD_ORDER.filter((m) => set.has(m)), ...rest]
  }, [routes])

  const [method, setMethod] = useState<string>(methods[0] ?? "")
  const [pathIndex, setPathIndex] = useState(0)

  const methodRoutes = useMemo(
    () => routes.filter((route) => route.method === method),
    [routes, method]
  )

  const selected = methodRoutes[pathIndex]

  const [params, setParams] = useState<Record<string, string>>({})
  const [query, setQuery] = useState<Record<string, string>>({})
  const [headers, setHeaders] = useState<Record<string, string>>({})
  const [body, setBody] = useState("")
  const [tab, setTab] = useState<RequestPane>("params")
  const [sending, setSending] = useState(false)
  const [result, setResult] = useState<RawResponse | null>(null)

  const changeMethod = (value: string) => {
    setMethod(value)
    setPathIndex(0)
    resetRequest(setParams, setQuery, setHeaders, setBody, setResult)
  }

  const changePath = (index: number) => {
    setPathIndex(index)
    resetRequest(setParams, setQuery, setHeaders, setBody, setResult)
  }

  const hasBody = ["POST", "PUT", "PATCH", "DELETE"].includes(selected?.method ?? "")

  const send = async () => {
    if (!selected) return
    const url = buildExplorerUrl(slug, selected.path, params, query)

    setSending(true)
    setResult(null)
    try {
      const init: { method: string; headers?: Record<string, string>; body?: string } = {
        method: selected.method,
      }
      const finalHeaders: Record<string, string> = {}
      Object.entries(headers).forEach(([key, value]) => {
        if (key.trim() && value.trim()) finalHeaders[key.trim()] = value
      })
      if (Object.keys(finalHeaders).length > 0) init.headers = finalHeaders
      if (hasBody && body.trim()) init.body = body

      const res = await rawRequest(url, init)
      setResult(res)
    } catch (err) {
      setResult({
        status: 0,
        body: err instanceof Error ? err.message : "请求失败",
        contentType: "text/plain",
        headers: {},
      })
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex min-h-0 flex-col">
      <div className="flex items-stretch gap-0 overflow-hidden rounded-md border">
        <div className="flex w-32 shrink-0 items-center border-r">
          <Select value={method} onValueChange={changeMethod}>
            <SelectTrigger className="h-full w-full rounded-none border-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {methods.map((m) => (
                <SelectItem key={m} value={m}>
                  <span className={cn("font-mono font-semibold", methodClass(m))}>{m}</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex min-w-0 flex-1 items-center border-r">
          <Select value={String(pathIndex)} onValueChange={(value) => changePath(Number(value))}>
            <SelectTrigger className="h-full w-full rounded-none border-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {methodRoutes.map((route, index) => (
                <SelectItem key={index} value={String(index)}>
                  <span className="font-mono">{route.path}</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button className="shrink-0 rounded-none" disabled={sending} onClick={() => void send()}>
          <Play /> {sending ? "发送中…" : "发送"}
        </Button>
      </div>

      <div className="mt-2 flex items-center gap-1 border-b">
        <RequestTab active={tab === "params"} onClick={() => setTab("params")}>
          参数
        </RequestTab>
        <RequestTab active={tab === "headers"} onClick={() => setTab("headers")}>
          请求头
        </RequestTab>
        {hasBody && (
          <RequestTab active={tab === "body"} onClick={() => setTab("body")}>
            请求体
          </RequestTab>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-3">
        {tab === "params" && selected && (
          <div className="grid gap-3">
            <KeyValueTable
              title="路径参数"
              rows={buildPathParamRows(selected.path, params, setParams)}
            />
            <KeyValueTable
              title="查询参数"
              rows={buildQueryRows(query, setQuery)}
              onAdd={() => setQuery((prev) => ({ ...prev, "": "" }))}
            />
          </div>
        )}

        {tab === "headers" && (
          <KeyValueTable
            title="请求头"
            rows={buildHeaderRows(headers, setHeaders)}
            onAdd={() => setHeaders((prev) => ({ ...prev, "": "" }))}
          />
        )}

        {tab === "body" && hasBody && (
          <Textarea
            value={body}
            placeholder='{"key": "value"}'
            className="min-h-48 font-mono"
            onChange={(event) => setBody(event.target.value)}
          />
        )}
      </div>

      {result && (
        <ResponseSection result={result} />
      )}
    </div>
  )
}

function ResponseSection({ result }: { result: RawResponse }) {
  const [formatted, setFormatted] = useState(false)
  const headers = Object.entries(result.headers ?? {})

  let language = "plaintext"
  if (result.contentType.includes("json")) language = "json"
  else if (result.contentType.includes("html")) language = "html"
  else if (result.contentType.includes("xml")) language = "xml"
  else if (result.contentType.includes("javascript")) language = "javascript"
  else if (result.contentType.includes("css")) language = "css"

  const isJson = result.contentType.includes("json")
  const displayBody = isJson && formatted ? prettyJson(result.body) : result.body

  return (
    <div className="mt-2 grid gap-3">
      <div className="flex items-center gap-2">
        <span className="text-muted-foreground text-xs font-medium">响应码</span>
        <span className={cn("font-mono text-xs font-semibold", statusClass(result.status))}>
          {result.status === 0 ? "发送失败" : result.status}
        </span>
        <span className="text-muted-foreground text-xs">{result.contentType}</span>
      </div>

      <div className="grid gap-2 rounded-md border p-3">
        <div className="text-muted-foreground text-xs font-medium">响应头</div>
        {headers.length === 0 ? (
          <div className="text-muted-foreground text-xs">无</div>
        ) : (
          <div className="grid max-h-40 gap-y-1 overflow-auto">
            {headers.map(([key, value]) => (
              <div key={key} className="grid grid-cols-[180px_1fr] gap-2 text-xs">
                <span className="text-muted-foreground truncate font-mono">{key}</span>
                <span className="min-w-0 truncate font-mono">{value}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="grid gap-2 rounded-md border p-3">
        <div className="flex items-center gap-1.5">
          <div className="text-muted-foreground text-xs font-medium">响应报文</div>
          {isJson && (
            <TooltipHint content={formatted ? "还原原始报文" : "格式化 JSON"}>
              <button
                type="button"
                className="text-muted-foreground hover:text-foreground inline-flex size-5 items-center justify-center rounded transition-colors"
                onClick={() => setFormatted((value) => !value)}
              >
                <Braces className="size-3.5" />
              </button>
            </TooltipHint>
          )}
        </div>
        {result.imageUrl ? (
          <img
            src={result.imageUrl}
            alt="响应图片预览"
            className="max-h-80 w-fit max-w-full rounded-md border"
          />
        ) : (
          <ResponseEditor language={language} value={displayBody} />
        )}
      </div>
    </div>
  )
}

function prettyJson(body: string): string {
  try {
    return JSON.stringify(JSON.parse(body), null, 2)
  } catch {
    return body
  }
}

function ResponseEditor({ language, value }: { language: string; value: string }) {
  const dark = useIsDark()
  const { font } = useTheme()

  return (
    <Editor
      path={`response.${language === "plaintext" ? "txt" : language}`}
      language={language}
      value={value}
      theme={dark ? "near-black" : "light"}
      beforeMount={(monaco) => {
        monaco.editor.defineTheme("near-black", {
          base: "vs-dark",
          inherit: true,
          rules: [],
          colors: {
            "editor.background": "#000000",
          },
        })
      }}
      options={{
        readOnly: true,
        minimap: { enabled: false },
        fontSize: 13,
        fontFamily: getMonacoFontFamily(font),
        scrollBeyondLastLine: false,
        automaticLayout: true,
        wordWrap: "on",
        domReadOnly: true,
      }}
      height="400px"
    />
  )
}

function resetRequest(
  setParams: (rows: Record<string, string>) => void,
  setQuery: (rows: Record<string, string>) => void,
  setHeaders: (rows: Record<string, string>) => void,
  setBody: (value: string) => void,
  setResult: (value: RawResponse | null) => void
) {
  setParams({})
  setQuery({})
  setHeaders({})
  setBody("")
  setResult(null)
}

function statusClass(status: number): string {
  if (status >= 200 && status < 300) return "text-green-600 dark:text-green-400"
  if (status === 0) return "text-destructive"

  return "text-orange-500"
}

function buildExplorerUrl(
  slug: string,
  template: string,
  params: Record<string, string>,
  query: Record<string, string>
): string {
  const path = fillPath(template, params)
  const qs = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value) qs.set(key, value)
  })
  const queryText = qs.toString()

  return `/br/${slug}${path}` + (queryText ? `?${queryText}` : "")
}

function buildPathParamRows(
  template: string,
  params: Record<string, string>,
  setParams: (rows: Record<string, string>) => void
): KeyValueRowProps[] {
  return template
    .split("/")
    .filter((segment) => segment.startsWith(":") || segment.startsWith("*"))
    .map((segment) => {
      const name = segment.slice(1)
      const wildcard = segment.startsWith("*")
      return {
        key: name,
        value: params[name] ?? "",
        keyReadonly: true,
        keyHint: wildcard ? "(通配)" : "",
        valuePlaceholder: wildcard ? "匹配后续任意路径段" : "单个路径参数",
        onChangeValue: (value) => setParams({ ...params, [name]: value }),
      }
    })
}

function buildQueryRows(
  query: Record<string, string>,
  setQuery: (rows: Record<string, string>) => void
): KeyValueRowProps[] {
  return Object.entries(query).map(([key, value]) => ({
    key,
    value,
    onChangeKey: (nextKey) => {
      const next = { ...query }
      const old = next[key]
      delete next[key]
      next[nextKey] = old ?? value
      setQuery(next)
    },
    onChangeValue: (nextValue) => setQuery({ ...query, [key]: nextValue }),
    onRemove: () => {
      const next = { ...query }
      delete next[key]
      setQuery(next)
    },
  }))
}

function buildHeaderRows(
  headers: Record<string, string>,
  setHeaders: (rows: Record<string, string>) => void
): KeyValueRowProps[] {
  return Object.entries(headers).map(([key, value]) => ({
    key,
    value,
    onChangeKey: (nextKey) => {
      const next = { ...headers }
      const old = next[key]
      delete next[key]
      next[nextKey] = old ?? value
      setHeaders(next)
    },
    onChangeValue: (nextValue) => setHeaders({ ...headers, [key]: nextValue }),
    onRemove: () => {
      const next = { ...headers }
      delete next[key]
      setHeaders(next)
    },
  }))
}

interface KeyValueRowProps {
  key: string
  value: string
  keyReadonly?: boolean
  keyHint?: string
  valuePlaceholder?: string
  onChangeKey?: (value: string) => void
  onChangeValue?: (value: string) => void
  onRemove?: () => void
}

function KeyValueTable({
  title,
  rows,
  onAdd,
}: {
  title: string
  rows: KeyValueRowProps[]
  onAdd?: () => void
}) {
  return (
    <div className="grid gap-2">
      <div className="text-muted-foreground text-xs font-medium">{title}</div>
      {rows.length === 0 && !onAdd ? (
        <div className="text-muted-foreground text-xs">无</div>
      ) : (
        <div className="grid gap-1.5">
          {rows.map((row, i) => (
            <div key={i} className="grid grid-cols-[1fr_1fr_auto] items-center gap-2">
              <Input
                value={row.key}
                readOnly={row.keyReadonly}
                placeholder={row.keyReadonly ? "参数名" : "键"}
                onChange={(event) => row.onChangeKey?.(event.target.value)}
              />
              <Input
                value={row.value}
                placeholder={row.valuePlaceholder ?? "值"}
                onChange={(event) => row.onChangeValue?.(event.target.value)}
              />
              {row.onRemove ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-8 text-muted-foreground hover:text-destructive"
                  aria-label="移除"
                  onClick={row.onRemove}
                >
                  <Trash2 />
                </Button>
              ) : (
                <div className="w-8" />
              )}
            </div>
          ))}
          {onAdd && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="w-fit"
              onClick={onAdd}
            >
              <Plus /> 添加
            </Button>
          )}
        </div>
      )}
    </div>
  )
}

function methodClass(method: string): string {
  if (method === "GET") return "text-green-600 dark:text-green-400"
  if (method === "POST") return "text-blue-600 dark:text-blue-400"
  if (method === "PUT" || method === "PATCH") return "text-orange-500"
  if (method === "DELETE") return "text-destructive"

  return "text-muted-foreground"
}

function RequestTab({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      className={cn(
        "border-b-2 px-2 py-1 text-xs transition-colors",
        active
          ? "border-primary text-foreground font-medium"
          : "text-muted-foreground border-transparent hover:text-foreground"
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}