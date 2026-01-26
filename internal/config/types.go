package config

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
