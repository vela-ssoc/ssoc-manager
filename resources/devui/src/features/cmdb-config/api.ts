import { http } from "@/lib/http"

import type { CMDBConfig, CMDBConfigForm } from "./types"

export const cmdbConfigApi = {
  get() {
    return http.get<CMDBConfig | null>("/cmdb/config")
  },

  update(data: CMDBConfigForm) {
    return http.post<void>("/cmdb/config", data)
  },
}