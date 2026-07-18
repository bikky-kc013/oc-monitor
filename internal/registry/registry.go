// Package registry provides dynamic model registry lookup backed by
// models.dev. It fetches model metadata (including context window sizes)
// and caches the result locally for offline use.
package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	fetchURL     = "https://models.dev/api.json"
	fetchTimeout = 5 * time.Second
	CacheTTL     = 24 * time.Hour
)

// apiProvider is the shape of each top-level entry in the models.dev JSON.
type apiProvider struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Models map[string]apiModel `json:"models"`
}

// apiModel is a single model within a provider.
type apiModel struct {
	ID    string `json:"id"`
	Limit struct {
		Context int `json:"context"`
		Input   int `json:"input,omitempty"`
		Output  int `json:"output,omitempty"`
	} `json:"limit"`
}

// modelEntry is a flattened registry record.
type modelEntry struct {
	Provider string
	ModelID  string
	Input    int // max input tokens (from limit.input)
	Context  int // context window (from limit.context)
}

// maxInput returns the best max input token estimate for this model.
func (e modelEntry) maxInput() int {
	if e.Input > 0 {
		return e.Input
	}
	return e.Context
}

// Client holds the parsed model registry and provides thread-safe lookups.
type Client struct {
	mu        sync.RWMutex
	entries   []modelEntry
	fetchedAt time.Time
}

// NewClient creates an empty Client. Call LoadFromDisk or Fetch before lookups.
func NewClient() *Client {
	return &Client{}
}

// Fetch downloads the registry from models.dev, parses it, and replaces the
// in-memory data. On success it also writes to disk. It returns an error if
// the request fails, but callers may silently ignore errors (existing data
// is preserved).
func (c *Client) Fetch() error {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(fetchURL)
	if err != nil {
		return fmt.Errorf("fetch registry: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch registry: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fetch registry: read body: %w", err)
	}
	entries, err := ParseJSON(body)
	if err != nil {
		return fmt.Errorf("fetch registry: parse: %w", err)
	}
	c.mu.Lock()
	c.entries = entries
	c.fetchedAt = time.Now()
	c.mu.Unlock()

	// best-effort cache write (ignore error)
	_ = SaveToDisk(body)

	return nil
}

// LoadFromDisk reads the cache file and parses it into the client.
// It also restores the fetchedAt timestamp from the file's mtime.
// Returns nil (no error) if the file doesn't exist — callers should
// treat an empty client as "no data yet".
func (c *Client) LoadFromDisk() error {
	path, err := cachePath()
	if err != nil {
		return nil // can't determine cache path, skip silently
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // file doesn't exist or unreadable, skip
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	entries, err := ParseJSON(data)
	if err != nil {
		return nil // corrupted cache, skip
	}
	c.mu.Lock()
	c.entries = entries
	c.fetchedAt = info.ModTime()
	c.mu.Unlock()
	return nil
}

// Stale reports whether the loaded data is older than CacheTTL.
func (c *Client) Stale() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fetchedAt.IsZero() {
		return true
	}
	return time.Since(c.fetchedAt) > CacheTTL
}

// ParseJSON decodes models.dev JSON into flat modelEntry slices.
func ParseJSON(data []byte) ([]modelEntry, error) {
	var raw map[string]apiProvider
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var entries []modelEntry
	for _, prov := range raw {
		for _, m := range prov.Models {
			entries = append(entries, modelEntry{
				Provider: prov.ID,
				ModelID:  m.ID,
				Input:    m.Limit.Input,
				Context:  m.Limit.Context,
			})
		}
	}
	return entries, nil
}

// SaveToDisk writes raw JSON bytes to the cache file atomically.
func SaveToDisk(data []byte) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// cachePath returns the OS cache location: <UserCacheDir>/oc-monitor/models.json
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "oc-monitor", "models.json"), nil
}
