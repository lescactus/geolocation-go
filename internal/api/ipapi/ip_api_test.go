package ipapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIPAPIClient(t *testing.T) {
	type args struct {
		baseURL string
		client  *http.Client
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
			if got := NewIPAPIClient(tt.args.baseURL, tt.args.client); !reflect.DeepEqual(got, tt.want) {
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
		default:
			w.WriteHeader(404)
		}
	}))

	// Close the http server
	defer server.Close()

	// Use Client & URL from the local test server
	c1 := NewIPAPIClient(server.URL, server.Client())
	g1, err := c1.Get(context.Background(), "/1.1.1.1")
	assert.NoError(t, err)
	assert.NotEmpty(t, g1)

	// Use Client & URL from the local test server
	c2 := NewIPAPIClient(server.URL, server.Client())
	g2, err := c2.Get(context.Background(), "/2.2.2.2")
	assert.NoError(t, err)
	assert.NotEmpty(t, g2)

	// Use Client & URL from the local test server
	c3 := NewIPAPIClient(server.URL, server.Client())
	g3, err := c3.Get(context.Background(), "/invalid-path")
	assert.Error(t, err)
	assert.Empty(t, g3)
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
	ch1 := make(chan error, 1)
	ch2 := make(chan error, 1)
	ch3 := make(chan error, 1)

	wg.Add(3)
	// Use Client & URL from the local test server
	c1 := NewIPAPIClient(server.URL, server.Client())
	c1.Status(context.TODO(), &wg, ch1)
	assert.NoError(t, <-ch1)

	// Use Client & URL from the local test server
	c2 := NewIPAPIClient(server.URL+"/invalid-path", server.Client())
	c2.Status(context.TODO(), &wg, ch2)
	assert.Error(t, <-ch2)

	// Use Client & URL from the local test server
	c3 := NewIPAPIClient(server.URL+"", server.Client())
	c3.Status(nil, &wg, ch3)
	assert.Error(t, <-ch3)
	wg.Wait()
}
