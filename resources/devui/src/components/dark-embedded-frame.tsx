import { useEffect, useRef } from "react"

import { useTheme } from "@/providers/theme-provider"

const DARK_FILTER = "invert(1) hue-rotate(180deg)"
const IMAGE_FILTER = "invert(1) hue-rotate(180deg)"

export function DarkEmbeddedFrame({
  src,
  title,
  mode = "invert",
}: {
  src: string
  title: string
  mode?: "invert" | "overlay"
}) {
  const { isDark } = useTheme()
  const iframeRef = useRef<HTMLIFrameElement | null>(null)

  useEffect(() => {
    if (mode !== "invert") return

    const patch = () => {
      const doc = iframeRef.current?.contentDocument
      if (!doc) return

      const id = "legacy-theme-patch"
      let style = doc.getElementById(id) as HTMLStyleElement | null
      if (isDark) {
        if (!style) {
          style = doc.createElement("style")
          style.id = id
          doc.head?.appendChild(style)
        }
        style.textContent = `img, video, picture, svg image, [style*="background-image"] { filter: ${IMAGE_FILTER}; }`
      } else if (style) {
        style.remove()
      }
    }

    patch()
    const iframe = iframeRef.current
    iframe?.addEventListener("load", patch)
    return () => iframe?.removeEventListener("load", patch)
  }, [isDark, mode])

  if (mode === "overlay") {
    return (
      <div className="relative h-full w-full">
        <iframe src={src} title={title} className="h-full w-full border-0" />
        {isDark && (
          <div className="pointer-events-none absolute inset-0 bg-black/40" />
        )}
      </div>
    )
  }

  return (
    <iframe
      ref={iframeRef}
      src={src}
      title={title}
      className="h-full w-full border-0"
      style={{ filter: isDark ? DARK_FILTER : "" }}
    />
  )
}