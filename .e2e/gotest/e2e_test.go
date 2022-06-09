package e2e

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
	"flag"

	"github.com/stretchr/testify/assert"
)

type GeoIP struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	City        string  `json:"city,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

const defaultBaseUrl = "http://127.0.0.1:8080"

var baseUrl = flag.String("baseurl", "http://127.0.0.1:8080", "Base URL of geolocation-go service")

func TestE2E(t *testing.T) {
	flag.Parse()
	if *baseUrl == "" {
		*baseUrl = defaultBaseUrl
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   []byte
		code   int
	}{
		{
			name:   "valid path - " + *baseUrl + "/rest/v1/1.1.1.1",
			url:    "" + *baseUrl + "/rest/v1/1.1.1.1",
			method: "GET",
			want:   []byte(`{"ip":"1.1.1.1","country_code":"AU","country_name":"Australia","city":"South Brisbane","latitude":-27.4766,"longitude":153.0166}`),
			code:   http.StatusOK,
		},
		{
			name:   "valid path - " + *baseUrl + "/rest/v1/2.2.2.2",
			url:    "" + *baseUrl + "/rest/v1/2.2.2.2",
			method: "GET",
			want:   []byte(`{"ip":"2.2.2.2","country_code":"FR","country_name":"France","city":"Paris","latitude":48.8566,"longitude":2.35222}`),
			code:   http.StatusOK,
		},
		{
			name:   "valid path - " + *baseUrl + "/rest/v1/3.3.3.3",
			url:    "" + *baseUrl + "/rest/v1/3.3.3.3",
			method: "GET",
			want:   []byte(`{"ip":"3.3.3.3","country_code":"US","country_name":"United States","city":"Chicago","latitude":41.8781,"longitude":-87.6298}`),
			code:   http.StatusOK,
		},
		{
			name:   "invalid path - " + *baseUrl + "/rest/v1/bla",
			url:    "" + *baseUrl + "/rest/v1/bla",
			method: "GET",
			want:   []byte(`{"status":"error","msg":"the provided IP is not a valid IPv4 address"}`),
			code:   http.StatusBadRequest,
		},
		{
			name:   "invalid path - " + *baseUrl + "/invalid",
			url:    "" + *baseUrl + "/invalid",
			method: "GET",
			want:   []byte(`{"status":"error","msg":"404 page not found"}`),
			code:   http.StatusNotFound,
		},
		{
			name:   "method not allowed - " + *baseUrl + "/rest/v1/1.1.1.1",
			url:    "" + *baseUrl + "/rest/v1/1.1.1.1",
			method: "POST",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			var geoip *GeoIP
			data, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, data)
			assert.Equal(t, tt.code, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type http header wasn't set to 'application/json'")
			assert.NotEmpty(t, resp.Header.Get("X-Request-Id"), "X-Request-Id http header wasn't set")
			assert.NoError(t, json.Unmarshal(data, &geoip))
		})
	}
}
