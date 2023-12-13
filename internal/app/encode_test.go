package app

import (
	"context"
	"github.com/GearFramework/urlshort/internal/config"
	"github.com/stretchr/testify/assert"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestEncodeURL(t *testing.T) {
	var err error
	if shortener == nil {
		shortener, err = NewShortener(config.GetConfig())
		assert.NoError(t, err)
	}
	shortener.ClearShortly()
	assert.Equal(t, 0, shortener.Store.Count())
	testURLs := []string{
		"http://ya.ru",
		"http://yandex.ru",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, testURI := range testURLs {
		shortURI, _ := shortener.EncodeURL(ctx, 1, testURI)
		assert.NotEmpty(t, shortURI)
		parsedURI, _ := url.ParseRequestURI(shortURI)
		assert.Equal(t, defShortLen, len(strings.TrimLeft(parsedURI.Path, "/")))
		assert.Regexp(t, regexp.MustCompile(`^/[a-zA-Z0-9]+$`), parsedURI.Path)
	}
}

func TestEncodeURLExists(t *testing.T) {
	var err error
	if shortener == nil {
		shortener, err = NewShortener(config.GetConfig())
		assert.NoError(t, err)
	}
	shortener.ClearShortly()
	assert.Equal(t, 0, shortener.Store.Count())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shortener.AddShortly(ctx, 1, "http://ya.ru", "dHGfdhj4")
	shortener.AddShortly(ctx, 1, "http://yandex.ru", "78gsshSd")
	assert.Equal(t, 2, shortener.Store.Count())
	testURLs := []struct {
		url  string
		want string
	}{
		{"http://ya.ru", shortener.Conf.ShortURLHost + "/dHGfdhj4"},
		{"http://yandex.ru", shortener.Conf.ShortURLHost + "/78gsshSd"},
	}
	for _, test := range testURLs {
		shortURL, _ := shortener.EncodeURL(ctx, 1, test.url)
		assert.Equal(t, test.want, shortURL)
	}
}
