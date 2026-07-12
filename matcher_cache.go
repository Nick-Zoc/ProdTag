package main

import "ProdTag/internal/core"

func buildMatcherCache(config AppConfig) MatcherCache { return core.BuildMatcherCache(config) }
func writeMatcherCache(path string, config AppConfig) error {
	return core.WriteMatcherCache(path, config)
}
func ensureMatcherCache(path string, config AppConfig) error {
	return core.EnsureMatcherCache(path, config)
}
func readMatcherCache(path string) (MatcherCache, error) { return core.ReadMatcherCache(path) }
