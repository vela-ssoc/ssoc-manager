import { http } from "@/lib/http"

import type { BuildInfo } from "@/features/buildinfo/types"
import type {
  Broker,
  BrokerConnectHistoryPages,
  BrokerForm,
  BrokerRecords,
  BrokerRoutes,
  BrokerSubstreams,
  BrokerItselfConfig,
  CronJobRecords,
  ItselfConfig,
} from "./types"

export const brokerApi = {
  list() {
    return http.get<BrokerRecords>("/brokers")
  },

  info(slug: string) {
    return http.get<Broker>("/broker", { slug })
  },

  effectConfig() {
    return http.get<ItselfConfig>("/itself/config/effect")
  },

  runningConfig(slug: string) {
    return http.get<BrokerItselfConfig>(`/br/${slug}/api/v3/itself/config`)
  },

  buildInfo(slug: string) {
    return http.get<BuildInfo>(`/br/${slug}/api/v3/itself/buildinfo`)
  },

  routes(slug: string, kind: "managers" | "agents" | "exposes" = "managers") {
    return http.get<BrokerRoutes>(`/br/${slug}/api/v3/itself/route/${kind}`)
  },

  softExitDeadline(slug: string) {
    return http.get<string | { data?: string }>(`/br/${slug}/api/v3/itself/exit`)
  },

  softExit(slug: string, seconds: number) {
    return http.post<unknown>(`/br/${slug}/api/v3/itself/exit`, { seconds })
  },

  softExitCancel(slug: string) {
    return http.delete<unknown>(`/br/${slug}/api/v3/itself/exit`)
  },

  connectHistories(slug: string, page: number, size: number) {
    return http.get<BrokerConnectHistoryPages>("/broker/connect-histories", {
      slug,
      page,
      size,
    })
  },

  substreams(slug: string) {
    return http.get<BrokerSubstreams>("/broker/substreams", { slug })
  },

  closeStream(slug: string) {
    return http.delete<unknown>("/broker/stream", { slug })
  },

  killSubstream(slug: string, id: string) {
    return http.delete<unknown>("/broker/substream", { slug, id })
  },

  crontabJobs(slug: string) {
    return http.get<CronJobRecords>(`/br/${slug}/api/v3/itself/crontab/jobs`)
  },

  crontabRemove(slug: string, id: string) {
    return http.delete<unknown>(`/br/${slug}/api/v3/itself/crontab/job`, { id })
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