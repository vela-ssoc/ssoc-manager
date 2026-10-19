import { useState } from "react"
import { NavLink, Outlet, useNavigate } from "react-router-dom"
import { Archive, Home, LogOut, Network, Radar, RadioTower, Settings, UserRound, type LucideIcon } from "lucide-react"

import { Logo, SSOCText } from "@/components/logo"
import { TooltipHint } from "@/components/tooltip-hint"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { clearToken } from "@/lib/http"
import { cn } from "@/lib/utils"
import { getAuthProfile } from "@/lib/auth-session"
import { legacyAuthApi } from "@/features/auth/api"

interface NavItem {
  title: string
  href: string
  icon: LucideIcon
}

const navItems: NavItem[] = [
  {
    title: "旧版界面",
    href: "/legacy",
    icon: Home,
  },
  {
    title: "探针节点",
    href: "/minions",
    icon: Radar,
  },
  {
    title: "中继节点",
    href: "/brokers",
    icon: RadioTower,
  },
  {
    title: "虚拟文件",
    href: "/zip-archives",
    icon: Archive,
  },
  {
    title: "反向代理",
    href: "/reverse-proxies",
    icon: Network,
  },
]

export function AppLayout() {
  const navigate = useNavigate()
  const profile = getAuthProfile()
  const [logoutConfirmOpen, setLogoutConfirmOpen] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)

  const logout = async () => {
    setLoggingOut(true)
    await legacyAuthApi.logout().catch(() => undefined)
    clearToken()
    navigate("/login", { replace: true })
  }

  return (
    <div className="flex h-svh w-full overflow-hidden bg-background">
      <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="flex h-12 items-center justify-between gap-2 border-b border-sidebar-border px-3">
          <div className="flex items-center gap-2">
            <Logo className="size-5 shrink-0" />
            <div className="brand-font flex min-w-0 flex-col justify-center">
              <SSOCText className="truncate text-lg font-bold leading-none" />
              <span className="text-muted-foreground truncate text-[8px] leading-tight tracking-tight select-none">
                Smart Security Operations Center
              </span>
            </div>
          </div>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto p-2">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.href}
                to={item.href}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-sidebar-accent text-sidebar-accent-foreground"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
                  )
                }
              >
                <Icon className="size-4 shrink-0" />
                <span className="truncate">{item.title}</span>
              </NavLink>
            )
          })}
        </nav>

        <div className="flex h-12 shrink-0 items-center border-t border-sidebar-border px-3">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md py-1 text-left hover:bg-sidebar-accent/50"
                >
                  <div className="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-full">
                    <UserRound className="size-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-medium">
                      {profile?.nickname || profile?.username || "系统用户"}
                    </div>
                    <div className="text-muted-foreground truncate text-xs">
                      {profile?.username || "用户账号"}
                    </div>
                  </div>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" className="w-48">
                <DropdownMenuLabel>
                  {profile?.nickname || profile?.username || "系统用户"}
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  variant="destructive"
                  className="cursor-pointer"
                  onSelect={() => setLogoutConfirmOpen(true)}
                >
                  <LogOut /> 退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <TooltipHint content="设置">
              <span className="flex">
                <NavLink
                  to="/settings"
                  aria-label="设置"
                  className={({ isActive }) =>
                    cn(
                      "my-auto flex size-8 shrink-0 self-center items-center justify-center rounded-md transition-colors",
                      isActive
                        ? "bg-sidebar-accent text-sidebar-accent-foreground"
                        : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
                    )
                  }
                >
                  <Settings className="size-4" />
                </NavLink>
              </span>
            </TooltipHint>
          </div>
        </div>
      </aside>

      <Dialog open={logoutConfirmOpen} onOpenChange={setLogoutConfirmOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>退出登录</DialogTitle>
            <DialogDescription>确定要退出当前账号吗？</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              disabled={loggingOut}
              onClick={() => setLogoutConfirmOpen(false)}
            >
              取消
            </Button>
            <Button disabled={loggingOut} onClick={() => void logout()}>
              {loggingOut ? "正在退出…" : "退出登录"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <div className="flex min-w-0 flex-1 flex-col">
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
