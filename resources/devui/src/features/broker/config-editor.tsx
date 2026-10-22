import { TriangleAlert } from "lucide-react"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

import type { BrokerConfig, LoggerConfig, ManagerConfig } from "./types"
import { emptyLoggerConfig, LOGGER_LEVELS } from "./types"

export function ConfigSection({
  title,
  enabled,
  inherit,
  onToggle,
  children,
}: {
  title: string
  enabled: boolean
  inherit: React.ReactNode
  onToggle: (enabled: boolean) => void
  children: React.ReactNode
}) {
  return (
    <div className="grid gap-2 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm">{title}</span>
        <button
          type="button"
          role="switch"
          aria-label={`覆写${title}`}
          aria-checked={enabled}
          data-state={enabled ? "checked" : "unchecked"}
          className="border-input data-[state=checked]:bg-primary data-[state=unchecked]:bg-input inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border transition-colors"
          onClick={() => onToggle(!enabled)}
        >
          <span
            className="bg-background pointer-events-none block size-4 rounded-full shadow-sm transition-transform data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0"
            data-state={enabled ? "checked" : "unchecked"}
          />
        </button>
      </div>
      {enabled ? (
        children
      ) : (
        <div className="text-muted-foreground text-xs">{inherit}</div>
      )}
    </div>
  )
}

