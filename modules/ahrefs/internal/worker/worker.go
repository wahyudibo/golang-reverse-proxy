package worker

import (
	"context"

	"github.com/go-redis/redis/v9"
)

// Worker defines methods that standard worker should have
type Worker interface {
	Start()
	Stop()
}

func New(headlessCtx context.Context, cache *redis.Client) *Manager {
	return &Manager{
		HeadlessCtx: headlessCtx,
		Cache:       cache,
	}
}

// Manager manages and passes shared properties to all workers
type Manager struct {
	Cache       *redis.Client
	HeadlessCtx context.Context
	workers     []Worker
}

func (manager *Manager) add(workers ...Worker) {
	manager.workers = append(manager.workers, workers...)
}

func (manager *Manager) register() {
	loginWorker := &loginWorker{
		Name:        "LOGIN",
		Cache:       manager.Cache,
		HeadlessCtx: manager.HeadlessCtx,
		StopCh:      make(chan bool),
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
