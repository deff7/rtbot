package bot

import (
	"sync"

	"github.com/deff7/rutracker/internal/rutracker"
)

type RTCollection interface {
	ListNext() []rutracker.TorrentFile
	HasNext() bool
	Len() int
}

type Session struct {
	Results RTCollection
}

type SessionManager interface {
	Get(int) Session
	Set(int, Session)
}

type sessionManager struct {
	mu            sync.RWMutex
	usersSessions map[int]Session
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		usersSessions: map[int]Session{},
	}
}

func (sm *sessionManager) Get(userID int) Session {
	sm.mu.RLock()
	s := sm.usersSessions[userID]
	sm.mu.RUnlock()
	return s
}

func (sm *sessionManager) Set(userID int, s Session) {
	sm.mu.Lock()
	sm.usersSessions[userID] = s
	sm.mu.Unlock()
}
