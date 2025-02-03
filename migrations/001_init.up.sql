CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    user_id int NOT NULL,
    chat_id TEXT NOT NULL,
    content TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP 
);
