package devmemory

// Schema is the SQLite database schema for the Dev Studio memory store.
// It uses a standard SQLite database, isolated from the main MakoClaw database.
const Schema = `
CREATE TABLE IF NOT EXISTS dev_memories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	content TEXT NOT NULL,
	metadata JSON,
	embedding BLOB,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
