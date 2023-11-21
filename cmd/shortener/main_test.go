package main

import (
	"encoding/json"
	"fmt"
	"github.com/GearFramework/urlshort/internal/app"
	"github.com/GearFramework/urlshort/internal/config"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/server"
	"github.com/GearFramework/urlshort/internal/server/handlers"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Test struct {
	name string
	t    *testing.T
	enc  *TestEncode
	dec  *TestDecode
}

type TestEncode struct {
	requestEncode    Req
	responseExpected RespExpected
	responseActual   RespActualEncode
	testEnc          func(t *testing.T, test *TestEncode)
}

type TestDecode struct {
	requestDecode    Req
	responseExpected RespExpected
	responseActual   RespActualDecode
	testDec          func(t *testing.T, test *TestDecode)
}

type Req struct {
	Method  string
	Target  string
	Body    string
	Headers map[string]string
}

type RespExpected struct {
	ResponseURL string
	StatusCode  int
}

type RespActualEncode struct {
	r           *http.Response
	ResponseURL string
}

type RespActualDecode struct {
	r          *http.Response
	StatusCode int
}

func (test *Test) test(t *testing.T, api pkg.APIShortener) {
	test.t = t
	if test.enc != nil {
		test.testEncode(api)
		return
	}
	if test.dec != nil {
		test.testDecode(api)
	}
}

func (test *Test) testEncode(api pkg.APIShortener) {
	request := httptest.NewRequest(
		test.enc.requestEncode.Method,
		test.enc.requestEncode.Target,
		strings.NewReader(test.enc.requestEncode.Body),
	)
	for head, value := range test.enc.requestEncode.Headers {
		request.Header.Add(head, value)
	}
	w := httptest.NewRecorder()
	s, err := server.NewServer(&config.ServiceConfig{Addr: "localhost:8080", LoggerLevel: "info"}, api)
	assert.NoError(test.t, err)
	s.InitRoutes()
	s.Router.ServeHTTP(w, request)
	response := w.Result()
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	assert.NoError(test.t, err)
	assert.Equal(test.t, test.enc.responseExpected.StatusCode, response.StatusCode)
	test.enc.responseActual = RespActualEncode{response, string(body)}
	if test.enc.testEnc != nil {
		test.enc.testEnc(test.t, test.enc)
	}
	if response.StatusCode == http.StatusCreated && test.dec != nil {
		test.dec.requestDecode.Target = string(body)
		test.testDecode(api)
	}
}

func (test *Test) testDecode(api pkg.APIShortener) {
	fmt.Println("Response url ", test.dec.requestDecode.Target)
	request := httptest.NewRequest(test.dec.requestDecode.Method, test.dec.requestDecode.Target, nil)
	w := httptest.NewRecorder()
	s, err := server.NewServer(&config.ServiceConfig{Addr: "localhost:8080", LoggerLevel: "info"}, api)
	assert.NoError(test.t, err)
	s.InitRoutes()
	s.Router.ServeHTTP(w, request)
	response := w.Result()
	defer response.Body.Close()
	assert.Equal(test.t, test.dec.responseExpected.StatusCode, response.StatusCode)
	if test.dec.testDec != nil {
		test.dec.responseActual = RespActualDecode{response, response.StatusCode}
		test.dec.testDec(test.t, test.dec)
	}
}

func getTests() []Test {
	return []Test{
		{
			name: "valid url encode valid method decode",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodPost,
					"/",
					"https://ya.ru",
					map[string]string{"Content-Type": "text/plain"},
				},
				responseExpected: RespExpected{StatusCode: http.StatusCreated},
				testEnc: func(t *testing.T, test *TestEncode) {
					assert.Regexp(t, "^http://localhost:8080/[a-zA-Z0-9]{8}$", test.responseActual.ResponseURL)
					assert.Equal(t, "text/plain", test.responseActual.r.Header.Get("Content-Type"))
				},
			},
			dec: &TestDecode{
				requestDecode:    Req{Method: http.MethodGet},
				responseExpected: RespExpected{"https://ya.ru", http.StatusTemporaryRedirect},
				testDec: func(t *testing.T, test *TestDecode) {
					assert.Equal(t, test.responseExpected.ResponseURL, test.responseActual.r.Header.Get("Location"))
				},
			},
		}, {
			name: "valid url encode invalid method decode",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodPost,
					"/",
					"https://yandex.ru",
					map[string]string{"Content-Type": "text/plain"},
				},
				responseExpected: RespExpected{StatusCode: http.StatusCreated},
				testEnc: func(t *testing.T, test *TestEncode) {
					assert.Regexp(t, "^http://localhost:8080/[a-zA-Z0-9]{8}$", test.responseActual.ResponseURL)
					assert.Equal(t, "text/plain", test.responseActual.r.Header.Get("Content-Type"))
				},
			},
			dec: &TestDecode{
				requestDecode:    Req{Method: http.MethodPut},
				responseExpected: RespExpected{StatusCode: http.StatusBadRequest},
			},
		}, {
			name: "invalid url",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodPost,
					"/",
					"https//ya.ru",
					map[string]string{"Content-Type": "text/plain"},
				},
				responseExpected: RespExpected{StatusCode: http.StatusBadRequest},
			},
		}, {
			name: "invalid request method",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodDelete,
					"/",
					"https://ya.ru",
					map[string]string{"Content-Type": "text/plain"},
				},
				responseExpected: RespExpected{StatusCode: http.StatusBadRequest},
			},
		}, {
			name: "invalid short url",
			dec: &TestDecode{
				requestDecode:    Req{Method: http.MethodGet, Target: "http://localhost:8080/8tbujofj"},
				responseExpected: RespExpected{StatusCode: http.StatusBadRequest},
			},
		}, {
			name: "json request encode",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodPost,
					"/api/shorten",
					`{"url":"https://ya.ru"}`,
					map[string]string{"Content-Type": "application/json"},
				},
				responseExpected: RespExpected{StatusCode: http.StatusConflict},
				testEnc: func(t *testing.T, test *TestEncode) {
					assert.Contains(t, test.responseActual.r.Header.Get("Content-Type"), "application/json")
					assert.NotEmpty(t, test.responseActual.ResponseURL)
					r := strings.NewReader(test.responseActual.ResponseURL)
					dec := json.NewDecoder(r)
					var req handlers.ResponseJSON
					err := dec.Decode(&req)
					assert.NoError(t, err)
					assert.NotEmpty(t, req.Result)
					assert.Regexp(t, "^http://localhost:8080/[a-zA-Z0-9]{8}$", req.Result)
				},
			},
		}, {
			name: "compress request encode",
			enc: &TestEncode{
				requestEncode: Req{
					http.MethodPost,
					"/api/shorten",
					`{"url":"https://ya.ru"}`,
					map[string]string{
						"Content-Type":    "application/json",
						"Accept-Encoding": "gzip",
					},
				},
				responseExpected: RespExpected{StatusCode: http.StatusConflict},
				testEnc: func(t *testing.T, test *TestEncode) {
					assert.Contains(t, test.responseActual.r.Header.Get("Content-Type"), "application/json")
					assert.Contains(t, test.responseActual.r.Header.Get("Content-Encoding"), "gzip")
				},
			},
		},
	}
}

func TestHandleServiceEncode(t *testing.T) {
	a, err := app.NewShortener(config.GetConfig())
	a.ClearShortly()
	assert.NoError(t, err)
	for _, test := range getTests() {
		t.Run(test.name, func(t *testing.T) {
			test.test(t, a)
		})
	}
}
