import { http } from "@/lib/http"

import type {
  RedisConfig,
  RedisConfigPayload,
  RedisDoResult,
  RedisInfoResult,
  RedisKeysPage,
  RedisSafeCommandsResult,
  RedisValue,
} from "./types"

export const redisConfigApi = {
  get() {
    return http.get<RedisConfig | null>("/redis-config")
  },

  upsert(payload: RedisConfigPayload) {
    return http.post<unknown>("/redis-config", payload)
  },

  info(sections: string[]) {
    return http.get<RedisInfoResult>("/redis/info", { section: sections })
  },

  do(args: string[], privilege = false) {
    return http.post<RedisDoResult>("/redis/do", { args, privilege })
  },

  safeCommands() {
    return http.get<RedisSafeCommandsResult>("/redis-config/safe-commands")
  },

  keys(cursor: number, match: string, size: number) {
    return http.get<RedisKeysPage>("/redis/keys", { cursor, match, size })
  },

  value(key: string, cursor: number, match: string, size: number) {
    return http.get<RedisValue>("/redis/value", {
      key,
      cursor,
      match,
      size,
    })
  },
}