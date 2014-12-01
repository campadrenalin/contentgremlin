package cgdb

import (
	"database/sql"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type ContentGremlinDB struct {
	*sql.DB
}

func Open(directory string) (ContentGremlinDB, error) {
	db, err := sql.Open("sqlite3", path.Join(directory, "db.sqlite"))
	return ContentGremlinDB{db}, err
}

func (db ContentGremlinDB) Init() error {
	_, err := db.Exec(`
		CREATE TABLE version_history (
			version INTEGER NOT NULL,
			timestamp STRING
		);
		INSERT INTO version_history VALUES (1, datetime());
	`)
	return err
}

func (db ContentGremlinDB) GetVersion() (version int, err error) {
	err = db.QueryRow(`
		SELECT version FROM version_history
		ORDER BY timestamp DESC
		LIMIT 1
	`).Scan(&version)
	return
}
