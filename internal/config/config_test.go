package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		params           []string
		envs             map[string]string
		expectedAddr     string
		expectedShortURL string
	}{
		{params: []string{"-a", ":8080", "-b", "http://localhost:8080"}, expectedAddr: ":8080", expectedShortURL: "http://localhost:8080"},
		{params: []string{"-a", ":8081", "-b", "http://localhost:8081"}, expectedAddr: ":8081", expectedShortURL: "http://localhost:8081"},
		{params: []string{"-a", ":8082"}, expectedAddr: ":8082", expectedShortURL: "http://localhost:8080"},
		{params: []string{"-a", ":8083", "-b", "http://localhost:8084"}, expectedAddr: ":8083", expectedShortURL: "http://localhost:8084"},
		{params: []string{"-b", "http://localhost:8484"}, expectedAddr: ":8080", expectedShortURL: "http://localhost:8484"},
		{envs: map[string]string{"SERVER_ADDRESS": ":9090"}, expectedAddr: ":9090", expectedShortURL: "http://localhost:8080"},
		{envs: map[string]string{"SERVER_ADDRESS": ":9090", "BASE_URL": "http://localhost:9090"}, expectedAddr: ":9090", expectedShortURL: "http://localhost:9090"},
		{envs: map[string]string{"BASE_URL": "http://localhost:9090"}, expectedAddr: ":8080", expectedShortURL: "http://localhost:9090"},
		{
			envs:             map[string]string{"SERVER_ADDRESS": ":9090", "BASE_URL": "http://localhost:9090"},
			params:           []string{"-a", ":8080", "-b", "http://localhost:8080"},
			expectedAddr:     ":9090",
			expectedShortURL: "http://localhost:9090",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": ":9090"},
			params:           []string{"-a", ":7070", "-b", "http://localhost:7070"},
			expectedAddr:     ":9090",
			expectedShortURL: "http://localhost:7070",
		},
		{
			envs:             map[string]string{"BASE_URL": "http://localhost:9090"},
			params:           []string{"-a", ":7070", "-b", "http://localhost:7070"},
			expectedAddr:     ":7070",
			expectedShortURL: "http://localhost:9090",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": "", "BASE_URL": ""},
			params:           []string{"-a", ":6060", "-b", "http://localhost:6060"},
			expectedAddr:     ":6060",
			expectedShortURL: "http://localhost:6060",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": "", "BASE_URL": ""},
			params:           []string{"-b", "http://localhost:6060"},
			expectedAddr:     ":8080",
			expectedShortURL: "http://localhost:6060",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": "", "BASE_URL": "http://localhost:9090"},
			params:           []string{"-a", ":6060"},
			expectedAddr:     ":6060",
			expectedShortURL: "http://localhost:9090",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": "", "BASE_URL": ""},
			params:           []string{"-a", ":6060"},
			expectedAddr:     ":6060",
			expectedShortURL: "http://localhost:8080",
		},
		{
			envs:             map[string]string{"SERVER_ADDRESS": ":9090", "BASE_URL": ""},
			params:           []string{"-b", "http://localhost:7070"},
			expectedAddr:     ":9090",
			expectedShortURL: "http://localhost:7070",
		},
	}
	var old []string
	copy(old, os.Args)
	for _, test := range tests {
		var buf strings.Builder
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		flag.CommandLine.SetOutput(&buf)
		os.Args = []string{os.Args[0]}
		if len(test.params) > 0 {
			os.Args = append(os.Args, test.params...)
		}
		for key, value := range test.envs {
			_ = os.Setenv(key, value)
		}
		conf := GetConfig()
		assert.Equal(t, test.expectedAddr, conf.Addr)
		assert.Equal(t, test.expectedShortURL, conf.ShortURLHost)
		for key := range test.envs {
			_ = os.Unsetenv(key)
		}
		copy(os.Args, old)
	}
}
