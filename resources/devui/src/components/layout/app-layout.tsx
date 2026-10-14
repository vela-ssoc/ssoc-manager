import { NavLink, Outlet } from "react-router-dom"
import { Archive } from "lucide-react"

import { Logo, SSOCText } from "@/components/logo"
import { ThemeToggle } from "@/components/theme-toggle"
import { cn } from "@/lib/utils"

interface NavItem {
  title: string
  href: string
  icon: typeof Archive
}

const navItems: NavItem[] = [
  {
    title: "归档文件",
    href: "/zip-archives",
    icon: Archive,
  },
]

export function AppLayout() {
  return (
    <div className="flex h-svh w-full overflow-hidden bg-background">
      <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="flex h-12 items-center justify-between gap-2 border-b border-sidebar-border px-3">
          <div className="flex items-center gap-2">
            <Logo className="size-7 shrink-0" />
            <SSOCText className="truncate text-base font-semibold" />
          </div>
          <ThemeToggle />
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto p-2">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.href}
                to={item.href}
                title={item.title}
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
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}