package store

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type UserStore interface {
	InsertUser(ctx context.Context, u *domain.User) error
	UpdateUser(ctx context.Context, u *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
}

type psqlUserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) UserStore {
	return &psqlUserStore{db: db}
}

func (s *psqlUserStore) InsertUser(ctx context.Context, u *domain.User) error {
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (s *psqlUserStore) UpdateUser(ctx context.Context, u *domain.User) error {
	if err := s.db.WithContext(ctx).Save(u).Error; err != nil {
		return err
	}
	return nil
}

func (s *psqlUserStore) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *psqlUserStore) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.GetUserByID(ctx, id)
}

func (s *psqlUserStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *psqlUserStore) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := s.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
