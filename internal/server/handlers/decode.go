package handlers

import (
	"github.com/GearFramework/urlshort/internal/app"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DecodeURL(ctx *gin.Context, api pkg.APIShortener) {
	code := ctx.Param("code")
	url, err := api.DecodeURL(ctx, code)
	if err == app.ErrShortURLIsDeleted {
		logger.Log.Errorf("%s\n", err.Error())
		ctx.Status(http.StatusGone)
		return
	}
	if err != nil {
		logger.Log.Errorf("%s\n", err.Error())
		ctx.Status(http.StatusBadRequest)
		return
	}
	logger.Log.Infof("Request short code: %s url: %s", code, url)
	ctx.Header("Location", url)
	ctx.Status(http.StatusTemporaryRedirect)
	ctx.Done()
}
