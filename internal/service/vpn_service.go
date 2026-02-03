package service

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"p2nova-vpn/internal/config"
	"p2nova-vpn/internal/domain"
	"p2nova-vpn/internal/repository"
)

type VPNService struct {
	sessionRepo *repository.SessionRepository
	wgService   *WireguardService
	config      *config.Config
	ipPool      *IPPool
}

type IPPool struct {
	mu        sync.Mutex
	network   *net.IPNet
	allocated map[string]bool
	lastIP    net.IP
}

func NewIPPool(cidr string) (*IPPool, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	// Start from network address
	startIP := network.IP.Mask(network.Mask)

	// Skip .0 (network) and .1 (server/gateway)
	// Start allocation from .2
	startIP = nextIP(startIP) // .0 -> .1
	startIP = nextIP(startIP) // .1 -> .2

	return &IPPool{
		network:   network,
		allocated: make(map[string]bool),
		lastIP:    startIP, // Now starts from .2
	}, nil
}
func (p *IPPool) Allocate() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Start from the next IP after network address
	ip := make(net.IP, len(p.lastIP))
	copy(ip, p.lastIP)

	// Try to find an available IP
	for i := 0; i < 253; i++ { // Avoid broadcast
		ip = nextIP(ip)

		if !p.network.Contains(ip) {
			return "", fmt.Errorf("IP pool exhausted")
		}

		ipStr := ip.String()
		if !p.allocated[ipStr] {
			p.allocated[ipStr] = true
			p.lastIP = ip
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no available IPs")
}

func (p *IPPool) Release(ipStr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.allocated, ipStr)
}

func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for j := len(next) - 1; j >= 0; j-- {
		next[j]++
		if next[j] != 0 {
			break
		}
	}

	return next
}

func NewVPNService(sessionRepo *repository.SessionRepository, wgService *WireguardService, cfg *config.Config) *VPNService {
	ipPool, err := NewIPPool(cfg.VPNSubnet) // e.g., "10.8.0.0/24"
	if err != nil {
		panic(err)
	}

	return &VPNService{
		sessionRepo: sessionRepo,
		wgService:   wgService,
		config:      cfg,
		ipPool:      ipPool,
	}
}

func (s *VPNService) Connect(serverCode string) (*domain.Session, error) {
	// 1. Check for any active session
	if active := s.sessionRepo.GetActiveSession(); active != nil {
		// found an existing connection, kill it (Disconnect)
		// This handles cleanup of WireGuard peers and IP release
		if err := s.Disconnect(active.SessionID); err != nil {
			return nil, fmt.Errorf("failed to disconnect existing session: %w", err)
		}
	}

	// 2. Proceed with new connection logic
	// Allocate IP for client
	clientIP, err := s.ipPool.Allocate()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Create WireGuard peer
	peerConfig, clientKey, err := s.wgService.AddPeer(clientIP)
	if err != nil {
		s.ipPool.Release(clientIP)
		return nil, fmt.Errorf("failed to add WireGuard peer: %w", err)
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
		return err
	}

	s.ipPool.Release(session.ClientIP)

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
