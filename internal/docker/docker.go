package docker

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Docker struct {
	ContainerName       string
	PathToDockerCompose string
}

func NewDocker() *Docker {
	return &Docker{
		PathToDockerCompose: viper.GetString("pathToDockerCompose"),
	}
}

// RecreateContainers stops, removes and recreates specified containers
// It uses docker-compose to handle the container lifecycle
func (d *Docker) RecreateContainers(containerNames []string) error {
	pathToDockerCompose := viper.GetString("pathToDockerCompose")
	uniqueContainers := d.UniqueContainerNames(containerNames)

	// Define the sequence of commands to execute
	commands := []struct {
		args       []string
		desc       string
		errMessage string
	}{
		{
			args:       []string{"rm", "-f"},
			desc:       "Stopping and removing containers",
			errMessage: "Error stopping and removing containers",
		},
		{
			args:       []string{"compose", "-f", pathToDockerCompose, "up", "-d", "--force-recreate"},
			desc:       "Recreating containers",
			errMessage: "Error recreating containers",
		},
	}

	// Execute each command in sequence
	for _, cmd := range commands {
		args := append(cmd.args, uniqueContainers...)

		execCmd := exec.Command("docker", args...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		fmt.Printf("%s: %s\n", cmd.desc, execCmd.String())
		if err := execCmd.Run(); err != nil {
			fmt.Printf("%s: %v\n", cmd.errMessage, err)
			return err
		}
	}

	fmt.Println("Docker containers recreated successfully!")
	return nil
}

// ReadComposeFile reads and parses the Docker compose file
func (d *Docker) ReadComposeFile() (map[string]interface{}, error) {
	data, err := os.ReadFile(d.PathToDockerCompose)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	return config, nil
}

// WriteComposeFile writes the updated compose configuration back to the file
func (d *Docker) WriteComposeFile(compose map[string]interface{}) error {
	data, err := yaml.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	if err := os.WriteFile(d.PathToDockerCompose, data, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// UpdateServiceEnv updates environment variables for a specific service in the compose file
// Returns the updated compose configuration and whether any changes were made
func (d *Docker) UpdateServiceEnv(compose map[string]interface{}, serviceName string, env map[string]interface{}) (map[string]interface{}, bool, error) {
	updated := false
	services, ok := compose["services"].(map[string]interface{})
	if !ok {
		return nil, updated, fmt.Errorf("services section not found")
	}

	service, ok := services[serviceName].(map[string]interface{})
	if !ok {
		return nil, updated, fmt.Errorf("service %s not found", serviceName)
	}

	serviceEnv, ok := service["environment"].(map[string]interface{})
	if !ok {
		return nil, updated, fmt.Errorf("environment section not found in service")
	}

	for key, value := range env {
		if serviceEnv[key] != value {
			serviceEnv[key] = value
			updated = true
		}
	}

	return compose, updated, nil
}

// UniqueContainerNames returns a sorted list of unique container names
func (d *Docker) UniqueContainerNames(containerNames []string) []string {
	unique := make(map[string]bool)
	for _, name := range containerNames {
		unique[name] = true
	}
	keys := make([]string, 0, len(unique))
	for k := range unique {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
