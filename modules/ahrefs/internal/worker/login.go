package worker

import (
	"context"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/headless/task"
)

var _ Worker = (*loginWorker)(nil)

type loginWorker struct {
	Name        string
	Config      *config.Config
	Cache       *redis.Client
	HeadlessCtx context.Context
	StopCh      chan bool
}

func nextInterval(min, max time.Duration) time.Duration {
	maxNs := max.Nanoseconds()
	minNs := min.Nanoseconds()

	interval := rand.Int63n(maxNs-minNs) + minNs
	return time.Duration(interval) * time.Nanosecond
}

func (w *loginWorker) Start() {
	log.Info().Msgf("[WORKER: %s] started", w.Name)

	rand.Seed(time.Now().UTC().UnixNano())

	min := 30 * time.Second
	max := 60 * time.Second

	ticker := time.NewTicker(min)

	for {
		select {
		case <-w.StopCh:
			ticker.Stop()
			log.Info().Msgf("[WORKER: %s] receiving stop signal", w.Name)
			return
		case <-ticker.C:
			ticker.Stop()
			ticker = time.NewTicker(nextInterval(min, max))

			timeoutCtx, timeoutCtxCancel := context.WithTimeout(w.HeadlessCtx, 2*time.Second)
			defer timeoutCtxCancel()

			if err := chromedp.Run(timeoutCtx, task.Ping()); err != nil {
				log.Error().Err(err).Msgf("[WORKER: %s] failed to run task.Ping()", w.Name)

				loginTimeoutCtx, loginTimeoutCtxCancel := context.WithTimeout(w.HeadlessCtx, 5*time.Second)
				defer loginTimeoutCtxCancel()

				var cookies []*network.Cookie
				if err := chromedp.Run(loginTimeoutCtx, task.Login(w.Config.AccountUsername, w.Config.AccountPassword, cookies)); err != nil {
					log.Error().Err(err).Msgf("[WORKER: %s] failed to run task.Login()", w.Name)
				}
			}
		}
	}
}

func (w *loginWorker) Stop() {
	close(w.StopCh)
	log.Info().Msgf("[WORKER: %s] stopped", w.Name)
}
