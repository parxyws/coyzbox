package dto

type RegisterUserRequest struct {
	Name       string `json:"name"       validate:"required,min=2,max=100"`
	Username   string `json:"username"   validate:"required,min=3,max=30,alphanum"`
	Email      string `json:"email"      validate:"required,email"`
	Password   string `json:"password"   validate:"required,min=8,max=128"`
	TenantName string `json:"tenant_name" validate:"required,min=2,max=100"`
}

type CreateMemberRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role"  validate:"required,oneof=admin member"`
}

type VerifyEmailRequest struct {
	ReferenceId string `json:"reference_id" validate:"required"`
	Otp         string `json:"otp"          validate:"required,len=6,numeric"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	SessionID    string `json:"session_id"    validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=30,alphanum"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=128"`
}

type UpdateEmailRequest struct {
	NewEmail string `json:"new_email"        validate:"required,email"`
}

type CommitUpdateEmailRequest struct {
	ReferenceId string `json:"reference_id" validate:"required"`
	Otp         string `json:"otp"          validate:"required,len=6,numeric"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"      validate:"required,email"`
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
