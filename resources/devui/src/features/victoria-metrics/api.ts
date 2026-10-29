import { http } from "@/lib/http"

import type {
  VictoriaMetricsConfig,
  VictoriaMetricsConfigForm,
} from "./types"

export const victoriaMetricsApi = {
  get() {
    return http.get<VictoriaMetricsConfig | null>("/victoria-metrics-config")
  },

  update(data: VictoriaMetricsConfigForm) {
    return http.post<unknown>("/victoria-metrics-config", data)
  },
}