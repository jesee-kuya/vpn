package service

import (
	"p2nova-vpn/internal/config"
	"p2nova-vpn/internal/domain"
	"p2nova-vpn/internal/repository"
)

type ServerService struct {
	serverRepo *repository.ServerRepository
	config     *config.Config
}

func NewServerService(serverRepo *repository.ServerRepository, cfg *config.Config) *ServerService {
	// Initialize servers from config
	for _, srv := range cfg.Servers {
		serverRepo.Store(&domain.Server{
			Code: srv.Code,
			Name: srv.Name,
			IP:   srv.IP,
		})
	}

	return &ServerService{
		serverRepo: serverRepo,
		config:     cfg,
	}
}

func (s *ServerService) ListServers() ([]*domain.Server, error) {
	return s.serverRepo.List(), nil
}

func (s *ServerService) GetServer(code string) (*domain.Server, error) {
	server := s.serverRepo.Get(code)
	if server == nil {
		return nil, domain.ErrServerNotFound
	}
	return server, nil
}
