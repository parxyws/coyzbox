package user

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/pkg/helper"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) GetProfile(ctx context.Context) (*UserProfileResponse, error) {
	id, ok := helper.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized: user context missing")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserProfileResponse{
		Id:                  user.Id,
		Name:                user.Name,
		Username:            user.Username,
		Email:               user.Email,
		IsVerified:          user.IsVerified,
		ForcePasswordChange: user.ForcePasswordChange,
		OnboardingCompleted: user.OnboardingCompleted,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UserProfileResponse, error) {
	id, ok := helper.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized: user context missing")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name == "" && req.Username == "" {
		return s.GetProfile(ctx)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Username != "" {
		user.Username = req.Username
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &UserProfileResponse{
		Id:                  user.Id,
		Name:                user.Name,
		Username:            user.Username,
		Email:               user.Email,
		IsVerified:          user.IsVerified,
		ForcePasswordChange: user.ForcePasswordChange,
		OnboardingCompleted: user.OnboardingCompleted,
	}, nil
}

func (s *Service) UpdatePassword(ctx context.Context, req *UpdatePasswordRequest) error {
	id, ok := helper.UserIDFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: user context missing")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.ForcePasswordChange = false

	return s.userRepo.Update(ctx, user)
}
