package worker

import (
	"github.com/go-redis/redis/v9"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/headless"
)

// Worker defines methods that standard worker should have
type Worker interface {
	Start()
	Stop()
}

func New(cfg *config.Config, headlessBrowser *headless.HeadlessBrowser, cache *redis.Client) *Manager {
	return &Manager{
		Config:          cfg,
		Cache:           cache,
		HeadlessBrowser: headlessBrowser,
	}
}

// Manager manages and passes shared properties to all workers
type Manager struct {
	Config          *config.Config
	Cache           *redis.Client
	HeadlessBrowser *headless.HeadlessBrowser
	workers         []Worker
}

func (manager *Manager) add(workers ...Worker) {
	manager.workers = append(manager.workers, workers...)
}

func (manager *Manager) register() {
	loginWorker := &loginWorker{
		Name:            "AHX:LOGIN",
		Config:          manager.Config,
		Cache:           manager.Cache,
		HeadlessBrowser: manager.HeadlessBrowser,
		StopCh:          make(chan bool),
	}

	manager.add(loginWorker)
}

// StartAll starts all worker in different goroutines
func (manager *Manager) StartAll() {
	manager.register()
	for _, worker := range manager.workers {
		go worker.Start()
	}
}

// StopAll stops all worker in different goroutines
func (manager *Manager) StopAll() {
	for _, worker := range manager.workers {
		worker.Stop()
	}
}
