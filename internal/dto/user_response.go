package dto

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

type UpdateEmailResponse struct {
	ReferenceId string `json:"reference_id"`
}

type CommitUpdateResponse struct {
}
