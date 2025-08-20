package cmd

import (
	"blockscout-vc/internal/client"
	"blockscout-vc/internal/config"
	"blockscout-vc/internal/heartbeat"
	"blockscout-vc/internal/server"
	"blockscout-vc/internal/subscription"
	"blockscout-vc/internal/worker"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StartSidecarCmd creates and returns the sidecar command.
// It monitors database changes and manages container configurations.
func StartSidecarCmd() *cobra.Command {
	startServer := &cobra.Command{
		Use:   "sidecar",
		Short: "Start sidecar service with token management",
		Long:  `Starts the sidecar service that monitors database changes, manages containers, and provides token management functionality`,
		PreRun: func(cmd *cobra.Command, args []string) {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			config.InitConfig(configPath)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a cancellable context
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Create the sidecar-injected.env file if it doesn't exist
			sidecarInjectedEnv := viper.GetString("pathToEnvFile")
			if sidecarInjectedEnv != "" {
				if _, err := os.Stat(sidecarInjectedEnv); os.IsNotExist(err) {
					file, err := os.Create(sidecarInjectedEnv)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error creating env file: %v\n", err)
					} else {
						if closeErr := file.Close(); closeErr != nil {
							fmt.Fprintf(os.Stderr, "Error closing env file: %v\n", closeErr)
						}
					}
				}
			}

			// Initialize and start HTTP server
			httpServer, err := server.NewServer()
			if err != nil {
				return fmt.Errorf("failed to initialize HTTP server: %w", err)
			}

			// Create error channel for HTTP server
			serverErrChan := make(chan error, 1)

			go func() {
				port := viper.GetString("httpPort")
				fmt.Printf("Starting HTTP server on port %s\n", port)
				fmt.Printf("Token management web interface available at: http://localhost:%s/\n", port)
				fmt.Printf("API endpoints available at: http://localhost:%s/api/v1/\n", port)
				if err := httpServer.Start(port); err != nil {
					fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
					serverErrChan <- err
				}
			}()

			// Initialize WebSocket client
			supabaseUrl := viper.GetString("supabaseUrl")
			supabaseRealtimeUrl := viper.GetString("supabaseRealtimeUrl")
			supabaseAnonKey := viper.GetString("supabaseAnonKey")
			if supabaseUrl != "" && supabaseRealtimeUrl != "" && supabaseAnonKey != "" {
				realtimeClient := client.New(supabaseRealtimeUrl, supabaseAnonKey)
				if err := realtimeClient.Connect(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to connect to Supabase realtime: %v\n", err)
					// Continue without realtime functionality rather than exiting
					fmt.Println("Continuing without realtime database monitoring...")
				} else {
					// Only defer Close if client was successfully created and connected
					defer func() {
						if closeErr := realtimeClient.Close(); closeErr != nil {
							fmt.Fprintf(os.Stderr, "Error closing realtime client: %v\n", closeErr)
						}
					}()

					// Initialize and start the worker
					worker := worker.New()
					worker.Start(ctx)

					// Initialize and start heartbeat service
					hb := heartbeat.New(realtimeClient, 30*time.Second)
					hb.Start()
					defer hb.Stop()

					// Initialize and start subscription service
					sub := subscription.New(realtimeClient)
					if err := sub.Subscribe(worker); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to subscribe to database changes: %v\n", err)
						fmt.Println("Continuing without database change monitoring...")
					} else {
						defer sub.Stop()
					}
				}
			}

			// Wait for interrupt signal or server error
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

			// Wait for interrupt signal or server error
			select {
			case <-interrupt:
				fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
			case <-ctx.Done():
				fmt.Println("\nContext cancelled, shutting down...")
			case err := <-serverErrChan:
				fmt.Printf("\nHTTP server failed to start: %v\n", err)
				fmt.Println("Shutting down due to server error...")
			}

			// Create shutdown context with timeout
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()

			// Shutdown HTTP server
			fmt.Println("Shutting down HTTP server...")
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Error shutting down HTTP server: %v\n", err)
			}

			fmt.Println("Shutdown complete.")
			return nil
		},
	}
	startServer.PersistentFlags().StringP("config", "c", "", "Path of the configuration file")
	return startServer
}
