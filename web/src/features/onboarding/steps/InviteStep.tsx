import { useState } from "react"
import { IconPlus, IconTrash, IconMail } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import type { OnboardingInviteData } from "../useOnboardingStore"

interface InviteStepProps {
  defaultValues: OnboardingInviteData
  onValidSubmit: (data: OnboardingInviteData) => void
  onSkip: () => void
  formId: string
}

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export function InviteStep({ defaultValues, onValidSubmit, onSkip, formId }: InviteStepProps) {
  const [emails, setEmails] = useState<string[]>(
    defaultValues.emails.length > 0 ? defaultValues.emails : [""]
  )
  const [errors, setErrors] = useState<Record<number, string>>({})

  function handleEmailChange(index: number, value: string) {
    setEmails((prev) => prev.map((e, i) => (i === index ? value : e)))
    if (errors[index]) {
      setErrors((prev) => {
        const next = { ...prev }
        delete next[index]
        return next
      })
    }
  }

  function addEmail() {
    setEmails((prev) => [...prev, ""])
  }

  function removeEmail(index: number) {
    setEmails((prev) => prev.filter((_, i) => i !== index))
    setErrors((prev) => {
      const next: Record<number, string> = {}
      Object.entries(prev).forEach(([k, v]) => {
        const ki = Number(k)
        if (ki !== index) {
          next[ki > index ? ki - 1 : ki] = v
        }
      })
      return next
    })
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const filled = emails.filter((e) => e.trim() !== "")

    const newErrors: Record<number, string> = {}
    emails.forEach((email, i) => {
      if (email.trim() !== "" && !isValidEmail(email)) {
        newErrors[i] = "Please enter a valid email address"
      }
    })

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    onValidSubmit({ emails: filled, skipped: false })
  }

  return (
    <form id={formId} onSubmit={handleSubmit}>
      <FieldGroup>
        {emails.map((email, index) => (
          <Field key={index} data-invalid={!!errors[index]}>
            {index === 0 && (
              <FieldLabel htmlFor={`onboarding-invite-email-${index}`}>
                Team member emails
              </FieldLabel>
            )}
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <IconMail className="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id={`onboarding-invite-email-${index}`}
                  type="email"
                  value={email}
                  onChange={(e) => handleEmailChange(index, e.target.value)}
                  placeholder="teammate@example.com"
                  aria-invalid={!!errors[index]}
                  className="pl-7"
                />
              </div>
              {emails.length > 1 && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => removeEmail(index)}
                  className="shrink-0 text-muted-foreground hover:text-destructive"
                >
                  <IconTrash />
                </Button>
              )}
            </div>
            {errors[index] && <FieldError>{errors[index]}</FieldError>}
          </Field>
        ))}
      </FieldGroup>

      <div className="mt-4 flex items-center justify-between">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={addEmail}
          className="gap-1.5 text-muted-foreground"
        >
          <IconPlus className="size-3.5" />
          Add another
        </Button>

        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onSkip}
          className="text-muted-foreground"
        >
          Skip for now
        </Button>
      </div>
    </form>
  )
}
