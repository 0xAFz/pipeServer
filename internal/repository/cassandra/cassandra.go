package cassandra

import (
	"github.com/gocql/gocql"
)

func NewCassandraSession(addr, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(addr)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	return cluster.CreateSession()
}
