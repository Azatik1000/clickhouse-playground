package docker

import (
	"github.com/docker/docker/client"
)

type Manager struct {
	*client.Client
}

func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// TODO: restore pull
	//reader, err := cli.ImagePull(context.Background(), executorImageName, types.ImagePullOptions{})
	//if err != nil {
	//	return nil, err
	//}

	// TODO: maybe remove
	//io.Copy(os.Stdout, reader)

	return &Manager{cli}, nil
}


