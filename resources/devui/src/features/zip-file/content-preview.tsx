import { useEffect, useMemo, useRef, useState } from "react"
import Editor from "@monaco-editor/react"
import type { editor } from "monaco-editor"
import {
  Columns2,
  Download,
  FileWarning,
  FolderOpen,
  Loader2,
  Save,
  Square,
} from "lucide-react"
import { toast } from "@/lib/toast"

import { Button } from "@/components/ui/button"
import { TooltipHint } from "@/components/tooltip-hint"
import { getMonacoFontFamily } from "@/lib/fonts"
import { useIsDark } from "@/features/zip-file/file-icon"
import { useTheme } from "@/providers/theme-provider"

import { zipFileApi } from "./api"
import { CsvTable, MarkdownPreview } from "./renderers"
import {
  fileExtension,
  formatFileSize,
  isImageFile,
  isTextFile,
  MAX_EDITABLE_SIZE,
  type ZipFile,
} from "./types"

const EDITABLE_LANGUAGES: Record<string, string> = {
  js: "javascript", jsx: "javascript", mjs: "javascript", cjs: "javascript",
  ts: "typescript", tsx: "typescript",
  json: "json", jsonc: "json", json5: "json",
  html: "html", htm: "html", vue: "html", svelte: "html", astro: "html",
  css: "css", scss: "scss", less: "less", sass: "scss",
  md: "markdown", markdown: "markdown",
  yaml: "yaml", yml: "yaml",
  xml: "xml", svg: "xml",
  sh: "shell", bash: "shell", zsh: "shell",
  py: "python", rb: "ruby", php: "php",
  sql: "sql", java: "java", c: "c", h: "c", cpp: "cpp", hpp: "cpp",
  cs: "csharp", rs: "rust", go: "go", kt: "kotlin", kts: "kotlin",
  swift: "swift", scala: "scala", lua: "lua", r: "r", dart: "dart",
  toml: "ini", ini: "ini", env: "ini", conf: "ini", properties: "ini",
  proto: "proto", graphql: "graphql", gql: "graphql",
  dockerfile: "dockerfile", makefile: "makefile", perl: "perl",
  ps1: "powershell", bat: "bat", cmd: "bat",
  diff: "diff", patch: "diff",
}

