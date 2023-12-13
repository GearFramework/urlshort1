package compresser

import (
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"sync"
)

type compressHandler struct {
	pool sync.Pool
}

func newCompressHandler() *compressHandler {
	handler := &compressHandler{
		pool: sync.Pool{
			New: func() interface{} {
				gz, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
				if err != nil {
					panic(err)
				}
				return gz
			},
		},
	}
	return handler
}

func (c *compressHandler) Handle(ctx *gin.Context) {
	if ctx.Request.Header.Get("Content-Encoding") == "gzip" {
		c.DecompressHandle(ctx)
	}
	if !c.canCompress(ctx.Request) {
		fmt.Println(ctx.Request.Header.Get("Accept-Encoding"))
		return
	}
	gz := c.pool.Get().(*gzip.Writer)
	defer c.pool.Put(gz)
	defer gz.Reset(io.Discard)
	gz.Reset(ctx.Writer)

	ctx.Header("Content-Encoding", "gzip")
	ctx.Header("Vary", "Accept-Encoding")
	ctx.Writer = &Compressor{ctx.Writer, gz}
	defer func() {
		gz.Close()
		ctx.Header("Content-Length", fmt.Sprint(ctx.Writer.Size()))
	}()
	ctx.Next()
}

func (c *compressHandler) DecompressHandle(ctx *gin.Context) {
	if ctx.Request.Body == nil {
		return
	}
	r, err := gzip.NewReader(ctx.Request.Body)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ctx.Request.Header.Del("Content-Encoding")
	ctx.Request.Header.Del("Content-Length")
	ctx.Request.Body = r
}

func (c *compressHandler) canCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	if !strings.Contains(req.Header.Get("Content-Type"), "text/html") &&
		!strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		return false
	}
	return true
}
