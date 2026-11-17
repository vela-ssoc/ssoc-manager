import {
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type ComponentType,
} from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import {
  A2uiSurface,
  MarkdownComponentContext,
  type MarkdownComponentProps,
  type ReactCatalogComponent,
} from "@a2ui/react"
import { MessageProcessor, type ActionPayload, type SurfaceModel } from "@a2ui/web_core"
import { injectBasicCatalogStyles } from "@a2ui/web_core/catalogs/basic/v1"
import type { AgentToRendererMessage } from "@a2ui/web_core/v1_0"

import { useIsDark } from "@/features/zip-file/file-icon"
import { chatCatalog } from "@/features/ai-chat/catalog"
import { cn } from "@/lib/utils"

if (typeof document !== "undefined") {
  injectBasicCatalogStyles()
}

function ChatMarkdown({ markdown }: MarkdownComponentProps) {
  return (
    <div className="markdown-body">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
    </div>
  )
}

function themeStyle(isDark: boolean): CSSProperties {
  return {
    colorScheme: isDark ? "dark" : "light",
    "--a2ui-color-background": "var(--card)",
    "--a2ui-color-on-background": "var(--foreground)",
    "--a2ui-color-surface": "var(--card)",
    "--a2ui-color-on-surface": "var(--card-foreground)",
    "--a2ui-color-primary": "var(--primary)",
    "--a2ui-color-primary-light": "var(--primary)",
    "--a2ui-color-primary-dark": "var(--primary)",
    "--a2ui-color-primary-hover":
      "color-mix(in srgb, var(--primary) 90%, transparent)",
    "--a2ui-color-on-primary": "var(--primary-foreground)",
    "--a2ui-color-secondary": "var(--muted)",
    "--a2ui-color-secondary-light": "var(--muted)",
    "--a2ui-color-secondary-dark": "var(--muted)",
    "--a2ui-color-secondary-hover": "var(--accent)",
    "--a2ui-color-on-secondary": "var(--foreground)",
    "--a2ui-color-border": "var(--border)",
    "--a2ui-color-input": "var(--background)",
    "--a2ui-color-on-input": "var(--foreground)",
    "--a2ui-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-border-width": "1px",
    "--a2ui-border": "1px solid var(--border)",
    "--a2ui-font-family-title": "inherit",
    "--a2ui-font-family-monospace": "var(--font-family)",
    "--a2ui-font-size": "0.875rem",
    "--a2ui-font-scale": "1.15",
    "--a2ui-line-height-body": "1.6",
    "--a2ui-line-height-headings": "1.35",
    "--a2ui-grid-base": "0.75rem",
    "--a2ui-text-color-text": "var(--foreground)",
    "--a2ui-text-caption-color": "var(--muted-foreground)",
    "--a2ui-text-a-color": "var(--primary)",
    "--a2ui-text-a-font-weight": "500",
    "--a2ui-label-font-size": "0.75rem",
    "--a2ui-label-font-weight": "500",
    "--a2ui-card-background": "var(--card)",
    "--a2ui-card-border": "1px solid var(--border)",
    "--a2ui-card-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-card-padding": "1rem",
    "--a2ui-card-margin": "0",
    "--a2ui-card-box-shadow": "none",
    "--a2ui-button-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-button-border": "1px solid var(--input)",
    "--a2ui-button-padding": "0.375rem 0.75rem",
    "--a2ui-button-margin": "0",
    "--a2ui-button-font-weight": "500",
    "--a2ui-textfield-border": "1px solid var(--input)",
    "--a2ui-textfield-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-textfield-padding": "0.375rem 0.625rem",
    "--a2ui-textfield-color-border-focus": "var(--ring)",
    "--a2ui-textfield-color-error": "var(--destructive)",
    "--a2ui-checkbox-border": "1px solid var(--input)",
    "--a2ui-checkbox-border-radius": "calc(var(--radius) - 4px)",
    "--a2ui-checkbox-gap": "0.5rem",
    "--a2ui-checkbox-margin": "0",
    "--a2ui-checkbox-color-error": "var(--destructive)",
    "--a2ui-choicepicker-chip-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-choicepicker-chip-padding": "0.25rem 0.625rem",
    "--a2ui-choicepicker-chip-background": "var(--muted)",
    "--a2ui-choicepicker-chip-background-selected": "var(--primary)",
    "--a2ui-choicepicker-chip-border": "1px solid var(--border)",
    "--a2ui-choicepicker-color-error": "var(--destructive)",
    "--a2ui-slider-margin": "0",
    "--a2ui-slider-track-color": "var(--muted)",
    "--a2ui-slider-thumb-color": "var(--primary)",
    "--a2ui-slider-color-error": "var(--destructive)",
    "--a2ui-divider-border": "1px solid var(--border)",
    "--a2ui-divider-spacing": "0.75rem",
    "--a2ui-list-padding": "0",
    "--a2ui-list-gap": "0.75rem",
    "--a2ui-row-gap": "0.75rem",
    "--a2ui-column-gap": "0.75rem",
    "--a2ui-tabs-border": "1px solid var(--border)",
    "--a2ui-tabs-header-background": "transparent",
    "--a2ui-tabs-header-background-active": "transparent",
    "--a2ui-tabs-header-color": "var(--muted-foreground)",
    "--a2ui-tabs-header-color-active": "var(--foreground)",
    "--a2ui-tabs-content-padding": "0.75rem 0 0",
    "--a2ui-modal-background": "var(--card)",
    "--a2ui-modal-border-radius": "calc(var(--radius) - 2px)",
    "--a2ui-modal-padding": "1.25rem",
    "--a2ui-modal-box-shadow": "0 10px 30px -10px rgb(0 0 0 / 0.3)",
    "--a2ui-modal-backdrop-bg": "rgb(0 0 0 / 0.5)",
    "--a2ui-image-border-radius": "calc(var(--radius) - 2px)",
  } as CSSProperties
}

