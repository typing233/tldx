package checker

import (
	"time"
)

type Result struct {
	Domain    string
	Available bool
	Error     error
	Duration  time.Duration
}

type Stats struct {
	Total       int
	Available   int
	Unavailable int
	Errors      int
	Duration    time.Duration
}
