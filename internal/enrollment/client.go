package enrollment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/gpu"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Request struct {
	EnrollmentToken string            `json:"enrollmentToken"`
	AgentVersion    string            `json:"agentVersion"`
	Inventory       gpu.HostInventory `json:"inventory"`
}

type Result struct {
	HostID          string `json:"hostId"`
	AgentCredential string `json:"agentCredential"`
}

func (c Client) Enroll(ctx context.Context, token, agentVersion string, inventory gpu.HostInventory) (Result, error) {
	if strings.TrimSpace(token) == "" {
		return Result{}, errors.New("enrollment token is required; set VUN_ENROLLMENT_TOKEN or use --token")
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return Result{}, errors.New("VUN API URL is not configured")
	}

	body, err := json.Marshal(Request{EnrollmentToken: token, AgentVersion: agentVersion, Inventory: inventory})
	if err != nil {
		return Result{}, fmt.Errorf("encode enrollment request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/enrollment/host", bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("create enrollment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("send enrollment request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var message struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&message)
		if message.Error == "" {
			message.Error = resp.Status
		}
		return Result{}, fmt.Errorf("enrollment API returned %s: %s", resp.Status, message.Error)
	}
	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Result{}, fmt.Errorf("decode enrollment response: %w", err)
	}
	if result.HostID == "" || result.AgentCredential == "" {
		return Result{}, errors.New("enrollment response missing hostId or agentCredential")
	}
	return result, nil
}
