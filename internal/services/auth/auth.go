package auth

import (
	"AuthGrpc/internal/cache"
	"AuthGrpc/internal/cache/local"
	"AuthGrpc/internal/lib/encryption/jwt"
	"AuthGrpc/internal/lib/encryption/token"
	"AuthGrpc/internal/pkg/storage"
	"AuthGrpc/internal/pkg/storage/sqlite"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log             *slog.Logger
	userProvider    storage.UserProvider
	userSaver       storage.UserSaver
	userUpdater     storage.UserUpdater
	tokenSaver      storage.TokenSaver
	tokenProvider   storage.TokenProvider
	tokenUpdater    storage.TokenUpdater
	cache           cache.Cacher
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func New(log *slog.Logger, storage *sqlite.Storage, cache *local.Cache, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, secretKey []byte) *Auth {
	return &Auth{
		userSaver:       storage,
		userProvider:    storage,
		userUpdater:     storage,
		tokenSaver:      storage,
		tokenProvider:   storage,
		tokenUpdater:    storage,
		log:             log,
		cache:           cache,
		secretKey:       secretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, login string, password string) (string, error) {
	const op = "auth.login"

	user, err := a.userProvider.GetUser(ctx, login)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	jwtToken, err := jwt.GenerateToken(ctx, user, a.accessTokenTTL, a.secretKey)
	if err != nil {
		a.log.Error("failed to generate token", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	hash := sha256.Sum256([]byte(jwtToken))
	key := hex.EncodeToString(hash[:])
	value := user.ID
	err = a.cache.Set(ctx, key, value, a.accessTokenTTL*time.Second)
	if err != nil {
		a.log.Error("failed to save in cache", err)
	}

	return jwtToken, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, login, phone, password string) (bool, error) {
	const op = "auth.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	err = a.userSaver.SaveUser(ctx, email, login, phone, passHash)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *Auth) LogOut(ctx context.Context, jwtToken string) (bool, error) {
	const op = "auth.LogOut"
	hash := sha256.Sum256([]byte(jwtToken))
	key := hex.EncodeToString(hash[:])

	if err := a.cache.Delete(ctx, key); err != nil {
		a.log.Error("failed to delete in cache", err)
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (a *Auth) ChangePassword(ctx context.Context, login, oldPassword, newPassword string) (bool, error) {
	const op = "auth.ChangePassword"
	user, err := a.userProvider.GetUser(ctx, login)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(oldPassword)); err != nil {
		a.log.Info("invalid credentials", err)
		return false, fmt.Errorf("%s: %w", op, err)
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	updates := map[string]interface{}{
		"pass_hash": passHash,
	}
	_, err = a.userUpdater.UpdateUser(ctx, user.ID, updates)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (a *Auth) ForgotPassword(ctx context.Context, login string) (string, error) {
	const op = "auth.ForgotPassword"
	user, err := a.userProvider.GetUser(ctx, login)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	resetToken, err := token.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	err = a.tokenSaver.SaveToken(ctx, user.ID, resetToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resetToken, nil
}

func (a *Auth) ResetPassword(ctx context.Context, token, newPassword string) (bool, error) {
	const op = "auth.ForgotPassword"

	resetToken, err := a.tokenProvider.GetToken(ctx, token)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	updates := map[string]interface{}{
		"pass_hash": passHash,
	}
	result, err := a.userUpdater.UpdateUser(ctx, resetToken.UserID, updates)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	err = a.tokenUpdater.DeleteToken(ctx, resetToken.ID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (a *Auth) UpdateUser(ctx context.Context, jwtToken string, email string, phone string) (bool, error) {
	const op = "auth.UpdateUSer"
	hash := sha256.Sum256([]byte(jwtToken))
	key := hex.EncodeToString(hash[:])
	user, ok := a.cache.Get(ctx, key)
	if !ok {
		a.log.Error("failed to find user in cache")
		return false, fmt.Errorf("%s: %w", op, ok)
	}
	updates := map[string]interface{}{
		"email": email,
		"phone": phone,
	}
	_, err := a.userUpdater.UpdateUser(ctx, user.(int64), updates)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}
