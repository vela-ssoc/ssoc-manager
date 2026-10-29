import { useCallback, useEffect, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { toast } from "@/lib/toast"
import { ArrowLeft, Loader2, Save } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { ConfigEditor } from "@/features/broker/config-editor"
import { ConfigDiff } from "@/features/broker/config-diff"
import { BasicTab } from "@/features/broker/basic-tab"
import { HistorySection } from "@/features/broker/history"
import { RouteApiSection } from "@/features/broker/route-api"
import { SubstreamsSection } from "@/features/broker/substreams"
import { BrokerCrontab } from "@/features/broker/crontab"
import { SoftwareSection } from "@/features/broker/software"
import { WebDavExplorer } from "@/features/broker/webdav-explorer"
import { brokerDavEndpoint } from "@/features/broker/webdav"
import { WebTerminal } from "@/components/web-terminal"
import { MetricsMonitoring } from "@/features/central/monitoring"
import { SystemProcessSection } from "@/features/system-process/system-process"
import { emptyConfig, toManagerConfig } from "@/features/broker/types"
import type {
  Broker,
  BrokerConfig,
  BrokerConnectHistory,
  BrokerForm,
  ExposeAddress,
  ManagerConfig,
} from "@/features/broker/types"
import { useServerEvent } from "@/features/server-events/use-server-events"

type Tab = "monitor" | "basic" | "config" | "history" | "substreams" | "crontab" | "software" | "files" | "processes" | "terminal" | "brokerRoutes" | "agentRoutes" | "exposeRoutes"

const TABS: { key: Tab; label: string; onlineOnly?: boolean }[] = [
  { key: "monitor", label: "运行监控" },
  { key: "basic", label: "基本信息" },
  { key: "config", label: "配置信息" },
  { key: "history", label: "连接历史" },
  { key: "substreams", label: "活跃子流", onlineOnly: true },
  { key: "crontab", label: "定时任务", onlineOnly: true },
  { key: "software", label: "软件信息", onlineOnly: true },
  { key: "files", label: "文件管理", onlineOnly: true },
  { key: "processes", label: "系统进程", onlineOnly: true },
  { key: "terminal", label: "虚拟终端", onlineOnly: true },
  { key: "brokerRoutes", label: "中心端侧接口", onlineOnly: true },
  { key: "agentRoutes", label: "探针侧接口", onlineOnly: true },
  { key: "exposeRoutes", label: "用户侧接口", onlineOnly: true },
]



export function BrokerDetailPage() {
  const { slug = "" } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [broker, setBroker] = useState<Broker | null>(null)
  const [managerConfig, setManagerConfig] = useState<ManagerConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [tab, setTab] = useState<Tab>("monitor")

  const [name, setName] = useState("")
  const [exposes, setExposes] = useState<ExposeAddress[]>([])
  const [config, setConfig] = useState<BrokerConfig>(emptyConfig)

  const [history, setHistory] = useState<BrokerConnectHistory[]>([])
  const [historyTotal, setHistoryTotal] = useState(0)
  const [historyPage, setHistoryPage] = useState(1)
  const [historySize, setHistorySize] = useState(24)
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

  useServerEvent(["broker.online", "broker.offline"], (event) => {
    const data = event.data as { slug?: string } | null
    if (!data?.slug || data.slug !== slug) return
    const online = event.event === "broker.online"
    setBroker((prev) => (prev ? { ...prev, status: online } : prev))
  })

  useEffect(() => {
    if (tab !== "config") return
    let cancelled = false
    brokerApi
      .effectConfig()
      .then((cfg) => {
        if (!cancelled) setManagerConfig(toManagerConfig(cfg))
      })
      .catch(() => {
        if (!cancelled) setManagerConfig(null)
      })

    return () => {
      cancelled = true
    }
  }, [tab])

  useEffect(() => {
    if (tab !== "history" || !broker) return
    void loadHistory(historyPage, historySize)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, broker])

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

  const title = [slug, broker?.name].filter(Boolean).join(" · ") || "—"

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 px-4 py-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft />
        </Button>
        <TooltipHint content={title}>
          <h2 className="min-w-0 truncate text-sm font-semibold">{title}</h2>
        </TooltipHint>
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

      <div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b px-4">
        {TABS.map((item) => {
          const disabled = Boolean(item.onlineOnly) && !broker?.status
          return (
            <button
              key={item.key}
              type="button"
              className={cn(
                "border-b-2 px-3 py-2 text-sm whitespace-nowrap transition-colors",
                tab === item.key
                  ? "border-primary text-foreground font-medium"
                  : "text-muted-foreground border-transparent hover:text-foreground",
                disabled && "cursor-not-allowed opacity-50 hover:text-muted-foreground"
              )}
              onClick={() => {
                if (disabled) {
                  toast.error("中继节点离线，无法访问该功能")
                  return
                }
                setTab(item.key)
              }}
            >
              {item.label}
            </button>
          )
        })}
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
          {tab === "monitor" && (
            <MetricsMonitoring appname="ssoc-broker" instance={slug} />
          )}

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
            <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
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

          {tab === "files" && (
            <WebDavExplorer
              {...brokerDavEndpoint(slug)}
              title={broker?.name || slug}
            />
          )}

          {tab === "processes" && (
            <SystemProcessSection slug={slug} />
          )}

          {tab === "terminal" && (
            <WebTerminal
              path={`/api/v3/br/${slug}/api/v3/system/pty`}
              title={broker?.name || slug}
            />
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

          {tab === "substreams" && (
            <SubstreamsSection slug={slug} />
          )}

          {tab === "crontab" && (
            <BrokerCrontab slug={slug} />
          )}

          {tab === "software" && (
            <SoftwareSection slug={slug} />
          )}

          {tab === "brokerRoutes" && (
            <RouteApiSection slug={slug} online={broker.status} kind="managers" />
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