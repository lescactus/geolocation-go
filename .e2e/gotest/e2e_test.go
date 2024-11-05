package e2e

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
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

var (
	baseUrl         = flag.String("baseurl", "http://127.0.0.1:8080", "Base URL of geolocation-go service")
	metricsUrl      = flag.String("metricsurl", "http://127.0.0.1:8080/metrics", "Metrics URL of the geolocation-go service")
	pprofUrl        = flag.String("pprofurl", "http://127.0.0.1:6060/debug/pprof/", "Pprof URL of the geolocation-go service")
	redisConnString = flag.String("redisconnstr", "redis://localhost:6379", "Redis connection string")

	rdb *redis.Client
)

func setup() {
	flag.Parse()
	setupRedisClient()
	resetRedis()
}

func TestE2E(t *testing.T) {
	setup()

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
			want:   []byte(`{"ip":"2.2.2.2","country_code":"FR","country_name":"France","city":"Issy-les-Moulineaux","latitude":48.8309,"longitude":2.26582}`),
			code:   http.StatusOK,
		},
		{
			name:   "valid path - " + *baseUrl + "/rest/v1/3.3.3.3",
			url:    "" + *baseUrl + "/rest/v1/3.3.3.3",
			method: "GET",
			want:   []byte(`{"ip":"3.3.3.3","country_code":"US","country_name":"United States","city":"Ashburn","latitude":39.0469,"longitude":-77.4903}`),
			code:   http.StatusOK,
		},
		{
			name:   "invalid path - " + *baseUrl + "/rest/v1/bla",
			url:    "" + *baseUrl + "/rest/v1/bla",
			method: "GET",
			want:   []byte(`{"status":"error","msg":"the provided ip is not a valid ipv4 address"}`),
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
			name:   "OPTIONS - method allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "OPTIONS",
			want:   []byte(``),
			code:   http.StatusOK,
		},
		{
			name:   "POST - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "POST",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "PUT - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "PUT",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "PATCH - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "PATCH",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "DELETE - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "DELETE",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "CONNECT - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "CONNECT",
			want:   []byte(`{"status":"error","msg":"405 method not allowed"}`),
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "TRACE - method not allowed - " + *baseUrl + "/rest/v1/4.4.4.4",
			url:    "" + *baseUrl + "/rest/v1/4.4.4.4",
			method: "TRACE",
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
			assert.NotEmpty(t, resp.Header.Get("X-Request-Id"), "X-Request-Id http header wasn't set")

			if tt.method == "OPTIONS" {
				assert.Equal(t, "GET, OPTIONS", resp.Header.Get("Allow"), "Allow http header wasn't set to 'GET, OPTIONS'")
			} else {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type http header wasn't set to 'application/json'")
				assert.NoError(t, json.Unmarshal(data, &geoip))
			}

			// Ensure the key has been stored in redis successfully.
			// As it is an asynchronous operation, using a retry allow
			// to wait until the server has saved the key in redis.
			if (tt.code == http.StatusOK) && (tt.method == "GET") {
				var ok bool
				var err error
				for retry := 0; retry < 5; retry++ {
					ip := getIP(tt.url)
					ok, err = isInRedis(ip)
					if ok {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}
				assert.Equal(t, true, ok)
				assert.NoError(t, err)
			} else {
				ip := getIP(tt.url)
				ok, err := isInRedis(ip)
				assert.Equal(t, false, ok)
				assert.Error(t, err)
			}
		})
	}
}

func setupRedisClient() {
	opt, err := redis.ParseURL(*redisConnString)
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(opt)
}

func resetRedis() {
	rdb.FlushAll(context.Background())
}

func getIP(url string) string {
	s := strings.Split(url, "/")
	return s[len(s)-1]
}

func isInRedis(key string) (bool, error) {
	_, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
