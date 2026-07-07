package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	handledEventLogFileName = "handled-events.jsonl"
	handledEventLogLimit    = 50
)

func (a *App) ListHandledEventLog() ([]RecentEventRecord, error) {
	return readHandledEventLog(handledEventLogLimit)
}

func appendHandledEventLog(record RecentEventRecord) error {
	if err := EnsureAppData(); err != nil {
		return err
	}

	path, err := handledEventLogPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create event log folder: %w", err)
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode handled event log record: %w", err)
	}
	data = append(data, '\n')

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open handled event log: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write handled event log: %w", err)
	}
	return nil
}

func readHandledEventLog(limit int) ([]RecentEventRecord, error) {
	path, err := handledEventLogPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []RecentEventRecord{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open handled event log: %w", err)
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
			return nil, fmt.Errorf("parse handled event log: %w", err)
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
	for index := len(records) - 1; index >= 0 && len(recent) < limit; index-- {
		recent = append(recent, records[index])
	}
	return recent, nil
}

func handledEventLogPath() (string, error) {
	paths, err := GetAppDataPaths()
	if err != nil {
		return "", err
	}
	return filepath.Join(paths.LogsDir, handledEventLogFileName), nil
}
