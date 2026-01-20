package repository

import (
	"sync"

	"p2nova-vpn/internal/domain"
)

type ServerRepository struct {
	mu      sync.RWMutex
	servers map[string]*domain.Server
}

func NewServerRepository() *ServerRepository {
	return &ServerRepository{
		servers: make(map[string]*domain.Server),
	}
}

func (r *ServerRepository) Store(server *domain.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers[server.Code] = server
}

func (r *ServerRepository) Get(code string) *domain.Server {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.servers[code]
}

func (r *ServerRepository) List() []*domain.Server {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := make([]*domain.Server, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}
	return servers
}
