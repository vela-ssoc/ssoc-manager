export type ProxyHeaders = Record<string, string>

export interface ReverseProxy {
  location: string
  name: string
  manager: boolean
  broker: boolean
  header: ProxyHeaders
  upstream: string
  created_at: string
  updated_at: string
}

export interface ReverseProxyRecords {
  records: ReverseProxy[]
}

export interface ReverseProxyForm {
  location: string
  name: string
  upstream: string
  header: ProxyHeaders
  manager: boolean
  broker: boolean
}
