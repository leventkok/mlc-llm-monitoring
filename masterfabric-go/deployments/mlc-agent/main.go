package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type switchRequest struct {
	ID           string `json:"id"`
	ProfileID    string `json:"profile_id"`
	RequestModel string `json:"request_model"`
	EngineModel  string `json:"engine_model"`
	LocalAdapter string `json:"local_adapter"`
	Status       string `json:"status"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("mlc-agent: ")

	apiBase := strings.TrimRight(envOr("API_BASE_URL", "https://mlc-llm-monitoring.onrender.com"), "/")
	apiKey := strings.TrimSpace(os.Getenv("MLC_API_KEY"))
	if apiKey == "" {
		log.Fatal("MLC_API_KEY is required")
	}
	deployDir := envOr("DEPLOYMENTS_DIR", "/deployments")
	envFile := filepath.Join(deployDir, envOr("ENV_FILE", ".env.hybrid"))
	poll := envDuration("POLL_INTERVAL", 15*time.Second)

	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("polling %s every %s (deploy dir %s)", apiBase, poll, deployDir)

	for {
		if err := tick(client, apiBase, apiKey, deployDir, envFile); err != nil {
			log.Printf("tick error: %v", err)
		}
		time.Sleep(poll)
	}
}

func tick(client *http.Client, apiBase, apiKey, deployDir, envFile string) error {
	req, err := http.NewRequest(http.MethodGet, apiBase+"/agent/model-switch/next", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-MLC-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("claim API %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Request *switchRequest `json:"request"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	if payload.Request == nil {
		return nil
	}
	sw := payload.Request
	log.Printf("switching to profile %s (%s)", sw.ProfileID, sw.EngineModel)

	if err := applyEnvFile(envFile, sw.EngineModel, sw.LocalAdapter); err != nil {
		report(client, apiBase, apiKey, sw.ID, "failed", err.Error())
		return err
	}

	if err := restartEngine(deployDir, envFile); err != nil {
		report(client, apiBase, apiKey, sw.ID, "failed", err.Error())
		return err
	}

	return report(client, apiBase, apiKey, sw.ID, "completed", "")
}

func applyEnvFile(path, engineModel, localAdapter string) error {
	lines, err := readLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lines = upsertEnv(lines, "MLC_MODEL", engineModel)
	if strings.TrimSpace(localAdapter) == "" {
		lines = removeEnvKey(lines, "MLC_LORA_ADAPTER")
	} else {
		lines = upsertEnv(lines, "MLC_LORA_ADAPTER", localAdapter)
	}
	return writeLines(path, lines)
}

func restartEngine(deployDir, envFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	args := []string{
		"compose",
		"-f", "docker-compose.hybrid.yml",
		"-f", "docker-compose.hybrid.real-mlc.yml",
		"--env-file", filepath.Base(envFile),
		"up", "-d", "--force-recreate", "mlc-engine", "mlc-llm",
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = deployDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("running: docker %s (cwd=%s)", strings.Join(args, " "), deployDir)
	return cmd.Run()
}

func report(client *http.Client, apiBase, apiKey, id, status, errMsg string) error {
	payload, _ := json.Marshal(map[string]string{
		"status":        status,
		"error_message": errMsg,
	})
	req, err := http.NewRequest(http.MethodPost, apiBase+"/agent/model-switch/"+id+"/report", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MLC-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("report API %d: %s", resp.StatusCode, string(b))
	}
	log.Printf("switch %s → %s", id, status)
	return nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func writeLines(path string, lines []string) error {
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	if len(lines) > 0 {
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func upsertEnv(lines []string, key, value string) []string {
	prefix := key + "="
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + value
			return lines
		}
	}
	return append(lines, prefix+value)
}

func removeEnvKey(lines []string, key string) []string {
	prefix := key + "="
	out := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			continue
		}
		out = append(out, line)
	}
	return out
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	if n, err := time.ParseDuration(raw + "s"); err == nil {
		return n
	}
	return fallback
}
