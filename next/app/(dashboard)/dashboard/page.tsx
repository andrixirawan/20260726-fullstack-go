"use client"

import { useAuthStore } from "@/store/auth.store"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Separator } from "@/components/ui/separator"
import { User, Mail, Calendar, Shield } from "lucide-react"
import Link from "next/link"
import { Button, buttonVariants } from "@/components/ui/button"

import { useEffect } from "react"

export default function DashboardPage() {
  const { user } = useAuthStore()

  useEffect(() => {
    document.title = "Dashboard | Fullstack App"
  }, [])

  const initials = user?.full_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) ?? "?"

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          Welcome back, {user?.full_name?.split(" ")[0] ?? "there"}! 👋
        </h1>
        <p className="text-muted-foreground">Here&apos;s an overview of your account.</p>
      </div>

      {/* Profile summary card */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center gap-5">
            <Avatar className="h-16 w-16">
              <AvatarImage src={user?.avatar_url} alt={user?.full_name} />
              <AvatarFallback className="text-lg font-semibold">{initials}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <h2 className="text-lg font-semibold">{user?.full_name}</h2>
              <p className="text-sm text-muted-foreground">{user?.email}</p>
              <div className="mt-2 flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                <div className="h-1.5 w-1.5 rounded-full bg-green-500" />
                Active Account
              </div>
            </div>
            <Link href="/profile" className={buttonVariants({ variant: "outline", size: "sm" })}>
              Edit Profile
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* Stats grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard
          icon={<User className="h-5 w-5 text-blue-500" />}
          label="Full Name"
          value={user?.full_name ?? "-"}
          bg="bg-blue-50 dark:bg-blue-950/30"
        />
        <StatCard
          icon={<Mail className="h-5 w-5 text-purple-500" />}
          label="Email"
          value={user?.email ?? "-"}
          bg="bg-purple-50 dark:bg-purple-950/30"
        />
        <StatCard
          icon={<Calendar className="h-5 w-5 text-orange-500" />}
          label="Member Since"
          value={
            user?.created_at
              ? new Date(user.created_at).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })
              : "-"
          }
          bg="bg-orange-50 dark:bg-orange-950/30"
        />
      </div>

      {/* Quick actions */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <Link href="/profile">
            <div className="flex items-center gap-3 rounded-lg p-3 hover:bg-muted/60 transition-colors cursor-pointer">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                <User className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-sm font-medium">Edit Profile</p>
                <p className="text-xs text-muted-foreground">Update your name and avatar</p>
              </div>
            </div>
          </Link>
          <Separator />
          <div className="flex items-center gap-3 rounded-lg p-3 hover:bg-muted/60 transition-colors cursor-pointer">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-green-100 dark:bg-green-900/30">
              <Shield className="h-4 w-4 text-green-600" />
            </div>
            <div>
              <p className="text-sm font-medium">Security</p>
              <p className="text-xs text-muted-foreground">Your account is secured with JWT</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function StatCard({
  icon,
  label,
  value,
  bg,
}: {
  icon: React.ReactNode
  label: string
  value: string
  bg: string
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${bg}`}>
            {icon}
          </div>
          <div className="min-w-0">
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-sm font-medium truncate">{value}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
