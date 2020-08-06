package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

var imageName = "yandex/clickhouse-server"

type Container struct {
	manager *Manager
	id      string
}

func NewContainer(port int, manager *Manager, alias string) (*Container, error) {
	ctx := context.Background()

	containerPort := nat.Port(fmt.Sprintf("%d/tcp", 8123))
	hostPort := fmt.Sprintf("%d", port)

	exposedPorts := nat.PortSet(
		map[nat.Port]struct{}{
			containerPort: {},
		},
	)

	portBindings := nat.PortMap(
		map[nat.Port][]nat.PortBinding{
			"8123/tcp": {nat.PortBinding{
				HostPort: hostPort,
			}},
		},
	)

	endpointsConfig := map[string]*network.EndpointSettings{
		"clickhouse-playground_default": &network.EndpointSettings{
			IPAMConfig:          nil,
			Links:               nil,
			Aliases:             []string{alias},
			NetworkID:           "",
			EndpointID:          "",
			Gateway:             "",
			IPAddress:           "",
			IPPrefixLen:         0,
			IPv6Gateway:         "",
			GlobalIPv6Address:   "",
			GlobalIPv6PrefixLen: 0,
			MacAddress:          "",
			DriverOpts:          nil,
		},
	}

	fmt.Println("creating a container")
	// TODO: add restart policy
	resp, err := manager.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings,
		NetworkMode: container.NetworkMode("clickhouse-playground_default"),
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, "")

	if err != nil {
		return nil, err
	}

	return &Container{manager, resp.ID}, nil
}

func (c *Container) GetIP() string {
	info, err := c.manager.ContainerInspect(context.Background(), c.id)
	if err != nil {
		panic(err)
	}

	return info.NetworkSettings.Networks["clickhouse-playground_default"].IPAddress
}

func (c *Container) Run(callback func()) error {
	ctx := context.Background()
	if err := c.manager.ContainerStart(ctx, c.id, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := c.manager.ContainerWait(ctx, c.id, container.WaitConditionNotRunning)
	go func() {
		// TODO: handle leak
		select {
		case err := <-errCh:
			if err != nil {
				callback()
			}
		case <-statusCh:
			callback()
		}
	}()

	return nil

	//out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	//if err != nil {
	//	panic(err)
	//}
	//
	//stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
