export interface VictoriaMetricsTenant {
  account_id: number
  project_id: number | null
}

export interface VictoriaMetricsConfig {
  id: string
  enabled: boolean
  address: string
  tenant: VictoriaMetricsTenant | null
  username: string
  password: string
  headers: Record<string, string>
  created_at: string
  updated_at: string
}

export interface VictoriaMetricsConfigForm {
  enabled: boolean
  address: string
  tenant_account: string
  tenant_project: string
  username: string
  password: string
  headers: Record<string, string>
}

export interface VictoriaMetricsConfigPayload {
  enabled: boolean
  address: string
  tenant: VictoriaMetricsTenant | null
  username: string
  password: string
  headers: Record<string, string>
}