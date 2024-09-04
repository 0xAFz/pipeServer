package repository

import (
	"fmt"
	"pipe/internal/entity"

	"github.com/gocql/gocql"
)

var _ Account = &AccountCassandraRepository{}

type AccountCassandraRepository struct {
	*CassandraCommonBehaviour
}

func NewAccountCassandraRepository(session *gocql.Session) *AccountCassandraRepository {
	return &AccountCassandraRepository{
		NewCassandraCommonBehaviour(session),
	}
}

func (r *AccountCassandraRepository) Save(user entity.User) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)
	batch.Query(`
		INSERT INTO users_by_id (user_id, private_id, pubkey, created_at) VALUES (?, ?, ?, ?)`,
		user.ID, user.PrivateID, user.PubKey, user.CreatedAt,
	)
	batch.Query(`
		INSERT INTO users_by_private_id (user_id, private_id, pubkey, created_at) VALUES (?, ?, ?, ?)`,
		user.ID, user.PrivateID, user.PubKey, user.CreatedAt,
	)

	if err := r.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *AccountCassandraRepository) SetPubKey(user entity.User) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)
	// batch.Query(`
	// 	DELETE FROM users_by_private_id WHERE private_id = ?`,
	// 	user.PrivateID,
	// )
	// batch.Query(`
	// 	INSERT INTO users_by_private_id (user_id, private_id, pubkey, created_at) VALUES (?, ?, ?, ?)`,
	// 	user.ID, user.PrivateID, user.PubKey, user.CreatedAt,
	// )
	batch.Query(`
		UPDATE users_by_id 
		SET pubkey = ? 
		WHERE user_id = ?`,
		user.PubKey, user.ID,
	)
	batch.Query(`
		UPDATE users_by_private_id 
		SET pubkey = ? 
		WHERE private_id = ?`,
		user.PubKey, user.PrivateID,
	)

	if err := r.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to update user pubkey: %w", err)
	}

	return nil
}

func (r *AccountCassandraRepository) Delete(user entity.User) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)
	batch.Query(`
		DELETE FROM users_by_id WHERE user_id = ?`,
		user.ID,
	)
	batch.Query(`
		DELETE FROM users_by_private_id WHERE private_id = ?`,
		user.PrivateID,
	)
	batch.Query(`
	DELETE FROM messages WHERE to_user = ?`,
		user.ID,
	)

	if err := r.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
