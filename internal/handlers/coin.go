package handlers

import (
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

	compose, err := h.docker.ReadComposeFile()
	if err != nil {
		result.Error = fmt.Errorf("failed to read compose file: %w", err)
		return result
	}
	frontendService := viper.GetString("frontendServiceName")
	backendService := viper.GetString("backendServiceName")
	statsService := viper.GetString("statsServiceName")

	// Define environment updates for each service
	updates := map[string]map[string]interface{}{
		frontendService: {
			"NEXT_PUBLIC_NETWORK_CURRENCY_SYMBOL": record.Coin,
		},
		backendService: {
			"COIN": record.Coin,
		},
		statsService: {
			"STATS_CHARTS__TEMPLATE_VALUES__NATIVE_COIN_SYMBOL": record.Coin,
		},
	}

	// Apply updates to each service
	for service, env := range updates {
		var updated bool
		compose, updated, err = h.docker.UpdateServiceEnv(compose, service, env)
		if err != nil {
			result.Error = fmt.Errorf("failed to update %s service environment: %w", service, err)
			return result
		}
		if updated {
			fmt.Printf("Updated %s service environment: %+v\n", service, env)
			result.ContainersToRestart = append(result.ContainersToRestart, service)
		}
	}

	if err = h.docker.WriteComposeFile(compose); err != nil {
		result.Error = fmt.Errorf("failed to write compose file: %w", err)
		return result
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
