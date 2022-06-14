package ipbase

import (
	"context"
	"fmt"
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

func TestNewNewIPBaseClient(t *testing.T) {
	type args struct {
		baseURL string
		apikey  string
		client  *http.Client
		logger  *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *IPBaseClient
	}{
		{
			name: "Empty base URL - empty apikey - empty http client",
			args: args{},
			want: &IPBaseClient{BaseURL: DefaultBaseURL, StatusURL: DefaultStatusURL},
		},
		{
			name: "Non empty base URL - empty apikey - empty http client",
			args: args{baseURL: "http://localhost:8080"},
			want: &IPBaseClient{BaseURL: "http://localhost:8080", StatusURL: DefaultStatusURL},
		},
		{
			name: "Non empty base URL - non empty apikey - empty http client",
			args: args{baseURL: "http://localhost:8080", apikey: "someapikey"},
			want: &IPBaseClient{BaseURL: "http://localhost:8080", StatusURL: DefaultStatusURL, apiKey: "someapikey"},
		},
		{
			name: "Non empty base URL - non empty apikey - non empty http client",
			args: args{baseURL: "http://localhost:8080", apikey: "someapikey", client: &http.Client{Timeout: 5 * time.Second}},
			want: &IPBaseClient{BaseURL: "http://localhost:8080", StatusURL: DefaultStatusURL, apiKey: "someapikey", Client: &http.Client{Timeout: 5 * time.Second}},
		},
		{
			name: "Empty base URL - empty apikey - non empty http client",
			args: args{client: &http.Client{Timeout: 5 * time.Second}},
			want: &IPBaseClient{BaseURL: DefaultBaseURL, StatusURL: DefaultStatusURL, Client: &http.Client{Timeout: 5 * time.Second}},
		},
		{
			name: "Empty base URL - non empty apikey - non empty http client",
			args: args{apikey: "someapikey", client: &http.Client{Timeout: 5 * time.Second}},
			want: &IPBaseClient{BaseURL: DefaultBaseURL, StatusURL: DefaultStatusURL, apiKey: "someapikey", Client: &http.Client{Timeout: 5 * time.Second}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIPBaseClient(tt.args.baseURL, tt.args.apikey, tt.args.client, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIPBaseClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPBaseClientGet(t *testing.T) {
	// Start local http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`ok`))
		case "/v2/info":
			q := r.URL.Query()

			switch q.Get("apikey") {
			case "":
				w.WriteHeader(401)
				w.Write([]byte(`Missing API key`))
			case "nil":
				// Says returned content is 50 but actuaklly send nil
				// resulting in ioutil.ReadAll() returning an error.
				w.Header().Add("Content-Length", "50")
				w.Write(nil)
			}

			ip := q.Get("ip")

			switch ip {
			case "1.1.1.1":
				w.Write([]byte(`{"data":{"timezone":{"id":"Australia/Sydney","current_time":"2022-04-25T18:56:38+10:00","code":"AEST","is_daylight_saving":false,"gmt_offset":36000},"ip":"1.1.1.1","type":"v4","connection":{"asn":13335,"organization":"CLOUDFLARENET","isp":"Cloudflare"},"location":{"geonames_id":2147714,"latitude":-33.86714172363281,"longitude":151.2071075439453,"zip":"2000","continent":{"code":"OC","name":"Oceania","name_translated":"Oceania"},"country":{"alpha2":"AU","alpha3":"AUS","calling_codes":["+61"],"currencies":[{"symbol":"AU$","name":"Australian Dollar","symbol_native":"$","decimal_digits":2,"rounding":0,"code":"AUD","name_plural":"Australian dollars"}],"emoji":"🇦🇺","ioc":"AUS","languages":[{"name":"English","name_native":"English"}],"name":"Australia","name_translated":"Australia","timezones":["Australia/Lord_Howe","Antarctica/Macquarie","Australia/Hobart","Australia/Currie","Australia/Melbourne","Australia/Sydney","Australia/Broken_Hill","Australia/Brisbane","Australia/Lindeman","Australia/Adelaide","Australia/Darwin","Australia/Perth","Australia/Eucla"],"is_in_european_union":false},"city":{"name":"Sydney","name_translated":"Sydney"},"region":{"fips":"AS-02","alpha2":"AU-NSW","name":"New South Wales","name_translated":"New South Wales"}}}}`))
			case "2.2.2.2":
				w.Write([]byte(`{"data":{"timezone":{"id":"Europe/Paris","current_time":"2022-06-13T16:08:58+02:00","code":"CEST","is_daylight_saving":true,"gmt_offset":7200},"ip":"2.2.2.2","type":"v4","connection":{"asn":3215,"organization":"Orange","isp":"Orange s.A."},"location":{"geonames_id":3036386,"latitude":48.91482162475586,"longitude":2.3812100887298584,"zip":"93300","continent":{"code":"EU","name":"Europe","name_translated":"Europe"},"country":{"alpha2":"FR","alpha3":"FRA","calling_codes":["+33"],"currencies":[{"symbol":"€","name":"Euro","symbol_native":"€","decimal_digits":2,"rounding":0,"code":"EUR","name_plural":"Euros"}],"emoji":"🇫🇷","ioc":"FRA","languages":[{"name":"French","name_native":"Français"}],"name":"France","name_translated":"France","timezones":["Europe/Paris"],"is_in_european_union":true},"city":{"name":"Aubervilliers","name_translated":"Aubervilliers"},"region":{"fips":"FR-11","alpha2":"FR-IDF","name":"Île-de-France","name_translated":"Île-de-France"}}}}`))
			case "3.3.3.3":
				w.Write([]byte(`thisisnotjson`))
			}

		case "/nil":
			// Says returned content is 50 but actuaklly send nil
			// resulting in ioutil.ReadAll() returning an error.
			w.Header().Add("Content-Length", "50")
			w.Write(nil)

		default:
			w.WriteHeader(404)
			w.Write([]byte(`ko`))
		}
	}))

	// Close the http server
	defer server.Close()

	t.Run("ipbase - /v2/info?ip=1.1.1.1", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPBaseClient(fmt.Sprintf("%s/v2/info?ip=", server.URL), "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "1.1.1.1")
		assert.NoError(t, err)
		assert.NotEmpty(t, g)
	})

	t.Run("ipbase - /v2/info?ip=2.2.2.2", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPBaseClient(fmt.Sprintf("%s/v2/info?ip=", server.URL), "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "2.2.2.2")
		assert.NoError(t, err)
		assert.NotEmpty(t, g)
	})

	t.Run("ipbase - /invalid-path", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPBaseClient(fmt.Sprintf("%s/invalid-path", server.URL), "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ipbase - /v2/info?ip=3.3.3.3", func(t *testing.T) {
		// Use Client & URL from the local test server
		c := NewIPBaseClient(fmt.Sprintf("%s/v2/info?ip=", server.URL), "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "3.3.3.3")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ipbase - invalid url - 01", func(t *testing.T) {
		// string([]byte{0x7f}) is a control character which will make NewRequestWithContext()
		// throw an error.
		c := NewIPBaseClient(string([]byte{0x7f}), "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "1.1.1.1")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ipbase - invalid url - 02", func(t *testing.T) {
		c := NewIPBaseClient("_invalidUrl_", "someapikey", server.Client(), &logger)
		g, err := c.Get(context.Background(), "1.1.1.1")
		assert.Error(t, err)
		assert.Empty(t, g)
	})

	t.Run("ipbase - /nil", func(t *testing.T) {
		c := NewIPBaseClient(fmt.Sprintf("%s/v2/info?ip=", server.URL), "nil", server.Client(), &logger)
		g, err := c.Get(context.Background(), "")
		fmt.Println(err)
		assert.Error(t, err)
		assert.Empty(t, g)
	})

}

func TestIPBaseClientStatus(t *testing.T) {
	// Start local http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`ok`))
		case "/v2/status":
			w.Write([]byte(`{"quotas":{"month":{"total":150,"used":2,"remaining":148}}}`))
		default:
			w.WriteHeader(404)
			w.Write([]byte(`ko`))
		}
	}))

	// Close the http server
	defer server.Close()

	var wg sync.WaitGroup

	wg.Add(4)

	t.Run("ipbase status - /v2/status", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPBaseClient(server.URL, "", server.Client(), &logger)
		c.StatusURL = fmt.Sprintf("%s/v2/status", server.URL)
		c.Status(context.Background(), &wg, ch)
		assert.NoError(t, <-ch)
	})

	t.Run("ipbase status - invalid url - 01", func(t *testing.T) {
		ch := make(chan error, 1)
		// string([]byte{0x7f}) is a control character which will make NewRequestWithContext()
		// throw an error.
		c := NewIPBaseClient(string([]byte{0x7f}), "", server.Client(), &logger)
		c.StatusURL = fmt.Sprintf("%s/v2/status", string([]byte{0x7f}))
		c.Status(context.Background(), &wg, ch)
		assert.Error(t, <-ch)
	})

	t.Run("ipbase status - invalid url - 02", func(t *testing.T) {
		ch := make(chan error, 1)
		c := NewIPBaseClient("_invalidUrl_", "", server.Client(), &logger)
		c.StatusURL = fmt.Sprintf("%s/v2/status", "_invalidUrl_")
		c.Status(context.Background(), &wg, ch)
		assert.Error(t, <-ch)
	})

	t.Run("ipbase status - /404", func(t *testing.T) {
		ch := make(chan error, 1)
		// Use Client & URL from the local test server
		c := NewIPBaseClient(server.URL, "", server.Client(), &logger)
		c.StatusURL = fmt.Sprintf("%s/404", server.URL)
		c.Status(context.Background(), &wg, ch)
		assert.Error(t, <-ch)
	})

	wg.Wait()
}
