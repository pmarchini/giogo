package limiter

import (
	"fmt"
	"os/exec"
	"strings"
)

// setupHTB sets up HTB (Hierarchical Token Bucket) qdisc for bandwidth limiting
// It creates a root qdisc and a class with the specified rate limit
func setupHTB(interfaceName string, classID uint32, rateBytesPerSec uint64) error {
	// Convert bytes per second to bits per second for tc
	rateBitsPerSec := rateBytesPerSec * 8
	
	// Format: 1:classID in hex
	classIDHex := fmt.Sprintf("1:%x", classID)
	
	// Check if root qdisc already exists
	checkCmd := exec.Command("tc", "qdisc", "show", "dev", interfaceName)
	output, _ := checkCmd.CombinedOutput()
	
	hasHTB := strings.Contains(string(output), "htb")
	
	// Add root qdisc if it doesn't exist
	if !hasHTB {
		// tc qdisc add dev <interface> root handle 1: htb default 30
		cmd := exec.Command("tc", "qdisc", "add", "dev", interfaceName, "root", "handle", "1:", "htb", "default", "30")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add HTB root qdisc: %v, output: %s", err, string(output))
		}
	}
	
	// Add class with bandwidth limit
	// tc class add dev <interface> parent 1: classid <classID> htb rate <rate>bit ceil <rate>bit
	rateStr := fmt.Sprintf("%dbit", rateBitsPerSec)
	cmd := exec.Command("tc", "class", "add", "dev", interfaceName, "parent", "1:", "classid", classIDHex, "htb", "rate", rateStr, "ceil", rateStr)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add HTB class: %v, output: %s", err, string(output))
	}
	
	// Add a filter to match packets with this classID
	// tc filter add dev <interface> parent 1: protocol ip prio 1 handle <classID> cgroup
	handleStr := fmt.Sprintf("%d", classID)
	filterCmd := exec.Command("tc", "filter", "add", "dev", interfaceName, "parent", "1:", "protocol", "ip", "prio", "1", "handle", handleStr, "cgroup")
	if output, err := filterCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add cgroup filter: %v, output: %s", err, string(output))
	}
	
	return nil
}

// cleanupHTB removes the HTB qdisc, which also removes all classes and filters
func cleanupHTB(interfaceName string) error {
	// tc qdisc del dev <interface> root
	cmd := exec.Command("tc", "qdisc", "del", "dev", interfaceName, "root")
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if this fails - the qdisc might already be gone
		// Return the error but don't make it fatal
		return fmt.Errorf("failed to delete HTB qdisc: %v, output: %s", err, string(output))
	}
	return nil
}

// GetDefaultInterface returns the default network interface name
// This is a simple implementation that returns "eth0" as default
// In production, this could be enhanced to detect the actual default interface
func GetDefaultInterface() string {
	// Try to find the default route interface
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "eth0" // fallback
	}
	
	// Parse output like: "default via 172.17.0.1 dev eth0"
	parts := strings.Fields(string(output))
	for i, part := range parts {
		if part == "dev" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	
	return "eth0" // fallback
}

