package pg

import (
	"database/sql"
	"log/slog"
	"unicode/utf8"

	_ "github.com/lib/pq"

	"github.com/mngn84/avito-cons/internal/config"
)

type PgClient struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPgClient(cfg *config.Config, logger *slog.Logger) (*PgClient, error) {
	db, err := sql.Open("postgres", cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	return &PgClient{
		db:     db,
		logger: logger,
	}, nil
}

func (c *PgClient) DB() *sql.DB {
	return c.db
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

	messages := []GptMsg{}
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

func (c *PgClient) GetAssistantId(userId int) (string, error) {
	c.logger.Info("GetAssistantId", "userId", userId)
	query := `SELECT asst_id FROM assistants WHERE user_id = $1`

	rows, err := c.db.Query(query, userId)
	if err != nil {
		c.logger.Error("GetAssistantId rows", "err", err)
		return "", err
	}
	defer rows.Close()

	asstId := ""
	if rows.Next() {
		err := rows.Scan(&asstId)
		if err != nil {
			c.logger.Error("GetAssistantId asstId", "err", err)
			return "", err
		}
	}

	return asstId, nil
}

func (c *PgClient) SaveAssistant(asstId, asstName string, userId int) error {
	c.logger.Info("SaveAssistantId", "asstId", asstId, "asstName", asstName, "userId", userId)

	query := `INSERT INTO assistants (asst_id, asst_name, user_id) VALUES ($1, $2, $3)`

	result, err := c.db.Exec(query, asstId, asstName, userId)
	if err != nil {
		c.logger.Error("SaveAssistantId", "err", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	c.logger.Info("SaveAssistantId", "rowsAffected", rowsAffected)

	return nil
}

func (c *PgClient) GetThreadId(chatId string) (string, error) {
	c.logger.Info("GetThreadId", "chatId", chatId)

	query := `SELECT thread_id FROM threads WHERE chat_id = $1`

	rows, err := c.db.Query(query, chatId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	threadId := ""
	if rows.Next() {
		err := rows.Scan(&threadId)
		if err != nil {
			return "", err
		}
	}
	return threadId, nil
}

func (c *PgClient) GetUserId(profileName string) (int, error) {
	c.logger.Info("GetUserId", "profileName", profileName)

	query := `SELECT user_id FROM profiles WHERE profile_name = $1`

	rows, err := c.db.Query(query, profileName)
	if err != nil {
		return 0, err
	}

	userId := 0
	if rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			return 0, err
		}
	}

	return userId, nil
}

func (c *PgClient) SaveThreadId(chatId string, threadId, asstId string) error {
	c.logger.Info("SaveThreadId", "chatId", chatId, "threadId", threadId)

	query := `INSERT INTO threads (chat_id, thread_id, asst_id) VALUES ($1, $2, $3)`

	_, err := c.db.Exec(query, chatId, threadId, asstId)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetStoreId(asstId string) (string, error) {
	c.logger.Info("GetStoreId", "asstId", asstId)

	query := `SELECT store_id FROM v_stores WHERE asst_id = $1`

	rows, err := c.db.Query(query, asstId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	storeId := ""
	if rows.Next() {
		err := rows.Scan(&storeId)
		if err != nil {
			return "", err
		}
	}

	c.logger.Info("GetStoreId", "storeId", storeId)
	return storeId, nil
}

func (c *PgClient) SaveStoreRecord(storeId, storeName, asstId string) error {
	c.logger.Info("SaveStoreRecord", "storeId", storeId, "storeName", storeName, "asstId", asstId)
	
	query := `INSERT INTO v_stores (store_id, store_name, asst_id) VALUES ($1, $2, $3)`
	
	result, err := c.db.Exec(query, storeId, storeName, asstId)
	if err != nil {
		c.logger.Error("SaveStoreId", "err", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.logger.Error("SaveStoreId", "err", err)
		return err
	}
	c.logger.Info("SaveStoreRecord", "rowsAffected", rowsAffected)

	return nil
}

func (c *PgClient) SaveFileRecord(fileId, fileName, fileType, storeId string) error {
	c.logger.Info("SaveFileRecord", "storeId", storeId, "fileId", fileId, "fileName", fileName, "fileType", fileType)

	if !utf8.ValidString(fileName) {
		c.logger.Info("SaveFileRecord", "fileName", fileName, "err", "invalid utf8 string")
		fileName = string([]rune(fileName))
	}

	query := `INSERT INTO files (file_id, file_name, file_type, store_id) VALUES ($1, $2, $3, $4)`

	result, err := c.db.Exec(query, fileId, fileName, fileType, storeId)
	if err != nil {
		c.logger.Error("SaveFileRecord", "err", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.logger.Error("SaveFileRecord", "err", err)
		return err
	}
	c.logger.Info("SaveFileRecord", "rowsAffected", rowsAffected)

	return nil
}

func (c *PgClient) GetOldFileId(storeId, fileName, fileType string) (string, error) {
	c.logger.Info("GetOldFile", "storeId", storeId, "fileName", fileName, "fileType", fileType)

	query := `SELECT file_id FROM v_files WHERE store_id = $1 AND file_name = $2 AND file_type = $3`

	rows, err := c.db.Query(query, storeId, fileName, fileType)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	oldFileId := ""
	if rows.Next() {
		err := rows.Scan(&oldFileId)
		if err != nil {
			return "", err
		}
	}
	c.logger.Info("GetOldFile", "oldFileId", oldFileId)
	return oldFileId, nil
}

func (c *PgClient) DeleteOldFile(storeId, fileId string) error {
	c.logger.Info("DeleteOldFile", "storeId", storeId, "fileId", fileId)

	query := `DELETE FROM v_files WHERE store_id = $1 AND file_id = $2`

	_, err := c.db.Exec(query, storeId, fileId)
	if err != nil {
		return err
	}

	return nil
}
