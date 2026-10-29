import { useEffect, useState } from "react"
import { LoaderCircle } from "lucide-react"

import { useIsDark } from "@/features/zip-file/file-icon"
import { toast } from "@/lib/toast"
import { cn } from "@/lib/utils"

import type { OAuthProvider } from "./api"

const QR_SIZE = 150

// iframe 内部页面以更宽视口渲染（保证二维码水平居中且不出现横向滚动条），
// 再通过负 marginLeft 把二维码对齐到可见窗口中央。
const FRAME_WIDTH = 320

// 登录系统二维码上方标题高度（`text-[16px]` 行高 24px + `mt-2` 8px + `mb-4` 16px ≈ 48px），
// 向上裁剪只露出二维码。
const TITLE_CROP = 49

// 可见窗口四周留白，避免二维码紧贴边框。
const PAD = 4

interface DongDongMessage {
  type?: string
  data?: string
}

function extractCode(redirectUri: string): string | null {
  try {
    const url = new URL(redirectUri, window.location.href)
    return url.searchParams.get("code")
  } catch {
    return null
  }
}

function buildAuthorizeUrl(provider: OAuthProvider): URL | null {
  if (
    !provider.authorization_endpoint ||
    !provider.client_id ||
    !provider.redirect_uri
  ) {
    return null
  }
  try {
    const url = new URL(provider.authorization_endpoint)
    url.searchParams.set("client_id", provider.client_id)
    url.searchParams.set("redirect_uri", provider.redirect_uri)
    return url
  } catch {
    return null
  }
}

export function OAuthQrLoading() {
  return (
    <div className="grid justify-items-center gap-3">
      <div
        className="relative flex items-center justify-center overflow-hidden rounded-md border bg-white"
        style={{ width: QR_SIZE + PAD * 2, height: QR_SIZE + PAD * 2 }}
      >
        <QrCorner className="top-3 left-3" delay="0ms" />
        <QrCorner className="top-3 right-3" delay="150ms" />
        <QrCorner className="bottom-3 left-3" delay="300ms" />
        <QrLoadingContent />
      </div>
    </div>
  )
}

export function OAuthQrLogin({
  provider,
  onCode,
}: {
  provider: OAuthProvider
  onCode: (code: string) => Promise<void>
}) {
  const [frameUrl, setFrameUrl] = useState("")
  const [authUrl, setAuthUrl] = useState("")
  const [loading, setLoading] = useState(true)
  const [iframeLoading, setIframeLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const dark = useIsDark()
  const {
    authorization_endpoint,
    client_id: clientId,
    redirect_uri: redirectUri,
  } = provider

  useEffect(() => {
    const url = buildAuthorizeUrl({
      authorization_endpoint,
      client_id: clientId,
      redirect_uri: redirectUri,
    })
    if (!url) {
      setAuthUrl("")
      setFrameUrl("")
      setLoading(false)
      return
    }
    setAuthUrl(url.toString())
    url.searchParams.set("qr", "1")
    url.searchParams.set("img_size", String(QR_SIZE))
    setIframeLoading(true)
    setFrameUrl(url.toString())
    setLoading(false)
  }, [authorization_endpoint, clientId, redirectUri])

  useEffect(() => {
    const handler = (event: MessageEvent) => {
      let msg: DongDongMessage
      try {
        msg = JSON.parse(event.data) as DongDongMessage
      } catch {
        return
      }
      if (msg.type !== "DongDongOAuth2Login" || !msg.data) return

      const code = extractCode(msg.data)
      if (!code) {
        toast.error("扫码登录返回的地址中缺少授权码")
        return
      }
      setSubmitting(true)
      void onCode(code).finally(() => setSubmitting(false))
    }

    window.addEventListener("message", handler)
    return () => window.removeEventListener("message", handler)
  }, [onCode])

  return (
    <div className="grid justify-items-center gap-3">
      {frameUrl ? (
        <div
          className="relative overflow-hidden rounded-md border bg-white"
          style={{
            width: QR_SIZE + PAD * 2,
            height: QR_SIZE + PAD * 2,
          }}
        >
          <iframe
            key={frameUrl}
            title="扫码登录"
            src={frameUrl}
            sandbox="allow-scripts allow-same-origin"
            className="block border-0"
            style={{
              width: FRAME_WIDTH,
              height: QR_SIZE + TITLE_CROP,
              marginTop: PAD - TITLE_CROP,
              marginLeft: PAD - (FRAME_WIDTH - QR_SIZE) / 2 - 1,
            }}
            onLoad={() => setIframeLoading(false)}
          />
          {dark && (
            <div className="pointer-events-none absolute -inset-px bg-black/50" />
          )}
          {iframeLoading && (
            <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/85 backdrop-blur-[1px]">
              <QrLoadingContent compact />
            </div>
          )}
          {submitting && (
            <div className="absolute -inset-px z-20 flex items-center justify-center bg-background/70">
              <div className="flex items-center gap-2 text-sm">
                <LoaderCircle className="animate-spin" />
                正在登录…
              </div>
            </div>
          )}
        </div>
      ) : (
        <div
          className="flex items-center justify-center rounded-md border bg-white"
          style={{ width: QR_SIZE + PAD * 2, height: QR_SIZE + PAD * 2 }}
        >
          {loading ? (
            <QrLoadingContent compact />
          ) : (
            <span className="text-muted-foreground max-w-48 text-center text-xs">
              未配置扫码登录入口，请先完成 OAuth 配置
            </span>
          )}
        </div>
      )}
      {authUrl && (
        <a
          href={authUrl}
          className="text-muted-foreground hover:text-primary text-xs underline underline-offset-2"
        >
          咚咚统一登录认证
        </a>
      )}
    </div>
  )
}

function QrLoadingContent({ compact = false }: { compact?: boolean }) {
  return (
    <div className="flex flex-col items-center gap-2">
      <span
        className={cn(
          "animate-spin rounded-full border-2 border-neutral-200 border-t-primary",
          compact ? "size-6" : "size-7"
        )}
      />
      <span
        className={cn(
          "text-neutral-500",
          compact ? "text-[10px]" : "text-[11px]"
        )}
      >
        正在加载二维码…
      </span>
    </div>
  )
}

function QrCorner({
  className,
  delay,
}: {
  className?: string
  delay?: string
}) {
  return (
    <span
      className={cn(
        "absolute flex size-6 animate-pulse items-center justify-center rounded-sm border-[3px] border-neutral-200",
        className
      )}
      style={{ animationDelay: delay }}
    >
      <span className="size-1.5 rounded-[1px] bg-neutral-300" />
    </span>
  )
}