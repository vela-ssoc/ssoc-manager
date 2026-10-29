import { useCallback, useEffect, useRef, useState, type RefObject } from "react"
import { toast } from "@/lib/toast"

interface DocumentPictureInPicture {
  requestWindow(options?: { width?: number; height?: number }): Promise<Window>
}

function getDocumentPiP(): DocumentPictureInPicture | undefined {
  return (window as unknown as { documentPictureInPicture?: DocumentPictureInPicture })
    .documentPictureInPicture
}

function copyStyles(pipDoc: Document) {
  for (const sheet of Array.from(document.styleSheets)) {
    try {
      const css = Array.from(sheet.cssRules)
        .map((rule) => rule.cssText)
        .join("\n")
      const style = pipDoc.createElement("style")
      style.textContent = css
      pipDoc.head.appendChild(style)
    } catch {
      // ignore cross-origin stylesheets
    }
  }
}

function prepareDocument(pipWindow: Window, title?: string) {
  const pipDoc = pipWindow.document
  copyStyles(pipDoc)
  pipDoc.documentElement.className = document.documentElement.className
  pipDoc.documentElement.style.cssText = document.documentElement.style.cssText
  pipDoc.documentElement.style.width = "100%"
  pipDoc.documentElement.style.height = "100%"
  pipDoc.title = title?.trim() || document.title
  pipDoc.body.style.margin = "0"
  pipDoc.body.style.width = "100%"
  pipDoc.body.style.height = "100%"
  pipDoc.body.style.overflow = "hidden"
  pipDoc.body.style.background = "var(--background)"

  return pipDoc
}

export function usePortalPictureInPicture(options?: {
  onResize?: () => void
  defaultWidth?: number
  defaultHeight?: number
}): {
  supported: boolean
  active: boolean
  container: HTMLElement | null
  open: (title?: string) => Promise<void>
} {
  const [active, setActive] = useState(false)
  const [container, setContainer] = useState<HTMLElement | null>(null)
  const [supported] = useState(() => Boolean(getDocumentPiP()))
  const onResizeRef = useRef(options?.onResize)

  useEffect(() => {
    onResizeRef.current = options?.onResize
  }, [options?.onResize])

  const open = useCallback(
    async (title?: string) => {
      const dpip = getDocumentPiP()
      if (!dpip) return

      try {
        const pipWindow = await dpip.requestWindow({
          width: options?.defaultWidth ?? 1024,
          height: options?.defaultHeight ?? 680,
        })
        const pipDoc = prepareDocument(pipWindow, title)

        const host = pipDoc.createElement("div")
        host.style.width = "100%"
        host.style.height = "100%"
        pipDoc.body.appendChild(host)

        setContainer(host)
        setActive(true)

        const refit = () => onResizeRef.current?.()
        const timer = window.setTimeout(refit, 60)
        pipWindow.addEventListener("resize", refit)
        pipWindow.addEventListener("pagehide", () => {
          window.clearTimeout(timer)
          setActive(false)
          setContainer(null)
        })
      } catch {
        toast.error("无法打开画中画窗口")
      }
    },
    [options?.defaultWidth, options?.defaultHeight]
  )

  return { supported, active, container, open }
}

export function usePictureInPicture(options: {
  containerRef: RefObject<HTMLElement | null>
  onResize?: () => void
  defaultWidth?: number
  defaultHeight?: number
}): { supported: boolean; active: boolean; open: (title?: string) => Promise<void> } {
  const { containerRef, onResize } = options
  const [active, setActive] = useState(false)
  const [supported] = useState(() => Boolean(getDocumentPiP()))
  const onResizeRef = useRef(onResize)

  useEffect(() => {
    onResizeRef.current = onResize
  }, [onResize])

  const open = useCallback(
    async (title?: string) => {
      const container = containerRef.current
      const dpip = getDocumentPiP()
      if (!container || !dpip) return

      try {
        const pipWindow = await dpip.requestWindow({
          width: options.defaultWidth ?? 960,
          height: options.defaultHeight ?? 640,
        })
        const pipDoc = prepareDocument(pipWindow, title)

        const parent = container.parentNode
        const prevWidth = container.style.width
        const prevHeight = container.style.height
        pipDoc.body.appendChild(container)
        container.style.width = "100%"
        container.style.height = "100%"
        setActive(true)

        const refit = () => onResizeRef.current?.()
        const timer = window.setTimeout(refit, 60)
        pipWindow.addEventListener("resize", refit)
        pipWindow.addEventListener("pagehide", () => {
          window.clearTimeout(timer)
          container.style.width = prevWidth
          container.style.height = prevHeight
          if (parent) parent.appendChild(container)
          setActive(false)
          window.setTimeout(refit, 60)
        })
      } catch {
        toast.error("无法打开画中画窗口")
      }
    },
    [containerRef, options.defaultWidth, options.defaultHeight]
  )

  return { supported, active, open }
}