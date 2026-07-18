"use client"

import { Loader2 } from "lucide-react"

export function LoadingScreen({ message = "Loading..." }: { message?: string }) {
  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center gap-4 bg-background">
      <div className="relative flex items-center justify-center">
        <div className="h-16 w-16 rounded-full border-4 border-muted" />
        <div className="absolute h-16 w-16 rounded-full border-4 border-primary border-t-transparent animate-spin" />
      </div>
      <p className="text-sm text-muted-foreground animate-pulse">{message}</p>
    </div>
  )
}

export function PageLoader() {
  return (
    <div className="flex h-[60vh] items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  )
}
