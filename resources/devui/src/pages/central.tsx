import { useState } from "react"

import { cn } from "@/lib/utils"
import { WebTerminal } from "@/components/web-terminal"
import { WebDavExplorer } from "@/features/broker/webdav-explorer"
import { centralDavEndpoint } from "@/features/broker/webdav"
import { CentralConfigSection } from "@/features/central/config"
import { CentralBasicInfo } from "@/features/central/basic-info"
import { MetricsMonitoring } from "@/features/central/monitoring"
import { CentralCrontab } from "@/features/central/crontab"
import { CentralRoutes } from "@/features/central/routes"
import { SystemProcessSection } from "@/features/system-process/system-process"

type Tab = "monitor" | "basic" | "config" | "crontab" | "brokerRoutes" | "exposeRoutes" | "files" | "processes" | "terminal"

const TABS: { key: Tab; label: string }[] = [
  { key: "monitor", label: "运行监控" },
  { key: "basic", label: "基本信息" },
  { key: "config", label: "配置信息" },
  { key: "crontab", label: "定时任务" },
  { key: "files", label: "文件管理" },
  { key: "processes", label: "系统进程" },
  { key: "terminal", label: "虚拟终端" },
  { key: "brokerRoutes", label: "中继节点侧接口" },
  { key: "exposeRoutes", label: "用户侧接口" },
]

export function CentralPage() {
  const [tab, setTab] = useState<Tab>("monitor")

  return (
    <div className="flex h-full flex-col">
      <div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b px-4">
        {TABS.map((item) => (
          <button
            key={item.key}
            type="button"
            className={cn(
              "border-b-2 px-3 py-2 text-sm whitespace-nowrap transition-colors",
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

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {tab === "monitor" && (
          <MetricsMonitoring />
        )}

        {tab === "basic" && (
          <CentralBasicInfo />
        )}

        {tab === "config" && (
          <CentralConfigSection />
        )}

        {tab === "crontab" && (
          <CentralCrontab />
        )}

        {tab === "brokerRoutes" && (
          <CentralRoutes kind="brokers" />
        )}

        {tab === "exposeRoutes" && (
          <CentralRoutes kind="exposes" />
        )}

        {tab === "files" && (
          <WebDavExplorer {...centralDavEndpoint()} title="中心端" />
        )}

        {tab === "processes" && (
          <SystemProcessSection />
        )}

        {tab === "terminal" && (
          <WebTerminal path="/api/v3/system/pty" title="中心端" />
        )}
      </div>
    </div>
  )
}