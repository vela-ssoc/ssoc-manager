import { http } from "@/lib/http"

import type { OAuthConfig, OAuthConfigForm } from "./types"

export const oauthConfigApi = {
  get() {
    return http.get<OAuthConfig | null>("/oauth/config")
  },

  update(data: OAuthConfigForm) {
    return http.post<unknown>("/oauth/config", data)
  },
}