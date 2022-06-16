package e2e

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithPprofEnabled(t *testing.T) {
	if os.Getenv("NO_PPROF") == "true" {
		t.Skip("geolocation-go is running without pprof enabled. Skipping")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	t.Run("Query pprof page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *pprofUrl, nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, data)
	})
}

func TestWithPprofDisabled(t *testing.T) {
	if os.Getenv("NO_PPROF") != "true" {
		t.Skip("geolocation-go is running with pprof enabled. Skipping")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Query pprof page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *pprofUrl, nil)

		resp, err := client.Do(req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
