package database

import (
	"database/sql"
	"caddy-manager/internal/config"
	
	_ "modernc.org/sqlite"
)

var db *sql.DB

func Init() error {
	var err error
	db, err = sql.Open("sqlite", config.DatabasePath)
	if err != nil {
		return err
	}

	// 创建表
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		target TEXT NOT NULL,
		ssl_enabled BOOLEAN DEFAULT 1,
		environment TEXT DEFAULT '',
		php_version TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		project_type TEXT NOT NULL,
		root_dir TEXT NOT NULL,
		exec_path TEXT,
		port INTEGER,
		start_command TEXT,
		auto_start BOOLEAN DEFAULT 0,
		status TEXT DEFAULT 'stopped',
		domains TEXT,
		ssl_enabled BOOLEAN DEFAULT 1,
		ssl_email TEXT,
		reverse_proxy_path TEXT,
		extra_headers TEXT,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		command TEXT NOT NULL,
		schedule TEXT NOT NULL,
		is_loop BOOLEAN DEFAULT 0,
		status TEXT DEFAULT 'waiting',
		last_run DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	// 插入默认设置
	db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('security_path', '')")
	db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('www_root', 'C:\\www')")
	
	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func GetDB() *sql.DB {
	return db
}

func IsFirstRun() bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return err != nil || count == 0
}
