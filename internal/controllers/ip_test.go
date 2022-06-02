package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/lescactus/geolocation-go/internal/repositories"
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
)

func init () {
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

func BenchmarkGetGeoIP_EntryNotInMemoryDB_EntryInRedis(b *testing.B) {
	route := "/ip/1.1.1.1"
	r := mux.NewRouter()

	// db
	mdb := repositories.NewInMemoryDB()
	rdb := &RedisMock{}

	// route regstration
	h := NewBaseHandler(mdb, rdb, nil)
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

	// route regstration
	h := NewBaseHandler(mdb, rdb, a)
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

	// route regstration
	h := NewBaseHandler(mdb, rdb, a)
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

	// route regstration
	h := NewBaseHandler(mdb, rdb, nil)
	r.HandleFunc(muxRoute, h.GetGeoIP).Methods("GET")

	req, _ := http.NewRequest("GET", route, nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(recorder, req)
	}
}
