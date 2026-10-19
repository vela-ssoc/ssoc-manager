import { legacyHttp } from "@/lib/http"

import type { MinionDetail, MinionFilter, MinionPage } from "./types"

export const minionApi = {
  page(current: number, size: number, keyword: string, filters: MinionFilter[] = []) {
    return legacyHttp.get<MinionPage>("/minions", {
      current,
      size,
      keyword,
      ...(filters.length ? { filters: JSON.stringify(filters) } : {}),
    })
  },

  detail(id: string) {
    return legacyHttp.get<MinionDetail>("/minion", { id })
  },

  logicalDelete(filters: MinionFilter[]) {
    return legacyHttp.delete<unknown>("/minion", { filters })
  },
}
