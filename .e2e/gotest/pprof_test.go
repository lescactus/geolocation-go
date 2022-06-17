package e2e

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithPprofEnabled(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Query pprof page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *pprofUrl, nil)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, data)
	})
}

func TestWithPprofDisabled(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Query pprof page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *pprofUrl, nil)

		resp, err := client.Do(req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
