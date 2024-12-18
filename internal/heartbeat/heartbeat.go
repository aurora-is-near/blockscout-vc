// Package heartbeat provides functionality for maintaining WebSocket connection health
package heartbeat

import (
	"blockscout-vc/internal/client"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type HeartbeatService struct {
	client   *client.Client
	interval time.Duration
	stopChan chan struct{}
}

type HeartbeatPayload struct {
	Event   string                 `json:"event"`
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
	Ref     string                 `json:"ref"`
}

func New(client *client.Client, interval time.Duration) *HeartbeatService {
	return &HeartbeatService{
		client:   client,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// sendHeartbeat sends a single heartbeat message through the WebSocket connection
func sendHeartbeat(conn *websocket.Conn) error {
	heartbeat := HeartbeatPayload{
		Event:   "heartbeat",
		Topic:   "phoenix",
		Payload: map[string]interface{}{},
		Ref:     uuid.New().String(),
	}
	return conn.WriteJSON(heartbeat)
}

// Start begins sending periodic heartbeat messages
func (h *HeartbeatService) Start() {
	ticker := time.NewTicker(h.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := sendHeartbeat(h.client.Conn); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				}
			case <-h.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop terminates the heartbeat service
func (h *HeartbeatService) Stop() {
	close(h.stopChan)
}
