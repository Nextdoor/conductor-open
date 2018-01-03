package data

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"

	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	postgresHost         = flags.EnvString("POSTGRES_HOST", "conductor-postgres")
	postgresPort         = flags.EnvString("POSTGRES_PORT", "5432")
	postgresUsername     = flags.EnvString("POSTGRES_USERNAME", "conductor")
	postgresPassword     = flags.EnvString("POSTGRES_PASSWORD", "conductor")
	postgresDatabaseName = flags.EnvString("POSTGRES_DATABASE_NAME", "conductor")
	postgresSSLMode      = flags.EnvString("POSTGRES_SSL_MODE", "disable")
)

type Postgres struct{ data }

func newPostgres() *Postgres {
	postgres := Postgres{}

	postgres.RegisterDB = func() error {
		return orm.RegisterDataBase("default", "postgres",
			fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				postgresHost, postgresPort, postgresUsername, postgresPassword,
				postgresDatabaseName, postgresSSLMode))
	}

	postgres.initialize()

	return &Postgres{}
}
