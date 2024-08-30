package repositories

import (
	"pipe/internal/entity"

	"github.com/gocql/gocql"
)

var _ CommonBehaviourRepository = &CassandraCommonBehaviour{}

type CassandraCommonBehaviour struct {
	session *gocql.Session
}

func NewCassandraCommonBehaviour(session *gocql.Session) *CassandraCommonBehaviour {
	return &CassandraCommonBehaviour{
		session: session,
	}
}

func (r *CassandraCommonBehaviour) ByID(ID int64) (entity.User, error) {
	user := entity.User{}
	err := r.session.Query("SELECT user_id, private_id, pubkey, created_at FROM users_by_id WHERE user_id = ?", ID).Scan(&user.ID, &user.PrivateID, &user.PubKey, &user.CreatedAt)
	if err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *CassandraCommonBehaviour) ByPrivateID(privateID string) (entity.User, error) {
	user := entity.User{}
	err := r.session.Query(`SELECT user_id, private_id, pubkey, created_at FROM users_by_private_id WHERE private_id = ?`, privateID).
		Consistency(gocql.One).
		Scan(&user.ID, &user.PrivateID, &user.PubKey, &user.CreatedAt)
	if err != nil {
		return entity.User{}, err
	}
	return user, nil
}
