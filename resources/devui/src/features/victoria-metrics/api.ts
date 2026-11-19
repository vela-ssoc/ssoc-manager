import { http } from "@/lib/http"

import type {
  VictoriaMetricsConfig,
  VictoriaMetricsConfigPayload,
} from "./types"

export const victoriaMetricsApi = {
  get() {
    return http.get<VictoriaMetricsConfig | null>("/victoria-metrics-config")
  },

  update(data: VictoriaMetricsConfigPayload) {
    return http.post<unknown>("/victoria-metrics-config", data)
  },
}