export function ContentPreview({
  archiveName,
  file,
  onChanged,
}: {
  archiveName: string
  file: ZipFile
  onChanged?: () => void
}) {
  const isDir = file.is_directory
  const isImage = isImageFile(file.file_name, file.mime_type)
  const isText = isTextFile(file.file_name, file.mime_type)
  const edited = file.file_size > MAX_EDITABLE_SIZE
  const ext = fileExtension(file.file_name)
  const isSvg = ext === "svg"
  const isMarkdown = ext === "md" || ext === "markdown"
  const isCsv = ext === "csv" || ext === "tsv"

  const [content, setContent] = useState<string>("")
  const [imageUrl, setImageUrl] = useState<string>("")
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)
  const [split, setSplit] = useState(true)
  const [csvEdit, setCsvEdit] = useState(false)
  const originalContentRef = useRef("")
  const previewRef = useRef<HTMLDivElement | null>(null)
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)

  useEffect(() => {
    if (isDir) return

    let cancelled = false
    setLoading(true)
    setContent("")
    originalContentRef.current = ""
    setImageUrl("")
    setDirty(false)

    const full = file.full_name
    zipFileApi
      .content(archiveName, full)
      .then((blob) => {
        if (cancelled) return
        if (isSvg) {
          blob.text().then((t) => {
            if (!cancelled) {
              const url = URL.createObjectURL(
                new Blob([t], { type: "image/svg+xml" })
              )
              setImageUrl(url)
            }
          })
        } else if (isImage) {
          setImageUrl(URL.createObjectURL(blob))
        } else if (isText) {
          blob.text().then((t) => {
            if (!cancelled) {
              originalContentRef.current = t
              setContent(t)
            }
          })
        }
      })
      .catch((err) => {
        if (!cancelled) {
          toast.error(err instanceof Error ? err.message : "加载失败")
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [archiveName, file.id, isDir])

  useEffect(() => {
    return () => {
      if (imageUrl) URL.revokeObjectURL(imageUrl)
    }
  }, [imageUrl])

  const language = useMemo(() => {
    const ext = fileExtension(file.file_name)
    return EDITABLE_LANGUAGES[ext] ?? "plaintext"
  }, [file.file_name])

  const handleEditorChange = (value: string) => {
    setContent(value)
    setDirty(value !== originalContentRef.current)
  }

  if (isDir) {
    return (
      <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-2">
        <FolderOpen className="size-10 opacity-40" />
        <span className="text-sm">目录</span>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="text-muted-foreground size-5 animate-spin" />
      </div>
    )
  }

  if (isImage) {
    return (
      <div className="flex h-full flex-col">
        <PreviewHeader archiveName={archiveName} file={file} />
        <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 overflow-auto p-4">
          {imageUrl ? (
            <img
              src={imageUrl}
              alt={file.file_name}
              className="max-h-full max-w-full object-contain"
            />
          ) : null}
          <div className="text-muted-foreground text-xs">
            {file.file_name} · {formatFileSize(file.file_size)}
          </div>
        </div>
      </div>
    )
  }

  if (!isText) {
    return (
      <div className="flex h-full flex-col">
        <PreviewHeader archiveName={archiveName} file={file} />
        <div className="text-muted-foreground flex min-h-0 flex-1 flex-col items-center justify-center gap-3">
          <FileWarning className="size-10 opacity-40" />
          <span className="text-sm">该类型文件不支持预览</span>
          <span className="text-xs">{formatFileSize(file.file_size)}</span>
        </div>
      </div>
    )
  }

  if (edited) {
    return (
      <div className="flex h-full flex-col">
        <PreviewHeader archiveName={archiveName} file={file} />
        <div className="text-muted-foreground flex min-h-0 flex-1 flex-col items-center justify-center gap-3">
          <FileWarning className="size-10 opacity-40" />
          <span className="text-sm">文件过大，暂不支持在线编辑</span>
          <span className="text-xs">
            {formatFileSize(file.file_size)}（上限 1 MB）
          </span>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between gap-2 border-b px-3 py-1">
        <div className="text-muted-foreground min-w-0 flex-1 truncate text-xs">
          {file.file_name} · {formatFileSize(file.file_size)}
        </div>
        <div className="flex items-center gap-1">
          <TooltipHint content="下载">
            <a
              href={zipFileApi.contentUrl(archiveName, file.full_name)}
              aria-label="下载"
              className="hover:bg-accent hover:text-accent-foreground flex size-7 items-center justify-center rounded-md"
            >
              <Download className="size-3.5" />
            </a>
          </TooltipHint>
          {isMarkdown && (
            <TooltipHint content={split ? "取消分屏" : "分屏预览"}>
              <Button
                variant={split ? "secondary" : "ghost"}
                size="icon"
                className="size-7"
                aria-label={split ? "取消分屏" : "分屏预览"}
                onClick={() => setSplit((v) => !v)}
              >
                <Columns2 />
              </Button>
            </TooltipHint>
          )}
          {isCsv && (
            <TooltipHint content={csvEdit ? "表格模式" : "源码编辑"}>
              <Button
                variant={csvEdit ? "secondary" : "ghost"}
                size="icon"
                className="size-7"
                aria-label={csvEdit ? "表格模式" : "源码编辑"}
                onClick={() => setCsvEdit((v) => !v)}
              >
                <Square />
              </Button>
            </TooltipHint>
          )}
          <Button
            size="sm"
            disabled={!dirty || saving}
            onClick={async () => {
              setSaving(true)
              try {
                await zipFileApi.update(archiveName, file.full_name, content, file.file_name)
                toast.success("保存成功")
                originalContentRef.current = content
                setDirty(false)
                onChanged?.()
              } catch (err) {
                toast.error(err instanceof Error ? err.message : "保存失败")
              } finally {
                setSaving(false)
              }
            }}
          >
            <Save /> {saving ? "保存中…" : "保存"}
          </Button>
        </div>
      </div>

      {isCsv && !csvEdit ? (
        <div className="min-h-0 flex-1">
          <CsvTable content={content} />
        </div>
      ) : isMarkdown && split ? (
        <div className="flex min-h-0 flex-1 divide-x divide-border">
          <div className="min-w-0 flex-1 bg-black">
            <CodeEditor
              path={file.full_name}
              language={language}
              value={content}
              onMount={(editor) => {
                editorRef.current = editor
                editor.onDidScrollChange(() => {
                  const preview = previewRef.current
                  if (!preview) return
                  const editorScrollable = editor.getScrollHeight() - editor.getLayoutInfo().height
                  const previewScrollable = preview.scrollHeight - preview.clientHeight
                  if (editorScrollable <= 0 || previewScrollable <= 0) return
                  const ratio = editor.getScrollTop() / editorScrollable
                  preview.scrollTop = ratio * previewScrollable
                })
              }}
              onChange={(value) => {
                handleEditorChange(value)
              }}
            />
          </div>
          <div className="min-w-0 flex-1 overflow-hidden">
            <MarkdownPreview content={content} containerRef={previewRef} />
          </div>
        </div>
      ) : isMarkdown ? (
        <div className="min-h-0 flex-1">
          <MarkdownPreview content={content} />
        </div>
      ) : (
        <div className="min-h-0 flex-1 bg-black">
          <CodeEditor
            path={file.full_name}
            language={language}
            value={content}
            onChange={(value) => {
              handleEditorChange(value)
            }}
          />
        </div>
      )}
    </div>
  )
}

function PreviewHeader({
  archiveName,
  file,
}: {
  archiveName: string
  file: ZipFile
}) {
  return (
    <div className="flex items-center justify-between gap-2 border-b px-3 py-1">
      <div className="text-muted-foreground min-w-0 flex-1 truncate text-xs">
        {file.file_name} · {formatFileSize(file.file_size)}
      </div>
      <TooltipHint content="下载">
        <a
          href={zipFileApi.contentUrl(archiveName, file.full_name)}
          aria-label="下载"
          className="hover:bg-accent hover:text-accent-foreground flex size-7 items-center justify-center rounded-md"
        >
          <Download className="size-3.5" />
        </a>
      </TooltipHint>
    </div>
  )
}

function CodeEditor({
  path,
  language,
  value,
  onChange,
  onMount,
}: {
  path: string
  language: string
  value: string
  onChange: (value: string) => void
  onMount?: (editor: editor.IStandaloneCodeEditor) => void
}) {
  const dark = useIsDark()
  const { font } = useTheme()

  return (
    <Editor
      path={path}
      language={language}
      value={value}
      theme={dark ? "vs-dark" : "light"}
      beforeMount={(monaco) => {
        monaco.editor.defineTheme("near-black", {
          base: "vs-dark",
          inherit: true,
          rules: [],
          colors: {
            "editor.background": "#000000",
          },
        })
      }}
      onMount={(editor, monaco) => {
        monaco.editor.setTheme(dark ? "near-black" : "light")
        editor.updateOptions({ theme: dark ? "near-black" : "light" })
        onMount?.(editor)
      }}
      options={{
        minimap: { enabled: false },
        fontSize: 13,
        fontFamily: getMonacoFontFamily(font),
        scrollBeyondLastLine: false,
        automaticLayout: true,
        wordWrap: "on",
        tabSize: 4,
      }}
      onChange={(value) => onChange(value ?? "")}
    />
  )
}
