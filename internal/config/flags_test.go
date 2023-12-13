package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		params           []string
		expectedAddr     string
		expectedShortURL string
	}{
		{params: []string{"-a", ":8080", "-b", "http://localhost:8080"}, expectedAddr: ":8080", expectedShortURL: "http://localhost:8080"},
		{params: []string{"-a", ":8081", "-b", "http://localhost:8081"}, expectedAddr: ":8081", expectedShortURL: "http://localhost:8081"},
		{params: []string{"-a", ":8082"}, expectedAddr: ":8082", expectedShortURL: "http://localhost:8080"},
		{params: []string{"-a", ":8083", "-b", "http://localhost:8084"}, expectedAddr: ":8083", expectedShortURL: "http://localhost:8084"},
		{params: []string{"-b", "http://localhost:8484"}, expectedAddr: ":8080", expectedShortURL: "http://localhost:8484"},
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
		conf := GetConfig()
		assert.Equal(t, test.expectedAddr, conf.Addr)
		assert.Equal(t, test.expectedShortURL, conf.ShortURLHost)
		copy(os.Args, old)
	}
}
