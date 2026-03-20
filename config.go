package main

import (
	"os"

	scfgs "codeberg.org/lindenii/go-scfgs"
)

type Config struct {
	Discord struct {
		Token       string `scfgs:"token"`
		AdminRoleID uint64 `scfgs:"admin_role_id"`
	} `scfgs:"discord"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := scfgs.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
