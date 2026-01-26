package config

import (
	"fmt"
	"os"

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
		ServerPublicKey:  getEnv("SERVER_PUBLIC_KEY", ""),
		ServerPrivateKey: getEnv("SERVER_PRIVATE_KEY", ""),
		ServerEndpoint:   getEnv("SERVER_ENDPOINT", ""),
		WGInterface:      getEnv("WG_INTERFACE", "wg0"),
		WGPort:           getEnv("WG_PORT", "51820"),
		VPNSubnet:        getEnv("VPN_SUBNET", "10.8.0.0/24"),
		DNSServers:       getEnv("DNS_SERVERS", "1.1.1.1, 8.8.8.8"),
	}

	// Get geo info for server configuration
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		serverIP = cfg.ServerEndpoint
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

	// Validate critical fields
	if cfg.ServerPublicKey == "" {
		return nil, fmt.Errorf("SERVER_PUBLIC_KEY environment variable is required")
	}

	if cfg.ServerEndpoint == "" {
		return nil, fmt.Errorf("SERVER_ENDPOINT environment variable is required")
	}

	// Log loaded configuration (for debugging)
	fmt.Println("✓ VPN Configuration Loaded:")
	fmt.Printf("  Server Public Key: %s\n", cfg.ServerPublicKey)
	fmt.Printf("  Server Endpoint: %s:%s\n", cfg.ServerEndpoint, cfg.WGPort)
	fmt.Printf("  WG Interface: %s\n", cfg.WGInterface)
	fmt.Printf("  VPN Subnet: %s\n", cfg.VPNSubnet)
	fmt.Printf("  DNS Servers: %s\n", cfg.DNSServers)

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCountryFlag(countryCode string) string {
	flags := map[string]string{
		"KE": "🇰🇪", "US": "🇺🇸", "GB": "🇬🇧", "UK": "🇬🇧",
		"DE": "🇩🇪", "FR": "🇫🇷", "SG": "🇸🇬", "JP": "🇯🇵",
	}
	if flag, ok := flags[countryCode]; ok {
		return flag
	}
	return "🌍"
}
