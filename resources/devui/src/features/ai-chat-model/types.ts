export interface AIChatModelConfig {
  id?: string
  enabled: boolean
  base_url: string
  api_key: string
  models: string[]
  prompt: string
  created_at?: string
  updated_at?: string
}

export interface AIChatModelForm {
  enabled: boolean
  base_url: string
  api_key: string
  models: string[]
  prompt: string
}

export interface UpstreamModel {
  id: string
  object?: string
  created?: number
  owned_by?: string
}

export interface UpstreamModelList {
  data?: UpstreamModel[]
}

export interface AIChatModelList {
  records: string[]
}