// setupIngressHTB sets up ingress traffic control using IFB device
// The pattern is: Redirect ingress → IFB → tc rules
func setupIngressHTB(interfaceName, ifbDeviceName string, classID uint32, rateBytesPerSec uint64) error {
	// Convert bytes per second to bits per second for tc
	rateBitsPerSec := rateBytesPerSec * 8
	
	// Format: 1:classID in hex
	classIDHex := fmt.Sprintf("1:%x", classID)
	
	// Step 1: Load IFB module if not already loaded
	modprobeCmd := exec.Command("modprobe", "ifb")
	modprobeCmd.CombinedOutput() // Ignore errors - module might already be loaded
	
	// Step 2: Create/bring up IFB device
	// Check if IFB device exists
	checkCmd := exec.Command("ip", "link", "show", ifbDeviceName)
	if _, err := checkCmd.CombinedOutput(); err != nil {
		// IFB device doesn't exist, create it
		createCmd := exec.Command("ip", "link", "add", "name", ifbDeviceName, "type", "ifb")
		if output, err := createCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create IFB device %s: %v, output: %s", ifbDeviceName, err, string(output))
		}
	}
	
	// Bring up the IFB device
	upCmd := exec.Command("ip", "link", "set", "dev", ifbDeviceName, "up")
	if output, err := upCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up IFB device %s: %v, output: %s", ifbDeviceName, err, string(output))
	}
	
	// Step 3: Redirect ingress traffic from main interface to IFB
	// First, add ingress qdisc on main interface
	ingressCmd := exec.Command("tc", "qdisc", "add", "dev", interfaceName, "ingress")
	ingressCmd.CombinedOutput() // Ignore error if already exists
	
	// Add filter to redirect ingress to IFB
	redirectCmd := exec.Command("tc", "filter", "add", "dev", interfaceName, "parent", "ffff:", 
		"protocol", "ip", "u32", "match", "u32", "0", "0", "flowid", "1:1", "action", "mirred", "egress", "redirect", "dev", ifbDeviceName)
	if output, err := redirectCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to redirect ingress to IFB: %v, output: %s", err, string(output))
	}
	
	// Step 4: Setup HTB on IFB device (same as egress)
	// Add root qdisc on IFB
	ifbQdiscCmd := exec.Command("tc", "qdisc", "add", "dev", ifbDeviceName, "root", "handle", "1:", "htb", "default", "30")
	if output, err := ifbQdiscCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add HTB root qdisc on IFB: %v, output: %s", err, string(output))
	}
	
	// Add class with bandwidth limit on IFB
	rateStr := fmt.Sprintf("%dbit", rateBitsPerSec)
	ifbClassCmd := exec.Command("tc", "class", "add", "dev", ifbDeviceName, "parent", "1:", "classid", classIDHex, "htb", "rate", rateStr, "ceil", rateStr)
	if output, err := ifbClassCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add HTB class on IFB: %v, output: %s", err, string(output))
	}
	
	// Add cgroup filter on IFB
	handleStr := fmt.Sprintf("%d", classID)
	ifbFilterCmd := exec.Command("tc", "filter", "add", "dev", ifbDeviceName, "parent", "1:", "protocol", "ip", "prio", "1", "handle", handleStr, "cgroup")
	if output, err := ifbFilterCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add cgroup filter on IFB: %v, output: %s", err, string(output))
	}
	
	return nil
}

// cleanupIngressHTB removes the ingress traffic control setup
func cleanupIngressHTB(interfaceName, ifbDeviceName string) error {
	var firstError error
	
	// Remove ingress qdisc from main interface
	ingressDelCmd := exec.Command("tc", "qdisc", "del", "dev", interfaceName, "ingress")
	if output, err := ingressDelCmd.CombinedOutput(); err != nil {
		firstError = fmt.Errorf("failed to delete ingress qdisc: %v, output: %s", err, string(output))
	}
	
	// Remove root qdisc from IFB device
	ifbDelCmd := exec.Command("tc", "qdisc", "del", "dev", ifbDeviceName, "root")
	if output, err := ifbDelCmd.CombinedOutput(); err != nil && firstError == nil {
		firstError = fmt.Errorf("failed to delete IFB qdisc: %v, output: %s", err, string(output))
	}
	
	// Bring down and delete IFB device
	downCmd := exec.Command("ip", "link", "set", "dev", ifbDeviceName, "down")
	downCmd.CombinedOutput() // Ignore errors
	
	delCmd := exec.Command("ip", "link", "del", ifbDeviceName)
	if output, err := delCmd.CombinedOutput(); err != nil && firstError == nil {
		firstError = fmt.Errorf("failed to delete IFB device: %v, output: %s", err, string(output))
	}
	
	return firstError
}
