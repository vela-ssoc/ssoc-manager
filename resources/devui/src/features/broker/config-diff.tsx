import { useEffect, useState } from "react"
import { Check, Loader2, TriangleAlert } from "lucide-react"

import { cn } from "@/lib/utils"
import { TooltipHint } from "@/components/tooltip-hint"
import { brokerApi } from "./api"
import type { BrokerConfig, LoggerConfig, ManagerConfig } from "./types"

interface Row {
  label: string
  target: string
  running: string
}

interface Section {
  title: string
  rows: Row[]
}

function resolveConfig(cfg: BrokerConfig, mgr: ManagerConfig | null): BrokerConfig {
  const logger: LoggerConfig = cfg.logger ?? {
    level: mgr?.logger?.level ?? "INFO",
    console: mgr?.logger?.console ?? false,
    filename: mgr?.logger?.filename ?? "",
    maxsize: mgr?.logger?.maxsize ?? 0,
    maxage: mgr?.logger?.maxage ?? 0,
    maxbackups: mgr?.logger?.maxbackups ?? 0,
    localtime: mgr?.logger?.localtime ?? false,
    compress: mgr?.logger?.compress ?? false,
  }

  return {
    ...cfg,
    server: cfg.server ?? (mgr?.server?.addr ? { addr: mgr.server.addr } : null),
    database: {
      uri: cfg.database?.uri || mgr?.database?.uri || "",
      dsn: cfg.database?.dsn || mgr?.database?.dsn || "",
    },
    logger,
  }
}

function onOff(value: boolean): string {
  return value ? "开" : "关"
}

function buildSections(target: BrokerConfig, running: BrokerConfig): Section[] {
  const server: Section = {
    title: "服务配置",
    rows: [
      { label: "监听地址", target: target.server?.addr ?? "", running: running.server?.addr ?? "" },
    ],
  }

  const database: Section = {
    title: "数据库配置",
    rows: [
      { label: "MongoDB 地址", target: target.database?.uri ?? "", running: running.database?.uri ?? "" },
      { label: "关系数据库地址", target: target.database?.dsn ?? "", running: running.database?.dsn ?? "" },
    ],
  }

  const logger = target.logger
  const rlogger = running.logger
  const logRows: Row[] = [
    { label: "级别", target: logger?.level ?? "", running: rlogger?.level ?? "" },
    { label: "控制台输出", target: onOff(logger?.console ?? false), running: onOff(rlogger?.console ?? false) },
    { label: "日志输出文件", target: logger?.filename ?? "", running: rlogger?.filename ?? "" },
    { label: "日志文件分割大小 MB", target: String(logger?.maxsize ?? 0), running: String(rlogger?.maxsize ?? 0) },
    { label: "保留天数", target: String(logger?.maxage ?? 0), running: String(rlogger?.maxage ?? 0) },
    { label: "文件数上限", target: String(logger?.maxbackups ?? 0), running: String(rlogger?.maxbackups ?? 0) },
    { label: "使用本地时间", target: onOff(logger?.localtime ?? false), running: onOff(rlogger?.localtime ?? false) },
    { label: "压缩历史日志", target: onOff(logger?.compress ?? false), running: onOff(rlogger?.compress ?? false) },
  ]
  const loggerSection: Section = { title: "日志配置", rows: logRows }

  return [server, database, loggerSection]
}

function hasDiff(sections: Section[]): boolean {
  return sections.some((section) =>
    section.rows.some((row) => row.target !== row.running)
  )
}

export function ConfigDiff({
  slug,
  config,
  managerConfig,
  online,
}: {
  slug: string
  config: BrokerConfig
  managerConfig: ManagerConfig | null
  online: boolean
}) {
  const [running, setRunning] = useState<BrokerConfig | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!online) {
      setRunning(null)
      return
    }

    let cancelled = false
    setLoading(true)
    brokerApi
      .runningConfig(slug)
      .then((data) => {
        if (!cancelled) setRunning(data.config)
      })
      .catch(() => {
        if (!cancelled) setRunning(null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [slug, online])

  if (!online) {
    return null
  }

  if (loading) {
    return (
      <div className="text-muted-foreground flex items-center gap-2 text-xs">
        <Loader2 className="size-3.5 animate-spin" />
        正在获取运行中配置…
      </div>
    )
  }

  if (!running) {
    return (
      <div className="text-muted-foreground text-xs">
        无法获取运行中配置，请确认节点在线且连接正常
      </div>
    )
  }

  const target = resolveConfig(config, managerConfig)
  const sections = buildSections(target, running)
  const differ = hasDiff(sections)

  return (
    <div className="grid gap-3 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium">配置生效验证</span>
        {differ ? (
          <TooltipHint content="存在未生效的变更，保存修改后需重启该中继节点服务才会生效">
            <span className="text-destructive inline-flex">
              <TriangleAlert className="size-3.5" />
            </span>
          </TooltipHint>
        ) : (
          <TooltipHint content="当前配置与生效配置一致">
            <span className="text-green-600 dark:text-green-400 inline-flex">
              <Check className="size-3.5" />
            </span>
          </TooltipHint>
        )}
      </div>

      {differ && (
        <p className="text-muted-foreground text-xs">
          已保存配置与节点运行中的配置不一致，保存修改后需重启该中继节点服务才会生效。
        </p>
      )}

      <div className="grid gap-3">
        {sections.map((section) => (
          <div key={section.title} className="grid gap-2">
            <div className="text-muted-foreground text-[11px] font-medium">
              {section.title}
            </div>
            <div className="overflow-hidden rounded-md border">
              <div className="bg-muted/40 text-muted-foreground grid grid-cols-[0.7fr_1fr_1fr] border-b px-2 py-1 text-[11px]">
                <span>配置项</span>
                <span>当前配置</span>
                <span>生效配置</span>
              </div>
              {section.rows.map((row) => {
                const diff = row.target !== row.running
                return (
                  <div
                    key={row.label}
                    className="grid grid-cols-[0.7fr_1fr_1fr] border-b px-2 py-1.5 text-xs last:border-b-0"
                  >
                    <span className="text-muted-foreground truncate pr-2">
                      {row.label}
                    </span>
                    <span className="truncate pr-2 font-mono">
                      {row.target || "—"}
                    </span>
                    <span
                      className={cn(
                        "truncate font-mono",
                        diff &&
                          "bg-destructive/15 rounded px-1 font-semibold text-destructive"
                      )}
                    >
                      {row.running || "—"}
                    </span>
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}