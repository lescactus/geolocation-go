package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/api/ipapi"
	"github.com/lescactus/geolocation-go/internal/config"
	"github.com/lescactus/geolocation-go/internal/controllers"
	"github.com/lescactus/geolocation-go/internal/repositories"
)

func main() {
	// Get application configuration
	cfg := config.New()

	// Create in-memory database
	mdb := repositories.NewInMemoryDB()

	// Create redis database client
	rdb, err := repositories.NewRedisDB(cfg.GetString("REDIS_CONNECTION_STRING"))
	if err != nil {
		log.Fatalln(err)
	}

	// Create http client
	httpClient := http.DefaultClient
	httpClient.Timeout = cfg.GetDuration("HTTP_CLIENT_TIMEOUT")

	// Create remote Geo IP API client
	var rApi api.GeoAPI

	switch cfg.GetString("GEOLOCATION_API") {
	case "ip-api":
		// Create ip-api client
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient)
	default:
		// Create ip-api client by default
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient)
	}

	// Create mux router and handler controller
	r := mux.NewRouter()
	h := controllers.NewBaseHandler(mdb, rdb, rApi)

	// Create http server
	s := &http.Server{
		Addr:              cfg.GetString("APP_ADDR"),
		Handler:           r,
		ReadTimeout:       cfg.GetDuration("SERVER_READ_TIMEOUT"),
		ReadHeaderTimeout: cfg.GetDuration("SERVER_READ_HEADER_TIMEOUT"),
		WriteTimeout:      cfg.GetDuration("SERVER_WRITE_TIMEOUT"),
	}

	r.Handle("/rest/v1/{ip}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.GetGeoIP))).Methods("GET")
	r.Handle("/ready", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.Healthz))).Methods("GET")
	r.Handle("/alive", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.Healthz))).Methods("GET")
	log.Println("Starting server ...")
	log.Fatal(s.ListenAndServe())
}
