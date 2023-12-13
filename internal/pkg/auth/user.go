package auth

import (
	"errors"
	"fmt"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const (
	TokenExpired = time.Hour * 24
	SecretKey    = "bu7HBJD&873HVHJdh*Jbhsfdfs8622Dsf"
)

var ErrNeedAuthorization = errors.New("требуется авторизация")
var ErrInvalidAuthorization = errors.New("отсутствует ID пользователя")

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWT(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpired)),
		},
		UserID: userID,
	})
	tk, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tk, nil
}

func GetUserIDFromJWT(tk string) int {
	claims, err := getClaims(tk)
	if err != nil {
		return -1
	}
	logger.Log.Infof("app user ID: %d", claims.UserID)
	return claims.UserID
}

func getClaims(tk string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tk, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil || !token.Valid {
		logger.Log.Error(err.Error())
		return nil, err
	}
	return claims, nil
}
