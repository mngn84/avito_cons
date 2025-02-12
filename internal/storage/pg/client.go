package pg

import (
	"database/sql"
	"log/slog"

	_ "github.com/lib/pq"

	"github.com/mngn84/avito-cons/internal/config"
)

type PgClient struct {
	db  *sql.DB
	logger *slog.Logger
}

func NewPgClient(cfg *config.Config, logger *slog.Logger) (*PgClient, error) {
	db, err := sql.Open("postgres", cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	return &PgClient{
		db:  db,
		logger: logger,
	}, nil
}

func (c *PgClient) GetMessages(limit int, chatId string) ([]GptMsg, error) {
	c.logger.Info("GetMessages", "chatId", chatId)

	query := `SELECT content, role
    FROM messages
     WHERE chat_id = $1
     ORDER BY created_at DESC
     LIMIT $2`

	rows, err := c.db.Query(query, chatId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c.logger.Info("GetMessages", "rows", rows)

	var messages []GptMsg
	for rows.Next() {
		var msg GptMsg
		err := rows.Scan(
			&msg.Content,
			&msg.Role,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	c.logger.Info("GetMessages", "messages", messages)

	return messages, nil
}

func (c *PgClient) SaveMsgPair(userMsg DbRow, gptMsg DbRow) error {
	c.logger.Info("SaveMsgPair", "userMsg", userMsg, "gptMsg", gptMsg)
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO messages (chat_id, user_id, content, role, created_at)
    VALUES ($1, $2, $3, $4, TO_TIMESTAMP($5))`

	_, err = tx.Exec(
		query,
		userMsg.ChatId,
		userMsg.UserId,
		userMsg.Content,
		userMsg.Role,
		userMsg.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		query,
		gptMsg.ChatId,
		gptMsg.UserId,
		gptMsg.Content,
		gptMsg.Role,
		gptMsg.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	c.logger.Info("SaveMsgPair", "tx", tx)
	return tx.Commit()
}

func (c *PgClient) DB() *sql.DB {
	return c.db
}
