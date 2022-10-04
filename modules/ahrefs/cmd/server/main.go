package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/router/proxy/root"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/router/proxy/static"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	cfg := new(config.Config)
	if err := cfg.ParseEnvVars(); err != nil {
		log.Fatal().Err(err).Msg("failed to parse environment variable")
	}

	// define route handler
	rootProxy, err := root.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate root proxy router")
	}
	staticProxy, err := static.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate static proxy router")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	r.HandleFunc("/ahx-static/*", staticProxy.Handler())
	r.HandleFunc("/*", rootProxy.Handler())

	port := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Fatal().Err(http.ListenAndServe(port, r)).Msg("failed to start server")
}
