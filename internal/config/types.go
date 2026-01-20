package config

type ServerConfig struct {
	Code string
	Name string
	IP   string
}

type Config struct {
	Port             string
	WireGuardIface   string
	WireGuardPort    int
	ServerPrivateKey string
	ServerPublicKey  string
	NetworkCIDR      string
	DNSServers       []string
	AllowedIPs       string
	Servers          []ServerConfig
}
