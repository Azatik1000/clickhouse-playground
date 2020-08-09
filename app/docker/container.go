package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"io"
)

var dbServerImageName = "yandex/clickhouse-server"
var dbClientImageName = "yandex/clickhouse-client"

type Container struct {
	manager *Manager
	Id      string
}

type DbServerContainer Container
type DbClientContainer Container

func NewDbServer(
	port int,
	manager *Manager,
	alias string,
	name string,
) (*DbServerContainer, error) {
	// TODO: can i move this configuration to dockerfile and import from it?
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
		"clickhouse-playground_default": {
			Aliases: []string{alias},
		},
	}

	fmt.Println("creating a container")
	// TODO: add restart policy
	resp, err := manager.ContainerCreate(context.Background(), &container.Config{
		Image:        dbServerImageName,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings,
		NetworkMode:  "clickhouse-playground_default",
		AutoRemove:   true,
	}, &network.NetworkingConfig{EndpointsConfig: endpointsConfig}, name)

	if err != nil {
		return nil, err
	}

	return &DbServerContainer{manager, resp.ID}, nil
}

func NewDbClient(
	manager *Manager,
	backendName string,
) (*DbClientContainer, error) {
	// TODO: remove code duplication
	//endpointsConfig := map[string]*network.EndpointSettings{
	//	"clickhouse-playground_default": {
	//		Aliases:             []string{alias},
	//	},
	//}

	fmt.Println("creating a container")
	// TODO: add restart policy
	resp, err := manager.ContainerCreate(context.Background(), &container.Config{
		Image:        dbClientImageName,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Cmd: []string{
			"--port=8123",
			fmt.Sprintf("--host=%s", backendName),
			"-nm",
		},
	}, &container.HostConfig{
		// TODO: remove hardcode
		NetworkMode: "clickhouse-playground_default",
		AutoRemove:  true,
		//Links:       []string{fmt.Sprintf("%s:%s", backendName, "clickhouse-server")},
	}, nil, "")

	if err != nil {
		return nil, err
	}

	return &DbClientContainer{manager, resp.ID}, nil
}

func (c *DbClientContainer) Write(input []byte) ([]byte, error) {
	hr, err := c.manager.ContainerAttach(context.Background(), c.Id, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})

	if err != nil {
		return nil, err
	}

	fmt.Println("start writing", input)

	n, err := io.Copy(hr.Conn, bytes.NewReader(input))
	fmt.Printf("wrote %d bytes to connection\n", n)

	if err != nil {
		return nil, err
	}

	if v, ok := hr.Conn.(interface{ CloseWrite() error }); ok {
		fmt.Println("closing write")
		v.CloseWrite()
	} else {
		fmt.Println("no such method")
	}

	fmt.Println("wrote")
	defer fmt.Println("read")

	//io.Copy(os.Stdout, hr.Reader)
	for {
		data := make([]byte, 1)
		n, err := hr.Reader.Read(data)
		fmt.Println(n, data, err)
	}

	//output, err := ioutil.ReadAll(hr.Reader)
	return nil, nil
}

func (c *Container) Run() (chan struct{}, error) {
	if err := c.manager.ContainerStart(context.Background(), c.Id, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	exited := make(chan struct{})
	statusCh, errCh := c.manager.ContainerWait(context.Background(), c.Id, container.WaitConditionNotRunning)
	go func() {
		// TODO: handle leak
		select {
		case err := <-errCh:
			exited <- struct{}{}
			if err != nil {
			}
		case <-statusCh:
			exited <- struct{}{}
		}
	}()

	return exited, nil
}

func (c *Container) Stop() error {
	return c.manager.ContainerStop(context.Background(), c.Id, nil)
}
