package version

import "spawn.dev/internal/buildinfo"

// Info returns the current build version.
func Info() string {
	return buildinfo.Version
}
