package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/adapter/db/mysql"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/router"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/router/proxy/app"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/router/proxy/static"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/worker"
	cacheClient "github.com/wahyudibo/golang-reverse-proxy/pkg/adapter/cache/redis"
	dbClient "github.com/wahyudibo/golang-reverse-proxy/pkg/adapter/db/mysql"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/headless"
)

func main() {
	ctx := context.Background()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate config")
	}

	headlessBrowser, err := headless.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate headless browser context")
	}
	defer headlessBrowser.Close()

	db, err := dbClient.New(ctx, &dbClient.Config{
		DBMySQLHost:     cfg.DBMySQLHost,
		DBMySQLPort:     cfg.DBMySQLPort,
		DBMySQLUser:     cfg.DBMySQLUser,
		DBMySQLPassword: cfg.DBMySQLPassword,
		DBMySQLDatabase: cfg.DBMySQLDatabase,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate database")
	}

	repo := dbmysql.New(db)

	cache, err := cacheClient.New(ctx, &cacheClient.Config{
		Address:  cfg.CacheRedisAddress,
		Password: cfg.CacheRedisPassword,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate cache")
	}

	worker := worker.New(cfg, headlessBrowser, cache)
	// start all workers
	worker.StartAll()
	defer worker.StopAll()

	// define route handler
	usageLimitRouteHandler := router.NewUsageLimitRouteHandler(repo)

	staticProxy, err := static.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate static proxy router")
	}
	appProxy, err := app.New(ctx, cfg, repo, cache)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate app proxy router")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	r.HandleFunc("/usage-limit", usageLimitRouteHandler.GetUsageLimit)
	r.HandleFunc("/ahx-static/*", staticProxy.Handler())
	r.HandleFunc("/*", appProxy.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.ProxyServerPort),
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	log.Info().Msgf("server started on port: %d", cfg.ProxyServerPort)

	<-stop
	log.Info().Msg("Receiving stop signal. Stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ProxyServerShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Msg("failed to shutdown server")
	}
	log.Info().Msg("server stopped")
}
