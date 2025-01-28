package subscription

import (
	"blockscout-vc/internal/client"
	"blockscout-vc/internal/docker"
	"blockscout-vc/internal/handlers"
	"blockscout-vc/internal/worker"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// Package subscription handles real-time database changes and container updates
type Subscription struct {
	client *client.Client
}

// PostgresChange represents a single database change subscription configuration
type PostgresChange struct {
	Event  string `json:"event"`
	Schema string `json:"schema"`
	Table  string `json:"table"`
	Filter string `json:"filter,omitempty"`
}

// SubscriptionPayload is the message sent to establish a real-time connection
type SubscriptionPayload struct {
	Event   string `json:"event"`
	Topic   string `json:"topic"`
	Payload struct {
		Config struct {
			Broadcast struct {
				Self bool `json:"self"`
			} `json:"broadcast"`
			PostgresChanges []PostgresChange `json:"postgres_changes"`
		} `json:"config"`
	} `json:"payload"`
	Ref string `json:"ref"`
}

// PostgresChanges represents a database change event received from Supabase
type PostgresChanges struct {
	Event   string `json:"event"`
	Payload struct {
		Data struct {
			Table  string          `json:"table"`
			Type   string          `json:"type"`
			Record handlers.Record `json:"record"`
		} `json:"data"`
	} `json:"payload"`
	Worker *worker.Worker
}

// New creates a new Subscription instance
func New(client *client.Client) *Subscription {
	return &Subscription{
		client: client,
	}
}

// Subscribe starts listening for database changes and handles container updates
func (s *Subscription) Subscribe(worker *worker.Worker) error {
	// Run initial check first to handle existing records
	if err := s.InitialCheck(worker); err != nil {
		return fmt.Errorf("failed initial check: %w", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Start listening for WebSocket messages
	go func() {
		for {
			_, message, err := s.client.Conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				os.Exit(1)
			}
			record, err := NewPostgresChanges(message, worker)
			if err != nil {
				log.Printf("Failed to handle payload: %v", err)
				continue
			}

			fmt.Printf("Received event: %s\n", record.Event)
			if record.Event == "postgres_changes" {
				table := viper.GetString("table")
				if record.Payload.Data.Table == table {
					if err := record.HandleMessage(); err != nil {
						log.Printf("Failed to handle message: %v", err)
					}
				} else {
					log.Printf("Unhandled table: %s", record.Payload.Data.Table)
				}
			}
		}
	}()

	table := viper.GetString("table")
	// Create subscription payload
	payload := SubscriptionPayload{
		Event: "phx_join",
		Topic: fmt.Sprintf("realtime:public:%s", table),
		Ref:   uuid.New().String(),
	}
	payload.Payload.Config.Broadcast.Self = true
	chainId := viper.GetInt("chainId")
	payload.Payload.Config.PostgresChanges = []PostgresChange{
		{
			Event:  "*",      // Listen to all events (INSERT, UPDATE, DELETE)
			Schema: "public", // Database schema
			Table:  table,    // Table name
			Filter: fmt.Sprintf("chain_id=eq.%d", chainId),
		},
	}

	// Send subscription request
	if err := s.client.Conn.WriteJSON(payload); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	fmt.Println("Subscribed to table changes.")
	return nil
}

// Stop closes the subscription connection
func (s *Subscription) Stop() {
	s.client.Close()
}

// NewPostgresChanges creates a PostgresChanges instance from a raw message
func NewPostgresChanges(message []byte, worker *worker.Worker) (*PostgresChanges, error) {
	var changes PostgresChanges
	if err := json.Unmarshal(message, &changes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	changes.Worker = worker
	return &changes, nil
}

// HandleMessage processes a database change event and updates containers if needed
func (p *PostgresChanges) HandleMessage() error {
	handlers := []handlers.Handler{
		handlers.NewCoinHandler(),
		handlers.NewImageHandler(),
		handlers.NewNameHandler(),
	}

	containersToRestart := []docker.Container{}
	for _, handler := range handlers {
		result := handler.Handle(&p.Payload.Data.Record)
		if result.Error != nil {
			return fmt.Errorf("handler error: %w", result.Error)
		}
		containersToRestart = append(containersToRestart, result.ContainersToRestart...)
	}

	if len(containersToRestart) > 0 {
		added := p.Worker.AddJob(containersToRestart)
		if !added {
			log.Printf("Job for containers %v already in queue", containersToRestart)
		}
	}

	return nil
}

// InitialCheck queries the database for existing record and processes it
// This ensures containers are properly configured on service startup
func (s *Subscription) InitialCheck(worker *worker.Worker) error {
	dbURL := viper.GetString("supabaseUrl")
	chainId := viper.GetInt("chainId")
	table := viper.GetString("table")

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Query the current state - limit 1 since there should be only one record
	query := fmt.Sprintf("SELECT id, name, base_token_symbol, chain_id, network_logo, network_logo_dark, favicon, created_at, updated_at FROM %s WHERE chain_id = $1 LIMIT 1", table)
	rows, err := db.Query(query, chainId)
	if err != nil {
		return fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	// Process each record using the same handlers as real-time updates
	for rows.Next() {
		var record handlers.Record
		err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.Coin,
			&record.ChainID,
			&record.LightLogoURL,
			&record.DarkLogoURL,
			&record.FaviconURL,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a PostgresChanges instance to reuse existing handler logic
		changes := &PostgresChanges{
			Event:  "postgres_changes",
			Worker: worker,
		}
		changes.Payload.Data.Record = record
		changes.Payload.Data.Table = table

		// Handle the record
		if err := changes.HandleMessage(); err != nil {
			log.Printf("Failed to handle initial record %d: %v", record.ID, err)
			continue
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}
