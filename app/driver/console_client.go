package driver

import "app/docker"

type consoleClientDriver struct {
	client *docker.DbClientContainer
}

func NewConsoleClientDriver(manager *docker.Manager, backendName string) (Driver, error) {
	client, err := docker.NewDbClient(manager, backendName)
	if err != nil {
		return nil, err
	}

	_, err = (*docker.Container)(client).Run()
	if err != nil {
		return nil, err
	}

	return &consoleClientDriver{client: client}, nil
}

func (c *consoleClientDriver) Exec(query string) (string, error) {
	output, err := c.client.Write([]byte(query))
	return string(output), err
}

func (c *consoleClientDriver) HealthCheck() error {
	// TODO: change
	return nil
	//panic("implement me")
}

func (c *consoleClientDriver) Close() error {
	// TODO: remove container
	return nil
}


