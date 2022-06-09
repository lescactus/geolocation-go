package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/api/ipapi"
	"github.com/lescactus/geolocation-go/internal/config"
	"github.com/lescactus/geolocation-go/internal/controllers"
	"github.com/lescactus/geolocation-go/internal/logger"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/hlog"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	httpmetricsmiddleware "github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func main() {
	// Get application configuration
	cfg := config.New()

	// logger
	logger := logger.New(
		cfg.GetString("LOGGER_LOG_LEVEL"),
		cfg.GetString("LOGGER_DURATION_FIELD_UNIT"),
		cfg.GetString("LOGGER_FORMAT"),
	)

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
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient, logger)
	default:
		// Create ip-api client by default
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient, logger)
	}

	// Create http router, middleware manager and handler controller
	r := httprouter.New()
	h := controllers.NewBaseHandler(mdb, rdb, rApi, logger)
	c := alice.New()

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

	// Register logging middleware
	c = c.Append(hlog.NewHandler(*logger))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RemoteAddrHandler("remote_client"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RequestIDHandler("req_id", "X-Request-ID"))

	// Prometheus middleware
	if cfg.GetBool("PROMETHEUS") {
		p := httpmetricsmiddleware.New(httpmetricsmiddleware.Config{
			Recorder:               metrics.NewRecorder(metrics.Config{HandlerIDLabel: "path"}),
			DisableMeasureInflight: true,
			DisableMeasureSize:     true,
		})

		// Reduce cardinality by setting the handler id
		c = c.Append(std.HandlerProvider("/rest/v1/:ip", p))
		r.Handler("GET", cfg.GetString("PROMETHEUS_PATH"), c.Then(promhttp.Handler()))
	}

	// Register routes
	r.Handler("GET", "/rest/v1/:ip", c.ThenFunc(h.GetGeoIP))
	r.Handler("GET", "/ready", c.ThenFunc(h.Healthz))
	r.Handler("GET", "/alive", c.ThenFunc(h.Healthz))

	// 404 and 405 custom handlers with middlewares
	r.NotFound = c.ThenFunc(h.NotFoundHandler)
	r.MethodNotAllowed = c.ThenFunc(h.MethodNotAllowedHandler)

	// Start server
	logger.Info().Msg("Starting server ...")
	if err := s.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Startup failed")
	}
}
