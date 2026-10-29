import { useCallback, useEffect, useState } from "react"
import { Loader2, RefreshCw, Route as RouteIcon } from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import { TooltipHint } from "@/components/tooltip-hint"
import { cn } from "@/lib/utils"
import { centralApi } from "./api"
import type { RouteInfo } from "./types"

const METHOD_ORDER = ["GET", "POST", "PUT", "PATCH", "DELETE"]

function sortRoutes(routes: RouteInfo[]): RouteInfo[] {
  const rank = (method: string) => {
    const index = METHOD_ORDER.indexOf(method)
    return index >= 0 ? index : METHOD_ORDER.length
  }
  return [...routes].sort(
    (a, b) => rank(a.method) - rank(b.method) || a.path.localeCompare(b.path)
  )
}

export function CentralRoutes({ kind }: { kind: "brokers" | "exposes" }) {
  const [routes, setRoutes] = useState<RouteInfo[]>([])
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result =
        kind === "brokers" ? await centralApi.brokerRoutes() : await centralApi.exposeRoutes()
      setRoutes(sortRoutes(result.records ?? []))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "加载接口列表失败")
    } finally {
      setLoading(false)
    }
  }, [kind])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">
          {kind === "brokers" ? "中继节点侧接口" : "用户侧接口"}
        </h3>
        <TooltipHint content="刷新">
          <Button
            variant="ghost"
            size="icon"
            className="size-7"
            aria-label="刷新"
            disabled={loading}
            onClick={() => void load()}
          >
            <RefreshCw className="size-3.5" />
          </Button>
        </TooltipHint>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading && routes.length === 0 ? (
          <div className="text-muted-foreground flex items-center gap-2 text-xs">
            <Loader2 className="size-3.5 animate-spin" />
            正在加载接口列表…
          </div>
        ) : routes.length === 0 ? (
          <div className="text-muted-foreground flex flex-col items-center justify-center gap-2 py-10 text-xs">
            <RouteIcon className="size-8 opacity-40" />
            <span>暂无接口</span>
          </div>
        ) : (
          <div className="overflow-hidden rounded-md border">
            <div className="bg-muted/40 text-muted-foreground grid grid-cols-[8rem_1fr] border-b px-2 py-1 text-[11px]">
              <span>方法</span>
              <span className="px-2">路径</span>
            </div>
            {routes.map((route, index) => (
              <div
                key={`${route.method}-${route.path}-${index}`}
                className="grid grid-cols-[8rem_1fr] items-center border-b px-2 py-1.5 text-xs last:border-b-0"
              >
                <span className={cn("font-mono font-semibold", methodClass(route.method))}>
                  {route.method}
                </span>
                <TooltipHint content={route.path}>
                  <span className="block truncate px-2 font-mono">{route.path}</span>
                </TooltipHint>
              </div>
            ))}
          </div>
        )}
      </div>
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