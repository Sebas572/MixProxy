package docker

import (
	"context"
	"log"
	"slices"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var networkName string = "mixproxy"
var skip []string = []string{"mix-admin-1", "mix-proxy-1", "mix-redis-1"}

type Containers struct {
	Name string
	Id   string
}

func GetContainers() []Containers {
	containersList := []Containers{}

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	networkDetail, err := cli.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for id, container := range networkDetail.Containers {
		if slices.Contains(skip, container.Name) {
			continue
		}

		add := Containers{
			Name: container.Name,
			Id:   string(id[:12]),
		}
		containersList = append(containersList, add)
	}

	return containersList
}
