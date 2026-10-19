import { useEffect, useRef } from "react"

import { useTheme } from "@/providers/theme-provider"

const DARK_FILTER = "invert(1) hue-rotate(180deg)"
const IMAGE_FILTER = "invert(1) hue-rotate(180deg)"

export function LegacyPage() {
  const { isDark } = useTheme()
  const iframeRef = useRef<HTMLIFrameElement | null>(null)

  useEffect(() => {
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
  }, [isDark])

  return (
    <iframe
      ref={iframeRef}
      src="/"
      title="旧版界面"
      className="h-full w-full border-0"
      style={{ filter: isDark ? DARK_FILTER : "" }}
    />
  )
}