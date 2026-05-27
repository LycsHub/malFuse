package schema

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DBExec interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type MaliciousPackage struct {
	Name      string
	Version   string
	Ecosystem string
	Published string
	Source    string
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func OpenReadOnly(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open database read-only: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database not accessible: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS malicious_packages (
		name TEXT NOT NULL,
		version TEXT,
		ecosystem TEXT NOT NULL,
		published TEXT NOT NULL DEFAULT '',
		source TEXT NOT NULL DEFAULT '',
		UNIQUE(name, ecosystem, version)
	);
	CREATE INDEX IF NOT EXISTS idx_malicious_packages_lookup ON malicious_packages(name, ecosystem, version);

	CREATE TABLE IF NOT EXISTS update_state (
		ecosystem TEXT NOT NULL UNIQUE,
		last_commit TEXT NOT NULL DEFAULT '',
		last_updated TEXT NOT NULL DEFAULT ''
	);
	`
	_, err := db.Exec(ddl)
	return err
}

func InsertOrReplace(db DBExec, p MaliciousPackage) error {
	var version interface{}
	if p.Version == "" {
		version = nil
	} else {
		version = p.Version
	}

	if p.Version == "" {
		db.Exec(`DELETE FROM malicious_packages WHERE name=? AND ecosystem=? AND version IS NULL`,
			p.Name, p.Ecosystem)
	} else {
		db.Exec(`DELETE FROM malicious_packages WHERE name=? AND ecosystem=? AND version=?`,
			p.Name, p.Ecosystem, p.Version)
	}

	_, err := db.Exec(
		`INSERT INTO malicious_packages (name, version, ecosystem, published, source)
		 VALUES (?, ?, ?, ?, ?)`,
		p.Name, version, p.Ecosystem, p.Published, p.Source,
	)
	return err
}

func Delete(db DBExec, name, ecosystem, version string) error {
	if version == "" {
		_, err := db.Exec(
			`DELETE FROM malicious_packages WHERE name=? AND ecosystem=? AND version IS NULL`,
			name, ecosystem,
		)
		return err
	}
	_, err := db.Exec(
		`DELETE FROM malicious_packages WHERE name=? AND ecosystem=? AND version=?`,
		name, ecosystem, version,
	)
	return err
}

func Lookup(db DBExec, name, ecosystem, version string) (bool, error) {
	var count int
	var err error
	if version == "" {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM malicious_packages
			 WHERE name=? AND ecosystem=?`,
			name, ecosystem,
		).Scan(&count)
	} else {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM malicious_packages
			 WHERE name=? AND ecosystem=? AND (version=? OR version IS NULL)`,
			name, ecosystem, version,
		).Scan(&count)
	}
	if err != nil {
		return false, fmt.Errorf("lookup query: %w", err)
	}
	return count > 0, nil
}

func GetUpdateState(db DBExec, ecosystem string) (lastCommit, lastUpdated string, err error) {
	err = db.QueryRow(
		`SELECT last_commit, last_updated FROM update_state WHERE ecosystem=?`,
		ecosystem,
	).Scan(&lastCommit, &lastUpdated)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return
}

func SetUpdateState(db DBExec, ecosystem, lastCommit, lastUpdated string) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO update_state (ecosystem, last_commit, last_updated)
		 VALUES (?, ?, ?)`,
		ecosystem, lastCommit, lastUpdated,
	)
	return err
}
