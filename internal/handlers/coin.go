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

	compose, err := h.docker.ReadComposeFile()
	if err != nil {
		result.Error = fmt.Errorf("failed to read compose file: %w", err)
		return result
	}

	// Define environment updates for each service
	updates := []docker.EnvUpdate{
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

	result.ContainersToRestart, err = h.docker.AppyAndUpdateEachService(compose, updates)
	if err != nil {
		result.Error = err
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
