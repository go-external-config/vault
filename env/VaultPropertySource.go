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

const VAULT_KEY_PREFIX = "vault."
const VAULT_VALUE_PREFIX = "vault:"
const vault_addr = "vault.addr"
const vault_token = "vault.token"
const vault_mount = "vault.mount"

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
	if strings.HasPrefix(key, VAULT_KEY_PREFIX) {
		switch key {
		case vault_addr:
			return false
		case vault_token:
			return false
		case vault_mount:
			return false
		default:
			return true
		}
	}
	for _, source := range this.environment.PropertySources() {
		if source.Properties() != nil && source.HasProperty(key) {
			return strings.HasPrefix(source.Property(key), VAULT_VALUE_PREFIX)
		}
	}
	return false
}

func (this *VaultPropertySource) Property(key string) string {
	if strings.HasPrefix(key, VAULT_KEY_PREFIX) {
		return this.resolveVaultProperty(fmt.Sprint(this.environment.ResolveRequiredPlaceholders(key[len(VAULT_KEY_PREFIX):])))
	}
	for _, source := range this.environment.PropertySources() {
		if source.Properties() != nil && source.HasProperty(key) {
			return this.resolveVaultProperty(fmt.Sprint(this.environment.ResolveRequiredPlaceholders(source.Property(key)[len(VAULT_VALUE_PREFIX):])))
		}
	}
	panic(err.NewIllegalArgumentException("No value present for " + key))
}

func (this *VaultPropertySource) resolveVaultProperty(property string) string {
	var mount, path, key string
	if strings.Contains(property, ":") {
		mount, path, _ = strings.Cut(property, ":")
	} else {
		mount = fmt.Sprint(this.environment.ResolveRequiredPlaceholders("${vault.mount:secret}"))
		path = property
	}
	path, key, found := strings.Cut(path, "#")
	lang.Assert(found, "Cannot resolve Vault property, expected path#key, got %s", property)
	return this.getSecretValue(mount, path, key)
}

func (this *VaultPropertySource) getSecretValue(mount, secretPath, key string) string {
	secret := optional.OfCommaErr(this.client.KVv2(mount).Get(context.Background(), secretPath)).OrElsePanic("Unable to get " + secretPath)
	value, ok := secret.Data[key]
	lang.Assert(ok, "No %s present in %s", key, secretPath)
	return fmt.Sprint(value)
}

func (this *VaultPropertySource) newClient() *vault.Client {
	config := vault.DefaultConfig()
	config.Address = this.environment.Property(vault_addr)
	client := optional.OfCommaErr(vault.NewClient(config)).OrElsePanic("Unable to initialize Vault client")
	client.SetToken(this.environment.Property(vault_token))
	return client
}

func (this *VaultPropertySource) Properties() map[string]string {
	return nil
}
