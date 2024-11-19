package scheduler

// BackpressurePolicy controls queue behavior under load.
type BackpressurePolicy struct {
	MaxQueueDepth int
	DropWhenFull  bool
}
