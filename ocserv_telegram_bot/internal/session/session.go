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

// Session holds per-chat conversational state (a tiny in-memory state machine).
// Each session is wiped after the configured TTL so abandoned flows do not
// leak credentials in memory.
type Session struct {
	State              State
	UpdatedAt          time.Time
	BufferUsername string
	BufferDesired  string
	BufferTargetID uint
	BufferPackage  uint
	UsernameAttemptsAt []time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
	ttl      time.Duration
}

func NewStore(ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Store{
		sessions: make(map[int64]*Session),
		ttl:      ttl,
	}
}

func (s *Store) Get(chatID int64) *Session {
	s.mu.RLock()
	sess, ok := s.sessions[chatID]
	s.mu.RUnlock()
	if !ok {
		return &Session{State: Idle}
	}
	if time.Since(sess.UpdatedAt) > s.ttl {
		s.mu.Lock()
		delete(s.sessions, chatID)
		s.mu.Unlock()
		return &Session{State: Idle}
	}
	// Return a value copy so callers can mutate without racing other goroutines
	// that read/write the same map entry between Get and Set.
	c := *sess
	return &c
}

func (s *Store) Set(chatID int64, sess *Session) {
	if sess == nil {
		return
	}
	sess.UpdatedAt = time.Now()
	s.mu.Lock()
	s.sessions[chatID] = sess
	s.mu.Unlock()
}

func (s *Store) Reset(chatID int64) {
	s.mu.Lock()
	delete(s.sessions, chatID)
	s.mu.Unlock()
}

// RegisterAttempt records a new username/password attempt. It returns false
// when the rate limit (5 attempts in 10 minutes) is exceeded.
func (s *Store) RegisterAttempt(chatID int64) bool {
	const window = 10 * time.Minute
	const maxAttempts = 5

	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[chatID]
	if !ok {
		sess = &Session{State: Idle, UpdatedAt: time.Now()}
		s.sessions[chatID] = sess
	}

	now := time.Now()
	cutoff := now.Add(-window)
	pruned := sess.UsernameAttemptsAt[:0]
	for _, t := range sess.UsernameAttemptsAt {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	sess.UsernameAttemptsAt = pruned
	if len(sess.UsernameAttemptsAt) >= maxAttempts {
		return false
	}
	sess.UsernameAttemptsAt = append(sess.UsernameAttemptsAt, now)
	sess.UpdatedAt = now
	return true
}
