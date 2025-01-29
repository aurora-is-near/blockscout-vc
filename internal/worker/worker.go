package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"blockscout-vc/internal/docker"
)

// Job represents a container recreation task with one or more containers
type Job struct {
	Containers []docker.Container
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
// Returns false if the job is already in queue or if containers is empty
// Returns true if the job was successfully added
func (w *Worker) AddJob(containers []docker.Container) bool {
	if len(containers) == 0 {
		return false
	}

	w.jobSetMux.Lock()
	defer w.jobSetMux.Unlock()

	key := w.makeKey(containers)
	if _, exists := w.jobSet[key]; exists {
		return false
	}

	w.jobSet[key] = struct{}{}
	w.jobs <- Job{Containers: containers}
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
			jobKey := w.makeKey(job.Containers)
			// Using an immediately invoked function to ensure defer runs after each job
			func() {
				defer func() {
					w.jobSetMux.Lock()
					delete(w.jobSet, jobKey)
					w.jobSetMux.Unlock()
				}()

				err := w.docker.RecreateContainers(job.Containers)
				if err != nil {
					log.Printf("failed to recreate containers: %v", err)
					return
				}

				// Add mandatory delay after successful recreation
				log.Printf("Container recreation completed, waiting 10 seconds before next job...")
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
					// Continue to next job after delay
				}
			}()
		}
	}
}

// makeKey creates a unique string key for a set of container names
// Uses docker.UniqueContainerNames to handle container name normalization
func (w *Worker) makeKey(containers []docker.Container) string {
	unique := w.docker.UniqueContainers(containers)
	return strings.Join(w.docker.GetContainerNames(unique), ",")
}
