package fs

// OverlayConfig controls read/write layer composition.
type OverlayConfig struct {
	LowerDir string
	UpperDir string
	WorkDir  string
}
