package driver

type Driver interface {
	Exec(query string) (string, error)
	HealthCheck() error
}



