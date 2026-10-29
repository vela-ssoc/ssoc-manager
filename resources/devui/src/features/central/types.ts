export interface CentralBootConfig {
  database: {
    uri: string
    dsn: string
  }
}

export interface CentralStaticConfig {
  spa: boolean
  dir: string
}

export interface CentralServerConfig {
  addr: string
  static: Record<string, CentralStaticConfig>
}

export interface CentralLoggerConfig {
  level: string
  console: boolean
  filename: string
  maxsize: number
  maxage: number
  maxbackups: number
  localtime: boolean
  compress: boolean
}

export interface CentralConfig {
  id: string
  server: CentralServerConfig | null
  logger: CentralLoggerConfig | null
  created_at: string
  updated_at: string
}

export interface CentralConfigResponse {
  boot: CentralBootConfig
  config: CentralConfig | null
}

export interface CentralConfigForm {
  server: CentralServerConfig
  logger: CentralLoggerConfig
}

export interface CronJobInfo {
  id: string
  name: string
  expr: string
  timeout: number
  immediate: boolean
  next?: string
  entry_id: number
}

export interface CronJobRecords {
  records: CronJobInfo[]
}

export interface RouteInfo {
  method: string
  path: string
}

export interface RouteRecords {
  records: RouteInfo[]
}