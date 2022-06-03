package controllers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var (
	OneOneOneOne = models.GeoIP{
		IP:          "1.1.1.1",
		CountryCode: "AU",
		CountryName: "Australia",
		City:        "South Brisbane",
		Latitude:    -27.4766,
		Longitude:   153.0166,
	}
	TwoTwoTwoTwo = models.GeoIP{
		IP:          "2.2.2.2",
		CountryCode: "FR",
		CountryName: "France",
		City:        "Paris",
		Latitude:    48.8566,
		Longitude:   2.35222,
	}
	ThreeThreeThree = models.GeoIP{
		IP:          "3.3.3.3",
		CountryCode: "US",
		CountryName: "United States",
		City:        "Chicago",
		Latitude:    41.8781,
		Longitude:   -87.6298,
	}
	muxRoute = "/ip/{ip}"

	logger = zerolog.New(os.Stdout).Level(zerolog.NoLevel)
)

func init() {
	// Make the logger silent
	log.SetOutput(ioutil.Discard)
}

// RedisMock implements models.GeoIPRepository
type RedisMock struct{}

func (m *RedisMock) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	if ip == "1.1.1.1" {
		return &OneOneOneOne, nil
	} else {
		return nil, fmt.Errorf("error: no value found for key %s", ip)
	}
}

func (m *RedisMock) Save(ctx context.Context, geoip *models.GeoIP) error { return nil }

func (m *RedisMock) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {}

// GeoAPIMock implements api.GeoIP
type GeoAPIMock struct{}

func (m *GeoAPIMock) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	switch ip {
	case "1.1.1.1":
		return &OneOneOneOne, nil
	case "2.2.2.2":
		return &TwoTwoTwoTwo, nil
	case "3.3.3.3":
		return &ThreeThreeThree, nil
	default:
		return nil, fmt.Errorf("error: error while fetch geo information for %s", ip)
	}
}

func (m *GeoAPIMock) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {}

func TestGetGeoIP(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []byte
		code int
	}{
		{
			name: "valid path - /rest/v1/1.1.1.1",
			path: "/rest/v1/1.1.1.1",
			want: []byte(`{"ip":"1.1.1.1","country_code":"AU","country_name":"Australia","city":"South Brisbane","latitude":-27.4766,"longitude":153.0166}`),
			code: 200,
		},
		{
			name: "valid path - /rest/v1/2.2.2.2",
			path: "/rest/v1/2.2.2.2",
			want: []byte(`{"ip":"2.2.2.2","country_code":"FR","country_name":"France","city":"Paris","latitude":48.8566,"longitude":2.35222}`),
			code: 200,
		},
		{
			name: "valid path - /rest/v1/3.3.3.3",
			path: "/rest/v1/3.3.3.3",
			want: []byte(`{"ip":"3.3.3.3","country_code":"US","country_name":"United States","city":"Chicago","latitude":41.8781,"longitude":-87.6298}`),
			code: 200,
		},
		{
			name: "valid path - /rest/v1/4.4.4.4",
			path: "/rest/v1/4.4.4.4",
			want: []byte(`error: couldn't retrieve geo IP information`),
			code: 500,
		},
		{
			name: "invalid path - /rest/v1/bla",
			path: "/rest/v1/bla",
			want: []byte(`error: the provided IP is not a valid IPv4 address`),
			code: 400,
		},
	}

	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	a := &GeoAPIMock{}

	// route registration
	h := NewBaseHandler(mdb, rdb, a, &logger)
	r.HandleFunc("/rest/v1/{ip}", h.GetGeoIP).Methods("GET")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			resp := recorder.Result()
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, data)
			assert.Equal(t, tt.code, resp.StatusCode)
		})
	}
}

func BenchmarkGetGeoIP_EntryNotInMemoryDB_EntryInRedis(b *testing.B) {
	route := "/ip/1.1.1.1"
	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}

	// route registration
	h := NewBaseHandler(mdb, rdb, nil, &logger)
	r.HandleFunc(muxRoute, h.GetGeoIP).Methods("GET")

	req, _ := http.NewRequest("GET", route, nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(recorder, req)
	}
}

func BenchmarkGetGeoIP_EntryNotInMemoryDB_EntryNotInRedis_EntryNotFoundByRemoteAPI(b *testing.B) {
	route := "/ip/4.4.4.4"
	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	a := &GeoAPIMock{}

	// route registration
	h := NewBaseHandler(mdb, rdb, a, &logger)
	r.HandleFunc(muxRoute, h.GetGeoIP).Methods("GET")

	req, _ := http.NewRequest("GET", route, nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(recorder, req)
	}
}

func BenchmarkGetGeoIP_EntryNotInMemoryDB_EntryNotInRedis_EntryFoundByRemoteAPI(b *testing.B) {
	route := "/ip/3.3.3.3"
	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	a := &GeoAPIMock{}

	// route registration
	h := NewBaseHandler(mdb, rdb, a, &logger)
	r.HandleFunc(muxRoute, h.GetGeoIP).Methods("GET")

	req, _ := http.NewRequest("GET", route, nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(recorder, req)
	}
}

func BenchmarkGetGeoIP_EntryInMemoryDB_EntryInRedis(b *testing.B) {
	route := "/ip/1.1.1.1"
	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}
	mdb.Save(context.Background(), &OneOneOneOne)

	// route registration
	h := NewBaseHandler(mdb, rdb, nil, &logger)
	r.HandleFunc(muxRoute, h.GetGeoIP).Methods("GET")

	req, _ := http.NewRequest("GET", route, nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(recorder, req)
	}
}
