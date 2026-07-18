"use client"

import { create } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"
import type { User } from "@/lib/api"

interface AuthState {
  token: string | null
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  hasCompletedOnboarding: boolean

  _hasHydrated: boolean

  setAuth: (token: string, user: User) => void
  setUser: (user: User) => void
  setLoading: (loading: boolean) => void
  completeOnboarding: () => void
  logout: () => void
  setHasHydrated: (state: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isLoading: false,
      isAuthenticated: false,
      hasCompletedOnboarding: false,
      _hasHydrated: false,

      setAuth: (token, user) =>
        set({
          token,
          user,
          isAuthenticated: true,
          isLoading: false,
        }),

      setUser: (user) => set({ user }),

      setLoading: (loading) => set({ isLoading: loading }),

      completeOnboarding: () => set({ hasCompletedOnboarding: true }),

      logout: () =>
        set({
          token: null,
          user: null,
          isAuthenticated: false,
          hasCompletedOnboarding: false,
        }),

      setHasHydrated: (state) => set({ _hasHydrated: state }),
    }),
    {
      name: "auth-storage",
      storage: createJSONStorage(() => localStorage),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true)
      },
      // Only persist token and onboarding status; user will be re-fetched
      partialize: (state) => ({
        token: state.token,
        hasCompletedOnboarding: state.hasCompletedOnboarding,
      }),
    },
  ),
)
