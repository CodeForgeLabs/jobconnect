package clock

import "time"

// RealClock implements application.Clock using system time.
type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
