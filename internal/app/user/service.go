package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*UserProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UserProfileResponse, error)
	UpdatePassword(ctx context.Context, userID string, req UpdatePasswordRequest) error
}

type userService struct {
	userStore store.UserStore
}

func NewUserService(userStore store.UserStore) UserService {
	return &userService{userStore: userStore}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*UserProfileResponse, error) {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapUserToResponse(u), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UserProfileResponse, error) {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	u.UpdatedAt = time.Now()

	if err := s.userStore.UpdateUser(ctx, u); err != nil {
		return nil, err
	}

	return mapUserToResponse(u), nil
}

func (s *userService) UpdatePassword(ctx context.Context, userID string, req UpdatePasswordRequest) error {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.CurrentPassword)); err != nil {
		return domain.ErrInvalidCredential
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.Password = string(hashedPassword)
	u.ForcePasswordChange = false
	u.UpdatedAt = time.Now()

	return s.userStore.UpdateUser(ctx, u)
}

func mapUserToResponse(user *domain.User) *UserProfileResponse {
	return &UserProfileResponse{
		Id:                  user.Id,
		Name:                user.Name,
		Username:            user.Username,
		Email:               user.Email,
		IsVerified:          user.IsVerified,
		ForcePasswordChange: user.ForcePasswordChange,
		OnboardingCompleted: user.OnboardingCompleted,
	}
}
