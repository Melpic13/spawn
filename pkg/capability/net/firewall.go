package net

// Policy defines basic allow/deny rules.
type Policy struct {
	Allowlist []string
	Denylist  []string
}
