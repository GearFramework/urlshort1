package server

import (
	"fmt"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/pkg/auth"
	"github.com/GearFramework/urlshort/internal/pkg/compresser"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (s *Server) logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		duration := logger.GetDurationInMilliseconds(start)
		logger.Log.Infoln(
			"uri", ctx.Request.RequestURI,
			"method", ctx.Request.Method,
			"status", ctx.Writer.Status(), // получаем перехваченный код статуса ответа
			"duration", fmt.Sprintf("%.4f ms", duration),
			"size", ctx.Writer.Size(), // получаем перехваченный размер ответа
		)
	}
}

func (s *Server) compress() gin.HandlerFunc {
	return compresser.NewCompressor()
}

const CookieParamName = "Athorization"

func (s *Server) auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userID int
		c, err := ctx.Request.Cookie(CookieParamName)
		// кука не установлена, создаём пользователя с новым userID
		if err != nil || c == nil || c.Value == "" {
			logger.Log.Infoln("empty token in cookie; need new token")
			if userID, err = s.AuthNewUser(ctx); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			ctx.Set(pkg.UserIDParamName, userID)
			ctx.Next()
			return
		}
		// пытаемся из токена получить userID
		logger.Log.Infof("token in cookie: %s", c.Value)
		if userID, err = s.AuthFromToken(ctx, c.Value); err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		logger.Log.Infof("Authorized user with ID: %d", userID)
		ctx.Set(pkg.UserIDParamName, userID)
		ctx.Next()
	}
}

func (s *Server) AuthNewUser(ctx *gin.Context) (int, error) {
	userID, token, err := s.api.CreateToken()
	if err != nil {
		logger.Log.Error(err.Error())
		return 0, err
	}
	s.setAuthCookie(ctx, token)
	logger.Log.Infof("Created user ID: %d", userID)
	return userID, nil
}

func (s *Server) AuthFromToken(ctx *gin.Context, token string) (int, error) {
	userID, err := s.api.Auth(token)
	if err != nil && err == auth.ErrInvalidAuthorization {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return 0, err
	}
	if err != nil && err == auth.ErrNeedAuthorization {
		return s.AuthNewUser(ctx)
	}
	s.setAuthCookie(ctx, token)
	return userID, err
}

func (s *Server) setAuthCookie(ctx *gin.Context, token string) {
	ctx.SetCookie(CookieParamName,
		token,
		int(auth.TokenExpired.Seconds()),
		"/",
		"localhost",
		true,
		true,
	)
}
