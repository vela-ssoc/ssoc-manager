import { http } from "@/lib/http"

import type {
  AIChatModelConfig,
  AIChatModelForm,
  AIChatModelList,
  UpstreamModelList,
} from "./types"

export const aiChatModelApi = {
  get() {
    return http.get<AIChatModelConfig | null>("/ai-chat-model")
  },

  update(data: AIChatModelForm) {
    return http.post<unknown>("/ai-chat-model", data)
  },

  models() {
    return http.get<AIChatModelList>("/ai-chat-model/models")
  },

  listModels(baseUrl: string, apiKey: string) {
    return http.post<UpstreamModelList>("/ai-chat-model/models", {
      base_url: baseUrl,
      api_key: apiKey,
    })
  },
}