import { useCallback, useEffect, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { toast } from "@/lib/toast"
import { ArrowLeft, Loader2, Radar } from "lucide-react"

import { TooltipHint } from "@/components/tooltip-hint"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { minionApi } from "@/features/minion/api"
import { MinionBasicInfo } from "@/features/minion/basic-info"
import { AgentConnectHistorySection } from "@/features/minion/connect-history"
import { AgentSubstreamsSection } from "@/features/minion/substreams"
import {
  MINION_STATUSES,
  MINION_STATUS_COLORS,
  MINION_STATUS_ONLINE,
  type MinionDetail,
} from "@/features/minion/types"

type Tab = "basic" | "history" | "substreams"

const TABS: { key: Tab; label: string; onlineOnly?: boolean }[] = [
  { key: "basic", label: "基本信息" },
  { key: "history", label: "连接历史" },
  { key: "substreams", label: "活跃子流", onlineOnly: true },
]

export function MinionDetailPage() {
  const { id = "" } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<MinionDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<Tab>("basic")

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setDetail(await minionApi.detail(id))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载探针节点失败")
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    void load()
  }, [load])

  const online = detail?.status === MINION_STATUS_ONLINE
  const title = detail?.inet || id || "—"
  const status = detail ? (MINION_STATUSES[detail.status] ?? "未知") : ""
  const statusColor = detail
    ? (MINION_STATUS_COLORS[detail.status] ?? "bg-muted-foreground/50")
    : ""

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 px-4 py-2">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft />
        </Button>
        <Radar className="text-muted-foreground size-4 shrink-0" />
        <TooltipHint content={title}>
          <h2 className="min-w-0 truncate font-mono text-sm font-semibold">
            {title}
          </h2>
        </TooltipHint>
        {detail && (
          <TooltipHint content={status}>
            <span
              role="status"
              aria-label={status}
              className="flex size-5 shrink-0 items-center justify-center"
            >
              <span className={cn("size-2.5 rounded-full", statusColor)} />
            </span>
          </TooltipHint>
        )}
        {detail?.unstable && <Badge variant="outline">内测</Badge>}
        {detail?.unload && <Badge variant="secondary">静默</Badge>}
        {loading && (
          <Loader2 className="text-muted-foreground size-3.5 shrink-0 animate-spin" />
        )}
      </div>

      <div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b px-4">
        {TABS.map((item) => {
          const disabled = Boolean(item.onlineOnly) && !online
          return (
            <button
              key={item.key}
              type="button"
              className={cn(
                "border-b-2 px-3 py-2 text-sm whitespace-nowrap transition-colors",
                tab === item.key
                  ? "border-primary text-foreground font-medium"
                  : "text-muted-foreground border-transparent hover:text-foreground",
                disabled &&
                  "cursor-not-allowed opacity-50 hover:text-muted-foreground"
              )}
              onClick={() => {
                if (disabled) {
                  toast.error("探针节点离线，无法访问该功能")
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

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {tab === "basic" &&
          (loading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="text-muted-foreground size-5 animate-spin" />
            </div>
          ) : detail ? (
            <MinionBasicInfo detail={detail} />
          ) : (
            <div className="text-muted-foreground py-16 text-center text-sm">
              未找到该探针节点
            </div>
          ))}
        {tab === "history" && <AgentConnectHistorySection agentId={id} />}
        {tab === "substreams" &&
          (detail?.broker_slug ? (
            <AgentSubstreamsSection
              agentId={id}
              brokerSlug={detail.broker_slug}
            />
          ) : (
            <div className="text-muted-foreground py-8 text-center text-xs">
              未找到该节点所属的中继节点，无法读取活跃子流
            </div>
          ))}
      </div>
    </div>
  )
}