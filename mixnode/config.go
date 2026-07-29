package mixnode

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Servers        []string `json:"servers"`
	Clients        []string `json:"clients"`
	PathLen        int      `json:"path_len"`
	BatchSize      int      `json:"batch_size"`
	ClientTickUs   int      `json:"client_tick_us"`
	FlushTimeoutMs int      `json:"flush_timeout_ms"`
	CryptoType     string   `json:"crypto_type"`
}

const (
	defaultBatchSize      = 10
	defaultClientTickUs   = 200000
	defaultFlushTimeoutMs = 1
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadConfig: " + err.Error())
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.New("LoadConfig: " + err.Error())
	}
	if len(cfg.Servers) == 0 {
		return nil, errors.New("LoadConfig: no servers in config")
	}
	if len(cfg.Clients) == 0 {
		return nil, errors.New("LoadConfig: no clients in config")
	}
	if cfg.PathLen <= 0 {
		cfg.PathLen = 3
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.ClientTickUs <= 0 {
		cfg.ClientTickUs = defaultClientTickUs
	}
	if cfg.FlushTimeoutMs <= 0 {
		cfg.FlushTimeoutMs = defaultFlushTimeoutMs
	}
	if cfg.CryptoType == "" {
		cfg.CryptoType = "ecc"
	}
	return &cfg, nil
}
