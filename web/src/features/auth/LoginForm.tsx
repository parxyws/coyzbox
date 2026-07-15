import { z } from "zod"
import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field.tsx"
import { Input } from "@/components/ui/input.tsx"
import { Button } from "@/components/ui/button.tsx"
import { Separator } from "@/components/ui/separator.tsx"
import {
  IconAlertSquareRounded,
  IconBrandGoogleHome,
} from "@tabler/icons-react"

const loginFormSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .pipe(z.email("Invalid email address")),
  password: z.string().min(1, "Password required"),
})

export function LoginForm() {
  const loginForm = useForm<z.infer<typeof loginFormSchema>>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  })

  function onSubmit(data: z.infer<typeof loginFormSchema>) {
    try {
      console.log(data)
    } catch (error) {
      console.error(error)
    }
  }

  return (
    <form
      id={"login-form"}
      onSubmit={loginForm.handleSubmit(onSubmit)}
      className={"flex flex-col gap-4"}
    >
      <FieldGroup>
        {/*--- email field ---*/}
        <Controller
          control={loginForm.control}
          name={"email"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"login-form-email"}>
                Email
                {fieldState.invalid && (
                  <FieldError>
                    <IconAlertSquareRounded className="inline size-3.5" />{" "}
                    {fieldState.error?.message}
                  </FieldError>
                )}
              </FieldLabel>
              <Input
                {...field}
                id={"login-form-email"}
                aria-invalid={fieldState.invalid}
                placeholder={"cozybox@example.com"}
              />
            </Field>
          )}
        />
        {/*--- password field ---*/}
        <Controller
          control={loginForm.control}
          name={"password"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"login-form-password"}>
                Password
                {fieldState.invalid && (
                  <FieldError>
                    <IconAlertSquareRounded className="inline size-3.5" />{" "}
                    {fieldState.error?.message}
                  </FieldError>
                )}
              </FieldLabel>
              <Input
                {...field}
                type={"password"}
                id={"login-form-password"}
                aria-invalid={fieldState.invalid}
              />
            </Field>
          )}
        />
      </FieldGroup>
      <Field>
        <Button type={"submit"} form={"login-form"} size={"lg"}>
          Submit
        </Button>
      </Field>
      <div className="relative flex items-center gap-2">
        <Separator className="flex-1" />
        <span className="shrink-0 px-2 text-sm font-medium text-muted-foreground">
          or
        </span>
        <Separator className="flex-1" />
      </div>
      <FieldGroup>
        <Field>
          <Button type={"button"} size={"lg"} variant={"outline"}>
            <IconBrandGoogleHome /> Sign in with Google
          </Button>
        </Field>
      </FieldGroup>
      <FieldGroup>
        <Field>
          <div className="text-center text-sm">
            Don&apos;t have an account?{" "}
            <a href="#" className="underline underline-offset-4">
              Sign Up
            </a>
          </div>
        </Field>
      </FieldGroup>
    </form>
  )
}
