package docker

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/spf13/viper"
)

type Docker struct {
	ContainerName       string
	PathToDockerCompose string
	ComposeFile         map[string]interface{}
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
	projectName := viper.GetString("projectName")
	uniqueContainers := d.UniqueContainers(containers)

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker executable not found: %w", err)
	}

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
			args: append([]string{"compose",
				"-f", pathToDockerCompose,
				"--project-name", projectName,
				"up", "-d", "--force-recreate", "--remove-orphans", "--no-deps"},
				serviceNames...),
			desc:       "Recreating containers",
			errMessage: "Error recreating containers",
		},
	}

	// Execute each command in sequence
	for _, cmd := range commands {
		execCmd := exec.Command(dockerPath, cmd.args...)
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
