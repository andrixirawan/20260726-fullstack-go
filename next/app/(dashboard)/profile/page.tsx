import { EditProfileForm } from "@/components/profile/edit-profile-form"

export const metadata = { title: "Profile" }

export default function ProfilePage() {
  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Profile Settings</h1>
        <p className="text-muted-foreground">Manage your account information and preferences.</p>
      </div>
      <EditProfileForm />
    </div>
  )
}
