// Package platform implements runtime-backed ports.
package platform

import "time"

// SystemClock implements port.Clock with the wall clock.
type SystemClock struct{}

// Now returns time.Now.
func (SystemClock) Now() time.Time { return time.Now() }
