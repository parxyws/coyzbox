package mail

type OTPData struct {
	Name string
	OTP  string
}

type ResetPasswordData struct {
	Name string
	URL  string
}
