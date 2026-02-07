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
