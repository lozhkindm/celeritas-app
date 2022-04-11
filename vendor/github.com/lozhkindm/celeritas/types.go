package celeritas

import "database/sql"

type initPaths struct {
	rootPath    string
	folderNames []string
}

type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type dbConfig struct {
	dsn      string
	database string
}

type database struct {
	DataType string
	Pool     *sql.DB
}

type redisConfig struct {
	host     string
	password string
	prefix   string
}
