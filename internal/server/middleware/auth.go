package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 10)), //TODO: равен константе
		},
		UserID: 1, //TODO:rand
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (int64, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func CreateCookie() (*http.Cookie, error) {
	token, err := BuildJWTString()
	if err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   20, //TODO: равен константе
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
	return ctx.Value(userIDKey).(int64)
}

func UpsertAuthCookie(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			existedCookie, err := req.Cookie("auth_token")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					log.Debug("no cookie")
					newCookie, err := CreateCookie()
					if err != nil {
						log.Error(err.Error())
					}
					http.SetCookie(res, newCookie)
					log.Debug("new cookie is set")
				} else {
					log.Error(err.Error())
				}
			} else {
				userID, err := ParseToken(existedCookie.Value)
				if err != nil {
					log.Error(err.Error())
					log.Debug("invalid token")
					newCookie, err := CreateCookie()
					if err != nil {
						log.Error(err.Error())
					}
					http.SetCookie(res, newCookie)
					log.Debug("new cookie is set")
				} else if userID == 0 {
					log.Debug("zero user")
				} else {
					ctx := context.WithValue(req.Context(), userIDKey, userID)
					req = req.WithContext(ctx)
				}
			}
			next.ServeHTTP(res, req)
		})
	}
}
