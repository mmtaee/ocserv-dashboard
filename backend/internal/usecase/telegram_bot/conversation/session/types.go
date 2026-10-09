package session

import (
	"sync"
	"time"
)

// State enumerates the possible multi-step flow states. Anything not listed
// here is treated as Idle (i.e. waiting for a top-level command).
type State int

const (
	Idle State = iota
	WaitingUsernameForLink
	WaitingPasswordForLink
	WaitingUsernameForNew
	WaitingPackageForNew
	WaitingNoteForNew
	WaitingPackageForRenew
	WaitingNoteForRenew
)

// Session holds per-chat conversational state. Each session is wiped after
// the configured TTL so abandoned flows do not retain credentials in memory.
type Session struct {
	State              State
	UpdatedAt          time.Time
	BufferUsername     string
	BufferDesired      string
	BufferTargetID     uint
	BufferPackage      uint
	UsernameAttemptsAt []time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
	ttl      time.Duration
}
