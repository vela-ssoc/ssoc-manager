export interface ZipArchive {
  id: string
  name: string
  description: string
  method: number
  password: string
  version: number
  created_at: string
  updated_at: string
}

export type PageQuery = {
  page?: number
  size?: number
  q?: string
}

export interface PageResult<T> {
  page: number
  size: number
  total: number
  records: T[]
}

export interface ZipArchiveForm {
  name: string
  description: string
  method: number
  password: string
}

// 0-不加密 1-ZipCrypto 2-AES-128 3-AES-192 4-AES-256
export const METHOD_OPTIONS = [
  { label: "ZipCrypto", value: 1 },
  { label: "AES-128", value: 2 },
  { label: "AES-192", value: 3 },
  { label: "AES-256", value: 4 },
] as const

export const NO_ENCRYPTION = 0

export function methodLabel(method: number): string {
  const option = METHOD_OPTIONS.find((o) => o.value === method)
  return option?.label ?? "不加密"
}