package net

import "time"

// RateLimit controls outgoing request budget.
type RateLimit struct {
	Requests int
	Per      time.Duration
}
