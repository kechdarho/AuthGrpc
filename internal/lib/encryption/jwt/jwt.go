package jwt

import (
	"AuthGrpc/internal/domain/models"
	"context"
	"crypto/x509"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type AccessToken struct {
	Token string
	Uid   int64
	Jti   string
	Nbf   time.Time
	Exp   int64
	Role  string
}

func GenerateToken(ctx context.Context, user models.User, duration time.Duration, secret []byte) (string, error) {
	if requestID, ok := ctx.Value("requestID").(string); ok {
		fmt.Println("Request ID:", requestID)
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(secret)
	if err != nil {
		return "", err
	}
	token := jwt.New(jwt.SigningMethodES256)
	token.Header["alg"] = "HS256"
	token.Header["typ"] = "JWT"
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["jti"] = uuid.New().String()
	claims["nbf"] = time.Now()
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["role"] = user.Role
	signToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return signToken, nil
}
