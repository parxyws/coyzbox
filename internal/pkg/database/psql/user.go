package psql

import (
	"context"
	"errors"

	"github.com/parxyws/cozybox/internal/domain"
	"gorm.io/gorm"
)

type UserRepo struct{ DB *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) Insert(ctx context.Context, u *domain.User) error {
	return r.DB.WithContext(ctx).Create(u).Error
}
func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	return r.DB.WithContext(ctx).Save(u).Error
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.DB.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.DB.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
