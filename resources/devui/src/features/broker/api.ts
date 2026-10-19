import { http } from "@/lib/http"

import type { Broker, BrokerForm, BrokerRecords, ManagerConfig } from "./types"

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