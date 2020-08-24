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
type DbServerContainer Container

func NewExecutor(
	manager *Manager,
	alias string,
	serverAlias string,
) (*ExecutorContainer, error) {
	// TODO: can i move this configuration to dockerfile and import from it?
	// TODO: load config
	restContainerPort := 8080

	exposedPorts := nat.PortSet(
		map[nat.Port]struct{}{
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
		Env: []string{
			fmt.Sprintf("HOST=%s", serverAlias), // TODO: right grammar?
		},
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
) (*DbServerContainer, error) {
	// TODO: remove code duplication
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
		//Mounts: []mount.Mount{{
		//	Type:   mount.TypeBind,
		//	Source: fmt.Sprintf("%s/mount", alias),
		//	Target: "",
		//}},
		//AutoRemove:  true,
	}, &network.NetworkingConfig{EndpointsConfig: endpointsConfig}, alias)

	if err != nil {
		return nil, err
	}

	return &DbServerContainer{manager, resp.ID}, nil
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

//func cleanDir(path string) {
//	dir, err := ioutil.ReadDir(path)
//	for _, d := range dir {
//		os.RemoveAll(path.Join([]string{"tmp", d.Name()}...))
//	}
//}

func (c *Container) Restart() error {
	// TODO: handle errors
	//inspection, _ := c.manager.ContainerInspect(context.Background(), c.Id)
	//dataPath := inspection.Mounts[0].Source

	//err := c.manager.ContainerStop(context.Background(), c.Id, nil)
	//if err != nil {
	//	return err
	//}
	//
	//c.manager.ContainerWait(context.Background(), c.Id, container.WaitConditionRemoved)

	//err = c.manager.ContainerRemove(context.Background(), c.Id, types.ContainerRemoveOptions{})
	//if err != nil {
	//	return err
	//}

	//time.Sleep(3 * time.Second)
	//
	//fmt.Printf("%#v\n", inspection.Mounts[0])
	//err = c.manager.VolumeRemove(context.Background(), inspection.Mounts[0].Name, true)
	//if err != nil {
	//	return err
	//}

	execInfo, err := c.manager.ContainerExecCreate(context.Background(), c.Id, types.ExecConfig{
		Cmd: []string{
			"sh",
			"-c",
			"rm -rf /var/lib/clickhouse/data /var/lib/clickhouse/metadata",
		},
	})

	if err != nil {
		return err
	}

	err = c.manager.ContainerExecStart(context.Background(), execInfo.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	//fmt.Printf("remove %s\n", path.Join(dataPath, "data"))
	//if err := os.RemoveAll(path.Join(dataPath)); err != nil {
	//	return err
	//}
	//
	//if err := os.RemoveAll(path.Join(dataPath, "metadata")); err != nil {
	//	return err
	//}

	//files, err := ioutil.ReadDir(dataPath)
	//if err != nil {
	//	fmt.Println(err)
	//	//log.Fatal(err)
	//} else {
	//	for _, f := range files {
	//		fmt.Println(f.Name())
	//	}
	//}

	return c.manager.ContainerRestart(context.Background(), c.Id, nil)
}
