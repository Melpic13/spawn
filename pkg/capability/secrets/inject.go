package secrets

import "context"

// Injection defines target env var + secret source.
type Injection struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// Inject resolves and returns env-style secret map.
func Inject(ctx context.Context, resolver Resolver, items []Injection) (map[string]string, error) {
	out := make(map[string]string, len(items))
	for _, item := range items {
		val, err := resolver.Resolve(ctx, item.Source)
		if err != nil {
			return nil, err
		}
		out[item.Name] = val
	}
	return out, nil
}
