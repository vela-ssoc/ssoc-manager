import { useCallback, useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "@/lib/toast"
import { ArrowDown, ArrowUp, Plus, Search } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { BrokerDialog } from "@/features/broker/broker-dialog"
import type { Broker, BrokerForm, ManagerConfig } from "@/features/broker/types"
import { formatBytes } from "@/features/broker/types"

export function BrokersPage() {
  const navigate = useNavigate()
  const [records, setRecords] = useState<Broker[]>([])
  const [managerConfig, setManagerConfig] = useState<ManagerConfig | null>(null)
  const [keyword, setKeyword] = useState("")
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result = await brokerApi.list()
      setRecords(result.records ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载中继节点失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    void brokerApi
      .config()
      .then(setManagerConfig)
      .catch(() => setManagerConfig(null))
  }, [])

  const filtered = useMemo(() => {
    const search = keyword.trim().toLowerCase()
    if (!search) return records
    return records.filter((item) =>
      [item.name, item.slug, ...item.exposes.map((e) => e.addr)].some((value) =>
        value?.toLowerCase().includes(search)
      )
    )
  }, [keyword, records])

  const create = async (form: BrokerForm) => {
    setSaving(true)
    try {
      await brokerApi.create(form)
      toast.success("中继节点已创建")
      setDialogOpen(false)
      await load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "创建中继节点失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-4">
      <div className="flex items-center gap-2">
        <div className="relative w-72 max-w-full">
          <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
          <Input
            value={keyword}
            placeholder="搜索名称、标识或地址"
            className="h-7 pl-9"
            onChange={(event) => setKeyword(event.target.value)}
          />
        </div>
        <Button size="sm" onClick={() => setDialogOpen(true)}>
          <Plus /> 新建
        </Button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto [scrollbar-gutter:stable]">
        {loading && records.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            加载中…
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            {records.length === 0 ? "暂无中继节点" : "没有匹配项"}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
            {filtered.map((broker) => (
              <div
                key={broker.id}
                role="button"
                tabIndex={0}
                onClick={() => navigate(`/brokers/${broker.slug}`)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") navigate(`/brokers/${broker.slug}`)
                }}
                className="flex cursor-pointer flex-col gap-3 rounded-md border bg-card p-3 transition-colors hover:bg-accent/20"
              >
                <div className="flex min-w-0 items-center gap-2">
                  <TooltipHint content={broker.name || broker.slug}>
                    <h2 className="min-w-0 truncate text-sm font-semibold">
                      {broker.name || broker.slug}
                    </h2>
                  </TooltipHint>
                  <TooltipHint content={broker.status ? "在线" : "离线"}>
                    <span
                      role="status"
                      aria-label={broker.status ? "在线" : "离线"}
                      className="inline-flex shrink-0 cursor-help items-center"
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
                  <div className="ml-auto flex shrink-0 items-center gap-2 text-[11px]">
                    <TrafficField
                      direction="up"
                      value={formatBytes(broker.tunnel_stat?.transmit_bytes ?? 0)}
                    />
                    <TrafficField
                      direction="down"
                      value={formatBytes(broker.tunnel_stat?.receive_bytes ?? 0)}
                    />
                  </div>
                </div>

                <div className="grid gap-x-3 gap-y-2 sm:grid-cols-2">
                  <BrokerField label="主机" value={broker.execute_stat?.hostname || "—"} />
                  <BrokerField
                    label="平台"
                    value={
                      [broker.execute_stat?.goos, broker.execute_stat?.goarch]
                        .filter(Boolean)
                        .join("/") || "—"
                    }
                  />
                  <BrokerField label="版本" value={broker.execute_stat?.version || "—"} mono />
                  <BrokerField label="网络" value={broker.tunnel_stat?.inet || "—"} />
                  <BrokerField
                    label="模块"
                    value={broker.tunnel_stat?.module || "—"}
                  />
                  <BrokerField
                    label="活跃子流数"
                    value={String(broker.active_streams ?? 0)}
                  />
                  <BrokerField
                    label="累计子流"
                    value={String(broker.tunnel_stat?.cumulative ?? 0)}
                  />
                  <BrokerField
                    label="远端地址"
                    value={broker.tunnel_stat?.remote_addr || "—"}
                    mono
                  />
                  <BrokerField
                    label="最近心跳"
                    value={
                      broker.heartbeat_at ? formatDateTime(broker.heartbeat_at) : "—"
                    }
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <BrokerDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        managerConfig={managerConfig}
        submitting={saving}
        onSubmit={create}
      />
    </div>
  )
}

function BrokerField({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px]">{label}</div>
      <TooltipHint content={value}>
        <div className={cn("truncate text-xs", mono && "font-mono")}>{value}</div>
      </TooltipHint>
    </div>
  )
}

function TrafficField({
  direction,
  value,
}: {
  direction: "up" | "down"
  value: string
}) {
  const Icon = direction === "up" ? ArrowUp : ArrowDown
  return (
    <span className="inline-flex items-center gap-1 font-mono">
      <Icon
        className={cn(
          "size-3",
          direction === "up" ? "text-green-500" : "text-orange-500"
        )}
      />
      <span>{value}</span>
    </span>
  )
}