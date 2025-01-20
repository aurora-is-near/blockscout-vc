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

type Container struct {
	Name        string
	ServiceName string
}

// RecreateContainers stops, removes and recreates specified containers
// It uses docker-compose to handle the container lifecycle
func (d *Docker) RecreateContainers(containers []Container) error {
	pathToDockerCompose := viper.GetString("pathToDockerCompose")
	uniqueContainers := d.UniqueContainers(containers)

	// Define the sequence of commands to execute
	containerNames := d.GetContainerNames(uniqueContainers)
	serviceNames := d.GetServiceNames(uniqueContainers)
	commands := []struct {
		args       []string
		desc       string
		errMessage string
	}{
		{
			args:       append([]string{"rm", "-f"}, containerNames...),
			desc:       "Stopping and removing containers",
			errMessage: "Error stopping and removing containers",
		},
		{
			args:       append([]string{"compose", "-f", pathToDockerCompose, "up", "-d", "--force-recreate", "--remove-orphans", "--no-deps"}, serviceNames...),
			desc:       "Recreating containers",
			errMessage: "Error recreating containers",
		},
	}

	// Execute each command in sequence
	for _, cmd := range commands {
		execCmd := exec.Command("docker", cmd.args...)
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

func (d *Docker) AppyAndUpdateEachService(compose map[string]interface{}, updates []EnvUpdate) ([]Container, error) {
	pathToCommonEnvironmentFile := viper.GetString("pathToCommonEnvironmentFile")

	if pathToCommonEnvironmentFile == "" {
		return d.applyAndUpdateEachServiceUsingCompose(compose, updates)
	}

	envViper := viper.New()
	envViper.SetConfigFile(pathToCommonEnvironmentFile)
	if err := envViper.ReadInConfig(); err != nil {
		err = fmt.Errorf("error loading %s env file: %w", pathToCommonEnvironmentFile, err)
		return nil, err
	}
	return d.applyAndUpdateEachServiceUsingEnvFile(envViper, updates)
}

func (d *Docker) applyAndUpdateEachServiceUsingEnvFile(variables *viper.Viper, updates []EnvUpdate) ([]Container, error) {
	var containers []Container
	for _, env := range updates {
		if (variables.GetString(env.Key) != env.Value) {
			variables.Set(env.Key, env.Value)

			fmt.Printf("Updated %s service environment: %+v\n", env.ServiceName, env)
			containers = append(containers, Container{
				Name:        env.ContainerName,
				ServiceName: env.ServiceName,
			})
		}
	}

	file, err := os.Create(variables.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("error opening .env file: %v", err)
	}
	defer file.Close()

	for _, key := range viper.AllKeys() {
		_, err := fmt.Fprintf(file, "%s=%s\n", key, viper.GetString(key))
		if err != nil {
			return nil, fmt.Errorf("error writing to .env file: %v", err)
		}
	}

	return containers, nil
}

func (d *Docker) applyAndUpdateEachServiceUsingCompose(compose map[string]interface{}, updates []EnvUpdate) ([]Container, error) {
	var containers []Container
	var err error
	
	// Apply updates to each service
	for _, env := range updates {
		var updated bool
		compose, updated, err = d.UpdateServiceEnv(compose, env.ServiceName, map[string]interface{}{
			env.Key: env.Value,
		})
		if err != nil {
			err = fmt.Errorf("failed to update %s service environment: %w", env.ServiceName, err)
			return nil, err
		}
		if updated {
			fmt.Printf("Updated %s service environment: %+v\n", env.ServiceName, env)
			containers = append(containers, Container{
				Name:        env.ContainerName,
				ServiceName: env.ServiceName,
			})
		}
	}

	if err = d.WriteComposeFile(compose); err != nil {
		err = fmt.Errorf("failed to write compose file: %w", err)
		return nil, err
	}
	return containers, nil
}

// UniqueContainerNames returns a sorted list of unique container names
func (d *Docker) UniqueContainers(containers []Container) []Container {
	unique := make(map[string]Container)
	for _, container := range containers {
		unique[container.Name] = container
	}
	uniqueContainers := make([]Container, 0, len(unique))
	for _, container := range unique {
		uniqueContainers = append(uniqueContainers, container)
	}
	return uniqueContainers
}

func (d *Docker) GetContainerNames(containers []Container) []string {
	names := make([]string, 0, len(containers))
	for _, container := range containers {
		names = append(names, container.Name)
	}
	sort.Strings(names)
	return names
}

func (d *Docker) GetServiceNames(containers []Container) []string {
	names := make([]string, 0, len(containers))
	for _, container := range containers {
		names = append(names, container.ServiceName)
	}
	sort.Strings(names)
	return names
}
