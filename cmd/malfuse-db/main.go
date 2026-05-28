package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"malFuse/internal/config"
	"malFuse/internal/db/output"
	"malFuse/internal/db/schema"
	"malFuse/internal/logger"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	mode := flag.String("mode", "direct", "update mode: direct or sql")
	dbPath := flag.String("db", "malfuse.db", "path to SQLite database")
	sqlOutput := flag.String("output", "", "SQL output file path (sql mode, default: updates-YYYYMMDDHHmm.sql)")
	repoDir := flag.String("repo", "ossf-malicious-packages", "path to git repo directory")
	flag.Parse()

	logger.Init(logger.Config{Level: "info", Format: "text", Output: "stdout"})

	if *mode != "direct" && *mode != "sql" {
		logger.Fatal("invalid mode", "mode", *mode)
	}

	if *mode == "sql" && *sqlOutput == "" {
		*sqlOutput = "updates-" + time.Now().Format("200601021504") + ".sql"
	}

	repoProxy := ""
	cfg, err := loadConfig(*configPath)
	if err == nil {
		repoProxy = cfg.RepoProxy
	}

	var db *sql.DB
	if *mode == "direct" {
		db, err = schema.Open(*dbPath)
		if err != nil {
			logger.Fatal("failed to open database", "error", err)
		}
		defer db.Close()
	}

	if err := output.RunUpdate(db, *repoDir, *mode, *sqlOutput, repoProxy); err != nil {
		logger.Fatal("update failed", "error", err)
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
