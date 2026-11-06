package handlers

import (
	"blockscout-vc/internal/docker"
	"blockscout-vc/internal/env"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func configureMetadataHandlers(t *testing.T) string {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	environmentPath := filepath.Join(t.TempDir(), "sidecar-injected.env")
	if err := os.WriteFile(environmentPath, []byte("UNRELATED_SETTING=preserved\n"), 0600); err != nil {
		t.Fatal(err)
	}
	viper.Set("pathToEnvFile", environmentPath)
	for _, serviceName := range []string{"frontend", "backend", "stats"} {
		viper.Set(serviceName+"ServiceName", serviceName)
		viper.Set(serviceName+"ContainerName", "test-"+serviceName)
	}
	return environmentPath
}

func applyMetadata(t *testing.T, record Record) []docker.Container {
	t.Helper()
	var containersToRestart []docker.Container
	for _, handler := range []Handler{NewNameHandler(), NewExplorerHandler()} {
		result := handler.Handle(&record)
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		containersToRestart = append(containersToRestart, result.ContainersToRestart...)
	}
	return containersToRestart
}

func readMetadataFile(t *testing.T, environmentPath string) []byte {
	t.Helper()
	contents, err := os.ReadFile(environmentPath)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func assertMetadata(t *testing.T, expectedValues map[string]string) {
	t.Helper()
	environment := env.NewEnv()
	if err := environment.ReadEnvFile(); err != nil {
		t.Fatal(err)
	}
	for key, expectedValue := range expectedValues {
		if actualValue := environment.EnvFile[key]; actualValue != expectedValue {
			t.Errorf("%s = %q, want %q", key, actualValue, expectedValue)
		}
	}
}

func assertRecreationServices(t *testing.T, containers []docker.Container, expectedServices ...string) {
	t.Helper()
	actualServices := make(map[string]bool)
	for _, container := range containers {
		actualServices[container.ServiceName] = true
		if container.Name != "test-"+container.ServiceName {
			t.Errorf("unexpected container identity: %+v", container)
		}
	}
	expectedServiceSet := make(map[string]bool)
	for _, serviceName := range expectedServices {
		expectedServiceSet[serviceName] = true
	}
	if !reflect.DeepEqual(actualServices, expectedServiceSet) {
		t.Errorf("recreation services = %v, want %v", actualServices, expectedServiceSet)
	}
}

func TestMetadataHandlersConverge(t *testing.T) {
	for _, explorerURL := range []string{
		"https://explorer.example.com",
		"https://explorer.example.com/",
		"https://explorer.example.com/path?query=value",
		"https://explorer.example.com:443/",
		"http://explorer.example.com/",
	} {
		t.Run(explorerURL, func(t *testing.T) {
			environmentPath := configureMetadataHandlers(t)
			record := Record{Name: "Example Network", ExplorerURL: explorerURL}
			if containers := applyMetadata(t, record); len(containers) == 0 {
				t.Fatal("initial metadata did not request container recreation")
			}
			expectedContents := readMetadataFile(t, environmentPath)
			for attempt := 0; attempt < 2; attempt++ {
				if containers := applyMetadata(t, record); len(containers) != 0 {
					t.Errorf("unchanged metadata requested recreation: %+v", containers)
				}
				if actualContents := readMetadataFile(t, environmentPath); !bytes.Equal(actualContents, expectedContents) {
					t.Error("unchanged metadata changed the environment file")
				}
			}
			assertMetadata(t, map[string]string{"UNRELATED_SETTING": "preserved"})
		})
	}
}

func TestNameChangesOnlyRecreateFrontend(t *testing.T) {
	for _, explorerURL := range []string{"https://explorer.example.com", "https://explorer.example.com/"} {
		t.Run(explorerURL, func(t *testing.T) {
			configureMetadataHandlers(t)
			viper.Set("proxyServiceName", "proxy")
			viper.Set("proxyContainerName", "test-proxy")
			record := Record{Name: "Example Network", ExplorerURL: explorerURL}
			applyMetadata(t, record)
			record.Name = "Renamed Network"
			assertRecreationServices(t, applyMetadata(t, record), "frontend")
			assertRecreationServices(t, applyMetadata(t, record))
		})
	}
}

func TestMetadataHandlersApplyChanges(t *testing.T) {
	environmentPath := configureMetadataHandlers(t)
	record := Record{Name: "Example Network", ExplorerURL: "https://explorer.example.com/"}
	applyMetadata(t, record)

	record.Name = "Renamed Network"
	assertRecreationServices(t, applyMetadata(t, record), "frontend")
	assertMetadata(t, map[string]string{
		"NEXT_PUBLIC_NETWORK_NAME":       "Renamed Network",
		"NEXT_PUBLIC_NETWORK_SHORT_NAME": "Renamed Network",
		"NEXT_PUBLIC_FEATURED_NETWORKS":  "[{'title':'Aurora','url':'https://explorer.aurora.dev/','group':'Mainnets'}, {'title':'Renamed Network','url':'https://explorer.example.com','group':'Mainnets', 'isActive':true}]",
	})

	record.ExplorerURL = "http://renamed.example.com/"
	assertRecreationServices(t, applyMetadata(t, record), "backend", "frontend", "stats")
	assertMetadata(t, map[string]string{
		"BLOCKSCOUT_HOST":               "renamed.example.com",
		"BLOCKSCOUT_HTTP_PROTOCOL":      "http",
		"NEXT_PUBLIC_FEATURED_NETWORKS": "[{'title':'Aurora','url':'https://explorer.aurora.dev/','group':'Mainnets'}, {'title':'Renamed Network','url':'http://renamed.example.com','group':'Mainnets', 'isActive':true}]",
		"UNRELATED_SETTING":             "preserved",
	})

	expectedContents := readMetadataFile(t, environmentPath)
	record.ExplorerURL = "http://renamed.example.com"
	if containers := applyMetadata(t, record); len(containers) != 0 {
		t.Errorf("equivalent URL requested recreation: %+v", containers)
	}
	if actualContents := readMetadataFile(t, environmentPath); !bytes.Equal(actualContents, expectedContents) {
		t.Error("equivalent URL changed the environment file")
	}
}

func TestExplorerChangesRecreateConfiguredProxy(t *testing.T) {
	configureMetadataHandlers(t)
	viper.Set("proxyServiceName", "proxy")
	viper.Set("proxyContainerName", "test-proxy")
	record := Record{Name: "Example Network", ExplorerURL: "https://explorer.example.com"}
	applyMetadata(t, record)
	record.ExplorerURL = "http://renamed.example.com"
	assertRecreationServices(t, applyMetadata(t, record), "backend", "frontend", "stats", "proxy")
	assertRecreationServices(t, applyMetadata(t, record))
}