function normalizeBindings(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(normalizeBindings)
  }
  if (!value || typeof value !== "object") {
    return value
  }

  const record = value as Record<string, unknown>
  const keys = Object.keys(record)
  if (keys.length === 1 && keys[0] === "path" && typeof record.path === "string") {
    return { "@path": record.path }
  }

  const next: Record<string, unknown> = {}
  for (const [key, item] of Object.entries(record)) {
    next[key] = normalizeBindings(item)
  }
  return next
}

function normalizeMessage(
  message: AgentToRendererMessage
): AgentToRendererMessage {
  const record = message as Record<string, unknown>
  const next: Record<string, unknown> = { ...record }
  if (record.version === "v1.0") {
    next.version = "1.0"
  }

  const create = record.createSurface as Record<string, unknown> | undefined
  if (create && Array.isArray(create.components)) {
    next.createSurface = {
      ...create,
      components: normalizeBindings(create.components),
    }
  }

  const update = record.updateComponents as Record<string, unknown> | undefined
  if (update && Array.isArray(update.components)) {
    next.updateComponents = {
      ...update,
      components: normalizeBindings(update.components),
    }
  }

  return next as unknown as AgentToRendererMessage
}

function isAlertsMessage(messages: AgentToRendererMessage[]) {
  for (const message of messages) {
    const create = (message as Record<string, unknown>).createSurface as
      | { components?: { component?: string }[] }
      | undefined
    if (!create || !Array.isArray(create.components)) continue
    const names = new Set(create.components.map((item) => item?.component))
    if (names.has("Icon") && names.has("List")) return true
  }
  return false
}

type FlatComponent = {
  component?: string
  id?: string
  child?: unknown
  children?: unknown
}

function isDetailMessage(messages: AgentToRendererMessage[]) {
  for (const message of messages) {
    const create = (message as Record<string, unknown>).createSurface as
      | { components?: FlatComponent[] }
      | undefined
    if (!create || !Array.isArray(create.components)) continue

    const components = create.components
    const card = components.find((item) => item?.component === "Card")
    const rootColumn = typeof card?.child === "string" ? card.child : ""
    const rowIds = new Set(
      components
        .filter((item) => item?.component === "Row")
        .map((item) => item?.id)
    )

    const grouped = components.some((item) => {
      if (item?.component !== "Column" || item.id === rootColumn) return false
      const children = item.children
      return (
        Array.isArray(children) &&
        children.some((id) => rowIds.has(id as string))
      )
    })
    if (grouped) return true
  }
  return false
}

export function A2uiMessages({
  messages,
  onAction,
}: {
  messages: AgentToRendererMessage[]
  onAction?: (action: ActionPayload) => void
}) {
  const actionRef = useRef(onAction)
  const [processor] = useState(
    // eslint-disable-next-line react-hooks/refs
    () =>
      new MessageProcessor<ReactCatalogComponent>(
        [chatCatalog],
        (action) => actionRef.current?.(action)
      )
  )
  const [surfaces, setSurfaces] = useState<
    SurfaceModel<ReactCatalogComponent>[]
  >(() => Array.from(processor.getSurfaces().values()))
  const processed = useRef(0)
  const isDark = useIsDark()

  useEffect(() => {
    actionRef.current = onAction
  }, [onAction])

  useEffect(() => {
    const sync = () =>
      setSurfaces(Array.from(processor.getSurfaces().values()))
    const created = processor.onSurfaceCreated(sync)
    const deleted = processor.onSurfaceDeleted(sync)
    return () => {
      created.unsubscribe()
      deleted.unsubscribe()
    }
  }, [processor])

  useEffect(() => {
    if (processed.current >= messages.length) return
    const pending = messages
      .slice(processed.current)
      .map(normalizeMessage)
    processed.current = messages.length
    const errors = processor.processMessagesWithErrors(pending)
    if (errors.length) {
      console.warn("A2UI 消息处理失败", errors)
    }
    setSurfaces(Array.from(processor.getSurfaces().values()))
  }, [messages, processor])

  if (!surfaces.length) return null

  return (
    <MarkdownComponentContext.Provider
      value={ChatMarkdown as ComponentType<MarkdownComponentProps>}
    >
      <div
        className={cn(
          "a2ui-surface flex flex-col gap-3",
          isAlertsMessage(messages) && "a2ui-alerts",
          isDetailMessage(messages) && "a2ui-detail"
        )}
        style={themeStyle(isDark)}
      >
        {surfaces.map((surface) => (
          <A2uiSurface key={surface.id} surface={surface} />
        ))}
      </div>
    </MarkdownComponentContext.Provider>
  )
}