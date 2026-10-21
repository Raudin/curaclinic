// Package clock provides an injectable time source so booking rules can be
// tested without sleeping.
package clock

import (
	"time"

	// Embed the IANA timezone database so containers without tzdata still
	// resolve CLINIC_TIMEZONE and doctor timezones.
	_ "time/tzdata"
)

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}

// Real is the system clock.
var Real Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// Fake is a controllable clock for tests.
type Fake struct {
	Time time.Time
}

// NewFake returns a fake clock pinned to the supplied time.
func NewFake(t time.Time) *Fake { return &Fake{Time: t.UTC()} }

// Now returns the pinned time.
func (f *Fake) Now() time.Time { return f.Time.UTC() }

// Advance moves the fake clock forward.
func (f *Fake) Advance(d time.Duration) { f.Time = f.Time.Add(d) }
