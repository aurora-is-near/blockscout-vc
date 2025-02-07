package cmd

import (
	"blockscout-vc/internal/client"
	"blockscout-vc/internal/config"
	"blockscout-vc/internal/heartbeat"
	"blockscout-vc/internal/subscription"
	"blockscout-vc/internal/worker"
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StartSidecarCmd creates and returns the sidecar command.
// It monitors database changes and manages container configurations.
func StartSidecarCmd() *cobra.Command {
	startServer := &cobra.Command{
		Use:   "sidecar",
		Short: "Start sidecar",
		Long:  `Starts sidecar to listen for changes in the database and recreate the containers`,
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
					if _, err := os.Create(sidecarInjectedEnv); err != nil {
						fmt.Fprintf(os.Stderr, "Error creating env file: %v\n", err)
					}
				}
			}

			// Initialize WebSocket client
			supabaseRealtimeUrl := viper.GetString("supabaseRealtimeUrl")
			supabaseAnonKey := viper.GetString("supabaseAnonKey")
			client := client.New(supabaseRealtimeUrl, supabaseAnonKey)
			if err := client.Connect(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer client.Close()

			// Initialize and start the worker
			worker := worker.New()
			worker.Start(ctx)

			// Initialize and start heartbeat service
			hb := heartbeat.New(client, 30*time.Second)
			hb.Start()
			defer hb.Stop()

			// Initialize and start subscription service
			sub := subscription.New(client)
			if err := sub.Subscribe(worker); err != nil {
				return fmt.Errorf("failed to subscribe: %w", err)
			}
			defer sub.Stop()

			// Wait for interrupt signal
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			<-interrupt
			fmt.Println("\nReceived interrupt signal, shutting down.")
			return nil
		},
	}
	startServer.PersistentFlags().StringP("config", "c", "", "Path of the configuration file")
	return startServer
}
