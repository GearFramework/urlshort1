package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func InvalidMethod(ctx *gin.Context) {
	ctx.Status(http.StatusBadRequest)
}
