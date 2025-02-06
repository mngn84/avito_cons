package pg

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mngn84/avito-cons/internal/config"
)

type SqliteClient struct {
	db  *sql.DB
	log *slog.Logger
}

func NewSqliteClient(cfg *config.Config, logger *slog.Logger) (*SqliteClient, error) {
	db, err := sql.Open("sqlite3", "database.db") // Файл базы SQLite
	if err != nil {
		return nil, err
	}

	client := &SqliteClient{
		db:  db,
		log: logger,
	}

	// Создание таблицы, если её нет
	if err := client.initSchema(); err != nil {
		return nil, err
	}

	return client, nil
}

// initSchema проверяет и создаёт таблицу, если её нет
func (c *SqliteClient) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		chat_id TEXT NOT NULL,
		content TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);
	`
	_, err := c.db.Exec(query)
	if err != nil {
		c.log.Error("Failed to initialize database schema", "error", err)
		return err
	}

	c.log.Info("Database schema initialized successfully")
	return nil
}

func (c *SqliteClient) GetMessages(limit int, chatId string) ([]GptMsg, error) {
	query := `SELECT content, role FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT ?`

	rows, err := c.db.Query(query, chatId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []GptMsg
	for rows.Next() {
		var msg GptMsg
		err := rows.Scan(&msg.Content, &msg.Role)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (c *SqliteClient) SaveMsgPair(userMsg DbRow, gptMsg DbRow) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO messages (chat_id, user_id, content, role, created_at) VALUES (?, ?, ?, ?, ?)`

	_, err = tx.Exec(query, userMsg.ChatId, userMsg.UserId, userMsg.Content, userMsg.Role, userMsg.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(query, gptMsg.ChatId, gptMsg.UserId, gptMsg.Content, gptMsg.Role, gptMsg.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *SqliteClient) DB() *sql.DB {
	return c.db
}
