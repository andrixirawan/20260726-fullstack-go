import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().min(1, "Email is required").email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
})

export const registerBaseSchema = z.object({
  full_name: z.string().min(1, "Full name is required").max(255, "Name too long"),
  email: z.string().min(1, "Email is required").email("Invalid email address"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(72, "Password too long"),
  confirm_password: z.string().min(1, "Please confirm your password"),
})

export const registerSchema = registerBaseSchema.refine((data) => data.password === data.confirm_password, {
  message: "Passwords do not match",
  path: ["confirm_password"],
})

export const updateProfileSchema = z.object({
  full_name: z.string().min(1, "Full name is required").max(255, "Name too long"),
  avatar_url: z.string().url("Invalid URL").optional().or(z.literal("")),
})

export type LoginFormValues = z.infer<typeof loginSchema>
export type RegisterFormValues = z.infer<typeof registerSchema>
export type UpdateProfileFormValues = z.infer<typeof updateProfileSchema>
