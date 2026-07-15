import { Fragment, useRef } from "react"
import { useNavigate } from "react-router-dom"
import { AnimatePresence, motion } from "framer-motion"
import { IconChevronLeft } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useOnboardingStore, STEPS } from "./useOnboardingStore"
import { ProfileStep } from "./steps/ProfileStep"
import { WorkspaceStep } from "./steps/WorkspaceStep"
import { InviteStep } from "./steps/InviteStep"
import { DoneStep } from "./steps/DoneStep"
import { createWorkspace, inviteMembers, completeOnboarding } from "@/api/onboarding"

const FORM_ID = "onboarding-step-form"

const slideVariants = {
  enter: (direction: number) => ({
    x: direction > 0 ? 40 : -40,
    opacity: 0,
  }),
  center: {
    x: 0,
    opacity: 1,
    transition: { duration: 0.3, ease: "easeOut" },
  },
  exit: (direction: number) => ({
    x: direction > 0 ? -40 : 40,
    opacity: 0,
    transition: { duration: 0.2, ease: "easeIn" },
  }),
}

const stepMeta: Record<number, { title: string; description: string }> = {
  0: {
    title: "Set up your profile",
    description: "Tell us a bit about yourself. You can always change this later.",
  },
  1: {
    title: "Create your workspace",
    description: "A workspace is where your team collaborates in Cozybox.",
  },
  2: {
    title: "Invite your team",
    description: "Great teams work together. Invite your colleagues to get started.",
  },
  3: {
    title: "You're all set!",
    description: "Your Cozybox workspace is ready to use.",
  },
}

export function OnboardingWizard() {
  const navigate = useNavigate()
  const store = useOnboardingStore()
  const {
    currentStep,
    totalSteps,
    formData,
    isSubmitting,
    setIsSubmitting,
    goNext,
    goPrev,
    setProfileData,
    setWorkspaceData,
    setInviteData,
  } = store

  const prevStepRef = useRef(currentStep)
  const directionRef = useRef(1)
  if (prevStepRef.current !== currentStep) {
    directionRef.current = currentStep > prevStepRef.current ? 1 : -1
    prevStepRef.current = currentStep
  }
  const direction = directionRef.current
  const isDoneStep = currentStep === totalSteps - 1
  const isInviteStep = currentStep === 2
  const meta = stepMeta[currentStep]

  async function handleFinish() {
    setIsSubmitting(true)
    try {
      await createWorkspace({
        name: formData.workspace.workspaceName,
        slug: formData.workspace.workspaceSlug,
        description: formData.workspace.workspaceDescription,
      })
      if (!formData.invite.skipped && formData.invite.emails.length > 0) {
        await inviteMembers({ emails: formData.invite.emails })
      }
      await completeOnboarding()
      navigate("/dashboard")
    } catch {
      // In a real app, show a toast. For now, still navigate.
      navigate("/dashboard")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex w-full max-w-md flex-col gap-0">
      {/* Header label */}
      <div className="mb-6 flex items-center justify-between">
        <span className="text-xs font-semibold tracking-widest text-muted-foreground uppercase">
          Cozybox
        </span>
        {!isDoneStep && (
          <span className="text-xs text-muted-foreground">
            Step {currentStep + 1} of {totalSteps - 1}
          </span>
        )}
      </div>

      {/* Step indicator */}
      {!isDoneStep && (
        <div className="mb-8 flex w-full">
          {STEPS.slice(0, -1).map((step, index) => {
            const isCompleted = index < currentStep
            const isActive = index === currentStep
            const isLast = index === STEPS.slice(0, -1).length - 1
            return (
              <Fragment key={step.id}>
                {/* Circle + label */}
                <div className="flex w-14 flex-col items-center gap-1.5">
                  <div
                    className={cn(
                      "flex size-7 items-center justify-center rounded-full border-2 text-xs font-semibold transition-all duration-300",
                      isCompleted
                        ? "border-primary bg-primary text-primary-foreground"
                        : isActive
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border bg-muted text-muted-foreground"
                    )}
                  >
                    {isCompleted ? (
                      <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                        <path d="M2 6l3 3 5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    ) : (
                      index + 1
                    )}
                  </div>
                  <span
                    className={cn(
                      "text-[10px] font-medium transition-colors",
                      isActive ? "text-primary" : "text-muted-foreground"
                    )}
                  >
                    {step.label}
                  </span>
                </div>

                {/* Connector line between circles */}
                {!isLast && (
                  <div
                    className={cn(
                      "mt-3.5 h-px flex-1 transition-all duration-500",
                      isCompleted ? "bg-primary" : "bg-border"
                    )}
                  />
                )}
              </Fragment>
            )
          })}
        </div>
      )}

      {/* Step title & description */}
      {!isDoneStep && (
        <div className="mb-6">
          <h1 className="text-xl font-semibold">{meta.title}</h1>
          <p className="mt-1 text-xs text-muted-foreground">{meta.description}</p>
        </div>
      )}

      {/* Animated step content */}
      <div className="min-h-[280px]">
        <AnimatePresence mode="wait" custom={direction}>
          <motion.div
            key={currentStep}
            custom={direction}
            variants={slideVariants}
            initial="enter"
            animate="center"
            exit="exit"
          >
            {currentStep === 0 && (
              <ProfileStep
                formId={FORM_ID}
                defaultValues={formData.profile}
                onValidSubmit={(data) => {
                  setProfileData(data)
                  goNext()
                }}
              />
            )}
            {currentStep === 1 && (
              <WorkspaceStep
                formId={FORM_ID}
                defaultValues={formData.workspace}
                onValidSubmit={(data) => {
                  setWorkspaceData(data)
                  goNext()
                }}
              />
            )}
            {currentStep === 2 && (
              <InviteStep
                formId={FORM_ID}
                defaultValues={formData.invite}
                onValidSubmit={(data) => {
                  setInviteData(data)
                  goNext()
                }}
                onSkip={() => {
                  setInviteData({ emails: [], skipped: true })
                  goNext()
                }}
              />
            )}
            {currentStep === 3 && (
              <DoneStep
                formData={formData}
                isSubmitting={isSubmitting}
                onFinish={handleFinish}
              />
            )}
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Footer navigation */}
      {!isDoneStep && (
        <div className="mt-8 flex items-center justify-between gap-3 border-t border-border pt-6">
          <Button
            type="button"
            variant="ghost"
            size="default"
            onClick={goPrev}
            disabled={currentStep === 0}
            className="gap-1"
          >
            <IconChevronLeft className="size-3.5" />
            Back
          </Button>

          <Button
            type="submit"
            form={FORM_ID}
            size="default"
          >
            {isInviteStep ? "Send Invites & Continue" : "Next"}
          </Button>
        </div>
      )}
    </div>
  )
}
