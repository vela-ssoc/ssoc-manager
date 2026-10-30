import { useState } from "react"
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom"
import {
  Activity,
  Archive,
  BookText,
  ChevronDown,
  ChevronsUpDown,
  Database,
  ExternalLink,
  HardDrive,
  Home,
  LogOut,
  MonitorPlay,
  Network,
  Radar,
  RadioTower,
  ServerCog,
  Settings,
  ShieldCheck,
  UserRound,
  type LucideIcon,
} from "lucide-react"

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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { clearToken } from "@/lib/http"
import { cn } from "@/lib/utils"
import { companyLabel, getAuthProfile } from "@/lib/auth-session"
import { legacyAuthApi } from "@/features/auth/api"
import { EventNotifications } from "@/features/server-events/event-notifier"
import { clearServerEventHistory } from "@/features/server-events/history"
import { NotificationCenter } from "@/features/server-events/notification-center"
import { useServerEvents } from "@/features/server-events/use-server-events"

interface NavItem {
  title: string
  href: string
  icon: LucideIcon
  external?: string
  children?: NavItem[]
}

const navItems: NavItem[] = [
  {
    title: "旧版界面",
    href: "/legacy",
    icon: Home,
    external: "/",
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
    title: "中心端",
    href: "/central",
    icon: ServerCog,
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
  {
    title: "配置管理",
    href: "/configs",
    icon: Settings,
    children: [
      { title: "身份验证", href: "/configs/oauth", icon: ShieldCheck },
    ],
  },
  {
    title: "数据源",
    href: "/datasources",
    icon: HardDrive,
    children: [
      { title: "CMDB", href: "/configs/cmdb", icon: Database },
      { title: "VictoriaMetrics", href: "/configs/victoria-metrics", icon: Activity },
    ],
  },
  {
    title: "大屏展示",
    href: "/dashboards",
    icon: MonitorPlay,
    children: [
      { title: "风险大屏", href: "/ssoc", icon: MonitorPlay, external: "/ssoc/" },
      { title: "蜜罐大屏", href: "/hone", icon: MonitorPlay, external: "/hone/" },
    ],
  },
  {
    title: "接口文档",
    href: "/oas3",
    icon: BookText,
    external: "/oas3/",
  },
]

export function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const profile = getAuthProfile()
  const displayName = profile?.nickname || profile?.username || "系统用户"
  const company = companyLabel(profile?.company)
  const accountLine =
    [company, profile?.username].filter(Boolean).join(" · ") || "用户账号"
  useServerEvents()
  const [logoutConfirmOpen, setLogoutConfirmOpen] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)
  const [openGroups, setOpenGroups] = useState<Set<string>>(() => {
    const initial = new Set<string>()
    for (const item of navItems) {
      if (item.children?.some((child) => child.href === location.pathname)) {
        initial.add(item.href)
      }
    }
    return initial
  })

  const toggleGroup = (href: string) => {
    setOpenGroups((prev) => {
      const next = new Set(prev)
      if (next.has(href)) next.delete(href)
      else next.add(href)
      return next
    })
  }

  const logout = async () => {
    setLoggingOut(true)
    await legacyAuthApi.logout().catch(() => undefined)
    clearServerEventHistory()
    clearToken()
    navigate("/login", { replace: true })
  }

  return (
    <div className="flex h-svh w-full overflow-hidden bg-background">
      <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="flex h-12 items-center justify-between gap-2 border-b border-sidebar-border px-3">
          <div className="flex min-w-0 items-center gap-2">
            <Logo className="size-5 shrink-0" />
            <div className="brand-font flex min-w-0 flex-col justify-center">
              <SSOCText className="truncate text-lg font-bold leading-none" />
              <span className="text-muted-foreground truncate text-[8px] leading-tight tracking-tight select-none">
                Smart Security Operations Center
              </span>
            </div>
          </div>
          <NotificationCenter />
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto p-2">
          {navItems.map((item) => {
            const Icon = item.icon

            if (!item.children?.length) {
              return (
                <div
                  key={item.href}
                  className={cn(
                    "group/item flex items-center rounded-md transition-colors",
                    location.pathname === item.href
                      ? "bg-sidebar-accent text-sidebar-accent-foreground"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
                  )}
                >
                  <NavLink
                    to={item.href}
                    className="flex min-w-0 flex-1 items-center gap-3 px-3 py-2 text-sm font-medium"
                  >
                    <Icon className="size-4 shrink-0" />
                    <span className="truncate">{item.title}</span>
                  </NavLink>
                  {item.external && (
                    <ExternalLinkButton
                      href={item.external}
                      revealClass="group-hover/item:opacity-100"
                    />
                  )}
                </div>
              )
            }

            const open = openGroups.has(item.href)
            const active = item.children.some(
              (child) => child.href === location.pathname
            )

            return (
              <div key={item.href}>
                <button
                  type="button"
                  className={cn(
                    "flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                    active
                      ? "text-sidebar-accent-foreground"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
                  )}
                  onClick={() => toggleGroup(item.href)}
                >
                  <Icon className="size-4 shrink-0" />
                  <span className="truncate">{item.title}</span>
                  <ChevronDown
                    className={cn(
                      "ml-auto size-4 shrink-0 transition-transform",
                      open && "rotate-180"
                    )}
                  />
                </button>
                {open && (
                  <div className="mt-1 space-y-1 pl-4">
                    {item.children.map((child) => {
                      const childActive = child.href === location.pathname
                      return (
                        <div
                          key={child.href}
                          className={cn(
                            "group/child flex items-center rounded-md transition-colors",
                            childActive
                              ? "bg-sidebar-accent text-sidebar-accent-foreground"
                              : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
                          )}
                        >
                          <NavLink
                            to={child.href}
                            className="flex min-w-0 flex-1 items-center px-3 py-2 text-sm"
                          >
                            <span className="truncate">{child.title}</span>
                          </NavLink>
                          {child.external && (
                            <ExternalLinkButton
                              href={child.external}
                              revealClass="group-hover/child:opacity-100"
                            />
                          )}
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </nav>

        <div className="shrink-0 border-t border-sidebar-border p-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-sidebar-accent/60 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <UserAvatar
                  url={profile?.avatar_url}
                  name={displayName}
                  className="size-8"
                />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm leading-tight font-medium">
                    {displayName}
                  </div>
                  <div className="text-muted-foreground mt-0.5 truncate text-xs leading-tight">
                    {accountLine}
                  </div>
                </div>
                <ChevronsUpDown className="text-muted-foreground size-3.5 shrink-0" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side="top"
              align="start"
              alignOffset={-8}
              sideOffset={0}
              className="w-60 duration-200 ease-out data-[state=closed]:duration-150"
            >
              <DropdownMenuItem
                className="cursor-pointer py-2"
                onSelect={() => navigate("/settings")}
              >
                个人中心
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                className="cursor-pointer py-2"
                onSelect={() => setLogoutConfirmOpen(true)}
              >
                <LogOut /> 退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </aside>

      <Dialog open={logoutConfirmOpen} onOpenChange={setLogoutConfirmOpen}>
        <DialogContent showCloseButton={false} className="sm:max-w-sm">
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
              {loggingOut ? "正在退出…" : "确定"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <div className="flex min-w-0 flex-1 flex-col">
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>

      <EventNotifications />
    </div>
  )
}

function UserAvatar({
  url,
  name,
  className,
  iconClassName,
}: {
  url?: string
  name: string
  className?: string
  iconClassName?: string
}) {
  const [failed, setFailed] = useState(false)

  if (url && !failed) {
    return (
      <img
        src={url}
        alt={name}
        className={cn("shrink-0 rounded-full object-cover", className)}
        onError={() => setFailed(true)}
      />
    )
  }

  return (
    <div
      className={cn(
        "bg-primary/10 text-primary flex shrink-0 items-center justify-center rounded-full",
        className
      )}
    >
      <UserRound className={iconClassName ?? "size-4"} />
    </div>
  )
}

function ExternalLinkButton({
  href,
  revealClass,
}: {
  href: string
  revealClass: string
}) {
  return (
    <TooltipHint content="新标签打开">
      <a
        href={href}
        target="_blank"
        rel="noreferrer"
        aria-label="新标签打开"
        className={cn(
          "mr-1 flex size-6 shrink-0 items-center justify-center rounded-md opacity-0 transition-opacity hover:bg-foreground/15 active:bg-foreground/25",
          revealClass
        )}
      >
        <ExternalLink className="size-3.5" />
      </a>
    </TooltipHint>
  )
}
