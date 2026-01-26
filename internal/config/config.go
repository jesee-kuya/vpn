package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"p2nova-vpn/internal/geo"
)

type Config struct {
	Port             string
	ServerPublicKey  string
	ServerPrivateKey string
	ServerEndpoint   string
	WGInterface      string
	WGPort           string
	VPNSubnet        string
	DNSServers       string
	Servers          []ServerConfig
}

type ServerConfig struct {
	Code string
	Name string
	IP   string
	Flag string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		ServerPrivateKey: getEnv("SERVER_PRIVATE_KEY", ""),
		ServerPublicKey:  getEnv("SERVER_PUBLIC_KEY", ""),
		ServerEndpoint:   getEnv("SERVER_ENDPOINT", ""),
		WGInterface:      getEnv("WG_INTERFACE", "wg0"),
		WGPort:           getEnv("WG_PORT", "51820"),
		VPNSubnet:        getEnv("VPN_SUBNET", "10.8.0.0/24"),
		DNSServers:       getEnv("DNS_SERVERS", "1.1.1.1, 8.8.8.8"),
	}

	// Get geo info for server configuration
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		serverIP = cfg.ServerEndpoint // fallback
	}

	geoInfo, err := geo.GetServerGeo(serverIP)
	if err != nil {
		geoInfo = &geo.GeoInfo{
			Country: "Unknown",
			City:    "Unknown",
			IP:      serverIP,
		}
	}

	cfg.Servers = []ServerConfig{
		{
			Code: geoInfo.Country,
			Name: geoInfo.City + " VPN",
			IP:   geoInfo.IP,
			Flag: getCountryFlag(geoInfo.Country),
		},
	}

	// Validate and auto-detect WireGuard keys if missing
	if err := cfg.validateAndSetupKeys(); err != nil {
		return nil, err
	}

	// Validate critical fields
	if cfg.ServerEndpoint == "" {
		return nil, fmt.Errorf("SERVER_ENDPOINT must be set (your server's public IP)")
	}

	if cfg.ServerPublicKey == "" {
		return nil, fmt.Errorf("server public key not found - set SERVER_PUBLIC_KEY env var or ensure WireGuard is running")
	}

	return cfg, nil
}

// validateAndSetupKeys checks if keys exist and tries to auto-detect them
func (cfg *Config) validateAndSetupKeys() error {
	// If public key is not set, try to get it from WireGuard interface
	if cfg.ServerPublicKey == "" {
		pubKey, err := getServerPublicKeyFromInterface(cfg.WGInterface)
		if err == nil && pubKey != "" {
			cfg.ServerPublicKey = pubKey
			fmt.Printf("✓ Auto-detected server public key from %s\n", cfg.WGInterface)
		} else {
			// Try to derive from private key if available
			if cfg.ServerPrivateKey != "" {
				pubKey, err := derivePublicKey(cfg.ServerPrivateKey)
				if err == nil {
					cfg.ServerPublicKey = pubKey
					fmt.Printf("✓ Derived public key from private key\n")
				}
			}
		}
	}

	return nil
}

// getServerPublicKeyFromInterface retrieves the public key from running WireGuard interface
func getServerPublicKeyFromInterface(interfaceName string) (string, error) {
	// Method 1: Get from running interface
	cmd := exec.Command("wg", "show", interfaceName, "public-key")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	// Method 2: Get from private key file (if exists)
	privateKeyPath := fmt.Sprintf("/etc/wireguard/%s.key", interfaceName)
	if _, err := os.Stat(privateKeyPath); err == nil {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("cat %s | wg pubkey", privateKeyPath))
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return strings.TrimSpace(string(output)), nil
		}
	}

	// Method 3: Parse config file
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", interfaceName)
	if _, err := os.Stat(configPath); err == nil {
		cmd := exec.Command("sh", "-c", 
			fmt.Sprintf("grep '^PrivateKey' %s | awk '{print $3}' | wg pubkey", configPath))
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return strings.TrimSpace(string(output)), nil
		}
	}

	return "", fmt.Errorf("could not retrieve server public key from interface %s", interfaceName)
}

// derivePublicKey derives public key from private key using wg command
func derivePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to derive public key: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getCountryFlag returns the flag emoji for a country code
func getCountryFlag(countryCode string) string {
	flags := map[string]string{
		"KE": "🇰🇪",
		"US": "🇺🇸",
		"GB": "🇬🇧",
		"UK": "🇬🇧",
		"DE": "🇩🇪",
		"FR": "🇫🇷",
		"SG": "🇸🇬",
		"JP": "🇯🇵",
		"CA": "🇨🇦",
		"AU": "🇦🇺",
		"NL": "🇳🇱",
	}
	
	if flag, ok := flags[strings.ToUpper(countryCode)]; ok {
		return flag
	}
	return "🌍"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}