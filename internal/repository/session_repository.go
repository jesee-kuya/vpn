package repository

import (
	"sync"

	"p2nova-vpn/internal/domain"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (r *SessionRepository) Store(session *domain.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.SessionID] = session
}

func (r *SessionRepository) Get(sessionID string) *domain.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sessions[sessionID]
}

func (r *SessionRepository) Update(session *domain.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.SessionID] = session
}

func (r *SessionRepository) GetActiveSession() *domain.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.Connected {
			return session
		}
	}
	return nil
}

func (r *SessionRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}
