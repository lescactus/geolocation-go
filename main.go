package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/geolocation-go/internal/api"
	"github.com/lescactus/geolocation-go/internal/api/ipapi"
	"github.com/lescactus/geolocation-go/internal/api/ipbase"
	"github.com/lescactus/geolocation-go/internal/chain"
	"github.com/lescactus/geolocation-go/internal/config"
	"github.com/lescactus/geolocation-go/internal/controllers"
	"github.com/lescactus/geolocation-go/internal/logger"
	"github.com/lescactus/geolocation-go/internal/repositories"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
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
	rdb, err := repositories.NewRedisDB(
		cfg.GetString("REDIS_CONNECTION_STRING"),
		cfg.GetDuration("REDIS_KEY_TTL"))
	if err != nil {
		log.Fatalln(err)
	}

	// Create the cacher chain
	chain := chain.New(logger)
	chain.Add("in-memory", mdb)
	chain.Add("redis", rdb)

	// Create http client
	httpClient := http.DefaultClient
	httpClient.Timeout = cfg.GetDuration("HTTP_CLIENT_TIMEOUT")

	// Create remote Geo IP API client
	var rApi api.GeoAPI

	switch cfg.GetString("GEOLOCATION_API") {
	case "ip-api":
		// Create ip-api client
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient, logger)
	case "ipbase":
		// Create ipbase client
		rApi = ipbase.NewIPBaseClient(cfg.GetString("IPBASE_BASE_URL"), cfg.GetString("IPBASE_API_KEY"), httpClient, logger)
	default:
		// Create ip-api client by default
		rApi = ipapi.NewIPAPIClient(cfg.GetString("IP_API_BASE_URL"), httpClient, logger)
	}

	// Create http router, middleware manager and handler controller
	r := httprouter.New()
	h := controllers.NewBaseHandler(chain, rApi, logger)
	h.CacheChain = chain
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
			if err := http.ListenAndServe(":6060", nil); err != nil {
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

	// OPTIONS method, 404 and 405 custom handlers with middlewares
	r.NotFound = c.ThenFunc(h.NotFoundHandler)
	r.MethodNotAllowed = c.ThenFunc(h.MethodNotAllowedHandler)
	r.GlobalOPTIONS = c.ThenFunc(h.OptionsHandler)

	// logger fields
	*logger = logger.With().Str("svc", config.AppName).Logger()

	// Start server
	go func() {
		logger.Info().Msgf("Starting server on address %s ...", cfg.GetString("APP_ADDR"))
		logger.Debug().
			Dict("config", zerolog.Dict(). // Don't show redis connection string in case password is provided
							Str("geolocation_api", cfg.GetString("GEOLOCATION_API")).
							Dict("server_config", zerolog.Dict().
								Dur("server_read_timeout", cfg.GetDuration("SERVER_READ_TIMEOUT")).
								Dur("server_read_header_timeout", cfg.GetDuration("SERVER_READ_HEADER_TIMEOUT")).
								Dur("server_write_timeout", cfg.GetDuration("SERVER_WRITE_TIMEOUT")),
				).
				Dict("logger_config", zerolog.Dict().
					Str("log_level", cfg.GetString("LOGGER_LOG_LEVEL")).
					Str("log_format", cfg.GetString("LOGGER_FORMAT")).
					Str("logger_duration_field_unit", cfg.GetString("LOGGER_DURATION_FIELD_UNIT")),
				).
				Dict("prometheus_config", zerolog.Dict().
					Bool("prometheus_enabled", cfg.GetBool("PROMETHEUS")).
					Str("prometheus_path", cfg.GetString("PROMETHEUS_PATH")),
				).
				Bool("pprof_enabled", cfg.GetBool("PPROF")),
			).
			Msg("With config")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Startup failed")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Blocking until receiving a shutdown signal
	sig := <-sigChan

	logger.Info().Msgf("Server received %s signal. Shutting down...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	// Attempting to gracefully shutdown the server
	if err := s.Shutdown(ctx); err != nil {
		logger.Warn().Msg("Failed to gracefully shutdown the server")
	}
}
