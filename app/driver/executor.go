package driver

import (
	"app/docker"
	"fmt"
	"net/url"
)

type consoleClientDriver struct {
	client *docker.DbClientContainer
	alias  string
}

func NewConsoleClientDriver(id int, manager *docker.Manager, backendName string) (Driver, error) {
	var driver consoleClientDriver
	driver.alias = fmt.Sprintf("client%d", id)

	client, err := docker.NewDbClient(
		manager,
		backendName,
		driver.alias,
	)

	if err != nil {
		return nil, err
	}

	_, err = (*docker.Container)(client).Run()
	if err != nil {
		return nil, err
	}

	driver.client = client
	return &driver, nil
}

func (c *consoleClientDriver) Exec(query string) (string, error) {
	fmt.Println("Exec:", query)
	endpoint, _ := url.Parse(
		fmt.Sprintf("http://%s:%d", c.alias, 9980),
	)
	dbDriver := NewHTTPDriver(endpoint)

	//output, err := c.client.Write([]byte(query))
	//runes := bytes.Runes(output)
	//for _, r := range runes {
	//	fmt.Print(string(r))
	//}
	//fmt.Println(string(output))
	//fmt.Printf("%+q", output)

	//return string(output), err
	return dbDriver.Exec(query)
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
