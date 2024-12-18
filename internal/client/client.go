// Package client provides WebSocket client functionality for Supabase Realtime
package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection to Supabase Realtime
type Client struct {
	apiKey   string
	endpoint string
	handlers map[string]func([]byte)
	Conn     *websocket.Conn // Public connection instance for external use
}

// New creates a new WebSocket client with the specified endpoint and API key
func New(endpoint, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		handlers: make(map[string]func([]byte)),
		Conn:     nil,
	}
}

// Connect establishes a WebSocket connection to the Supabase Realtime server
// It configures the connection with the necessary headers and authentication
func (c *Client) Connect() error {
	header := http.Header{}
	header.Add("Authorization", "Bearer "+c.apiKey)

	dialer := websocket.Dialer{
		EnableCompression: true,
	}

	conn, resp, err := dialer.Dial(c.endpoint+"?apikey="+c.apiKey, header)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP Response Status: %s", resp.Status)
			log.Printf("HTTP Response Headers: %v", resp.Header)
		}
		log.Fatalf("Failed to connect to Realtime server: %v", err)
	}
	c.Conn = conn

	fmt.Println("Connected to Supabase Realtime!")
	return nil
}

// Close terminates the WebSocket connection
func (c *Client) Close() error {
	return c.Conn.Close()
}
