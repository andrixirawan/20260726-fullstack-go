"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { CheckCircle, User, Sparkles, ArrowRight, Loader2 } from "lucide-react"

import { useAuthStore } from "@/store/auth.store"
import { Progress } from "@/components/ui/progress"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

const STEPS = [
  {
    id: 1,
    title: "Welcome to Fullstack Go! 👋",
    description: "We're excited to have you here. Let's quickly walk you through what you can do.",
    icon: Sparkles,
    content: (
      <div className="space-y-4">
        <div className="rounded-lg border bg-muted/40 p-4 space-y-3">
          <Feature icon="🔐" title="Secure Authentication" desc="Your account is protected with JWT tokens and bcrypt encryption." />
          <Feature icon="👤" title="Profile Management" desc="Customize your profile and upload a photo anytime." />
          <Feature icon="📁" title="File Uploads" desc="Upload files securely with our backend storage system." />
        </div>
      </div>
    ),
  },
  {
    id: 2,
    title: "Your account is ready",
    description: "Here's a quick summary of your new account.",
    icon: User,
    content: null, // Injected at render
  },
  {
    id: 3,
    title: "You're all set! 🎉",
    description: "Everything is configured. Let's go to your dashboard.",
    icon: CheckCircle,
    content: (
      <div className="flex flex-col items-center gap-4 py-6">
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
          <CheckCircle className="h-10 w-10 text-green-600 dark:text-green-400" />
        </div>
        <p className="text-center text-sm text-muted-foreground">
          Your profile is set up and you&apos;re ready to explore. Click below to go to your dashboard.
        </p>
      </div>
    ),
  },
]

function Feature({ icon, title, desc }: { icon: string; title: string; desc: string }) {
  return (
    <div className="flex items-start gap-3">
      <span className="text-xl">{icon}</span>
      <div>
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-muted-foreground">{desc}</p>
      </div>
    </div>
  )
}

export default function OnboardingPage() {
  const router = useRouter()
  const { user, completeOnboarding } = useAuthStore()
  const [step, setStep] = useState(0)
  const [completing, setCompleting] = useState(false)

  const currentStep = STEPS[step]
  const progress = ((step + 1) / STEPS.length) * 100

  const handleNext = async () => {
    if (step < STEPS.length - 1) {
      setStep((s) => s + 1)
    } else {
      setCompleting(true)
      completeOnboarding()
      await new Promise((r) => setTimeout(r, 600))
      router.push("/dashboard")
    }
  }

  const StepIcon = currentStep.icon

  // Dynamic content for step 2
  const stepContent =
    step === 1 ? (
      <div className="rounded-lg border bg-muted/40 p-4 space-y-3">
        <InfoRow label="Name" value={user?.full_name ?? "-"} />
        <InfoRow label="Email" value={user?.email ?? "-"} />
        <InfoRow label="Status" value="Active" />
        <InfoRow label="Joined" value={user?.created_at ? new Date(user.created_at).toLocaleDateString() : "-"} />
      </div>
    ) : currentStep.content

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background to-muted/30 p-4">
      <div className="w-full max-w-lg">
        {/* Progress */}
        <div className="mb-8 space-y-2">
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>Step {step + 1} of {STEPS.length}</span>
            <span>{Math.round(progress)}% complete</span>
          </div>
          <Progress value={progress} />
        </div>

        <Card className="overflow-hidden">
          <CardContent className="p-8">
            {/* Step icon */}
            <div className="mb-6 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10">
              <StepIcon className="h-7 w-7 text-primary" />
            </div>

            <h2 className="text-2xl font-bold tracking-tight">{currentStep.title}</h2>
            <p className="mt-2 mb-6 text-sm text-muted-foreground">{currentStep.description}</p>

            {stepContent}

            <div className="mt-8 flex items-center justify-between">
              {step > 0 ? (
                <Button variant="ghost" onClick={() => setStep((s) => s - 1)} disabled={completing}>
                  Back
                </Button>
              ) : (
                <div />
              )}
              <Button onClick={handleNext} disabled={completing} className="gap-2">
                {completing ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Taking you in...
                  </>
                ) : step === STEPS.length - 1 ? (
                  <>
                    Go to Dashboard
                    <ArrowRight className="h-4 w-4" />
                  </>
                ) : (
                  <>
                    Next
                    <ArrowRight className="h-4 w-4" />
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Step dots */}
        <div className="mt-6 flex justify-center gap-2">
          {STEPS.map((_, i) => (
            <div
              key={i}
              className={`h-2 rounded-full transition-all duration-300 ${
                i === step ? "w-6 bg-primary" : i < step ? "w-2 bg-primary/50" : "w-2 bg-muted"
              }`}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}
