package ipapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var logger = zerolog.New(os.Stdout).Level(zerolog.NoLevel)

func TestNewIPAPIClient(t *testing.T) {
	type args struct {
		baseURL string
		client  *http.Client
		logger  *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *IPAPIClient
	}{
		{
			name: "Empty base URL - empty http client",
			args: args{},
			want: &IPAPIClient{BaseURL: DefaultBaseURL},
		},
		{
			name: "Non empty base URL - empty http client",
			args: args{baseURL: "http://localhost:8080"},
			want: &IPAPIClient{BaseURL: "http://localhost:8080"},
		},
		{
			name: "Non empty base URL - non empty http client",
			args: args{baseURL: "http://localhost:8080", client: &http.Client{Timeout: 5 * time.Second}},
			want: &IPAPIClient{BaseURL: "http://localhost:8080", Client: &http.Client{Timeout: 5 * time.Second}},
		},
		{
			name: "Empty base URL - non empty http client",
			args: args{client: &http.Client{Timeout: 5 * time.Second}},
			want: &IPAPIClient{BaseURL: DefaultBaseURL, Client: &http.Client{Timeout: 5 * time.Second}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIPAPIClient(tt.args.baseURL, tt.args.client, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIPAPIClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPAPIClientGet(t *testing.T) {
	// Start local http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/1.1.1.1":
			w.Write([]byte(`{"status":"success","country":"Australia","countryCode":"AU","region":"QLD","regionName":"Queensland","city":"South Brisbane","zip":"4101","lat":-27.4766,"lon":153.0166,"timezone":"Australia/Brisbane","isp":"Cloudflare, Inc","org":"APNIC and Cloudflare DNS Resolver project","as":"AS13335 Cloudflare, Inc.","query":"1.1.1.1"}`))
		case "/2.2.2.2":
			w.Write([]byte(`{"status":"success","country":"France","countryCode":"FR","region":"IDF","regionName":"Île-de-France","city":"Paris","zip":"75000","lat":48.8566,"lon":2.35222,"timezone":"Europe/Paris","isp":"France Telecom Orange","org":"","as":"AS3215 Orange S.A.","query":"2.2.2.2"}`))
		case "/3.3.3.3":
			w.Write([]byte(`thisisnotjson`))
		case "/4.4.4.4":
			// Says returned content is 50 but actuaklly send nil
			// resulting in ioutil.ReadAll() returning an error.
			w.Header().Add("Content-Length", "50")
			w.Write(nil)
		default:
			w.WriteHeader(404)
		}
	}))

	// Close the http server
	defer server.Close()

	t.Run("ip-api - /1.1.1.1", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		g, err := c.Get(context.Background(), "/1.1.1.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, g)
	})

	t.Run("ip-api - /2.2.2.2", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		g, err := c.Get(context.Background(), "/2.2.2.2")
		assert.NoError(t, err)
		assert.NotEmpty(t, g)
	})

	t.Run("ip-api - /invalid-path", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		g, err := c.Get(context.Background(), "/invalid-path")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ip-api - /3.3.3.3", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		g, err := c.Get(context.Background(), "/3.3.3.3")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ip-api - invalid url - 01", func(t *testing.T) {
		// string([]byte{0x7f}) is a control character which will make NewRequestWithContext()
		// throw an error.
		c := NewIPAPIClient(string([]byte{0x7f}), server.Client(), &logger)
		g, err := c.Get(context.Background(), "/")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ip-api - invalid url - 02", func(t *testing.T) {
		c := NewIPAPIClient("_invalidUrl_", server.Client(), &logger)
		g, err := c.Get(context.Background(), "/")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ip-api - /4.4.4.4", func(t *testing.T) {
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		g, err := c.Get(context.Background(), "/4.4.4.4")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

}

func TestIPAPIClientStatus(t *testing.T) {
	// Start local http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`ok`))
		case "/1.1.1.1":
			w.Write([]byte(`{"status":"success","country":"Australia","countryCode":"AU","region":"QLD","regionName":"Queensland","city":"South Brisbane","zip":"4101","lat":-27.4766,"lon":153.0166,"timezone":"Australia/Brisbane","isp":"Cloudflare, Inc","org":"APNIC and Cloudflare DNS Resolver project","as":"AS13335 Cloudflare, Inc.","query":"1.1.1.1"}`))
		case "/2.2.2.2":
			w.Write([]byte(`{"status":"success","country":"France","countryCode":"FR","region":"IDF","regionName":"Île-de-France","city":"Paris","zip":"75000","lat":48.8566,"lon":2.35222,"timezone":"Europe/Paris","isp":"France Telecom Orange","org":"","as":"AS3215 Orange S.A.","query":"2.2.2.2"}`))
		default:
			w.WriteHeader(404)
			w.Write([]byte(`ko`))
		}
	}))

	// Close the http server
	defer server.Close()

	var wg sync.WaitGroup

	wg.Add(4)

	t.Run("ip-api status - /", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL, server.Client(), &logger)
		c.Status(context.Background(), &wg, ch)
		assert.NoError(t, <-ch)
	})

	t.Run("ip-api status - /invalid-path", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL+"/invalid-path", server.Client(), &logger)
		c.Status(context.Background(), &wg, ch)
		assert.Error(t, <-ch)
	})

	t.Run("ip-api status - invalid url", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPAPIClient("_invalidUrl_", server.Client(), &logger)
		c.Status(context.Background(), &wg, ch)
		assert.Error(t, <-ch)
	})

	t.Run("ip-api status - nil context", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPAPIClient(server.URL+"", server.Client(), &logger)
		c.Status(nil, &wg, ch)
		assert.Error(t, <-ch)
	})

	wg.Wait()
}
