CREATE TABLE IF NOT EXISTS todo (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at  TEXT DEFAULT CURRENT_TIMESTAMP,
	updated_at  TEXT DEFAULT CURRENT_TIMESTAMP,
	description TEXT NOT NULL,
	completed   BOOL DEFAULT FALSE,
  due_date    TEXT,
	priority    INTEGER DEFAULT 5 NOT NULL CHECK (priority IN (1, 2, 3, 4, 5)),
	status      TEXT DEFAULT 'pending'
);
