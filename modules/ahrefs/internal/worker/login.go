package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"

	redisClient "github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/adapter/cache/redis"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/headless/task"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/headless"
)

var _ Worker = (*loginWorker)(nil)

type loginWorker struct {
	Name            string
	Config          *config.Config
	Cache           *redis.Client
	HeadlessBrowser *headless.HeadlessBrowser
	StopCh          chan bool
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

			var requireLogin bool

			pingCtx, pingCtxCancel := context.WithTimeout(w.HeadlessBrowser.Context, 2*time.Second)
			defer pingCtxCancel()

			if err := chromedp.Run(pingCtx, task.Ping()); err != nil {
				log.Error().Err(err).Msgf("[WORKER: %s] failed to run task.Ping(). Might required re-login", w.Name)

				requireLogin = true
			}

			if requireLogin {
				cookies := make([]*network.Cookie, 0)
				if err := chromedp.Run(w.HeadlessBrowser.Context, task.Login(w.Config.AccountUsername, w.Config.AccountPassword, &cookies)); err != nil {
					log.Error().Err(err).Msgf("[WORKER: %s] failed to run task.Login()", w.Name)
				}

				cookiesJSON, err := json.Marshal(cookies)
				if err != nil {
					log.Error().Err(err).Msgf("[WORKER: %s] failed to marshal cookies to JSON", w.Name)
				}
				w.Cache.Set(context.Background(), fmt.Sprintf("%s:cookies", redisClient.Prefix), string(cookiesJSON), 0)
			}
		}
	}
}

func (w *loginWorker) Stop() {
	close(w.StopCh)
	log.Info().Msgf("[WORKER: %s] stopped", w.Name)
}
