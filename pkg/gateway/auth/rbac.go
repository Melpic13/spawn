package auth

import "fmt"

// RBAC authorizes actions by role.
type RBAC struct {
	Roles map[string]map[string]bool
}

// Allow checks whether role can execute action.
func (r RBAC) Allow(role, action string) error {
	if perms, ok := r.Roles[role]; ok {
		if perms[action] {
			return nil
		}
	}
	return fmt.Errorf("rbac: role %q not allowed for %q", role, action)
}
