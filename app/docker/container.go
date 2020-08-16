package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

var executorImageName = "clickhouse-executor"
var serverImageName = "yandex/clickhouse-server"

type Container struct {
	manager *Manager
	Id      string
}

type ExecutorContainer Container
type DbClientContainer Container

func NewExecutor(
	manager *Manager,
	alias string,
) (*ExecutorContainer, error) {
	// TODO: can i move this configuration to dockerfile and import from it?
	// TODO: load config
	httpContainerPort := 8123
	nativeContainerPort := 9000
	restContainerPort := 8080

	exposedPorts := nat.PortSet(
		map[nat.Port]struct{}{
			nat.Port(fmt.Sprintf("%d/tcp", httpContainerPort)):   {},
			nat.Port(fmt.Sprintf("%d/tcp", nativeContainerPort)): {},
			nat.Port(fmt.Sprintf("%d/tcp", restContainerPort)):   {},
		},
	)

	//portBindings := nat.PortMap(
	//	map[nat.Port][]nat.PortBinding{
	//		nat.Port(fmt.Sprintf("%d/tcp", restContainerPort)): {nat.PortBinding{
	//			HostPort: fmt.Sprintf("%d", restHostPort),
	//		}},
	//		nat.Port(fmt.Sprintf("%d/tcp", nativeContainerPort)): {nat.PortBinding{
	//			HostPort: fmt.Sprintf("%d", nativeHostPort),
	//		}},
	//	},
	//)

	endpointsConfig := map[string]*network.EndpointSettings{
		"clickhouse-playground_default": {
			Aliases: []string{alias},
		},
	}

	fmt.Println("creating a container")
	// TODO: add restart policy
	resp, err := manager.ContainerCreate(context.Background(), &container.Config{
		Image:        executorImageName,
		ExposedPorts: exposedPorts,
		//Cmd:          []string{"-nm"},
	}, &container.HostConfig{
		//PortBindings: portBindings,
		NetworkMode: "clickhouse-playground_default",
		//AutoRemove:  true,
	}, &network.NetworkingConfig{EndpointsConfig: endpointsConfig}, alias)

	if err != nil {
		return nil, err
	}

	return &ExecutorContainer{manager, resp.ID}, nil
}

func NewServer(
	manager *Manager,
	alias string,
) (*ExecutorContainer, error) {
	// TODO: can i move this configuration to dockerfile and import from it?
	// TODO: load config
	nativeContainerPort := 9000

	exposedPorts := nat.PortSet(
		map[nat.Port]struct{}{
			nat.Port(fmt.Sprintf("%d/tcp", nativeContainerPort)): {},
		},
	)

	//portBindings := nat.PortMap(
	//	map[nat.Port][]nat.PortBinding{
	//		nat.Port(fmt.Sprintf("%d/tcp", restContainerPort)): {nat.PortBinding{
	//			HostPort: fmt.Sprintf("%d", restHostPort),
	//		}},
	//		nat.Port(fmt.Sprintf("%d/tcp", nativeContainerPort)): {nat.PortBinding{
	//			HostPort: fmt.Sprintf("%d", nativeHostPort),
	//		}},
	//	},
	//)

	endpointsConfig := map[string]*network.EndpointSettings{
		"clickhouse-playground_default": {
			Aliases: []string{alias},
		},
	}

	fmt.Println("creating a container")
	// TODO: add restart policy
	resp, err := manager.ContainerCreate(context.Background(), &container.Config{
		Image:        serverImageName,
		ExposedPorts: exposedPorts,
		//Cmd:          []string{"-nm"},
	}, &container.HostConfig{
		//PortBindings: portBindings,
		NetworkMode: "clickhouse-playground_default",
		//AutoRemove:  true,
	}, &network.NetworkingConfig{EndpointsConfig: endpointsConfig}, alias)

	if err != nil {
		return nil, err
	}

	return &ExecutorContainer{manager, resp.ID}, nil
}

//func NewDbClient(
//	manager *Manager,
//	backendName string,
//	alias string,
//) (*DbClientContainer, error) {
//	// TODO: remove code duplication
//	// TODO: can i move this configuration to dockerfile and import from it?
//	execContainerPort := 9980
//	//execHostPort := 9980
//
//	exposedPorts := nat.PortSet(
//		map[nat.Port]struct{}{
//			nat.Port(fmt.Sprintf("%d/tcp", execContainerPort)): {},
//		},
//	)
//
//	//portBindings := nat.PortMap(
//	//	map[nat.Port][]nat.PortBinding{
//	//		nat.Port(fmt.Sprintf("%d/tcp", execContainerPort)): {nat.PortBinding{
//	//			HostPort: fmt.Sprintf("%d", execHostPort),
//	//		}},
//	//	},
//	//)
//
//	endpointsConfig := map[string]*network.EndpointSettings{
//		"clickhouse-playground_default": {
//			Aliases: []string{alias},
//		},
//	}
//
//	fmt.Println("creating a container")
//	// TODO: add restart policy
//	resp, err := manager.ContainerCreate(context.Background(), &container.Config{
//		Image:        dbClientImageName,
//		AttachStdin:  true,
//		AttachStdout: true,
//		AttachStderr: true,
//		//Tty:          true,
//		OpenStdin: true,
//		Cmd: []string{
//			"--port=9000",
//			fmt.Sprintf("--host=%s", backendName),
//			"-nm",
//			"-f",
//			"JSON",
//		},
//		ExposedPorts: exposedPorts,
//	}, &container.HostConfig{
//		// TODO: remove hardcode
//		NetworkMode: "clickhouse-playground_default",
//		//PortBindings: portBindings,
//		//AutoRemove:  true,
//		//Links:       []string{fmt.Sprintf("%s:%s", backendName, "clickhouse-server")},
//	}, &network.NetworkingConfig{EndpointsConfig: endpointsConfig}, "")
//
//	if err != nil {
//		return nil, err
//	}
//
//	return &DbClientContainer{manager, resp.ID}, nil
//}

//func (c *DbClientContainer) Write(input []byte) ([]byte, error) {
//	hr, err := c.manager.ContainerAttach(context.Background(), c.Id, types.ContainerAttachOptions{
//		Stream: true,
//		Stdin:  true,
//		Stdout: true,
//		Stderr: true,
//	})
//
//	//out, err := c.manager.ContainerAttach(context.Background(), c.Id, types.ContainerAttachOptions{
//	//	Stream: true,
//	//	Stdout: true,
//	//	Stderr: true,
//	//})
//
//	if err != nil {
//		return nil, err
//	}
//
//	_, err = io.Copy(hr.Conn, bytes.NewReader(input))
//	if err != nil {
//		return nil, err
//	}
//
//	fmt.Println("wrote")
//
//	//err = hr.CloseWrite()
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	then := time.Now().Add(7 * time.Second)
//	hr.Conn.SetReadDeadline(then)
//
//	//io.Copy(os.Stdout, hr.Reader)
//	//for i := 0; i < 100; i++ {
//	//	data := make([]byte, 1)
//	//	n, err := out.Reader.Read(data)
//	//	fmt.Println(n, data, err)
//	//}
//
//	// TODO: change strategy
//	// TODO: handle error
//	output, _ := ioutil.ReadAll(hr.Reader)
//	//fmt.Println(string(output))
//	return output, nil
//	//return nil, nil
//}

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
