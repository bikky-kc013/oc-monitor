package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bikky/oc-monitor/internal/registry"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	Path string
	reg  *registry.Client
}

func Open() (*DB, error) {
	dbPath, err := resolvePath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"opencode db not found at %s\nrun opencode at least once, or set OPENCODE_DB", dbPath)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(2)
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("could not open %s: %w", dbPath, err)
	}
	return &DB{conn: conn, Path: dbPath}, nil
}

func resolvePath() (string, error) {
	if v := os.Getenv("OPENCODE_DB"); v != "" && v != ":memory:" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		p := filepath.Join(xdg, "opencode", "opencode.db")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return filepath.Join(home, ".local", "share", "opencode", "opencode.db"), nil
}

func (d *DB) Close() { d.conn.Close() }
