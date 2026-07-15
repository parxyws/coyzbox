import { OnboardingWizard } from "@/features/onboarding/OnboardingWizard"

export default function OnboardingPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card px-8 py-10 shadow-sm">
        <OnboardingWizard />
      </div>
    </div>
  )
}
