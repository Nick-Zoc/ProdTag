package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const HandledEventLogFileName = "handled-events.jsonl"
const HandledEventLogLimit = 50
const HandledEventLogMaxLines = 500

func HandledEventLogPath() (string, error) {
	paths, err := GetAppDataPaths()
	if err != nil {
		return "", err
	}
	return filepath.Join(paths.LogsDir, HandledEventLogFileName), nil
}
func AppendHandledEventLog(record RecentEventRecord) error {
	if err := EnsureAppData(); err != nil {
		return err
	}
	path, err := HandledEventLogPath()
	if err != nil {
		return err
	}
	return withFileLock(path+".lock", func() error { return appendAndCap(path, record, HandledEventLogMaxLines) })
}
func appendAndCap(path string, record RecentEventRecord, maxLines int) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return capEventLog(path, maxLines)
}
func withFileLock(path string, action func() error) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = file.Close()
			defer os.Remove(path)
			return action()
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > 10*time.Second {
			_ = os.Remove(path)
			continue
		}
		if time.Now().After(deadline) {
			return errors.New("event log is busy")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func capEventLog(path string, maxLines int) error {
	if maxLines <= 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= maxLines {
		return nil
	}
	retained := strings.Join(lines[len(lines)-maxLines:], "\n") + "\n"
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(retained), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func ReadHandledEventLog(limit int) ([]RecentEventRecord, error) {
	path, err := HandledEventLogPath()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []RecentEventRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	records := []RecentEventRecord{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record RecentEventRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read handled event log: %w", err)
	}
	if limit <= 0 || limit > len(records) {
		limit = len(records)
	}
	recent := make([]RecentEventRecord, 0, limit)
	for i := len(records) - 1; i >= 0 && len(recent) < limit; i-- {
		recent = append(recent, records[i])
	}
	return recent, nil
}