export function ConfigEditor({
  config,
  managerConfig,
  onChange,
}: {
  config: BrokerConfig
  managerConfig: ManagerConfig | null
  onChange: (config: BrokerConfig) => void
}) {
  const toggleServer = (on: boolean) =>
    onChange({ ...config, server: on ? { addr: "" } : null })

  const toggleLogger = (on: boolean) =>
    onChange({ ...config, logger: on ? emptyLoggerConfig() : null })

  const toggleDatabase = (on: boolean) =>
    onChange({ ...config, database: on ? { uri: "", dsn: "" } : null })

  const updateLogger = (patch: Partial<LoggerConfig>) =>
    onChange({
      ...config,
      logger: config.logger ? { ...config.logger, ...patch } : config.logger,
    })

  return (
    <div className="grid gap-3">
      <ConfigSection
        title="服务配置"
        enabled={Boolean(config.server)}
        onToggle={toggleServer}
        inherit={
          managerConfig?.server?.addr ? (
            <div className="space-y-0.5">
              <div>继承中心端配置</div>
              <div>监听地址 {managerConfig.server.addr}</div>
            </div>
          ) : (
            "继承中心端配置"
          )
        }
      >
        <div className="grid gap-1.5">
          <Label htmlFor="broker-server-addr">监听地址</Label>
          <Input
            id="broker-server-addr"
            value={config.server?.addr ?? ""}
            placeholder="0.0.0.0:9000"
            onChange={(event) =>
              onChange({ ...config, server: { addr: event.target.value } })
            }
          />
        </div>
      </ConfigSection>

      <ConfigSection
        title="数据库配置"
        enabled={Boolean(config.database)}
        onToggle={toggleDatabase}
        inherit={
          managerConfig?.database ? (
            <div className="space-y-0.5">
              <div>继承中心端配置</div>
              {managerConfig.database.uri && (
                <div>MongoDB 地址 {managerConfig.database.uri}</div>
              )}
              {managerConfig.database.dsn && (
                <div>关系数据库地址 {managerConfig.database.dsn}</div>
              )}
            </div>
          ) : (
            "继承中心端配置"
          )
        }
      >
        <div className="grid gap-1.5">
          <Label htmlFor="broker-database-uri">MongoDB 地址</Label>
          <Input
            id="broker-database-uri"
            value={config.database?.uri ?? ""}
            placeholder="留空使用中心端默认"
            onChange={(event) =>
              onChange({
                ...config,
                database: { uri: event.target.value, dsn: config.database?.dsn ?? "" },
              })
            }
          />
          <Label htmlFor="broker-database-dsn">关系数据库地址</Label>
          <Input
            id="broker-database-dsn"
            value={config.database?.dsn ?? ""}
            placeholder="MySQL 或 openGauss，留空使用中心端默认"
            onChange={(event) =>
              onChange({
                ...config,
                database: { uri: config.database?.uri ?? "", dsn: event.target.value },
              })
            }
          />
          <div className="border-destructive/30 bg-destructive/5 flex gap-2 rounded-md border p-2.5 text-xs">
            <TriangleAlert className="text-destructive mt-px size-3.5 shrink-0" />
            <div className="text-muted-foreground space-y-1 leading-relaxed">
              <p>必须和中心端使用同一个数据库服务，通常不需要填写。</p>
              <p>
                仅当中继节点网络受限、需要通过四层代理才能访问时，此时填写中继节点视角下的连接地址。
              </p>
              <p className="text-destructive">地址填错会导致中继节点服务启动异常，请谨慎填写。</p>
            </div>
          </div>
        </div>
      </ConfigSection>

      <ConfigSection
        title="日志配置"
        enabled={Boolean(config.logger)}
        onToggle={toggleLogger}
        inherit={
          managerConfig?.logger ? (
            <div className="space-y-0.5">
              <div>继承中心端配置</div>
              <div>级别 {managerConfig.logger.level || "INFO"}</div>
              <div>控制台输出 {managerConfig.logger.console ? "开" : "关"}</div>
              {managerConfig.logger.filename && (
                <div>日志输出文件 {managerConfig.logger.filename}</div>
              )}
            </div>
          ) : (
            "继承中心端配置"
          )
        }
      >
        <div className="grid gap-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="grid gap-1.5">
              <Label>级别</Label>
              <Select
                value={config.logger?.level || "INFO"}
                onValueChange={(value) => updateLogger({ level: value })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {LOGGER_LEVELS.map((level) => (
                    <SelectItem key={level} value={level}>
                      {level}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-1.5">
              <Label>控制台输出</Label>
              <Select
                value={config.logger?.console ? "on" : "off"}
                onValueChange={(value) => updateLogger({ console: value === "on" })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="off">关</SelectItem>
                  <SelectItem value="on">开</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="broker-log-filename">日志输出文件</Label>
            <Input
              id="broker-log-filename"
              value={config.logger?.filename ?? ""}
              placeholder={
                managerConfig?.logger?.filename
                  ? `留空表示不输出日志文件，推荐写 ${managerConfig.logger.filename}`
                  : "留空表示不输出日志文件"
              }
              onChange={(event) => updateLogger({ filename: event.target.value })}
            />
          </div>

          <div className="grid gap-3 rounded-md border p-3">
            <span className="text-xs font-medium">日志轮转</span>
            <div className="grid grid-cols-3 gap-3">
              <div className="grid gap-1.5">
                <Label htmlFor="broker-log-maxsize">日志文件分割大小 MB</Label>
                <Input
                  id="broker-log-maxsize"
                  type="number"
                  min={0}
                  value={String(config.logger?.maxsize ?? 0)}
                  onChange={(event) =>
                    updateLogger({ maxsize: Number(event.target.value) || 0 })
                  }
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="broker-log-maxage">保留天数</Label>
                <Input
                  id="broker-log-maxage"
                  type="number"
                  min={0}
                  value={String(config.logger?.maxage ?? 0)}
                  onChange={(event) =>
                    updateLogger({ maxage: Number(event.target.value) || 0 })
                  }
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="broker-log-maxbackups">文件数上限</Label>
                <Input
                  id="broker-log-maxbackups"
                  type="number"
                  min={0}
                  value={String(config.logger?.maxbackups ?? 0)}
                  onChange={(event) =>
                    updateLogger({ maxbackups: Number(event.target.value) || 0 })
                  }
                />
              </div>
            </div>

            <div className="flex flex-wrap gap-x-5 gap-y-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  className="size-4 accent-primary"
                  checked={config.logger?.localtime ?? false}
                  onChange={(event) =>
                    updateLogger({ localtime: event.target.checked })
                  }
                />
                使用本地时间
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  className="size-4 accent-primary"
                  checked={config.logger?.compress ?? false}
                  onChange={(event) =>
                    updateLogger({ compress: event.target.checked })
                  }
                />
                压缩历史日志
              </label>
            </div>
          </div>
        </div>
      </ConfigSection>
    </div>
  )
}