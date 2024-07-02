package fs

import "time"

// Snapshot holds metadata for filesystem snapshots.
type Snapshot struct {
	ID        string
	CreatedAt time.Time
	Path      string
}
