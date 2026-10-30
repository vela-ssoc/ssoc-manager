export interface AuthSession {
  id: string
  company: string
  username: string
  nickname: string
  token: string
  avatar_url: string
}

export type AuthProfile = Omit<AuthSession, "token">

const PROFILE_KEY = "auth-profile"

export const AUTH_PROFILE_KEY = PROFILE_KEY

const COMPANY_LABELS: Record<string, string> = {
  group: "集团",
  security: "证券",
  ttfund: "天天基金",
  futures: "期货",
  dcfund: "东财基金",
}

export function companyLabel(company?: string | null): string {
  if (!company) return ""
  return COMPANY_LABELS[company] ?? company
}

export function saveAuthSession(session: AuthSession) {
  try {
    localStorage.setItem("token", session.token)
    const { id, company, username, nickname, avatar_url } = session
    localStorage.setItem(
      PROFILE_KEY,
      JSON.stringify({
        id,
        company,
        username,
        nickname,
        avatar_url,
      } satisfies AuthProfile)
    )
  } catch {
    // Ignore unavailable local storage.
  }
}

export function getAuthProfile(): AuthProfile | null {
  try {
    const value = localStorage.getItem(PROFILE_KEY)
    if (!value) return null
    const profile = JSON.parse(value) as AuthProfile
    if (typeof profile.username !== "string") return null
    return profile
  } catch {
    return null
  }
}

export function clearAuthProfile() {
  try {
    localStorage.removeItem(PROFILE_KEY)
  } catch {
    // Ignore unavailable local storage.
  }
}