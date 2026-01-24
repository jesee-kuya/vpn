package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os/exec"

	"p2nova-vpn/internal/config"

	"golang.org/x/crypto/curve25519"
)

type WireguardService struct {
	config *config.Config
	wg     string
}

func NewWireguardService(cfg *config.Config) *WireguardService {
	return &WireguardService{
		config: cfg,
		wg:     cfg.WGInterface, // e.g., "wg0"
	}
}

func (s *WireguardService) generateKeys() (privateKey, publicKey string, err error) {
	// Generate private key
	var private [32]byte
	if _, err := rand.Read(private[:]); err != nil {
		return "", "", err
	}

	privateKey = base64.StdEncoding.EncodeToString(private[:])

	// Derive public key
	var public [32]byte
	curve25519.ScalarBaseMult(&public, &private)
	publicKey = base64.StdEncoding.EncodeToString(public[:])

	return privateKey, publicKey, nil
}

func (s *WireguardService) AddPeer(clientIP string) (peerConfig string, publicKey string, err error) {
	// Generate client keys
	privateKey, pubKey, err := s.generateKeys()
	if err != nil {
		return "", "", err
	}

	// Add peer to WireGuard using wg command
	cmd := exec.Command("wg", "set", s.wg,
		"peer", pubKey,
		"allowed-ips", fmt.Sprintf("%s/32", clientIP))

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("failed to add peer: %s - %v", output, err)
	}

	// Generate client config
	peerConfig = s.generatePeerConfig(privateKey, clientIP)

	return peerConfig, pubKey, nil
}

func (s *WireguardService) RemovePeer(publicKey string) error {
	cmd := exec.Command("wg", "set", s.wg, "peer", publicKey, "remove")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove peer: %s - %v", output, err)
	}
	return nil
}

func (s *WireguardService) generatePeerConfig(privateKey, clientIP string) string {
	// This is the complete config the client needs
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s:%s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25`,
		privateKey,
		clientIP,
		s.config.DNSServers, // e.g., "1.1.1.1, 8.8.8.8"
		s.config.ServerPublicKey,
		s.config.ServerEndpoint, // Your VPN server's public IP
		s.config.WGPort,         // Usually 51820
	)
}
