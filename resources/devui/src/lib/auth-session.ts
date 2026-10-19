export interface AuthSession {
  id: number
  username: string
  nickname: string
  token: string
  issued_at: string
}

export type AuthProfile = Omit<AuthSession, "token">

const PROFILE_KEY = "auth-profile"

export function saveAuthSession(session: AuthSession) {
  try {
    localStorage.setItem("token", session.token)
    const { id, username, nickname, issued_at } = session
    localStorage.setItem(
      PROFILE_KEY,
      JSON.stringify({ id, username, nickname, issued_at } satisfies AuthProfile)
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
