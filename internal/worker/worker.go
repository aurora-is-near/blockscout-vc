package worker

import (
	"context"
	"log"
	"strings"
	"sync"

	"blockscout-vc/internal/docker"
)

// Job represents a container recreation task with one or more containers
type Job struct {
	ContainerNames []string
}

// Worker manages a queue of container recreation jobs,
// ensuring sequential processing and preventing duplicate jobs
type Worker struct {
	docker    *docker.Docker
	jobs      chan Job            // Buffered channel for job queue
	jobSet    map[string]struct{} // Set of unique jobs currently in queue
	jobSetMux sync.Mutex          // Mutex to protect the job set
}

// New creates a new Worker instance with a job buffer of 100
func New() *Worker {
	return &Worker{
		docker:    docker.NewDocker(),
		jobs:      make(chan Job, 100),
		jobSet:    make(map[string]struct{}),
		jobSetMux: sync.Mutex{},
	}
}

// Start begins processing jobs in a separate goroutine
// The worker will continue until the context is cancelled
func (w *Worker) Start(ctx context.Context) {
	go w.process(ctx)
}

// AddJob adds a new container recreation job to the queue
// Returns false if the job is already in queue or if containerNames is empty
// Returns true if the job was successfully added
func (w *Worker) AddJob(containerNames []string) bool {
	if len(containerNames) == 0 {
		return false
	}

	w.jobSetMux.Lock()
	defer w.jobSetMux.Unlock()

	key := w.makeKey(containerNames)
	if _, exists := w.jobSet[key]; exists {
		return false
	}

	w.jobSet[key] = struct{}{}
	w.jobs <- Job{ContainerNames: containerNames}
	return true
}

// process is the main job processing loop
// It handles one job at a time and removes completed jobs from the set
func (w *Worker) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobs:
			err := w.docker.RecreateContainers(job.ContainerNames)
			if err != nil {
				log.Printf("failed to recreate containers: %v", err)
				continue
			}

			w.jobSetMux.Lock()
			delete(w.jobSet, w.makeKey(job.ContainerNames))
			w.jobSetMux.Unlock()
		}
	}
}

// makeKey creates a unique string key for a set of container names
// Uses docker.UniqueContainerNames to handle container name normalization
func (w *Worker) makeKey(containerNames []string) string {
	unique := w.docker.UniqueContainerNames(containerNames)
	return strings.Join(unique, ",")
}
