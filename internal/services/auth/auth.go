package auth

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/cache"
	"sso/internal/domain/models"
	"sso/internal/lib/encryption/jwt"
	"sso/internal/pkg/storage/sqlite"
	"time"
)

type Auth struct {
	log             *slog.Logger
	userProvider    UserProvider
	userSaver       UserSaver
	cache           Cacher
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type (
	UserSaver interface {
		SaveUser(
			ctx context.Context,
			email string,
			login string,
			phone string,
			passHash []byte,
		) (id int64, err error)
	}
	UserProvider interface {
		User(ctx context.Context, login string) (models.User, error)
	}

	Cacher interface {
		Set(
			ctx context.Context,
			key string,
			value interface{},
			duration time.Duration,
		) error
		Get(
			ctx context.Context,
			key string,
		) (interface{}, bool)
		Delete(
			ctx context.Context,
			key string,
		) error
	}
)

func New(log *slog.Logger, storage *sqlite.Storage, cache *cache.Cache, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, secretKey []byte) *Auth {
	return &Auth{
		userSaver:       storage,
		userProvider:    storage,
		log:             log,
		cache:           cache,
		secretKey:       secretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, login string, password string) (string, error) {
	const op = "auth.login"

	user, err := a.userProvider.User(ctx, login)

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)
	}
	authToken, err := jwt.GenerateToken(ctx, user, a.accessTokenTTL, a.secretKey)
	if err != nil {
		a.log.Error("failed to generate token", err)
	}

	err = a.cache.Set(ctx, authToken.Jti, authToken.Token, time.Duration(authToken.Exp)*time.Second)
	if err != nil {
		a.log.Error("failed to save in cache", err)
	}

	return authToken.Token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, login, phone, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	result, err := a.userSaver.SaveUser(ctx, email, login, phone, passHash)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (a *Auth) LogOut(ctx context.Context, token string) (bool, error) {
	panic("implement me")
}

/*
func (a *Auth) ValidateToken(ctx context.Context, token string) (bool, error) {
	const op = "auth.ValidateToken"
	err := jwt.ValidateToken(ctx, []byte(token), a.secretKey)
	if err != nil {
		return false, fmt.Errorf("%s:%w", op, err)
	}
	return result, nil
}
*/
