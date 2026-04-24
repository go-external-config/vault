# VaultPropertySource

Vault provides centralized, well-audited privileged access and secret management for mission-critical data whether you deploy systems on-premises, in the cloud, or in a hybrid environment.

With a modular design based around a growing plugin ecosystem, Vault lets you integrate with your existing systems and customize your application workflow. ([more](https://developer.hashicorp.com/vault/docs/about-vault/what-is-vault))

cmd/app/main.go

    import (
    	"github.com/go-external-config/go/env"
    	vault "github.com/go-external-config/vault/env"
    )

    var _ = env.Instance().WithPropertySource(vault.NewVaultPropertySource())

    func main() {
    	defer err.Recover()
    	fmt.Println(env.Value[string]("${db.pass}"))
    }

config/application.yaml

    db:
    	pass: VAULT:my-secret-password:password
    	# pass: VAULT:secret:my-secret-password:password

    vault:
    	address: http://127.0.0.1:8200
    	token: generated
    	# mount: secret
