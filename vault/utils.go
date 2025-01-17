package vault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func connect(ctx context.Context, d *plugin.QueryData) (*api.Client, error) {
	addr := os.Getenv("VAULT_ADDR")
	tkn := os.Getenv("VAULT_TOKEN")

	vaultConfig := GetConfig(d.Connection)
	if &vaultConfig != nil {
		if vaultConfig.Address != nil {
			addr = *vaultConfig.Address
		}
		if vaultConfig.Token != nil {
			tkn = *vaultConfig.Token
		}
	}

	if addr == "" {
		return nil, errors.New("Vault Address must be set either in VAULT_ADDR environment variable or in connection configuration file.")
	}

	if tkn == "" {
		return nil, errors.New("Vault Token must be set either in VAULT_TOKEN environment variable or in connection configuration file.")
	}

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := api.NewClient(&api.Config{Address: addr, HttpClient: httpClient})

	if err != nil {
		return nil, errors.New(err.Error())
	}

	client.SetToken(tkn)

	return client, nil
}

// Util func to replace any double / with single ones, used to make concatting paths easier
func replaceDoubleSlash(url string) string {
	return strings.ReplaceAll(url, "//", "/")
}

// Util func to obtain filtered mounts from all mounts
func filterMounts(in map[string]*api.MountOutput, mountType string) map[string]*api.MountOutput {
	filtered := map[string]*api.MountOutput{}

	for key, mount := range in {
		if mount.Type == mountType {
			filtered[key] = mount
		}
	}

	return filtered
}

// Util func to obtain []string by key from map[string]interface
func getValues(in map[string]interface{}, key string) []string {
	if in[key] == nil {
		return []string{}
	}

	var out []string
	for _, s := range in[key].([]interface{}) {
		out = append(out, fmt.Sprintf("%s", s.(string)))
	}

	return out
}

// Converts and api.Secret object into a slice of strings containing all secret paths
func getSecretAsStrings(s *api.Secret) []string {
	if s == nil || s.Data["keys"] == nil || len(s.Data["keys"].([]interface{})) == 0 {
		return []string{}
	}
	var secrets []string
	for _, s := range s.Data["keys"].([]interface{}) {
		secrets = append(secrets, fmt.Sprintf("%s", s.(string)))

	}
	return secrets
}

// Transforms
func convertTimestamp(_ context.Context, input *transform.TransformData) (interface{}, error) {
	return time.Unix(input.Value.(int64), 0), nil
}
