const API_BASE = "/api/v3"

export class ApiError extends Error {
  status: number
  detail?: ProblemDetail

  constructor(status: number, message: string, detail?: ProblemDetail) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.detail = detail
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

function buildUrl(path: string, query?: Record<string, unknown>): string {
  const url = path.startsWith("/") ? API_BASE + path : path
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

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown
  query?: Record<string, unknown>
}

async function request<T>(path: string, init: RequestOptions = {}): Promise<T> {
  const { query, body, headers, ...rest } = init
  const url = buildUrl(path, query)

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

  const response = await fetch(url, {
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

// 下载文件：如果后端不支持在 URL 上携带 token，
// 则回退为 fetch + blob（通过 Header 携带认证），保持调用方 API 一致。
export async function download(
  path: string,
  query: Record<string, unknown> = {}
): Promise<Blob> {
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set("Authorization", token)

  const response = await fetch(buildUrl(path, query), { headers })
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