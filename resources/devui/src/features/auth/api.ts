import type { AuthSession } from "@/lib/auth-session"
import { legacyHttp } from "@/lib/http"

interface AuthValidResponse {
  uid: string
  bind: boolean
}

interface TotpResponse {
  url: string
  secret?: string
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
