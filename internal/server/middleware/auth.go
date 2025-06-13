package middleware

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func BuildJWTString(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (int64, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	return claims.UserID, err
}

func CreateCookie(userID int64) (*http.Cookie, error) {
	token, err := BuildJWTString(userID)
	if err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	return &cookie, nil
}

type ctxKey int

const (
	userIDKey ctxKey = iota
)

func GetUserID(ctx context.Context) int64 {
	v, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0
	}
	return v
}

func UpsertAuthCookie(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			existedCookie, err := req.Cookie("Authorization")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					log.Debug("no cookie")
					userID := int64(rand.Intn(1000))
					newCookie, err := CreateCookie(userID)
					if err != nil {
						log.Error(err.Error())
					}
					http.SetCookie(res, newCookie)
					ctx := context.WithValue(req.Context(), userIDKey, userID)
					req = req.WithContext(ctx)
					log.Debug("new cookie is set")
				} else {
					log.Error(err.Error())
				}
			} else {
				userID, err := ParseToken(existedCookie.Value)
				if err != nil {
					log.Error(err.Error())
					log.Debug("invalid token")
					newCookie, err := CreateCookie(userID)
					if err != nil {
						log.Error(err.Error())
					}
					http.SetCookie(res, newCookie)
					log.Debug("new cookie is set")
				}
				if userID != 0 {
					ctx := context.WithValue(req.Context(), userIDKey, userID)
					req = req.WithContext(ctx)
				}
			}
			next.ServeHTTP(res, req)
		})
	}
}
