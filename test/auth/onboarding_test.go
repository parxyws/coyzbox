package auth_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/domain"
)

func TestOnboarding_SuccessWithImage(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	userID := "user-123"
	tenantID := "tenant-1"

	req := &auth.OnboardingRequest{
		Email:        "org@example.com",
		Phone:        "+1234567890",
		AddressLine1: "123 Main St",
		City:         "Metropolis",
		State:        "NY",
		PostalCode:   "10001",
		Country:      "US",
		TaxId:        "TAX-001",
	}

	image := &domain.UploadInput{
		Object:      bytes.NewReader([]byte("fake-image-data")),
		ObjectName:  "logo.png",
		ObjectSize:  1024,
		ContentType: "image/png",
	}

	m.tenantRepo.On("GetByID", ctx, tenantID).Return(&domain.Tenant{
		Id:   tenantID,
		Name: "Test Corp",
	}, nil)

	m.userRepo.On("GetByID", ctx, userID).Return(&domain.User{
		Id:                  userID,
		OnboardingCompleted: false,
	}, nil)

	m.fileStorage.On("PutObject", ctx, mock.MatchedBy(func(in domain.UploadInput) bool {
		return in.ObjectName == "logo.png"
	})).Return("logos/logo.png", nil)

	m.orgRepo.On("Insert", ctx, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.TenantId == tenantID && o.LogoS3Key == "logos/logo.png"
	})).Return(nil)

	m.userRepo.On("Update", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Id == userID && u.OnboardingCompleted == true
	})).Return(nil)

	err := svc.CompleteOnboarding(ctx, userID, tenantID, req, image)
	require.NoError(t, err)
	m.tenantRepo.AssertExpectations(t)
	m.userRepo.AssertExpectations(t)
	m.fileStorage.AssertExpectations(t)
	m.orgRepo.AssertExpectations(t)
}

func TestOnboarding_SuccessWithoutImage(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	userID := "user-123"
	tenantID := "tenant-1"

	req := &auth.OnboardingRequest{
		Email: "org@example.com",
		City:  "Metropolis",
	}

	m.tenantRepo.On("GetByID", ctx, tenantID).Return(&domain.Tenant{Id: tenantID, Name: "Test Corp"}, nil)
	m.userRepo.On("GetByID", ctx, userID).Return(&domain.User{Id: userID, OnboardingCompleted: false}, nil)

	// No fileStorage.PutObject call expected

	m.orgRepo.On("Insert", ctx, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.TenantId == tenantID && o.LogoS3Key == ""
	})).Return(nil)

	m.userRepo.On("Update", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.OnboardingCompleted == true
	})).Return(nil)

	err := svc.CompleteOnboarding(ctx, userID, tenantID, req, nil)
	require.NoError(t, err)
}

func TestOnboarding_TenantNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.tenantRepo.On("GetByID", ctx, "tenant-unknown").Return(nil, domain.ErrNotFound)

	err := svc.CompleteOnboarding(ctx, "user-123", "tenant-unknown", &auth.OnboardingRequest{}, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOnboarding_UserNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.tenantRepo.On("GetByID", ctx, "tenant-1").Return(&domain.Tenant{Id: "tenant-1", Name: "Test Corp"}, nil)
	m.userRepo.On("GetByID", ctx, "user-unknown").Return(nil, domain.ErrNotFound)

	err := svc.CompleteOnboarding(ctx, "user-unknown", "tenant-1", &auth.OnboardingRequest{}, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOnboarding_ImageUploadFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.tenantRepo.On("GetByID", ctx, "tenant-1").Return(&domain.Tenant{Id: "tenant-1", Name: "Test Corp"}, nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123"}, nil)
	m.fileStorage.On("PutObject", ctx, mock.Anything).Return("", errors.New("s3 error"))

	err := svc.CompleteOnboarding(ctx, "user-123", "tenant-1", &auth.OnboardingRequest{}, &domain.UploadInput{
		Object:      bytes.NewReader([]byte("data")),
		ObjectName:  "logo.png",
		ObjectSize:  100,
		ContentType: "image/png",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload image")
}

func TestOnboarding_OrgInsertFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.tenantRepo.On("GetByID", ctx, "tenant-1").Return(&domain.Tenant{Id: "tenant-1", Name: "Test Corp"}, nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123"}, nil)
	m.orgRepo.On("Insert", ctx, mock.Anything).Return(errors.New("db error"))

	err := svc.CompleteOnboarding(ctx, "user-123", "tenant-1", &auth.OnboardingRequest{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save organization")
}

func TestOnboarding_UserUpdateFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.tenantRepo.On("GetByID", ctx, "tenant-1").Return(&domain.Tenant{Id: "tenant-1", Name: "Test Corp"}, nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123"}, nil)
	m.orgRepo.On("Insert", ctx, mock.Anything).Return(nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	err := svc.CompleteOnboarding(ctx, "user-123", "tenant-1", &auth.OnboardingRequest{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to complete onboarding")
}
