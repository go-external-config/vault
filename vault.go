package vault

import (
	config "github.com/go-external-config/go/env"
	"github.com/go-external-config/vault/env"
)

func init() {
	config.RegisterPropertySource(env.NewVaultPropertySource())
}
