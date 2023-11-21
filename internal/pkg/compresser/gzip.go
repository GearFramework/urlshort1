package compresser

import (
	"github.com/gin-gonic/gin"
	"io"
)

type Compressor struct {
	gin.ResponseWriter
	Writer io.Writer
}

func NewCompressor() gin.HandlerFunc {
	return newCompressHandler().Handle
}

func (c *Compressor) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

func (c *Compressor) WriteString(s string) (int, error) {
	c.Header().Del("Content-Length")
	return c.Writer.Write([]byte(s))
}

func (c *Compressor) WriteHeader(code int) {
	c.Header().Del("Content-Length")
	c.ResponseWriter.WriteHeader(code)
}
