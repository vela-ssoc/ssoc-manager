import { http } from "@/lib/http"

import type { BuildInfo } from "@/features/buildinfo/types"
import type {
  CentralConfig,
  CentralConfigForm,
  CentralConfigResponse,
  CronJobRecords,
  RouteRecords,
} from "./types"

export const centralApi = {
  effectConfig() {
    return http.get<CentralConfigResponse>("/itself/config/effect")
  },

  latestConfig() {
    return http.get<CentralConfig | null>("/itself/config")
  },

  updateConfig(data: CentralConfigForm) {
    return http.post<unknown>("/itself/config", data)
  },

  buildInfo() {
    return http.get<BuildInfo>("/itself/buildinfo")
  },

  exitDeadline() {
    return http.get<{ data: string }>("/itself/exit")
  },

  exit(seconds: number) {
    return http.post<unknown>("/itself/exit", { seconds })
  },

  exitCancel() {
    return http.delete<unknown>("/itself/exit")
  },

  crontabJobs() {
    return http.get<CronJobRecords>("/itself/crontab/jobs")
  },

  crontabRemove(id: string) {
    return http.delete<unknown>("/itself/crontab/job", { id })
  },

  brokerRoutes() {
    return http.get<RouteRecords>("/itself/route/brokers")
  },

  exposeRoutes() {
    return http.get<RouteRecords>("/itself/route/exposes")
  },
}