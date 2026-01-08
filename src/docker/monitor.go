package docker

import (
	"context"
	"log"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var networkName string = "mixproxy"
var skip []string = []string{"mix-admin-1", "mix-proxy-1", "mix-redis-1"}
var cli *client.Client
var ctx context.Context

type Containers struct {
	Name   string
	Id     string
	FullId string
	State  string
}

func init() {
	var err error
	ctx = context.Background()

	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
}

func GetContainers(all bool) []Containers {
	containersList := []Containers{}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	for _, cont := range containers {
		if slices.Contains(skip, cont.Names[0][1:]) {
			continue
		}

		if cont.NetworkSettings == nil || cont.NetworkSettings.Networks == nil {
			continue
		}

		if _, exists := cont.NetworkSettings.Networks[networkName]; !exists {
			continue
		}

		if !all && cont.State != "running" {
			continue
		}

		add := Containers{
			Name:  cont.Names[0][1:], // Remove leading '/'
			Id:    cont.ID[:12],
			State: cont.State,
		}
		containersList = append(containersList, add)
	}

	return containersList
}

func StopContainer(id string) {
	cli.ContainerStop(ctx, id, container.StopOptions{})
}

func StartContainer(id string) {
	cli.ContainerStart(ctx, id, container.StartOptions{})
}

func CreateContainer(image string, name string) {
	config := &container.Config{
		Image: image,
	}
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
	if err != nil {
		log.Fatal(err)
	}
	cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func CreateContainerFromTemplate(image string, name string, port int, envs []string) {
	config := &container.Config{
		Image: image,
		Env:   envs,
	}
	hostConfig := &container.HostConfig{}
	// containerPort := strconv.Itoa(port) + "/tcp"
	// PortBindings: nat.PortMap{
	// 	nat.Port(containerPort): []nat.PortBinding{{HostPort: strconv.Itoa(port)}},
	// },

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
	if err != nil {
		log.Fatal(err)
	}
	cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}
