export interface CMDBConfig {
  endpoint: string
  enabled: boolean
  key: string
  secret: string
  created_at: string
  updated_at: string
}

export interface CMDBConfigForm {
  endpoint: string
  enabled: boolean
  key: string
  secret: string
}