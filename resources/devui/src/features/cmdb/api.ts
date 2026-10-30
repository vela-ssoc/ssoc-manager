import { http } from "@/lib/http"

import type {
  CMDBAssetPage,
  CMDBAssetQuery,
  CMDBDepartmentRecords,
} from "./types"

export const cmdbApi = {
  departments() {
    return http.get<CMDBDepartmentRecords>("/cmdb/departments")
  },

  assets(query: CMDBAssetQuery) {
    return http.get<CMDBAssetPage>("/cmdb/assets", {
      page: query.page,
      size: query.size,
      q: query.q,
      department: query.department,
      ci_type: query.ciTypes?.length ? query.ciTypes : undefined,
    })
  },
}