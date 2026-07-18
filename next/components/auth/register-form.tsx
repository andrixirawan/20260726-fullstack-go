"use client"

import { useRouter } from "next/navigation"
import { useForm } from "@tanstack/react-form"
import { useState } from "react"
import { Eye, EyeOff, Loader2 } from "lucide-react"

import { registerSchema, registerBaseSchema, type RegisterFormValues } from "@/lib/schemas"
import { authApi, ApiError } from "@/lib/api"
import { useAuthStore } from "@/store/auth.store"
import { toast } from "@/components/ui/sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function RegisterForm() {
  const router = useRouter()
  const { setAuth } = useAuthStore()
  const [showPassword, setShowPassword] = useState(false)

  const form = useForm({
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
      confirm_password: "",
    } as RegisterFormValues,
    onSubmit: async ({ value }) => {
      const result = registerSchema.safeParse(value)
      if (!result.success) {
        toast.error(result.error.issues[0]?.message ?? "Validation failed")
        return
      }

      try {
        const response = await authApi.register({
          email: result.data.email,
          password: result.data.password,
          full_name: result.data.full_name,
        })
        setAuth(response.token, response.user)
        toast.success("Account created!", { description: "Welcome! Let's set up your profile." })
        router.push("/onboarding")
      } catch (err) {
        if (err instanceof ApiError) {
          toast.error(err.message)
        } else {
          toast.error("An unexpected error occurred")
        }
      }
    },
  })

  const getFieldError = (errors: unknown[]) =>
    errors.length > 0 ? String(errors[0]) : null

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      className="space-y-4"
    >
      <form.Field
        name="full_name"
        validators={{
          onBlur: ({ value }) => {
            const r = registerBaseSchema.shape.full_name.safeParse(value)
            return r.success ? undefined : r.error.issues[0]?.message
          },
        }}
        children={(field) => (
          <div className="space-y-1.5">
            <Label htmlFor="full_name">Full Name</Label>
            <Input
              id="full_name"
              placeholder="John Doe"
              value={field.state.value}
              onBlur={field.handleBlur}
              onChange={(e) => field.handleChange(e.target.value)}
              disabled={form.state.isSubmitting}
            />
            {field.state.meta.errors.length > 0 && (
              <p className="text-xs text-destructive">{getFieldError(field.state.meta.errors)}</p>
            )}
          </div>
        )}
      />

      <form.Field
        name="email"
        validators={{
          onBlur: ({ value }) => {
            const r = registerBaseSchema.shape.email.safeParse(value)
            return r.success ? undefined : r.error.issues[0]?.message
          },
        }}
        children={(field) => (
          <div className="space-y-1.5">
            <Label htmlFor="reg-email">Email</Label>
            <Input
              id="reg-email"
              type="email"
              placeholder="you@example.com"
              value={field.state.value}
              onBlur={field.handleBlur}
              onChange={(e) => field.handleChange(e.target.value)}
              disabled={form.state.isSubmitting}
            />
            {field.state.meta.errors.length > 0 && (
              <p className="text-xs text-destructive">{getFieldError(field.state.meta.errors)}</p>
            )}
          </div>
        )}
      />

      <form.Field
        name="password"
        validators={{
          onBlur: ({ value }) => {
            const r = registerBaseSchema.shape.password.safeParse(value)
            return r.success ? undefined : r.error.issues[0]?.message
          },
        }}
        children={(field) => (
          <div className="space-y-1.5">
            <Label htmlFor="reg-password">Password</Label>
            <div className="relative">
              <Input
                id="reg-password"
                type={showPassword ? "text" : "password"}
                placeholder="Min. 8 characters"
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                disabled={form.state.isSubmitting}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                tabIndex={-1}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {field.state.meta.errors.length > 0 && (
              <p className="text-xs text-destructive">{getFieldError(field.state.meta.errors)}</p>
            )}
          </div>
        )}
      />

      <form.Field
        name="confirm_password"
        validators={{
          onBlur: ({ value, fieldApi }) => {
            const password = fieldApi.form.getFieldValue("password")
            if (value !== password) return "Passwords do not match"
            return undefined
          },
        }}
        children={(field) => (
          <div className="space-y-1.5">
            <Label htmlFor="confirm_password">Confirm Password</Label>
            <Input
              id="confirm_password"
              type="password"
              placeholder="Re-enter your password"
              value={field.state.value}
              onBlur={field.handleBlur}
              onChange={(e) => field.handleChange(e.target.value)}
              disabled={form.state.isSubmitting}
            />
            {field.state.meta.errors.length > 0 && (
              <p className="text-xs text-destructive">{getFieldError(field.state.meta.errors)}</p>
            )}
          </div>
        )}
      />

      <Button type="submit" className="w-full" disabled={form.state.isSubmitting}>
        {form.state.isSubmitting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Creating account...
          </>
        ) : (
          "Create Account"
        )}
      </Button>
    </form>
  )
}
