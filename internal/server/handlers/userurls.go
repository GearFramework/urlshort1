package handlers

import (
	"encoding/json"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetUserURLs(ctx *gin.Context, api pkg.APIShortener) {
	userID, ok := ctx.Get(pkg.UserIDParamName)
	if !ok {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	userURLs := api.GetUserURLs(ctx, userID.(int))
	if len(userURLs) == 0 {
		ctx.JSON(http.StatusNoContent, userURLs)
		return
	}
	ctx.JSON(http.StatusOK, userURLs)
}

func DeleteUserURLs(ctx *gin.Context, api pkg.APIShortener) {
	if !strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
		logger.Log.Errorf(
			"invalid request header: Content-Type %s\n",
			ctx.Request.Header.Get("Content-Type"),
		)
		ctx.Status(http.StatusBadRequest)
		return
	}
	userID, ok := ctx.Get(pkg.UserIDParamName)
	if !ok {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	defer ctx.Request.Body.Close()
	dec := json.NewDecoder(ctx.Request.Body)
	var codes []string
	if err := dec.Decode(&codes); err != nil {
		logger.Log.Errorln("invalid urls in request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	api.DeleteUserURLs(ctx, userID.(int), codes)
	ctx.Status(http.StatusAccepted)
}
