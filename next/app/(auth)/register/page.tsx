import Link from "next/link"
import { RegisterForm } from "@/components/auth/register-form"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"

export const metadata = { title: "Create Account" }

export default function RegisterPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted/30 p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-xl">
            F
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Create an account</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Join us today and get started in minutes
          </p>
        </div>

        <Card>
          <CardHeader className="pb-4">
            <CardTitle className="text-lg">Sign Up</CardTitle>
            <CardDescription>Fill in your details to create your account</CardDescription>
          </CardHeader>
          <CardContent>
            <RegisterForm />
          </CardContent>
          <CardFooter className="justify-center">
            <p className="text-sm text-muted-foreground">
              Already have an account?{" "}
              <Link href="/login" className="font-medium text-primary hover:underline">
                Sign in
              </Link>
            </p>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
