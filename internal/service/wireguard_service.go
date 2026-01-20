package service

import (
	"fmt"
	"os/exec"
	"strings"

	"p2nova-vpn/internal/config"
	"p2nova-vpn/pkg/wireguard"
)

type WireguardService struct {
	config *config.Config
	wg     *wireguard.Interface
}

func NewWireguardService(cfg *config.Config) *WireguardService {
	wg := wireguard.NewInterface(cfg.WireGuardIface)

	svc := &WireguardService{
		config: cfg,
		wg:     wg,
	}

	// Initialize WireGuard interface
	if err := svc.initialize(); err != nil {
		fmt.Printf("Warning: WireGuard initialization failed: %v\n", err)
	}

	return svc
}

func (s *WireguardService) initialize() error {
	// Generate keys if needed
	if s.config.ServerPrivateKey == "" {
		privateKey, publicKey, err := s.generateKeys()
		if err != nil {
			return err
		}
		s.config.ServerPrivateKey = privateKey
		s.config.ServerPublicKey = publicKey
	}

	// Setup interface
	return s.wg.Setup(s.config.ServerPrivateKey, s.config.WireGuardPort, s.config.NetworkCIDR)
}

func (s *WireguardService) AddPeer(clientIP string) (peerConfig string, publicKey string, err error) {
	// Generate client keys
	privateKey, pubKey, err := s.generateKeys()
	if err != nil {
		return "", "", err
	}

	// Add peer to WireGuard
	if err := s.wg.AddPeer(pubKey, clientIP); err != nil {
		return "", "", err
	}

	// Generate client config
	peerConfig = s.generatePeerConfig(privateKey, s.config.ServerPublicKey, clientIP)

	return peerConfig, pubKey, nil
}

func (s *WireguardService) RemovePeer(publicKey string) error {
	return s.wg.RemovePeer(publicKey)
}

func (s *WireguardService) generateKeys() (privateKey, publicKey string, err error) {
	cmd := exec.Command("wg", "genkey")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(output))

	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(output))

	return privateKey, publicKey, nil
}

func (s *WireguardService) generatePeerConfig(privateKey, serverPublicKey, clientIP string) string {
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = %s
PersistentKeepalive = 25`,
		privateKey,
		clientIP,
		strings.Join(s.config.DNSServers, ", "),
		serverPublicKey,
		s.config.Servers[0].IP,
		s.config.WireGuardPort,
		s.config.AllowedIPs,
	)
}
