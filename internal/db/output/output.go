package output

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"malFuse/internal/db/ingest"
	"malFuse/internal/db/schema"
)

func UpsertPackages(db *sql.DB, pkgs []ingest.ParsedPackage) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, p := range pkgs {
		mp := schema.MaliciousPackage{
			Name:      p.Name,
			Version:   p.Version,
			Ecosystem: p.Ecosystem,
			Published: p.Published,
			Source:    p.ID,
		}
		if err := schema.InsertOrReplace(tx, mp); err != nil {
			return fmt.Errorf("insert %s: %w", p.Name, err)
		}
	}

	return tx.Commit()
}

func DeletePackages(db *sql.DB, pkgs []ingest.ParsedPackage) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, p := range pkgs {
		if err := schema.Delete(tx, p.Name, p.Ecosystem, p.Version); err != nil {
			return fmt.Errorf("delete %s: %w", p.Name, err)
		}
	}

	return tx.Commit()
}

func GenerateInsertSQL(pkgs []ingest.ParsedPackage) string {
	var b strings.Builder
	for _, p := range pkgs {
		version := "NULL"
		if p.Version != "" {
			version = fmt.Sprintf("'%s'", escapeSQL(p.Version))
		}
		b.WriteString(fmt.Sprintf(
			"INSERT OR REPLACE INTO malicious_packages (name, version, ecosystem, published, source) VALUES ('%s', %s, '%s', '%s', '%s');\n",
			escapeSQL(p.Name), version, escapeSQL(p.Ecosystem), escapeSQL(p.Published), escapeSQL(p.ID),
		))
	}
	return b.String()
}

func GenerateDeleteSQL(pkgs []ingest.ParsedPackage) string {
	var b strings.Builder
	for _, p := range pkgs {
		if p.Version == "" {
			b.WriteString(fmt.Sprintf(
				"DELETE FROM malicious_packages WHERE name='%s' AND ecosystem='%s' AND version IS NULL;\n",
				escapeSQL(p.Name), escapeSQL(p.Ecosystem),
			))
		} else {
			b.WriteString(fmt.Sprintf(
				"DELETE FROM malicious_packages WHERE name='%s' AND ecosystem='%s' AND version='%s';\n",
				escapeSQL(p.Name), escapeSQL(p.Ecosystem), escapeSQL(p.Version),
			))
		}
	}
	return b.String()
}

func WriteSQLFile(path string, inserts, deletes []ingest.ParsedPackage) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	f.WriteString("-- malFuse DB incremental update\n")
	if len(deletes) > 0 {
		f.WriteString(GenerateDeleteSQL(deletes))
	}
	if len(inserts) > 0 {
		f.WriteString(GenerateInsertSQL(inserts))
	}
	return nil
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
