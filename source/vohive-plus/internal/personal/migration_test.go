package personal

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyMigrationPreservesSourceAndMessages(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "legacy #?.db")
	destination := filepath.Join(dir, "personal.db")
	connection, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: source}).String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = connection.Exec(`
        CREATE TABLE sms (id integer PRIMARY KEY AUTOINCREMENT, imsi text, iccid text, peer text, local_phone text, sender text, recipient text, content text, type integer, status integer, timestamp datetime, created_at datetime);
        INSERT INTO sms VALUES (7,'test-imsi','890000001','10086','','10086','','fixture message',1,1,'2026-07-01 12:00:00','2026-07-01 12:00:00');
        CREATE TABLE devices (imei text PRIMARY KEY, alias text);
        INSERT INTO devices VALUES ('fixture-device','My modem');
    `)
	if err != nil {
		t.Fatal(err)
	}
	connection.Close()
	original, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	var report bytes.Buffer
	if err := MigrateDatabase(source, destination, &report); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(original) != sha256.Sum256(after) {
		t.Fatal("source database changed")
	}
	target, err := sql.Open("sqlite", destination)
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	var content, alias string
	if err := target.QueryRow("SELECT content FROM sms WHERE id=7").Scan(&content); err != nil {
		t.Fatal(err)
	}
	if err := target.QueryRow("SELECT alias FROM devices WHERE imei='fixture-device'").Scan(&alias); err != nil {
		t.Fatal(err)
	}
	if content != "fixture message" || alias != "My modem" {
		t.Fatal("legacy records lost")
	}
	if err := MigrateDatabase(source, destination, &report); err == nil {
		t.Fatal("overwrote destination")
	}
	if err := MigrateDatabase(source, source, &report); err == nil {
		t.Fatal("allowed in-place migration")
	}
}

func TestMigrationRejectsCorruptSource(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "broken.db")
	if err := os.WriteFile(source, []byte("not a database"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := MigrateDatabase(source, filepath.Join(dir, "copy.db"), &bytes.Buffer{}); err == nil {
		t.Fatal("accepted corrupt source")
	}
}
