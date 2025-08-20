package env

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

type Env struct {
	PathToEnvFile string
	EnvFile       map[string]string
}

func NewEnv() *Env {
	return &Env{
		PathToEnvFile: viper.GetString("pathToEnvFile"),
		EnvFile:       make(map[string]string),
	}
}

// ReadEnvFile reads and parses the environment file
func (e *Env) ReadEnvFile() error {
	file, err := os.Open(e.PathToEnvFile)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close env file: %v\n", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		e.EnvFile[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning env file: %w", err)
	}

	return nil
}

// WriteEnvFile writes the environment variables back to the file
func (e *Env) WriteEnvFile() error {
	file, err := os.Create(e.PathToEnvFile)
	if err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close env file: %v\n", closeErr)
		}
	}()

	writer := bufio.NewWriter(file)

	// Sort keys for consistent output
	keys := make([]string, 0, len(e.EnvFile))
	for k := range e.EnvFile {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := e.EnvFile[key]
		// Add quotes if value contains spaces
		if strings.Contains(value, " ") {
			value = fmt.Sprintf(`"%s"`, value)
		}

		line := fmt.Sprintf("%s=%s\n", key, value)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write line to env file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush env file: %w", err)
	}

	return nil
}

// UpdateEnvVars updates environment variables in the env file
// Returns whether any changes were made
func (e *Env) UpdateEnvVars(updates map[string]string) (bool, error) {
	err := e.ReadEnvFile()
	if err != nil {
		return false, fmt.Errorf("failed to read env file: %w", err)
	}

	updated := false
	for key, newValue := range updates {
		if currentValue, exists := e.EnvFile[key]; !exists || currentValue != newValue {
			e.EnvFile[key] = newValue
			updated = true
		}
	}

	if updated {
		if err := e.WriteEnvFile(); err != nil {
			return false, fmt.Errorf("failed to write env file: %w", err)
		}
	}

	return updated, nil
}
