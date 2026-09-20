package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Enrollment struct {
	HostID          string `json:"hostId"`
	AgentCredential string `json:"agentCredential"`
}

func SaveEnrollment(enrollment Enrollment) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	data, err := json.MarshalIndent(enrollment, "", "  ")
	if err != nil {
		return fmt.Errorf("encode enrollment state: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write enrollment state: %w", err)
	}
	return nil
}

func Path() (string, error) {
	if path := os.Getenv("VUN_STATE_PATH"); path != "" {
		return path, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve state directory: %w", err)
	}
	return filepath.Join(configDir, "vun", "state.json"), nil
}
