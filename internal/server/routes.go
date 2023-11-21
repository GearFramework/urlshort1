package server

import (
	"github.com/GearFramework/urlshort/internal/server/handlers"
	"github.com/gin-gonic/gin"
)

func (s *Server) InitRoutes() {
	s.Router.POST("/", func(ctx *gin.Context) { handlers.EncodeURL(ctx, s.api) })
	s.Router.GET("/:code", func(ctx *gin.Context) { handlers.DecodeURL(ctx, s.api) })
	s.Router.POST("/api/shorten", func(ctx *gin.Context) { handlers.EncodeURLFromJSON(ctx, s.api) })
	s.Router.POST("/api/shorten/batch", func(ctx *gin.Context) { handlers.BatchEncodeURLs(ctx, s.api) })
	s.Router.GET("/api/user/urls", func(ctx *gin.Context) { handlers.GetUserURLs(ctx, s.api) })
	s.Router.DELETE("/api/user/urls", func(ctx *gin.Context) { handlers.DeleteUserURLs(ctx, s.api) })
	s.Router.GET("/ping", func(ctx *gin.Context) { handlers.Ping(ctx, s.api) })
	s.Router.NoRoute(handlers.InvalidMethod)
}
