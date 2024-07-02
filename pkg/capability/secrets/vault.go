package secrets

import "context"

// VaultResolver is a minimal placeholder vault resolver.
type VaultResolver struct {
	Address string
}

// Resolve resolves vault:// refs.
func (v VaultResolver) Resolve(_ context.Context, ref string) (string, error) {
	return "vault-secret:" + ref + "@" + v.Address, nil
}
