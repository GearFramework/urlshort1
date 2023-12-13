package server

import (
	"github.com/GearFramework/urlshort/internal/app"
	"github.com/GearFramework/urlshort/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

var a *app.ShortApp

func TestRoutes(t *testing.T) {
	var err error
	if a == nil {
		a, err = app.NewShortener(config.GetConfig())
	}
	assert.NoError(t, err)
	s, err := NewServer(a.Conf, a)
	assert.NoError(t, err)
	s.InitRoutes()
	tests := map[string][]struct {
		pathExpected string
		valid        bool
	}{
		"POST": {
			{"/:id", false},
			{"/", true},
			{"/short/:code", false},
			{"/api/shorten", true},
		},
		"GET": {
			{"/:id", false},
			{"/", false},
			{"/:code", true},
			{"/api/shorten", false},
		},
	}
	routes := s.Router.Routes()
	for method, paths := range tests {
		for _, test := range paths {
			exists := isRouteExists(method, test.pathExpected, routes)
			assert.Equal(t, test.valid, exists)
		}
	}
}

func isRouteExists(method, path string, routes gin.RoutesInfo) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
