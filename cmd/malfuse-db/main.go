package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"malFuse/internal/config"
	"malFuse/internal/db/output"
	"malFuse/internal/db/schema"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	mode := flag.String("mode", "direct", "update mode: direct or sql")
	dbPath := flag.String("db", "malfuse.db", "path to SQLite database")
	sqlOutput := flag.String("output", "", "SQL output file path (sql mode, default: updates-YYYYMMDDHHmm.sql)")
	repoDir := flag.String("repo", "ossf-malicious-packages", "path to git repo directory")
	flag.Parse()

	if *mode != "direct" && *mode != "sql" {
		log.Fatalf("invalid mode: %s (must be 'direct' or 'sql')", *mode)
	}

	if *mode == "sql" && *sqlOutput == "" {
		*sqlOutput = "updates-" + time.Now().Format("200601021504") + ".sql"
	}

	log.SetFlags(0)

	repoProxy := ""
	cfg, err := loadConfig(*configPath)
	if err == nil {
		repoProxy = cfg.RepoProxy
	}

	var db *sql.DB
	if *mode == "direct" {
		db, err = schema.Open(*dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
	}

	if err := output.RunUpdate(db, *repoDir, *mode, *sqlOutput, repoProxy); err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	fmt.Println("Update complete.")
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return config.Load(data)
}
