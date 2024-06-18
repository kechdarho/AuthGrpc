package storage

import (
	"AuthGrpc/internal/domain/models"
	"context"
	"time"
)

type (
	UserSaver interface {
		SaveUser(ctx context.Context, email string, login string, phone string, passHash []byte) error
	}

	UserProvider interface {
		GetUser(ctx context.Context, login string) (models.User, error)
	}

	UserUpdater interface {
		UpdateUser(ctx context.Context, id int64, updates map[string]interface{}) (bool, error)
	}

	TokenSaver interface {
		SaveToken(ctx context.Context, id int64, token string, expiresAt time.Time) error
	}

	TokenProvider interface {
		GetToken(ctx context.Context, token string) (models.ResetToken, error)
	}
	TokenUpdater interface {
		DeleteToken(ctx context.Context, id int64) error
	}
)
