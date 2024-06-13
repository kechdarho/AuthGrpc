package jwt

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"reflect"
	"sso/internal/domain/models"

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

func GenerateToken(ctx context.Context, user models.User, duration time.Duration, secret []byte) (*AccessToken, error) {
	privateKey, err := x509.ParsePKCS8PrivateKey(secret)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	accessToken := &AccessToken{
		Token: signToken,
		Uid:   user.ID,
		Jti:   claims["jti"].(string),
		Nbf:   claims["nbf"].(time.Time),
		Exp:   claims["exp"].(int64),
		Role:  user.Role,
	}

	return accessToken, nil
}

func ValidateToken(ctx context.Context, cacheAccessToken *jwt.Token, requestAccessToken *jwt.Token) error {
	exp, ok := requestAccessToken.Claims.(jwt.MapClaims)["exp"].(float64)
	if !ok {
		return errors.New("token mismatch")
	}
	if time.Now().Unix() > int64(exp) {
		return errors.New("token exp mismatch")
	}

	if !reflect.DeepEqual(cacheAccessToken.Header, requestAccessToken.Header) {
		return errors.New("token mismatch")
	}
	if !reflect.DeepEqual(cacheAccessToken.Claims, requestAccessToken.Claims) {
		return errors.New("token mismatch")
	}
	if !reflect.DeepEqual(cacheAccessToken.Signature, requestAccessToken.Signature) {
		return errors.New("token mismatch")
	}

	return nil
}

func DecodeAccessToken(ctx context.Context, accessToken string, secretKey []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
