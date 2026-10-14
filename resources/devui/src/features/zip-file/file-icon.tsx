import { useEffect, useState } from "react"

const ICON_BASE = `${import.meta.env.BASE_URL}/icons/`

const IMAGE_EXTENSIONS = new Set([
  "png", "jpg", "jpeg", "gif", "webp", "bmp", "ico", "avif", "svg",
])

const FONT_EXTENSIONS = new Set([
  "ttf", "otf", "woff", "woff2", "eot",
])

const ARCHIVE_EXTENSIONS = new Set([
  "zip", "rar", "7z", "tar", "gz", "tgz", "bz2", "xz", "lz", "lz4",
  "zst", "cab", "iso", "jar", "war", "ear",
])

// 监听 dark/light 模式变化（ThemeProvider 通过切换 html.dark class 实现）。
export function useIsDark(): boolean {
  const [dark, setDark] = useState(() =>
    document.documentElement.classList.contains("dark")
  )

  useEffect(() => {
    const observer = new MutationObserver(() => {
      setDark(document.documentElement.classList.contains("dark"))
    })
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    })
    return () => observer.disconnect()
  }, [])

  return dark
}

function iconUrl(name: string, dark: boolean): string {
  return `${ICON_BASE}${name}${dark ? "_dark" : ""}.svg`
}

// 根据文件类型映射图标名（public/icons 下的基础名，不含 _dark 后缀）。
function fileIconBase(fileName: string, isDir: boolean): string {
  if (isDir) return "folder"

  const lower = fileName.toLowerCase()

  // 特殊文件优先精确匹配
  if (lower === "go.mod") return "goMod"
  if (lower === "go.sum") return "goSum"
  if (lower === "go.work") return "goWork"
  if (lower.endsWith(".schema.json")) return "jsonSchema"
  if (lower === ".gitignore") return "ignored"
  if (lower === "dockerfile") return "docker"

  const dot = lower.lastIndexOf(".")
  const ext = dot >= 0 ? lower.slice(dot + 1) : ""

  if (IMAGE_EXTENSIONS.has(ext)) return "image"
  if (FONT_EXTENSIONS.has(ext)) return "font"
  if (ARCHIVE_EXTENSIONS.has(ext)) return "archive"

  switch (ext) {
    case "go":
      return "go"
    case "java":
      return "java"
    case "js":
    case "mjs":
    case "cjs":
      return "javaScript"
    case "jsx":
      return "jsx"
    case "json":
    case "jsonc":
    case "json5":
      return "json"
    case "sql":
      return "sql"
    case "txt":
    case "text":
    case "log":
      return "text"
    case "md":
    case "markdown":
      return "markdown"
    case "css":
      return "css"
    case "html":
    case "htm":
    case "xhtml":
      return "html"
    case "yaml":
    case "yml":
      return "yaml"
    case "sh":
    case "bash":
    case "zsh":
    case "fish":
    case "ksh":
      return "shell"
    case "tsx":
      return "tsx"
    case "ts":
      return "typeScript"
    case "csv":
      return "csv"
    case "lua":
      return "lua"
    case "properties":
      return "properties"
    case "rust":
    case "rs":
      return "rust"
    case "kotlin":
    case "kt":
    case "kts":
      return "kotlin"
    case "ruby":
    case "rb":
    case "erb":
      return "ruby"
    case "http":
    case "rest":
      return "http"
    case "php":
      return "php"
    case "cpp":
    case "cc":
    case "cxx":
    case "hpp":
    case "hh":
    case "hxx":
      return "cpp"
    case "jsp":
      return "jsp"
    case "py":
    case "pyw":
      return "python"
    default:
      return "anyType"
  }
}

function TreeIcon({
  name,
  dark,
  className,
}: {
  name: string
  dark: boolean
  className?: string
}) {
  const [fellBack, setFellBack] = useState(false)

  useEffect(() => {
    setFellBack(false)
  }, [name, dark])

  const src = dark && !fellBack ? iconUrl(name, true) : iconUrl(name, false)

  return (
    <img
      src={src}
      alt=""
      draggable={false}
      className={className}
      onError={() => {
        if (dark && !fellBack) setFellBack(true)
      }}
    />
  )
}

export function FileIcon({
  name,
  isDir,
  className,
}: {
  name: string
  isDir: boolean
  className?: string
}) {
  const dark = useIsDark()
  return (
    <TreeIcon
      name={fileIconBase(name, isDir)}
      dark={dark}
      className={className}
    />
  )
}

// fold=收起（当前展开），unfold=展开（当前折叠）。
export function CollapseIcon({
  open,
  className,
}: {
  open: boolean
  className?: string
}) {
  const dark = useIsDark()
  return <TreeIcon name={open ? "fold" : "unfold"} dark={dark} className={className} />
}