package personal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/yibaiba/hideck/internal/db"
)

// MigrateDatabase snapshots an existing SQLite database through its read-only
// connection, including committed WAL data. All schema migrations target the
// new copy; this command never opens modem ports or starts background services.
func MigrateDatabase(source, destination string, output io.Writer) error {
	src, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	dst, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if src == dst {
		return fmt.Errorf("source and destination must differ")
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		return fmt.Errorf("destination must not exist")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	sourceDB, err := sql.Open("sqlite", sqliteReadOnlyURI(src))
	if err != nil {
		return err
	}
	defer sourceDB.Close()
	sourceDB.SetMaxOpenConns(1)
	if err := integrityCheck(sourceDB); err != nil {
		return fmt.Errorf("source integrity: %w", err)
	}
	if _, err := sourceDB.Exec("VACUUM INTO ?", dst); err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	if err := os.Chmod(dst, 0600); err != nil {
		return err
	}
	snapshotDB, err := sql.Open("sqlite", sqliteReadOnlyURI(dst))
	if err != nil {
		return err
	}
	before, err := tableCounts(snapshotDB)
	snapshotDB.Close()
	if err != nil {
		return err
	}
	if err := db.Init(dst); err != nil {
		return fmt.Errorf("migrate copy: %w", err)
	}
	targetDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	defer targetDB.Close()
	if err := integrityCheck(targetDB); err != nil {
		return fmt.Errorf("target integrity: %w", err)
	}
	after, err := tableCounts(targetDB)
	if err != nil {
		return err
	}
	for _, table := range []string{"devices", "sms"} {
		if before[table] != after[table] {
			return fmt.Errorf("%s count changed: %d -> %d; keep the legacy service", table, before[table], after[table])
		}
	}
	return json.NewEncoder(output).Encode(map[string]any{
		"mode": "offline_migration", "source_unchanged": true,
		"integrity": "ok", "before": before, "after": after,
	})
}

func sqliteReadOnlyURI(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: "mode=ro"}).String()
}

func integrityCheck(connection *sql.DB) error {
	var result string
	if err := connection.QueryRow("PRAGMA quick_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("%s", result)
	}
	return nil
}

func tableCounts(connection *sql.DB) (map[string]int64, error) {
	counts := map[string]int64{}
	for _, table := range []string{"devices", "sms", "sms_contacts", "card_policies", "sim_cards", "proxy_instances"} {
		var exists int
		if err := connection.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			continue
		}
		var count int64
		if err := connection.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			return nil, err
		}
		counts[table] = count
	}
	return counts, nil
}
