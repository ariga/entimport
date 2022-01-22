package mux

import (
	"database/sql"
	"net/url"

	atlasmysql "ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"

	"entgo.io/ent/dialect"
	"github.com/go-sql-driver/mysql"
)

func init() {
	Default.RegisterProvider(mysqlProvider, "mysql")
	Default.RegisterProvider(postgresProvider, "postgres", "postgresql")
}

func mysqlProvider(dsn string) (*ImportDriver, error) {
	db, err := sql.Open(dialect.MySQL, dsn)
	if err != nil {
		return nil, err
	}
	drv, err := atlasmysql.Open(db)
	if err != nil {
		return nil, err
	}
	// dsn example: root:pass@tcp(localhost:3308)/test?parseTime=True
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return &ImportDriver{
		Closer:     db,
		Inspector:  drv,
		Dialect:    dialect.MySQL,
		SchemaName: cfg.DBName,
	}, nil
}

func postgresProvider(dsn string) (*ImportDriver, error) {
	dsn = "postgres://" + dsn
	db, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}
	drv, err := postgres.Open(db)
	if err != nil {
		return nil, err
	}
	// dsn example: postgresql://user:pass@localhost:5432/atlas?search_path=some_schema
	parsed, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	schemaName := "public"
	if s := parsed.Query().Get("search_path"); s != "" {
		schemaName = s
	}
	return &ImportDriver{
		Closer:     db,
		Inspector:  drv,
		Dialect:    dialect.Postgres,
		SchemaName: schemaName,
	}, nil
}
