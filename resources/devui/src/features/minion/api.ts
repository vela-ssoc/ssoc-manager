import { http, legacyHttp } from "@/lib/http"

import type {
  MinionDetail,
  MinionFilter,
  MinionNodePage,
  MinionPage,
} from "./types"

export const minionApi = {
  page(current: number, size: number, keyword: string, filters: MinionFilter[] = []) {
    return legacyHttp.get<MinionPage>("/minions", {
      current,
      size,
      keyword,
      ...(filters.length ? { filters: JSON.stringify(filters) } : {}),
    })
  },

  pageNodes(page: number, size: number, keyword: string, department?: string) {
    return http.get<MinionNodePage>("/minions", {
      page,
      size,
      q: keyword,
      department,
    })
  },

  detail(id: string) {
    return legacyHttp.get<MinionDetail>("/minion", { id })
  },

  logicalDelete(filters: MinionFilter[]) {
    return legacyHttp.delete<unknown>("/minion", { filters })
  },
}