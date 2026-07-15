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

const registerFormSchema = z
  .object({
    username: z.string().min(1, "Username is required"),
    email: z
      .string()
      .min(1, "Email is required")
      .pipe(z.email("Invalid email address")),
    password: z
      .string()
      .min(1, "Password required")
      .min(8, { message: "Password must be at least 8 characters long" })
      .regex(/[A-Z]/, {
        message: "Password must contain at least one uppercase letter",
      })
      .regex(/[a-z]/, {
        message: "Password must contain at least one lowercase letter",
      })
      .regex(/[0-9]/, { message: "Password must contain at least one number" })
      .regex(/[^A-Za-z0-9]/, {
        message: "Password must contain at least one special character",
      }),
    confirmPassword: z
      .string()
      .min(1, "Password required")
      .min(8, { message: "Password must be at least 8 characters long" })
      .regex(/[A-Z]/, {
        message: "Password must contain at least one uppercase letter",
      })
      .regex(/[a-z]/, {
        message: "Password must contain at least one lowercase letter",
      })
      .regex(/[0-9]/, { message: "Password must contain at least one number" })
      .regex(/[^A-Za-z0-9]/, {
        message: "Password must contain at least one special character",
      }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  })

export function RegisterForm() {
  const registerForm = useForm<z.infer<typeof registerFormSchema>>({
    resolver: zodResolver(registerFormSchema),
    defaultValues: {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  })

  function onSubmit(data: z.infer<typeof registerFormSchema>) {
    try {
      console.log(data)
    } catch (error) {
      console.log(error)
    }
  }

  return (
    <form
      id={"register-form"}
      onSubmit={registerForm.handleSubmit(onSubmit)}
      className={"flex flex-col gap-4"}
    >
      <FieldGroup>
        {/*--- username field ---*/}
        <Controller
          control={registerForm.control}
          name={"username"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"register-form-username"}>
                Username
                {fieldState.invalid && (
                  <FieldError>
                    <IconAlertSquareRounded className="inline size-3.5" />{" "}
                    {fieldState.error?.message}
                  </FieldError>
                )}
              </FieldLabel>
              <Input
                {...field}
                type={"text"}
                id={"register-form-username"}
                aria-invalid={fieldState.invalid}
                placeholder={"cozybox"}
              />
            </Field>
          )}
        />
        {/*--- email field ---*/}
        <Controller
          control={registerForm.control}
          name={"email"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"register-form-email"}>
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
                id={"register-form-email"}
                aria-invalid={fieldState.invalid}
                placeholder={"cozybox@example.com"}
              />
            </Field>
          )}
        />
        {/*--- password field ---*/}
        <Controller
          control={registerForm.control}
          name={"password"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"register-form-password"}>
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
                id={"register-form-password"}
                aria-invalid={fieldState.invalid}
              />
            </Field>
          )}
        />
        {/*--- confirm password field ---*/}
        <Controller
          control={registerForm.control}
          name={"confirmPassword"}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={"register-form-confirmPassword"}>
                Confirm Password
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
                id={"register-form-confirmPassword"}
                aria-invalid={fieldState.invalid}
              />
            </Field>
          )}
        />
      </FieldGroup>
      <Field>
        <Button type={"submit"} form={"register-form"} size={"lg"}>
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
            <IconBrandGoogleHome /> Sign up with Google
          </Button>
        </Field>
      </FieldGroup>
      <FieldGroup>
        <Field>
          <div className="text-sm text-center">
            Already have an account?{" "}
            <a href="#" className="underline underline-offset-4">
              Log In
            </a>
          </div>
        </Field>
      </FieldGroup>
    </form>
  )
}
