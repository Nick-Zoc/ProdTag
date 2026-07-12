package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func BuildMatcherCache(config AppConfig) MatcherCache {
	events, broad := map[string]bool{}, map[string]bool{}
	candidates := make([]MatcherCacheCandidate, 0, len(config.Rules))
	for _, rule := range config.Rules {
		if !rule.Enabled {
			continue
		}
		event := strings.TrimSpace(rule.EventType)
		if event == "" {
			continue
		}
		events[event] = true
		mode := strings.TrimSpace(rule.MatchMode)
		if mode == "" {
			mode = "any"
		}
		pattern := strings.TrimSpace(rule.CommandPattern)
		if mode == "any" || pattern == "" {
			broad[event] = true
		}
		candidates = append(candidates, MatcherCacheCandidate{RuleID: rule.ID, EventType: event, Pattern: pattern, MatchMode: mode, HasExitCode: rule.ExitCode != nil})
	}
	return MatcherCache{Version: ConfigVersion, Complete: true, EnabledEventTypes: sortedKeys(events), BroadEventTypes: sortedKeys(broad), Candidates: candidates, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
}
func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func WriteMatcherCache(path string, config AppConfig) error {
	if err := WriteJSON(path, BuildMatcherCache(config)); err != nil {
		return fmt.Errorf("write matcher cache: %w", err)
	}
	return nil
}
func EnsureMatcherCache(path string, config AppConfig) error {
	cache, err := ReadMatcherCache(path)
	if err == nil && cache.UpdatedAt >= config.UpdatedAt {
		return nil
	}
	return WriteMatcherCache(path, config)
}
func ReadMatcherCache(path string) (MatcherCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MatcherCache{}, err
	}
	var cache MatcherCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return MatcherCache{}, err
	}
	if cache.Version != ConfigVersion || !cache.Complete || cache.EnabledEventTypes == nil {
		return MatcherCache{}, fmt.Errorf("unsupported matcher cache version %d", cache.Version)
	}
	return cache, nil
}
