package ft

import (
	"fmt"
)

var FT_CGROUP_V1_SUPPORT = "false"

// CheckFeature is a helper function to check the status of a feature toggle
func CheckFeature(feature string) bool {
	switch feature {
	// This feature toggle is used to enable/disable cgroup v1 support
	case "FT_CGROUP_V1_SUPPORT":
		return FT_CGROUP_V1_SUPPORT == "true"
	default:
		fmt.Printf("Unknown feature: %s\n", feature)
		return false
	}
}
