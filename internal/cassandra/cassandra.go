package cassandra

import (
	"pipe/internal/config"

	"github.com/gocql/gocql"
)

func NewCassandraSession() (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.AppConfig.CassandraHost)
	cluster.Keyspace = config.AppConfig.CassandraKeyspace
	cluster.Consistency = gocql.Quorum

	return cluster.CreateSession()
}
