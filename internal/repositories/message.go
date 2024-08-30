package repositories

import (
	"fmt"
	"pipe/internal/entity"

	"github.com/gocql/gocql"
)

var _ Message = &MessageCassandraRepository{}

type MessageCassandraRepository struct {
	*CassandraCommonBehaviour
}

func NewMessageCassandraRepository(session *gocql.Session) *MessageCassandraRepository {
	return &MessageCassandraRepository{
		NewCassandraCommonBehaviour(session),
	}
}

func (m *MessageCassandraRepository) ByUserID(ID int64) ([]entity.Message, error) {
	messages := []entity.Message{}
	iter := m.session.Query(`SELECT message_id, text, date 
	FROM messages WHERE to_user = ? ORDER BY date DESC LIMIT 100`, ID).Iter()
	var message entity.Message
	for iter.Scan(&message.ID, &message.Text, &message.Date) {
		messages = append(messages, message)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *MessageCassandraRepository) DeleteAllByUserID(ID int64) error {
	if err := m.session.Query(`DELETE FROM messages WHERE to_user = ?`, ID).Exec(); err != nil {
		return fmt.Errorf("failed to delete all message: %w", err)
	}
	return nil
}

func (m *MessageCassandraRepository) Send(message entity.Message) error {
	batch := m.session.NewBatch(gocql.LoggedBatch)
	batch.Query(`
		INSERT INTO messages (message_id, from_user, to_user, text, date) VALUES (?, ?, ?, ?, ?) USING TTL 1800`,
		message.ID, message.FromUser, message.ToUser, message.Text, message.Date,
	)

	if err := m.session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
