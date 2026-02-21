package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// VaultResolver resolves vault:// refs against HashiCorp Vault HTTP API.
type VaultResolver struct {
	Address string
	Token   string
	Client  *http.Client
}

// Resolve resolves vault:// refs.
func (v VaultResolver) Resolve(ctx context.Context, ref string) (string, error) {
	if strings.TrimSpace(v.Address) == "" {
		return "", fmt.Errorf("resolve vault secret: address is required")
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("resolve vault secret: parse ref: %w", err)
	}
	if u.Scheme != "vault" {
		return "", fmt.Errorf("resolve vault secret: invalid scheme %q", u.Scheme)
	}
	secretPath := path.Clean(path.Join(u.Host, u.Path))
	secretPath = strings.TrimPrefix(secretPath, "/")
	if secretPath == "" || secretPath == "." {
		return "", fmt.Errorf("resolve vault secret: secret path is required")
	}
	secretKey := strings.TrimPrefix(u.Fragment, "#")
	if secretKey == "" {
		secretKey = "value"
	}

	token := strings.TrimSpace(v.Token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("VAULT_TOKEN"))
	}
	if token == "" {
		return "", fmt.Errorf("resolve vault secret: VAULT_TOKEN is required")
	}

	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	endpoint := strings.TrimSuffix(v.Address, "/") + "/v1/" + secretPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("resolve vault secret: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("resolve vault secret: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("resolve vault secret: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("resolve vault secret: decode response: %w", err)
	}
	values := extractVaultData(payload)
	if values == nil {
		return "", fmt.Errorf("resolve vault secret: response missing data")
	}
	if value, ok := values[secretKey]; ok {
		return fmt.Sprintf("%v", value), nil
	}
	for _, value := range values {
		return fmt.Sprintf("%v", value), nil
	}
	return "", fmt.Errorf("resolve vault secret: no values in secret")
}

func extractVaultData(payload map[string]interface{}) map[string]interface{} {
	root, _ := payload["data"].(map[string]interface{})
	if root == nil {
		return nil
	}
	// KV v2 nests secret fields in data.data.
	if nested, ok := root["data"].(map[string]interface{}); ok {
		return nested
	}
	// KV v1 stores fields in data.
	return root
}
