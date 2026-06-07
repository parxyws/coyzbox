package auth

type RegisterUserRequest struct {
	Name       string `json:"name"       validate:"required,min=2,max=100"`
	Username   string `json:"username"   validate:"required,min=3,max=30,alphanum"`
	Email      string `json:"email"      validate:"required,email"`
	Password   string `json:"password"   validate:"required,min=8,max=128"`
	TenantName string `json:"tenant_name" validate:"required,min=2,max=100"`
}

type VerifyEmailRequest struct {
	ReferenceId string `json:"reference_id" validate:"required"`
	Otp         string `json:"otp"          validate:"required,len=6,numeric"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	SessionID    string `json:"session_id"    validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	ReferenceId string `json:"reference_id" validate:"required"`
	Token       string `json:"token" validate:"required"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
}

type OnboardingRequest struct {
	Email           string `json:"email" validate:"email"`
	Phone           string `json:"phone" validate:"number"`
	AddressLine1    string `json:"address_line_1"`
	AddressLine2    string `json:"address_line_2"`
	City            string `json:"city"`
	State           string `json:"state"`
	PostalCode      string `json:"postal_code"`
	Country         string `json:"country"`
	TaxId           string `json:"tax_number"`
	LogoS3URL       string `json:"logo_s3_url"`
	DefaultCurrency string `json:"default_currency"`
}

type UserResponse struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Username            string `json:"username"`
	Email               string `json:"email"`
	ForcePasswordChange bool   `json:"force_password_change"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
}

type RegisterResponse struct {
	ReferenceId string `json:"reference_id"`
}

type VerifyEmailResponse struct {
	ReferenceId string `json:"reference_id"`
}

type UserAuthenticateResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	TokenType    string         `json:"token_type"`
	User         UserResponse   `json:"user"`
	Tenant       TenantResponse `json:"tenant"`
}

type TenantResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}
