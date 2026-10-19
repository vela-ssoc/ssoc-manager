import { useCallback, useEffect, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { toast } from "@/lib/toast"
import { ArrowLeft, Loader2, Plus, Save, Trash2 } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { InfoTooltip } from "@/components/info-tooltip"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { ConfigEditor } from "@/features/broker/config-editor"
import { emptyConfig } from "@/features/broker/types"
import type {
  Broker,
  BrokerConfig,
  BrokerForm,
  ExposeAddress,
  ManagerConfig,
} from "@/features/broker/types"

export function BrokerDetailPage() {
  const { slug = "" } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [broker, setBroker] = useState<Broker | null>(null)
  const [managerConfig, setManagerConfig] = useState<ManagerConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const [name, setName] = useState("")
  const [exposes, setExposes] = useState<ExposeAddress[]>([])
  const [config, setConfig] = useState<BrokerConfig>(emptyConfig)

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

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    void brokerApi
      .config()
      .then(setManagerConfig)
      .catch(() => setManagerConfig(null))
  }, [])

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
        <h2 className="min-w-0 flex-1 truncate text-sm font-semibold">
          {broker?.name || slug}
        </h2>
        <Button size="sm" disabled={saving || loading} onClick={() => void save()}>
          <Save /> {saving ? "保存中…" : "保存"}
        </Button>
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
          <div className="mx-auto flex max-w-3xl flex-col gap-4">
            <div className="flex items-center gap-2">
              <TooltipHint content={broker.status ? "在线" : "离线"}>
                <span
                  role="status"
                  aria-label={broker.status ? "在线" : "离线"}
                  className="inline-flex size-6 items-center justify-center"
                >
                  <span
                    className={cn(
                      "size-2.5 rounded-full",
                      broker.status
                        ? "bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.65)]"
                        : "bg-muted-foreground/50"
                    )}
                  />
                </span>
              </TooltipHint>
              <span className="text-sm">{broker.status ? "在线" : "离线"}</span>
              <span className="text-muted-foreground ml-auto text-xs">
                最近心跳{" "}
                {broker.heartbeat_at ? formatDateTime(broker.heartbeat_at) : "—"}
              </span>
            </div>

            <section className="grid gap-3 rounded-md border bg-card p-4">
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
              </div>
            </section>

            <section className="grid gap-3 rounded-md border bg-card p-4">
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
                  onClick={() => setExposes((rows) => [...rows, { name: "", addr: "" }])}
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
                      onChange={(event) => updateExpose(index, { name: event.target.value })}
                    />
                    <Input
                      aria-label={`暴露地址 ${index + 1} 地址`}
                      value={row.addr}
                      placeholder="例如 broker.example.com:9000"
                      onChange={(event) => updateExpose(index, { addr: event.target.value })}
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-8 text-muted-foreground hover:text-destructive"
                      aria-label="移除地址"
                      onClick={() => setExposes((rows) => rows.filter((_, i) => i !== index))}
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

            <section className="grid gap-3 rounded-md border bg-card p-4">
              <div className="flex items-center gap-1.5">
                <h3 className="text-sm font-semibold">高级配置</h3>
                <InfoTooltip label="高级配置说明">
                  <p>修改保存后，需要重启该中继节点服务才会生效。</p>
                </InfoTooltip>
              </div>
              <ConfigEditor
                config={config}
                managerConfig={managerConfig}
                onChange={setConfig}
              />
            </section>
          </div>
        </div>
      )}
    </div>
  )
}