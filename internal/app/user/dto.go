package user

type UpdateProfileRequest struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=30,alphanum"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=128"`
}

type UserProfileResponse struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Username            string `json:"username"`
	Email               string `json:"email"`
	IsVerified          bool   `json:"is_verified"`
	ForcePasswordChange bool   `json:"force_password_change"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
}
