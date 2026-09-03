package database

import (
	"database/sql"
	"log"
	_ "embed"
)

//go:embed schema.sql
var schema string

func Migrate(db *sql.DB) {
	log.Println("[db-migrate] Running migrations.")
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("[db-migrate] Migrations failed: %v", err)
	}

	// Ensure project_id column exists if table was created before this migration
	rows, err := db.Query("PRAGMA table_info(todo)")
	if err == nil {
		defer rows.Close()
		hasProjectID := false
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue sql.NullString
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
				if name == "project_id" {
					hasProjectID = true
					break
				}
			}
		}
		if !hasProjectID {
			if _, err := db.Exec("ALTER TABLE todo ADD COLUMN project_id TEXT DEFAULT 'todos' NOT NULL"); err != nil {
				log.Printf("[db-migrate] Alter table error: %v", err)
			}
		}
	}

	log.Println("[db-migrate] Done.")
}
