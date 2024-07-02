package fs

// Mount defines a virtual mount point.
type Mount struct {
	Path   string
	Source string
	Mode   string
	Quota  string
}
