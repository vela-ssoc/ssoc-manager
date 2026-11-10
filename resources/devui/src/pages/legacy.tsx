import { useSearchParams } from "react-router-dom"

import { DarkEmbeddedFrame } from "@/components/dark-embedded-frame"

export function LegacyPage() {
  const [params] = useSearchParams()
  const requested = params.get("path") ?? ""
  const src =
    requested.startsWith("/") && !requested.startsWith("//") ? requested : "/"

  return <DarkEmbeddedFrame src={src} title="旧版界面" />
}