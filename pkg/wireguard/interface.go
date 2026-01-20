package wireguard

import (
	"fmt"
	"os/exec"
	"strings"
)

type Interface struct {
	name string
}

func NewInterface(name string) *Interface {
	return &Interface{name: name}
}

func (i *Interface) Setup(privateKey string, port int, cidr string) error {
	// Create interface
	if err := i.exec("ip", "link", "add", "dev", i.name, "type", "wireguard"); err != nil {
		// Interface might already exist
		fmt.Printf("Interface may already exist: %v\n", err)
	}

	// Set private key
	cmd := exec.Command("wg", "set", i.name, "private-key", "/dev/stdin", "listen-port", fmt.Sprintf("%d", port))
	cmd.Stdin = strings.NewReader(privateKey)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}

	// Assign IP address
	if err := i.exec("ip", "addr", "add", cidr, "dev", i.name); err != nil {
		fmt.Printf("IP may already be assigned: %v\n", err)
	}

	// Bring up interface
	if err := i.exec("ip", "link", "set", "up", "dev", i.name); err != nil {
		return fmt.Errorf("failed to bring up interface: %w", err)
	}

	// Enable IP forwarding
	if err := i.exec("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Setup NAT
	if err := i.setupNAT(); err != nil {
		fmt.Printf("Warning: NAT setup failed: %v\n", err)
	}

	return nil
}

func (i *Interface) AddPeer(publicKey, allowedIP string) error {
	return i.exec("wg", "set", i.name, "peer", publicKey, "allowed-ips", allowedIP+"/32")
}

func (i *Interface) RemovePeer(publicKey string) error {
	return i.exec("wg", "set", i.name, "peer", publicKey, "remove")
}

func (i *Interface) setupNAT() error {
	// Get default interface
	output, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return err
	}

	parts := strings.Fields(string(output))
	if len(parts) < 5 {
		return fmt.Errorf("failed to determine default interface")
	}

	defaultIface := parts[4]

	// Setup iptables NAT
	return i.exec("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", defaultIface, "-j", "MASQUERADE")
}

func (i *Interface) exec(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed: %w - %s", name, err, string(output))
	}
	return nil
}
