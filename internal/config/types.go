package config

type ServerConfig struct {
	Code string
	Name string
	IP   string
}

type Config struct {
	Port             string
	ServerPublicKey  string
	ServerPrivateKey string
	ServerEndpoint   string
	WGInterface      string
	WGPort           string
	VPNSubnet        string
	DNSServers       string
	Servers 		[]ServerConfig
}
