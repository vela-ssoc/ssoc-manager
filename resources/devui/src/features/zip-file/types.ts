export interface ZipFile {
  id: string
  archive_id: string
  grid_id: string
  directory: string
  file_name: string
  full_name: string
  file_size: number
  file_mode: number
  mime_type: string
  is_directory: boolean
  version: number
  created_at: string
  updated_at: string
}

export interface ZipFileListResult {
  records: ZipFile[]
}

export interface TreeNode {
  name: string
  path: string
  file?: ZipFile
  children: TreeNode[]
}

// 文本类扩展名（可在线编辑 / Monaco 预览）
export const TEXT_EXTENSIONS = new Set([
  "txt", "log", "md", "markdown", "go", "js", "jsx", "ts", "tsx", "mjs",
  "cjs", "json", "jsonc", "json5", "yaml", "yml", "toml", "ini", "conf",
  "env", "xml", "html", "htm", "css", "scss", "less", "sass", "sql",
  "sh", "bash", "zsh", "fish", "bat", "cmd", "ps1", "py", "rb", "php",
  "java", "c", "h", "cpp", "hpp", "cc", "cs", "rs", "kt", "kts", "swift",
  "scala", "lua", "pl", "pm", "r", "dart", "vue", "svelte", "astro",
  "graphql", "gql", "proto", "dockerfile", "makefile", "properties",
  "csv", "tsv", "svg", "gitignore", "gitattributes", "editorconfig",
  "cmd", "gradle", "lock", "diff", "patch", "apacheconf", "nginx",
])

// 图片类扩展名
export const IMAGE_EXTENSIONS = new Set([
  "png", "jpg", "jpeg", "gif", "webp", "bmp", "ico", "avif", "svg",
])

// 单文件可编辑大小上限：1MB
export const MAX_EDITABLE_SIZE = 1024 * 1024

export function fileExtension(name: string): string {
  const idx = name.lastIndexOf(".")
  if (idx < 0) return ""
  return name.slice(idx + 1).toLowerCase()
}

export function isTextFile(name: string, mimeType?: string): boolean {
  const ext = fileExtension(name)
  if (ext && TEXT_EXTENSIONS.has(ext)) return true
  if (!mimeType) return false
  return (
    mimeType.startsWith("text/") ||
    mimeType === "application/json" ||
    mimeType === "application/xml" ||
    mimeType === "application/javascript" ||
    mimeType === "application/x-httpd-php" ||
    mimeType === "application/graphql" ||
    mimeType === "application/x-yaml" ||
    mimeType.includes("+json") ||
    mimeType.includes("+xml")
  )
}

export function isImageFile(name: string, mimeType?: string): boolean {
  const ext = fileExtension(name)
  // svg 既可文本编辑也可图片预览，优先当作图片
  if (ext && IMAGE_EXTENSIONS.has(ext)) return true
  return !!mimeType && mimeType.startsWith("image/")
}

export function sortFiles(a: ZipFile, b: ZipFile): number {
  if (a.is_directory !== b.is_directory) return a.is_directory ? -1 : 1
  return a.full_name.localeCompare(b.full_name)
}

export function isSubPath(root: string, candidate: string): boolean {
  return candidate.startsWith(root === "/" || root === "" ? "" : root + "/")
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  }
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
}