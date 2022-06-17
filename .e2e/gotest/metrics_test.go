package e2e

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func populateMetrics() error {
	requests := []struct {
		ip string
	}{{ip: "1.1.1.1"},{ip: "2.2.2.2"},{ip: "3.3.3.3"},{ip: "4.4.4.4"},{ip: "5.5.5.5"},{ip: "6.6.6.6"},}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, r := range requests {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/v1/%s", *baseUrl, r.ip), nil)
		if err != nil {
			return err
		}

		_, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestWithMetricsEnabled(t *testing.T) {
	err := populateMetrics()
	if err != nil {
		t.Fatalf(err.Error())
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
				atLeast: 6,
			},
			{
				name:    "in_memory_items_read_total",
				metric:  "in_memory_items_read_total",
				atLeast: 0,
			},
			{
				name:    "redis_items_saved_total",
				metric:  "redis_items_saved_total",
				atLeast: 6,
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

func TestWithMetricsDisabled(t *testing.T) {
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
		assert.Equal(t, data, []byte(`{"status":"error","msg":"404 page not found"}`))
		assert.NoError(t, err)
	})
}
