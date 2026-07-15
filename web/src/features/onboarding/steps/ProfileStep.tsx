import { useEffect, useRef } from "react"
import { z } from "zod"
import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { IconUser, IconUpload } from "@tabler/icons-react"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { OnboardingProfileData } from "../useOnboardingStore"

const profileSchema = z.object({
  fullName: z.string().min(2, { message: "Full name must be at least 2 characters" }),
  displayName: z.string().min(2, { message: "Display name must be at least 2 characters" }),
  avatarPreview: z.string().optional(),
})

type ProfileFormValues = z.infer<typeof profileSchema>

interface ProfileStepProps {
  defaultValues: OnboardingProfileData
  onValidSubmit: (data: OnboardingProfileData) => void
  formId: string
}

export function ProfileStep({ defaultValues, onValidSubmit, formId }: ProfileStepProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues,
  })

  const avatarPreview = form.watch("avatarPreview")

  // Auto-fill display name from full name if display name is empty
  const fullName = form.watch("fullName")
  useEffect(() => {
    if (!form.getValues("displayName")) {
      const parts = fullName.trim().split(" ")
      form.setValue("displayName", parts[0] ?? "")
    }
  }, [fullName]) // eslint-disable-line react-hooks/exhaustive-deps

  function handleAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (ev) => {
      form.setValue("avatarPreview", ev.target?.result as string)
    }
    reader.readAsDataURL(file)
  }

  return (
    <form id={formId} onSubmit={form.handleSubmit(onValidSubmit)}>
      {/* Avatar picker */}
      <div className="mb-6 flex flex-col items-center gap-3">
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          className="group relative flex size-20 items-center justify-center overflow-hidden rounded-full border-2 border-dashed border-border bg-muted transition-colors hover:border-primary"
        >
          {avatarPreview ? (
            <img src={avatarPreview} alt="Avatar preview" className="size-full object-cover" />
          ) : (
            <IconUser className="size-8 text-muted-foreground transition-colors group-hover:text-primary" />
          )}
          <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
            <IconUpload className="size-5 text-white" />
          </div>
        </button>
        <p className="text-xs text-muted-foreground">Click to upload a profile photo (optional)</p>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleAvatarChange}
        />
      </div>

      <FieldGroup>
        {/* Full name */}
        <Controller
          control={form.control}
          name="fullName"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="onboarding-full-name">Full Name</FieldLabel>
              <Input
                {...field}
                id="onboarding-full-name"
                type="text"
                placeholder="Jane Doe"
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />

        {/* Display name */}
        <Controller
          control={form.control}
          name="displayName"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="onboarding-display-name">Display Name</FieldLabel>
              <Input
                {...field}
                id="onboarding-display-name"
                type="text"
                placeholder="jane"
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
      </FieldGroup>
    </form>
  )
}
