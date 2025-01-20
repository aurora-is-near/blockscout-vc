package docker

type EnvUpdate struct {
	ServiceName   string
	Key           string
	Value         string
	ContainerName string
}
