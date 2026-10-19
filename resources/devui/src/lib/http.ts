import { clearAuthProfile } from "@/lib/auth-session"

const API_BASE = "/api/v3"
const LEGACY_API_BASE = "/api/v1"

export class ApiError extends Error {
  status: number
  detail?: ProblemDetail
  kind: "http" | "network" | "aborted"

  constructor(
    status: number,
    message: string,
    detail?: ProblemDetail,
    kind: "http" | "network" | "aborted" = "http"
  ) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.detail = detail
    this.kind = kind
  }
}

export interface ProblemDetail {
  type?: string
  title?: string
  status?: number
  detail?: string
  instance?: string
}

export function getToken(): string | null {
  try {
    return localStorage.getItem("token")
  } catch {
    return null
  }
}

export function setToken(token: string) {
  try {
    localStorage.setItem("token", token)
  } catch {
    // Ignore unavailable local storage.
  }
}

export function clearToken() {
  try {
    localStorage.removeItem("token")
  } catch {
    // Ignore unavailable local storage.
  }
  clearAuthProfile()
}

function redirectToLogin() {
  clearToken()
  const base = import.meta.env.BASE_URL.endsWith("/")
    ? import.meta.env.BASE_URL
    : `${import.meta.env.BASE_URL}/`
  const loginPath = `${base}login`
  if (window.location.pathname !== loginPath) {
    window.location.replace(loginPath)
  }
}

function buildUrl(
  path: string,
  query?: Record<string, unknown>,
  apiBase = API_BASE
): string {
  const url = path.startsWith("/") ? apiBase + path : path
  if (!query) return url

  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === "") continue
    params.set(key, String(value))
  }
  const qs = params.toString()
  return qs ? `${url}?${qs}` : url
}

async function parseError(response: Response): Promise<ApiError> {
  if (response.status === 401) redirectToLogin()

  let detail: ProblemDetail | undefined
  let message = `请求失败（${response.status}）`
  try {
    const contentType = response.headers.get("content-type") ?? ""
    if (contentType.includes("json")) {
      detail = (await response.json()) as ProblemDetail
      message = detail?.detail || detail?.title || message
    }
  } catch {
    // ignore
  }
  return new ApiError(response.status, message, detail)
}

async function fetchResponse(url: string, init: RequestInit): Promise<Response> {
  try {
    return await fetch(url, init)
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw new ApiError(0, "请求已取消", undefined, "aborted")
    }
    throw new ApiError(
      0,
      "网络连接失败，请检查网络连接后重试",
      undefined,
      "network"
    )
  }
}

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown
  query?: Record<string, unknown>
}

async function request<T>(
  path: string,
  init: RequestOptions = {},
  apiBase = API_BASE
): Promise<T> {
  const { query, body, headers, ...rest } = init
  const url = buildUrl(path, query, apiBase)

  const finalHeaders = new Headers(headers)
  const token = getToken()
  if (token) {
    // 携带在 Authorization 中，无需 Scheme 前缀
    finalHeaders.set("Authorization", token)
  }
  const isForm = body instanceof FormData
  if (body !== undefined && !isForm) {
    finalHeaders.set("Content-Type", "application/json")
  }

  const response = await fetchResponse(url, {
    ...rest,
    headers: finalHeaders,
    body:
      body === undefined
        ? undefined
        : isForm
          ? (body as FormData)
          : JSON.stringify(body),
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  const contentType = response.headers.get("content-type") ?? ""
  if (contentType.includes("application/json")) {
    return (await response.json()) as T
  }

  return (await response.text()) as unknown as T
}

export const http = {
  get: <T>(path: string, query?: Record<string, unknown>) =>
    request<T>(path, { method: "GET", query }),
  post: <T>(path: string, body?: unknown, query?: Record<string, unknown>) =>
    request<T>(path, { method: "POST", body, query }),
  put: <T>(path: string, body?: unknown, query?: Record<string, unknown>) =>
    request<T>(path, { method: "PUT", body, query }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PATCH", body }),
  delete: <T>(path: string, query?: Record<string, unknown>) =>
    request<T>(path, { method: "DELETE", query }),
}

export const legacyHttp = {
  get: <T>(path: string, query?: Record<string, unknown>) =>
    request<T>(path, { method: "GET", query }, LEGACY_API_BASE),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body }, LEGACY_API_BASE),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body }, LEGACY_API_BASE),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PATCH", body }, LEGACY_API_BASE),
  delete: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "DELETE", body }, LEGACY_API_BASE),
}

// 下载文件：如果后端不支持在 URL 上携带 token，
// 则回退为 fetch + blob（通过 Header 携带认证），保持调用方 API 一致。
export async function download(
  path: string,
  query: Record<string, unknown> = {}
): Promise<Blob> {
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set("Authorization", token)

  const response = await fetchResponse(buildUrl(path, query), { headers })
  if (!response.ok) throw await parseError(response)

  return response.blob()
}

// 构造带认证信息的完整 URL（适用于 GET 下载接口，token 通过 query 的 authorization 参数传递，交给浏览器直接下载）。
export function buildAuthUrl(
  path: string,
  query: Record<string, unknown> = {}
): string {
  const params: Record<string, unknown> = { ...query }
  const token = getToken()
  if (token) params.authorization = token
  return buildUrl(path, params)
}
