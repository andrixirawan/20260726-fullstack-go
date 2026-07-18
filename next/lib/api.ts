import { useAuthStore } from "@/store/auth.store"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

interface RequestOptions extends RequestInit {
  skipAuth?: boolean
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { skipAuth = false, ...fetchOptions } = options
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(fetchOptions.headers as Record<string, string>),
  }

  if (!skipAuth) {
    const token = useAuthStore.getState().token
    if (token) {
      headers["Authorization"] = `Bearer ${token}`
    }
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...fetchOptions,
    headers,
  })

  if (response.status === 401) {
    // Token expired or invalid — logout and redirect
    useAuthStore.getState().logout()
    if (typeof window !== "undefined") {
      window.location.href = "/login"
    }
    throw new Error("Unauthorized")
  }

  const data = await response.json()

  if (!response.ok) {
    throw new ApiError(data.error || "Something went wrong", response.status, data.details)
  }

  return data as T
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: Record<string, string>,
  ) {
    super(message)
    this.name = "ApiError"
  }
}

export interface User {
  id: string
  email: string
  full_name: string
  avatar_url: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

// Auth endpoints
export const authApi = {
  register: (data: { email: string; password: string; full_name: string }) =>
    request<AuthResponse>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
      skipAuth: true,
    }),

  login: (data: { email: string; password: string }) =>
    request<AuthResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
      skipAuth: true,
    }),

  me: () => request<User>("/api/v1/auth/me"),
}

// User endpoints
export const userApi = {
  updateProfile: (data: Partial<Pick<User, "full_name" | "avatar_url">>) =>
    request<User>("/api/v1/auth/me", {
      method: "PATCH",
      body: JSON.stringify(data),
    }),

  uploadAvatar: async (file: File): Promise<{ url: string }> => {
    const token = useAuthStore.getState().token
    const formData = new FormData()
    formData.append("file", file)

    const response = await fetch(`${API_BASE_URL}/api/v1/upload`, {
      method: "POST",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: formData,
    })

    const data = await response.json()
    if (!response.ok) throw new ApiError(data.error || "Upload failed", response.status)
    return data
  },
}
