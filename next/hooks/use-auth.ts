"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useAuthStore } from "@/store/auth.store"
import { authApi } from "@/lib/api"

/**
 * Bootstraps auth state from stored token on app load.
 * Fetches /me if token exists but user is null.
 */
export function useAuthBootstrap() {
  const { token, user, setAuth, setLoading, logout, _hasHydrated } = useAuthStore()

  useEffect(() => {
    if (!_hasHydrated) return
    if (!token) return
    if (user) return // Already hydrated

    setLoading(true)
    authApi
      .me()
      .then((fetchedUser) => {
        setAuth(token, fetchedUser)
      })
      .catch(() => {
        logout()
      })
      .finally(() => {
        setLoading(false)
      })
  }, [token, user, setAuth, setLoading, logout, _hasHydrated])
}

/**
 * Redirects to /login if user is not authenticated.
 */
export function useRequireAuth() {
  const { token, isLoading, _hasHydrated } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    if (!_hasHydrated) return
    if (!isLoading && !token) {
      router.replace("/login")
    }
  }, [token, isLoading, router, _hasHydrated])

  return useAuthStore()
}

/**
 * Redirects to /dashboard if user is already authenticated.
 */
export function useRedirectIfAuthenticated() {
  const { token, _hasHydrated } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    if (!_hasHydrated) return
    if (token) {
      router.replace("/dashboard")
    }
  }, [token, router, _hasHydrated])
}
