package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const BakiDir = ".baki"
const SnapshotsDir = "snapshots"
const BackupsDir = "backups"

type Config struct {
	Device       string `json:"device"`
	CreatedAt    string `json:"created_at"`
	SnapshotNext int    `json:"snapshot_next"`
}

type Snapshot struct {
	ID         int         `json:"id"`
	Message    string      `json:"message"`
	Timestamp  string      `json:"timestamp"`
	Partitions []Partition `json:"partitions"`
}

func InitBaki() error {
	if _, err := os.Stat(BakiDir); err == nil {
		return fmt.Errorf(".baki already exists in this directory")
	}

	device, _ := adbExec("shell", "getprop", "ro.product.model")
	device = sanitize(device)

	dirs := []string{
		BakiDir,
		filepath.Join(BakiDir, SnapshotsDir),
		filepath.Join(BakiDir, BackupsDir),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	cfg := Config{
		Device:       device,
		CreatedAt:    time.Now().Format(time.RFC3339),
		SnapshotNext: 1,
	}
	return saveConfig(cfg)
}

func LoadBaki() (*Config, error) {
	if _, err := os.Stat(BakiDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a baki project (run 'baki init' first)")
	}
	return loadConfig()
}

func saveConfig(cfg Config) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(filepath.Join(BakiDir, "config.json"), data, 0644)
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(filepath.Join(BakiDir, "config.json"))
	if err != nil {
		return nil, err
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return &cfg, nil
}

func SaveSnapshot(snap Snapshot) error {
	cfg, err := LoadBaki()
	if err != nil {
		return err
	}

	snap.ID = cfg.SnapshotNext
	snap.Timestamp = time.Now().Format(time.RFC3339)

	data, _ := json.MarshalIndent(snap, "", "  ")
	filename := filepath.Join(BakiDir, SnapshotsDir, fmt.Sprintf("%d.json", snap.ID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	cfg.SnapshotNext++
	return saveConfig(*cfg)
}

func LoadSnapshot(id int) (*Snapshot, error) {
	filename := filepath.Join(BakiDir, SnapshotsDir, fmt.Sprintf("%d.json", id))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	json.Unmarshal(data, &snap)
	return &snap, nil
}

func ListSnapshots() ([]Snapshot, error) {
	dir := filepath.Join(BakiDir, SnapshotsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var snaps []Snapshot
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var snap Snapshot
		json.Unmarshal(data, &snap)
		snaps = append(snaps, snap)
	}
	return snaps, nil
}

func LatestSnapshot() (*Snapshot, error) {
	snaps, err := ListSnapshots()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, fmt.Errorf("no snapshots yet")
	}
	latest := snaps[0]
	for _, s := range snaps[1:] {
		if s.ID > latest.ID {
			latest = s
		}
	}
	return &latest, nil
}

func DeleteSnapshot(id int) error {
	files := []string{
		filepath.Join(BakiDir, SnapshotsDir, fmt.Sprintf("%d.json", id)),
		filepath.Join(BakiDir, BackupsDir, fmt.Sprintf("%d", id)),
	}
	for _, f := range files {
		os.RemoveAll(f)
	}
	return nil
}

func sanitize(s string) string {
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
