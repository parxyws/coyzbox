package user

import "context"

type UserService interface {
	GetUser(ctx context.Context)
	UpdateAccount(ctx context.Context)
	DeleteAccount(ctx context.Context)
}
