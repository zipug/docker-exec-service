package postgres_utils

import "fmt"

func ParseURI(user, password, host, dbName, sslMode string, port int) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbName,
		sslMode,
	)
}
