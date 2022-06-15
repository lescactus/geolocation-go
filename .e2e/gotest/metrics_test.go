package e2e

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

func parseMF(r io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func TestWithMetricsEnabled(t *testing.T) {
	if os.Getenv("NO_METRICS") == "true" {
		t.Skip("geolocation-go is running without metrics enabled. Skipping")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Scrape metrics page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *metricsUrl, nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, data)

		// From: https://github.com/prometheus/pushgateway#command-line
		//
		// Note that in the text protocol, each line has to end with a line-feed character (aka 'LF' or '\n').
		// Ending a line in other ways, e.g. with 'CR' aka '\r', 'CRLF' aka '\r\n', or just the end of the packet,
		// will result in a protocol error.
		//
		// Appending a '\n' to the body reponse
		data = append(data, '\n')
		mf, err := parseMF(bytes.NewReader(data))
		if err != nil {
			t.Fatalf(err.Error())
		}
		assert.NoError(t, err)

		tests := []struct {
			name    string
			metric  string
			atLeast float64
		}{
			{
				name:    "in_memory_items_saved_total",
				metric:  "in_memory_items_saved_total",
				atLeast: 3,
			},
			{
				name:    "in_memory_items_read_total",
				metric:  "in_memory_items_read_total",
				atLeast: 0,
			},
			{
				name:    "redis_items_saved_total",
				metric:  "redis_items_saved_total",
				atLeast: 3,
			},
			{
				name:    "redis_items_read_total",
				metric:  "redis_items_read_total",
				atLeast: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := mf[tt.metric]
				v := m.GetMetric()[0].GetCounter().GetValue()

				assert.GreaterOrEqual(t, v, tt.atLeast)
			})
		}
	})
}

func TestWithoutMetricsEnabled(t *testing.T) {
	if os.Getenv("NO_METRICS") != "true" {
		t.Skip("geolocation-go is running with metrics enabled. Skipping")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Scrape metrics page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", *metricsUrl, nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Empty(t, data)
		assert.NoError(t, err)
	})
}
