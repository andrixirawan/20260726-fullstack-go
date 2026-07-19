"use client"

import { useForm } from "@tanstack/react-form"
import { useState } from "react"
import { Loader2, Upload, Trash2 } from "lucide-react"
import { cn } from "@/lib/utils"

import { updateProfileSchema, type UpdateProfileFormValues } from "@/lib/schemas"
import { userApi, ApiError, API_BASE_URL } from "@/lib/api"
import { useAuthStore } from "@/store/auth.store"
import { toast } from "@/components/ui/sonner"
import { Button, buttonVariants } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export function EditProfileForm() {
  const { user, setUser } = useAuthStore()
  const [uploading, setUploading] = useState(false)
  const [saving, setSaving] = useState(false)

  const form = useForm({
    defaultValues: {
      full_name: user?.full_name ?? "",
      avatar_url: user?.avatar_url ?? "",
    } as UpdateProfileFormValues,
    onSubmit: async ({ value }) => {
      const result = updateProfileSchema.safeParse(value)
      if (!result.success) {
        toast.error(result.error.issues[0]?.message ?? "Validation failed")
        return
      }
      setSaving(true)
      try {
        const updated = await userApi.updateProfile(result.data)
        setUser(updated)
        toast.success("Profile updated successfully!")
      } catch (err) {
        if (err instanceof ApiError) {
          toast.error(err.message)
        } else {
          toast.error("Failed to update profile")
        }
      } finally {
        setSaving(false)
      }
    },
  })

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const result = await userApi.uploadAvatar(file)
      const avatarUrl = `${API_BASE_URL}${result.url}`
      form.setFieldValue("avatar_url", avatarUrl)
      const updated = await userApi.updateProfile({ avatar_url: avatarUrl })
      setUser(updated)
      toast.success("Avatar uploaded!")
    } catch {
      toast.error("Failed to upload avatar")
    } finally {
      setUploading(false)
    }
  }

  const handleRemoveAvatar = async () => {
    setUploading(true)
    try {
      form.setFieldValue("avatar_url", "")
      const updated = await userApi.updateProfile({ avatar_url: "" })
      setUser(updated)
      toast.success("Avatar removed!")
    } catch {
      toast.error("Failed to remove avatar")
    } finally {
      setUploading(false)
    }
  }

  const initials = user?.full_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) ?? "?"

  const getFieldError = (errors: unknown[]) =>
    errors.length > 0 ? String(errors[0]) : null

  return (
    <div className="space-y-6">
      {/* Avatar Section */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Profile Picture</CardTitle>
          <CardDescription>Upload a photo to personalize your account</CardDescription>
        </CardHeader>
        <CardContent className="flex items-center gap-6">
          <Avatar className="h-20 w-20">
            <AvatarImage src={user?.avatar_url} alt={user?.full_name} />
            <AvatarFallback className="text-xl">{initials}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="avatar-upload" className="cursor-pointer">
                <span className={cn(buttonVariants({ variant: "outline", size: "sm" }), uploading && "opacity-50 pointer-events-none")}>
                  {uploading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Upload className="mr-2 h-4 w-4" />
                  )}
                  {uploading ? "Uploading..." : "Upload Photo"}
                </span>
              </Label>
              {user?.avatar_url && (
                <Button
                  type="button"
                  variant="destructive"
                  size="sm"
                  onClick={handleRemoveAvatar}
                  disabled={uploading}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Remove
                </Button>
              )}
            </div>
            <input
              id="avatar-upload"
              type="file"
              accept=".jpg,.jpeg,.png,.gif,.webp"
              onChange={handleAvatarUpload}
              className="hidden"
            />
            <p className="text-xs text-muted-foreground">JPG, PNG, GIF, WebP. Max 10MB.</p>
          </div>
        </CardContent>
      </Card>

      {/* Profile Info Section */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Personal Information</CardTitle>
          <CardDescription>Update your name and contact details</CardDescription>
        </CardHeader>
        <CardContent>
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
                  const r = updateProfileSchema.shape.full_name.safeParse(value)
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
                    disabled={saving}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-destructive">
                      {getFieldError(field.state.meta.errors)}
                    </p>
                  )}
                </div>
              )}
            />

            <div className="space-y-1.5">
              <Label>Email Address</Label>
              <Input value={user?.email ?? ""} disabled className="bg-muted/50" />
              <p className="text-xs text-muted-foreground">Email cannot be changed</p>
            </div>

            <Separator />

            <div className="flex justify-end">
              <Button type="submit" disabled={saving}>
                {saving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  "Save Changes"
                )}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Account Info */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Account Details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">User ID</span>
            <span className="font-mono text-xs">{user?.id}</span>
          </div>
          <Separator />
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Member since</span>
            <span>
              {user?.created_at
                ? new Date(user.created_at).toLocaleDateString("en-US", {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })
                : "-"}
            </span>
          </div>
          <Separator />
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Status</span>
            <span className="inline-flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full bg-green-500" />
              Active
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
