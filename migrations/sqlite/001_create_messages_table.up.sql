CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    chat_id TEXT NOT NULL,
    content TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP
);
