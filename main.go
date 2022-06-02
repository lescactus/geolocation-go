package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/api/ipapi"
	"github.com/lescactus/geolocation-go/internal/config"
	"github.com/lescactus/geolocation-go/internal/controllers"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func main() {
	// Get application configuration
	cfg := config.New()

	// logger
	zerolog.DurationFieldUnit = time.Second
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

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
	h := controllers.NewBaseHandler(mdb, rdb, rApi, &logger)

	// Create http server
	s := &http.Server{
		Addr:              cfg.GetString("APP_ADDR"),
		Handler:           handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r), // recover from panics and print recovery stack
		ReadTimeout:       cfg.GetDuration("SERVER_READ_TIMEOUT"),
		ReadHeaderTimeout: cfg.GetDuration("SERVER_READ_HEADER_TIMEOUT"),
		WriteTimeout:      cfg.GetDuration("SERVER_WRITE_TIMEOUT"),
	}

	// pprof
	if cfg.GetBool("PPROF") {
		logger.Info().Msg("Starting pprof server ...")
		// start pprof server
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				logger.Fatal().Err(err).Msg("Startup pprof server failed")
			}
		}()
	}

	// Register middleware
	r.Use(hlog.NewHandler(logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	r.Use(hlog.RefererHandler("referer"))
	r.Use(hlog.RemoteAddrHandler("remote_client"))
	r.Use(hlog.UserAgentHandler("user_agent"))
	r.Use(hlog.RequestIDHandler("req_id", "Request-ID"))

	// Register routes
	r.HandleFunc("/rest/v1/{ip}", h.GetGeoIP).Methods("GET")
	r.HandleFunc("/ready", h.Healthz).Methods("GET")
	r.HandleFunc("/alive", h.Healthz).Methods("GET")

	// Start server
	logger.Info().Msg("Starting server ...")
	if err := s.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Startup failed")
	}
}
