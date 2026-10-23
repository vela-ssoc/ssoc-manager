export interface ExposeAddress {
  name: string
  addr: string
}

export interface BrokerServerConfig {
  addr: string
}

export interface BrokerDatabaseConfig {
  uri: string
  dsn: string
}

export interface LoggerConfig {
  level: string
  console: boolean
  filename: string
  maxsize: number
  maxage: number
  maxbackups: number
  localtime: boolean
  compress: boolean
}

export interface BrokerConfig {
  server: BrokerServerConfig | null
  logger: LoggerConfig | null
  database: BrokerDatabaseConfig | null
}

export interface ManagerConfig {
  server: { addr: string }
  database: { uri: string; dsn: string; level: string }
  logger: LoggerConfig
}

export interface BrokerExecuteStat {
  goos: string
  goarch: string
  version: string
  unstable: boolean
  customized: string
  pid: number
  args: string[]
  hostname: string
  workdir: string
  executable: string
}

export interface BrokerTunnelStat {
  library: string
  module: string
  inet: string
  connected_at: string
  cumulative: number
  local_addr: string
  remote_addr: string
  receive_bytes: number
  transmit_bytes: number
}

export interface Broker {
  id: string
  slug: string
  name: string
  exposes: ExposeAddress[]
  config: BrokerConfig
  status: boolean
  tunnel_stat: BrokerTunnelStat | null
  execute_stat: BrokerExecuteStat | null
  active_streams: number
  heartbeat_at: string
  disconnect_at: string
  created_at: string
  updated_at: string
}

export interface BrokerRecords {
  records: Broker[]
}

export interface BrokerConnectHistory {
  id: string
  broker_id: string
  execute_stat: BrokerExecuteStat
  tunnel_stat: BrokerTunnelStat
  seconds: number
  disconnected_at: string
}

export interface BrokerConnectHistoryPages {
  page: number
  size: number
  total: number
  records: BrokerConnectHistory[]
}

export interface BrokerRouteInfo {
  method: string
  path: string
}

export interface BrokerRoutes {
  records: BrokerRouteInfo[]
}

export interface BrokerSubstream {
  id: string
  rx: number
  tx: number
  established_at: string
}

export interface BrokerSubstreams {
  records: BrokerSubstream[]
}

export interface BrokerForm {
  slug: string
  name: string
  exposes: ExposeAddress[]
  config: BrokerConfig
}

export const LOGGER_LEVELS = ["DEBUG", "INFO", "WARN", "ERROR"] as const

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
}

export function emptyLoggerConfig(): LoggerConfig {
  return {
    level: "INFO",
    console: false,
    filename: "",
    maxsize: 100,
    maxage: 180,
    maxbackups: 10000,
    localtime: true,
    compress: true,
  }
}

export function emptyConfig(): BrokerConfig {
  return {
    server: null,
    logger: null,
    database: null,
  }
}