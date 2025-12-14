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
	log.Println("[db-migrate] Done.")
}
