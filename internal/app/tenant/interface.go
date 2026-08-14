package tenant

import (
	"context"

	"github.com/parxyws/cozybox/internal/domain"
)

type FileStorage interface {
	PutObject(ctx context.Context, input domain.UploadInput) (key string, err error)
}
