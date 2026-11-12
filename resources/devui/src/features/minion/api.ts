import { http, legacyHttp } from "@/lib/http"

import type {
  AgentConnectHistoryPage,
  AgentSubstreams,
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

  connectHistories(id: string, page: number, size: number) {
    return http.get<AgentConnectHistoryPage>("/agent/connect-histories", {
      id,
      page,
      size,
    })
  },

  // 探针侧接口经中心端内部代理转发到所在中继节点，真实路由为 /api/v3/agent/*。
  substreams(brokerSlug: string, id: string) {
    return http.get<AgentSubstreams>(
      `/br/${brokerSlug}/api/v3/agent/substreams`,
      { id }
    )
  },

  killSubstream(brokerSlug: string, id: string, streamId: string) {
    return http.delete<unknown>(
      `/br/${brokerSlug}/api/v3/agent/substream`,
      { id, stream_id: streamId }
    )
  },

  logicalDelete(filters: MinionFilter[]) {
    return legacyHttp.delete<unknown>("/minion", { filters })
  },
}