import { http } from "@/lib/http"

import type { ReverseProxyForm, ReverseProxyRecords } from "./types"

export const reverseProxyApi = {
  list() {
    return http.get<ReverseProxyRecords>("/reverse-proxies")
  },

  create(data: ReverseProxyForm) {
    return http.post<unknown>("/reverse-proxy", data)
  },

  update(data: ReverseProxyForm) {
    return http.put<unknown>("/reverse-proxy", data)
  },

  remove(location: string) {
    return http.delete<unknown>("/reverse-proxy", { location })
  },
}
