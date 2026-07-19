"use client"

import { useEffect } from "react"
import { useRouter, usePathname } from "next/navigation"
import Link from "next/link"
import { LayoutDashboard, User, LogOut, Menu, BookOpen } from "lucide-react"

import { useAuthStore } from "@/store/auth.store"
import { useAuthBootstrap } from "@/hooks/use-auth"
import { authApi } from "@/lib/api"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { LoadingScreen } from "@/components/shared/loading-screen"
import { Separator } from "@/components/ui/separator"

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { token, user, isLoading, logout, setAuth, _hasHydrated } = useAuthStore()

  // Bootstrap user from token on first load
  useAuthBootstrap()

  useEffect(() => {
    if (!_hasHydrated) return
    if (!token && !isLoading) {
      router.replace("/login")
    }
  }, [token, isLoading, router, _hasHydrated])

  if (!_hasHydrated || isLoading || (!user && token)) {
    return <LoadingScreen message="Loading your workspace..." />
  }

  if (!token) return null

  const initials = user?.full_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) ?? "?"

  const handleLogout = () => {
    logout()
    router.push("/login")
  }

  const navItems = [
    { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/blog", label: "Blog", icon: BookOpen },
    { href: "/profile", label: "Profile", icon: User },
  ]

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <aside className="hidden w-60 flex-col border-r bg-card md:flex">
        {/* Logo */}
        <div className="flex h-16 items-center gap-3 px-6 border-b">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
            F
          </div>
          <span className="font-semibold text-sm">Fullstack App</span>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-1 p-4">
          {navItems.map(({ href, label, icon: Icon }) => (
            <Link key={href} href={href}>
              <div
                className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  pathname === href
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                }`}
              >
                <Icon className="h-4 w-4" />
                {label}
              </div>
            </Link>
          ))}
        </nav>

        <Separator />

        {/* User info */}
        <div className="p-4">
          <div className="flex items-center gap-3 rounded-lg p-2">
            <Avatar className="h-8 w-8">
              <AvatarImage src={user?.avatar_url} />
              <AvatarFallback className="text-xs">{initials}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user?.full_name}</p>
              <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleLogout}
            className="mt-2 w-full justify-start text-muted-foreground hover:text-destructive"
          >
            <LogOut className="mr-2 h-4 w-4" />
            Sign Out
          </Button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        {/* Mobile topbar */}
        <header className="flex h-16 items-center justify-between border-b px-4 md:hidden">
          <span className="font-semibold">Fullstack App</span>
          <Button variant="ghost" size="icon">
            <Menu className="h-5 w-5" />
          </Button>
        </header>

        <div className="p-6 md:p-8">{children}</div>
      </main>
    </div>
  )
}
