import { RegisterForm } from "@/features/auth/RegisterForm"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"

export default function RegisterPage() {
  return (
    <div className="flex h-screen items-center justify-center">
      <div>
        <div className={"flex items-end justify-end"}>
          <div
            className={
              "w-[15%] rounded-t-lg bg-primary py-1 text-center font-bold text-muted"
            }
          >
            COZYBOX
          </div>
        </div>
        <Card className="flex max-w-5xl min-w-[75%] flex-col overflow-hidden rounded-tr-none py-0 sm:flex-row sm:gap-0">
          {/* Banner image */}
          <div className="grow-0">
            <img
              src="https://images.unsplash.com/photo-1615990531332-e8725ae1e3a3?q=80&w=1180&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
              alt="banner"
              className="size-full object-cover"
            />
          </div>

          {/* Form side */}
          <div className="flex flex-col gap-4 px-4 py-8 sm:min-w-1/2">
            {/*<div className="flex w-full items-center justify-center">*/}
            {/*  <img src={cozyboxLogo} alt="Cozybox logo" className="w-35" />*/}
            {/*</div>*/}
            <div>
              <CardHeader className={"pb-2"}>
                <CardTitle className="">Register Your Account</CardTitle>
                <CardDescription>
                  Create an account to start using Cozybox.
                </CardDescription>
              </CardHeader>

              <CardContent>
                <RegisterForm />
              </CardContent>

              <CardFooter></CardFooter>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
