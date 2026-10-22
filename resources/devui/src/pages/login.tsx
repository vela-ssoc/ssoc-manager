import { useState, type FormEvent } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import { ArrowLeft, Copy, Eye, EyeOff, LoaderCircle } from "lucide-react"
import { QRCodeSVG } from "qrcode.react"

import { Logo, SSOCText } from "@/components/logo"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { InputOTP } from "@/components/ui/input-otp"
import { Label } from "@/components/ui/label"
import { TooltipHint } from "@/components/tooltip-hint"
import { toast } from "@/lib/toast"
import { saveAuthSession } from "@/lib/auth-session"
import { legacyAuthApi } from "@/features/auth/api"
import { OAuthQrLogin } from "@/features/auth/oauth-login"

type LoginMode = "password" | "qrcode"

interface ReturnLocation {
  pathname?: string
  search?: string
  hash?: string
}

type LoginStep = "credentials" | "totp"

export function LoginPage() {
  const [mode, setMode] = useState<LoginMode>("qrcode")
  const [step, setStep] = useState<LoginStep>("credentials")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [uid, setUid] = useState("")
  const [totpUrl, setTotpUrl] = useState("")
  const [totpSecret, setTotpSecret] = useState("")
  const [needsBinding, setNeedsBinding] = useState(false)
  const [code, setCode] = useState("")
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const validateCredentials = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!username.trim() || !password) return

    setLoading(true)
    try {
      const result = await legacyAuthApi.validate(username.trim(), password)
      setUid(result.uid)
      setNeedsBinding(!result.bind)
      setTotpUrl("")
      setTotpSecret("")
      setCode("")

      if (!result.bind) {
        const totp = await legacyAuthApi.totp(result.uid)
        setTotpUrl(totp.url)
        setTotpSecret(totp.secret ?? "")
      }
      setStep("totp")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "用户名或密码验证失败")
    } finally {
      setLoading(false)
    }
  }

  const submitCode = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!uid || !/^\d{6}$/.test(code)) {
      toast.error("请输入 6 位数字验证码")
      return
    }

    setLoading(true)
    try {
      const session = await legacyAuthApi.submit(uid, code)
      if (!session.token) throw new Error("登录响应中没有 Token")

      saveAuthSession(session)
      const from = (location.state as { from?: ReturnLocation } | null)?.from
      const destination = from?.pathname && from.pathname !== "/login"
        ? `${from.pathname}${from.search ?? ""}${from.hash ?? ""}`
        : "/legacy"
      navigate(destination, { replace: true })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "验证码校验失败")
    } finally {
      setLoading(false)
    }
  }

  const submitOauth = async (code: string) => {
    const session = await legacyAuthApi.oauth(code)
    if (!session.token) throw new Error("登录响应中没有 Token")

    saveAuthSession(session)
    const from = (location.state as { from?: ReturnLocation } | null)?.from
    const destination = from?.pathname && from.pathname !== "/login"
      ? `${from.pathname}${from.search ?? ""}${from.hash ?? ""}`
      : "/legacy"
    navigate(destination, { replace: true })
  }

  return (
    <main className="flex min-h-svh items-center justify-center bg-background p-4">
      <div className="bg-card w-full max-w-sm rounded-md border p-6 shadow-sm">
        <div className="mb-6 flex items-center gap-3">
          <Logo className="size-8 shrink-0" />
          <div className="brand-font flex flex-col">
            <SSOCText className="text-lg font-bold leading-none" />
            <span className="text-muted-foreground text-[8px] leading-tight">
              Smart Security Operations Center
            </span>
          </div>
        </div>

        <div className="bg-muted mb-6 flex gap-0.5 rounded-md p-0.5">
          <button
            type="button"
            className={`flex-1 rounded px-3 py-1 text-xs font-medium transition-colors ${
              mode === "qrcode"
                ? "bg-card shadow-sm"
                : "text-muted-foreground"
            }`}
            onClick={() => setMode("qrcode")}
          >
            咚咚扫码
          </button>
          <button
            type="button"
            className={`flex-1 rounded px-3 py-1 text-xs font-medium transition-colors ${
              mode === "password"
                ? "bg-card shadow-sm"
                : "text-muted-foreground"
            }`}
            onClick={() => setMode("password")}
          >
            账密登录
          </button>
        </div>

        {mode === "qrcode" ? (
          <OAuthQrLogin onCode={submitOauth} />
        ) : step === "credentials" ? (
          <form className="grid gap-4" onSubmit={validateCredentials}>
            <div className="grid gap-2">
              <Label htmlFor="login-username">用户名</Label>
              <Input
                id="login-username"
                autoComplete="username"
                autoFocus
                value={username}
                placeholder="请输入用户名"
                onChange={(event) => setUsername(event.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="login-password">密码</Label>
              <div className="relative">
                <Input
                  id="login-password"
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  value={password}
                  placeholder="请输入密码"
                  className="pr-10"
                  onChange={(event) => setPassword(event.target.value)}
                />
                <TooltipHint content={showPassword ? "隐藏密码" : "显示密码"}>
                  <button
                    type="button"
                    aria-label={showPassword ? "隐藏密码" : "显示密码"}
                    className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 flex size-6 -translate-y-1/2 items-center justify-center rounded"
                    onClick={() => setShowPassword((visible) => !visible)}
                  >
                    {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                  </button>
                </TooltipHint>
              </div>
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={loading || !username.trim() || !password}
            >
              {loading && <LoaderCircle className="animate-spin" />}
              登录
            </Button>
          </form>
        ) : (
          <form className="grid gap-4" onSubmit={submitCode}>
            {needsBinding && totpUrl && (
              <div className="grid w-full min-w-0 justify-items-center gap-3 overflow-hidden rounded-md bg-muted/30 p-3">
                <QRCodeSVG
                  value={totpUrl}
                  size={160}
                  level="H"
                  includeMargin
                  bgColor="#ffffff"
                  fgColor="#000000"
                  className="block h-auto max-w-full"
                />
                <p className="text-muted-foreground text-center text-xs">
                    请使用身份验证器扫描二维码，输入 6 位验证码验证通过后自动完成绑定。
                </p>
                {totpSecret && (
                  <div className="w-full min-w-0 rounded-md border bg-card p-2.5">
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground text-xs">密钥</span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 shrink-0"
                        aria-label="复制密钥"
                        onClick={async () => {
                          try {
                            await navigator.clipboard.writeText(totpSecret)
                            toast.success("密钥已复制")
                          } catch {
                            toast.error("复制失败，请手动复制密钥")
                          }
                        }}
                      >
                        <Copy className="size-3.5" />
                      </Button>
                    </div>
                    <code className="mt-1 block w-full truncate text-xs leading-relaxed">
                      {totpSecret}
                    </code>
                  </div>
                )}
              </div>
            )}
            {!needsBinding && (
              <p className="text-muted-foreground text-sm">
                请输入身份验证器生成的 6 位验证码
              </p>
            )}
            <div className="grid gap-2">
              <Label htmlFor="login-totp">验证码</Label>
              <InputOTP
                id="login-totp"
                aria-label="6 位验证码"
                value={code}
                onChange={(value) => setCode(value.replace(/\D/g, "").slice(0, 6))}
              />
            </div>
            <div className="grid grid-cols-[auto_1fr] gap-2">
              <Button
                type="button"
                variant="outline"
                disabled={loading}
                onClick={() => {
                  setStep("credentials")
                  setUid("")
                  setTotpUrl("")
                  setTotpSecret("")
                  setCode("")
                }}
              >
                <ArrowLeft /> 返回
              </Button>
              <Button type="submit" disabled={loading || code.length !== 6}>
                {loading && <LoaderCircle className="animate-spin" />}
                {needsBinding ? "绑定并登录" : "登录"}
              </Button>
            </div>
          </form>
        )}
      </div>
    </main>
  )
}
