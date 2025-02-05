package handlers

import (
	"blockscout-vc/internal/docker"
	"fmt"

	"github.com/spf13/viper"
)

// MaxCoinLength defines the maximum allowed length for a coin symbol
const MaxCoinLength = 20

type CoinHandler struct {
	BaseHandler
}

func NewCoinHandler() *CoinHandler {
	return &CoinHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// Handle processes coin-related changes and updates service configurations
func (h *CoinHandler) Handle(record *Record) HandlerResult {
	result := HandlerResult{}

	if err := h.validateCoin(record.Coin); err != nil {
		result.Error = fmt.Errorf("invalid coin: %w", err)
		return result
	}

	updates := []EnvUpdate{
		{
			ServiceName:   viper.GetString("frontendServiceName"),
			Key:           "NEXT_PUBLIC_NETWORK_CURRENCY_SYMBOL",
			Value:         record.Coin,
			ContainerName: viper.GetString("frontendContainerName"),
		},
		{
			ServiceName:   viper.GetString("backendServiceName"),
			Key:           "COIN",
			Value:         record.Coin,
			ContainerName: viper.GetString("backendContainerName"),
		},
		{
			ServiceName:   viper.GetString("statsServiceName"),
			Key:           "STATS_CHARTS__TEMPLATE_VALUES__NATIVE_COIN_SYMBOL",
			Value:         record.Coin,
			ContainerName: viper.GetString("statsContainerName"),
		},
	}

	// Apply updates to each service
	for _, env := range updates {
		updated, err := h.UpdateServiceEnv(env.ServiceName, map[string]string{
			env.Key: env.Value,
		})
		if err != nil {
			result.Error = fmt.Errorf("failed to update %s service environment: %w", env.ServiceName, err)
			return result
		}
		if updated {
			fmt.Printf("Updated %s service environment: %+v\n", env.ServiceName, env)
			result.ContainersToRestart = append(result.ContainersToRestart, docker.Container{
				Name:        env.ContainerName,
				ServiceName: env.ServiceName,
			})
		}
	}

	return result
}

// validateCoin checks if the coin symbol meets the required criteria
func (h *CoinHandler) validateCoin(coin string) error {
	if coin == "" {
		return fmt.Errorf("coin symbol cannot be empty")
	}
	if len(coin) == 0 {
		return fmt.Errorf("coin symbol cannot be empty")
	}
	if len(coin) > MaxCoinLength {
		return fmt.Errorf("coin symbol length cannot exceed %d characters", MaxCoinLength)
	}
	return nil
}
