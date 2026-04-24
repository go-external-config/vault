package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-errr/go/err"
	"github.com/go-external-config/go/env"
	"github.com/go-external-config/go/lang"
	"github.com/go-external-config/go/util/optional"
	vault "github.com/hashicorp/vault/api"
)

const VAULT = "VAULT:"

type VaultPropertySource struct {
	environment *env.Environment
	client      *vault.Client
}

func NewVaultPropertySource() *VaultPropertySource {
	ps := &VaultPropertySource{}
	ps.environment = env.Instance()
	ps.client = ps.newClient()
	return ps
}

func (this *VaultPropertySource) Name() string {
	return "VaultPropertySource"
}

func (this *VaultPropertySource) HasProperty(key string) bool {
	for _, source := range this.environment.PropertySources() {
		if source.Properties() != nil && source.HasProperty(key) {
			return strings.HasPrefix(source.Property(key), VAULT)
		}
	}
	return false
}

func (this *VaultPropertySource) Property(key string) string {
	for _, source := range this.environment.PropertySources() {
		if source.Properties() != nil && source.HasProperty(key) {
			value := fmt.Sprint(this.environment.ResolveRequiredPlaceholders(source.Property(key)))
			parts := strings.Split(value, ":")
			lang.AssertState(len(parts) == 3 || len(parts) == 4, "Expected %smount:secret:key, got %s", VAULT, value)
			if len(parts) == 3 {
				mount := fmt.Sprint(this.environment.ResolveRequiredPlaceholders("${vault.mount:secret}"))
				return this.getSecretValue(mount, parts[1], parts[2])
			}
			return this.getSecretValue(parts[1], parts[2], parts[3])
		}
	}
	panic(err.NewIllegalArgumentException("No value present for " + key))
}

func (this *VaultPropertySource) getSecretValue(mount, secretPath, key string) string {
	secret := optional.OfCommaErr(this.client.KVv2(mount).Get(context.Background(), secretPath)).OrElsePanic("Unable to get secret")
	value, ok := secret.Data[key]
	lang.AssertState(ok, "No %s present in %s", key, secretPath)
	result, ok := value.(string)
	lang.AssertState(ok, "Value type assertion failed: %T %#v", value, value)
	return result
}

func (this *VaultPropertySource) newClient() *vault.Client {
	config := vault.DefaultConfig()
	config.Address = this.environment.Property("vault.address")
	client := optional.OfCommaErr(vault.NewClient(config)).OrElsePanic("Unable to initialize Vault client")
	client.SetToken(this.environment.Property("vault.token"))
	return client
}

func (this *VaultPropertySource) Properties() map[string]string {
	return nil
}
