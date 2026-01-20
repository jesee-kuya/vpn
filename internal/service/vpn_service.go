package service

import (
	"fmt"
	"math/rand"
	"time"

	"p2nova-vpn/internal/config"
	"p2nova-vpn/internal/domain"
	"p2nova-vpn/internal/repository"
)

type VPNService struct {
	sessionRepo *repository.SessionRepository
	wgService   *WireguardService
	config      *config.Config
}

func NewVPNService(sessionRepo *repository.SessionRepository, wgService *WireguardService, cfg *config.Config) *VPNService {
	return &VPNService{
		sessionRepo: sessionRepo,
		wgService:   wgService,
		config:      cfg,
	}
}

func (s *VPNService) Connect(serverCode string) (*domain.Session, error) {
	// Check if already connected
	if active := s.sessionRepo.GetActiveSession(); active != nil {
		return nil, domain.ErrAlreadyConnected
	}

	// Allocate IP for client
	clientIP := s.allocateIP()

	// Create WireGuard peer
	peerConfig, clientKey, err := s.wgService.AddPeer(clientIP)
	if err != nil {
		return nil, fmt.Errorf("failed to add peer: %w", err)
	}

	// Create session
	session := domain.NewSession(serverCode, clientIP, peerConfig, clientKey)
	s.sessionRepo.Store(session)

	return session, nil
}

func (s *VPNService) Disconnect(sessionID string) error {
	session := s.sessionRepo.Get(sessionID)
	if session == nil {
		return domain.ErrSessionNotFound
	}

	// Remove WireGuard peer
	if err := s.wgService.RemovePeer(session.ClientKey); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	session.Connected = false
	session.EndTime = time.Now().Unix()
	s.sessionRepo.Update(session)

	return nil
}

func (s *VPNService) GetStatus() (*domain.VPNStatus, error) {
	session := s.sessionRepo.GetActiveSession()

	if session == nil {
		return &domain.VPNStatus{Connected: false}, nil
	}

	duration := time.Now().Unix() - session.StartTime

	return &domain.VPNStatus{
		Connected: true,
		Server:    session.ServerCode,
		Duration:  duration,
		IP:        session.ClientIP,
	}, nil
}

func (s *VPNService) GetSpeed() *domain.SpeedTest {
	// Simulate speed test - replace with actual implementation
	return &domain.SpeedTest{
		Download: 20.0 + rand.Float64()*30.0,
		Upload:   5.0 + rand.Float64()*15.0,
		Latency:  20 + rand.Intn(50),
	}
}

func (s *VPNService) allocateIP() string {
	// Simple IP allocation - in production, use proper IPAM
	return fmt.Sprintf("10.8.0.%d", 2+s.sessionRepo.Count())
}
