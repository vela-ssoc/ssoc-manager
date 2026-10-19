export interface ExposeAddress {
  name: string
  addr: string
}

export interface BrokerServerConfig {
  addr: string
}

export interface BrokerDatabaseConfig {
  uri: string
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
  logger: { level: string; console: boolean; filename: string }
}

export interface Broker {
  id: string
  slug: string
  name: string
  exposes: ExposeAddress[]
  config: BrokerConfig
  status: boolean
  heartbeat_at: string
  created_at: string
  updated_at: string
}

export interface BrokerRecords {
  records: Broker[]
}

export interface BrokerForm {
  slug: string
  name: string
  exposes: ExposeAddress[]
  config: BrokerConfig
}

export const LOGGER_LEVELS = ["DEBUG", "INFO", "WARN", "ERROR"] as const

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