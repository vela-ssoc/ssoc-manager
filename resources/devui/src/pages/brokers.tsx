import { useCallback, useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "@/lib/toast"
import { Plus, Search } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { formatDateTime } from "@/lib/format"
import { cn } from "@/lib/utils"
import { brokerApi } from "@/features/broker/api"
import { BrokerDialog } from "@/features/broker/broker-dialog"
import type { Broker, BrokerForm, ManagerConfig } from "@/features/broker/types"

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
                  <TooltipHint content={broker.status ? "在线" : "离线"}>
                    <span
                      role="status"
                      aria-label={broker.status ? "在线" : "离线"}
                      className="inline-flex size-6 shrink-0 cursor-help items-center justify-center"
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
                  <TooltipHint content={broker.name || broker.slug}>
                    <h2 className="min-w-0 truncate text-sm font-semibold">
                      {broker.name || broker.slug}
                    </h2>
                  </TooltipHint>
                </div>

                <div className="grid gap-y-2">
                  <BrokerField label="标识" value={broker.slug || "—"} mono />
                  <BrokerField
                    label="最近心跳"
                    value={
                      broker.heartbeat_at ? formatDateTime(broker.heartbeat_at) : "—"
                    }
                  />
                </div>

                <ExposeTable exposes={broker.exposes} />
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

function ExposeTable({ exposes }: { exposes: Broker["exposes"] }) {
  const items = exposes?.filter((item) => item.addr) ?? []
  if (items.length === 0) {
    return (
      <div className="text-muted-foreground text-xs">暂无暴露地址</div>
    )
  }

  return (
    <div className="overflow-hidden rounded-md border">
      <table className="w-full table-fixed text-xs">
        <thead>
          <tr className="bg-muted/40 text-muted-foreground border-b">
            <th className="w-1/3 px-2 py-1.5 text-left font-medium">名称</th>
            <th className="px-2 py-1.5 text-left font-medium">地址</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, index) => (
            <tr key={index} className="border-b last:border-b-0">
              <td className="px-2 py-1.5">
                <TooltipHint content={item.name || "—"}>
                  <div className="truncate">{item.name || "—"}</div>
                </TooltipHint>
              </td>
              <td className="px-2 py-1.5">
                <TooltipHint content={item.addr}>
                  <div className="truncate font-mono">{item.addr}</div>
                </TooltipHint>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}