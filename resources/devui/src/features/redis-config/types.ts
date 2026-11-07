export interface RedisConfig {
  id?: string
  enabled: boolean
  addresses: string[] | null
  username: string
  password: string
  cluster_mode: boolean
  master_name: string
  db: number
  created_at?: string
  updated_at?: string
}

export interface RedisConfigForm {
  enabled: boolean
  addresses: string[]
  username: string
  password: string
  clusterMode: boolean
  masterName: string
  db: number
}

export interface RedisConfigPayload {
  enabled: boolean
  addresses: string[]
  username: string
  password: string
  cluster_mode: boolean
  master_name: string
  db: number
}

export interface RedisKeyMeta {
  key: string
  type: string
  ttl: number
  memory: number
}

export interface RedisKeysPage {
  keys: RedisKeyMeta[]
  cursor: number
}

export interface RedisValue {
  key: string
  type: string
  data: unknown
  cursor: number
}

export interface RedisInfoResult {
  data?: string
}

export interface RedisDoResult {
  data?: unknown
}

export interface RedisSafeCommandsResult {
  data?: string[]
}

export function redisTtl(ttl: number): {
  seconds: number
  permanent: boolean
  missing: boolean
} {
  const seconds = ttl / 1e9
  return {
    seconds,
    permanent: seconds === -1,
    missing: seconds === -2,
  }
}