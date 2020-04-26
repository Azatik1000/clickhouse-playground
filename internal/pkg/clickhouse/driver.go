package clickhouse

type Driver interface {
	Exec(query string) (string, error)
	HealthCheck() bool
}



