package bot

import "sync"

type State int

const (
	StateStart State = iota
	StateWaitQuery
	StateWaitCommand
)

type StateManager interface {
	Get(int) State
	Set(int, State)
}

type stateManager struct {
	mu         sync.RWMutex
	usersState map[int]State
}

func newStateManager() *stateManager {
	return &stateManager{
		usersState: map[int]State{},
	}
}

func (sm *stateManager) Get(userID int) State {
	sm.mu.RLock()
	s := sm.usersState[userID]
	sm.mu.RUnlock()
	return s
}

func (sm *stateManager) Set(userID int, state State) {
	sm.mu.Lock()
	sm.usersState[userID] = state
	sm.mu.Unlock()
}
