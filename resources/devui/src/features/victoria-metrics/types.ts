export interface VictoriaMetricsConfig {
  id: string
  enabled: boolean
  address: string
  tenant_id: string
  username: string
  password: string
  headers: Record<string, string>
  created_at: string
  updated_at: string
}

export interface VictoriaMetricsConfigForm {
  enabled: boolean
  address: string
  tenant_id: string
  username: string
  password: string
  headers: Record<string, string>
}