import { http } from "@/lib/http"

import type {
  Broker,
  BrokerConfig,
  BrokerConnectHistoryPages,
  BrokerForm,
  BrokerRecords,
  BrokerRoutes,
  ManagerConfig,
} from "./types"

export const brokerApi = {
  list() {
    return http.get<BrokerRecords>("/brokers")
  },

  info(slug: string) {
    return http.get<Broker>("/broker", { slug })
  },

  config() {
    return http.get<ManagerConfig>("/config")
  },

  runningConfig(slug: string) {
    return http.get<BrokerConfig>(`/br/${slug}/api/v3/config/boot`)
  },

  routes(slug: string, kind: "brokers" | "agents" | "exposes" = "brokers") {
    return http.get<BrokerRoutes>(`/br/${slug}/api/v3/route/${kind}`)
  },

  softExitDeadline(slug: string) {
    return http.get<string | { data?: string }>(`/br/${slug}/api/v3/soft/exit`)
  },

  softExit(slug: string, seconds: number) {
    return http.post<unknown>(`/br/${slug}/api/v3/soft/exit`, { seconds })
  },

  softExitCancel(slug: string) {
    return http.delete<unknown>(`/br/${slug}/api/v3/soft/exit`)
  },

  connectHistories(slug: string, page: number, size: number) {
    return http.get<BrokerConnectHistoryPages>("/broker/connect-histories", {
      slug,
      page,
      size,
    })
  },

  create(data: BrokerForm) {
    return http.post<unknown>("/broker", toPayload(data))
  },

  update(data: BrokerForm) {
    return http.put<unknown>("/broker", toPayload(data))
  },
}

function toPayload(data: BrokerForm) {
  return {
    slug: data.slug,
    name: data.name,
    exposes: data.exposes.filter((item) => item.name.trim() && item.addr.trim()),
    config: {
      server: data.config.server ?? null,
      logger: data.config.logger ?? null,
      database: data.config.database ?? null,
    },
  }
}