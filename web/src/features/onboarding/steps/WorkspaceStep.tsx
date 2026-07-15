import { useEffect, useRef } from "react"
import { z } from "zod"
import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import type { OnboardingWorkspaceData } from "../useOnboardingStore"

const workspaceSchema = z.object({
  workspaceName: z.string().min(2, { message: "Workspace name must be at least 2 characters" }),
  workspaceSlug: z
    .string()
    .min(2, { message: "Slug must be at least 2 characters" })
    .regex(/^[a-z0-9-]+$/, { message: "Slug can only contain lowercase letters, numbers, and hyphens" }),
  workspaceDescription: z.string().optional(),
})

type WorkspaceFormValues = z.infer<typeof workspaceSchema>

interface WorkspaceStepProps {
  defaultValues: OnboardingWorkspaceData
  onValidSubmit: (data: OnboardingWorkspaceData) => void
  formId: string
}

function toSlug(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
}

export function WorkspaceStep({ defaultValues, onValidSubmit, formId }: WorkspaceStepProps) {
  const form = useForm<WorkspaceFormValues>({
    resolver: zodResolver(workspaceSchema),
    defaultValues,
  })

  const workspaceName = form.watch("workspaceName")

  const isSlugManualRef = useRef(false)

  // Auto-generate slug from workspace name while the user hasn't manually edited it
  useEffect(() => {
    if (!isSlugManualRef.current) {
      form.setValue("workspaceSlug", toSlug(workspaceName), { shouldValidate: true })
    }
  }, [workspaceName, form])

  return (
    <form id={formId} onSubmit={form.handleSubmit(onValidSubmit)}>
      <FieldGroup>
        {/* Workspace name */}
        <Controller
          control={form.control}
          name="workspaceName"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="onboarding-workspace-name">Workspace Name</FieldLabel>
              <Input
                {...field}
                id="onboarding-workspace-name"
                type="text"
                placeholder="Acme Corp"
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />

        {/* Workspace slug */}
        <Controller
          control={form.control}
          name="workspaceSlug"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="onboarding-workspace-slug">Workspace URL</FieldLabel>
              <div className="flex items-center gap-0 rounded-md border border-input bg-muted/40 text-xs">
                <span className="select-none border-r border-input px-2 py-1.5 text-muted-foreground">
                  cozybox.app/
                </span>
                <input
                  {...field}
                  id="onboarding-workspace-slug"
                  type="text"
                  placeholder="acme-corp"
                  aria-invalid={fieldState.invalid}
                  className="h-7 min-w-0 flex-1 bg-transparent px-2 py-0.5 text-xs outline-none placeholder:text-muted-foreground"
                  onChange={(e) => {
                    isSlugManualRef.current = true
                    field.onChange(e)
                  }}
                />
              </div>
              <FieldDescription>This will be your unique workspace URL.</FieldDescription>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />

        {/* Description */}
        <Controller
          control={form.control}
          name="workspaceDescription"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="onboarding-workspace-description">
                Description{" "}
                <span className="font-normal text-muted-foreground">(optional)</span>
              </FieldLabel>
              <Textarea
                {...field}
                id="onboarding-workspace-description"
                placeholder="What does your team work on?"
                rows={3}
                className="resize-none"
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
