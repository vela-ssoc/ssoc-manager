import type { AuthSession } from "@/lib/auth-session"
import { http, legacyHttp } from "@/lib/http"

interface AuthValidResponse {
  uid: string
  bind: boolean
}

interface TotpResponse {
  url: string
  secret?: string
}

export interface OAuthProvider {
  client_id: string
  redirect_uri: string
  authorization_endpoint: string
}

export const oauthProviderApi = {
  get(origin: string) {
    return http.get<OAuthProvider | null>("/oauth/provider", { origin })
  },
}

export const oauthApi = {
  dong(code: string, redirectUri: string) {
    return http.post<AuthSession>("/oauth/dong", {
      code,
      redirect_uri: redirectUri,
    })
  },
}

export const legacyAuthApi = {
  validate(username: string, password: string) {
    return legacyHttp.post<AuthValidResponse>("/auth/valid", { username, password })
  },

  totp(uid: string) {
    return legacyHttp.post<TotpResponse>("/auth/totp", { uid })
  },

  submit(uid: string, code: string) {
    return legacyHttp.post<AuthSession>("/auth/submit", { uid, code })
  },

  logout() {
    return legacyHttp.delete<unknown>("/logout")
  },
}
