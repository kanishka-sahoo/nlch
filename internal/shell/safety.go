// Package shell provides safety checks for dangerous commands.
package shell

import (
	"strings"
)

// List of dangerous commands requiring extra confirmation.
var dangerousCommands = []string{
	"rm", "dd", "mkfs", "shutdown", "reboot", "init 0", "halt",
}

// IsDangerousCommand returns true if the command is considered dangerous.
func IsDangerousCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, danger := range dangerousCommands {
		if strings.Contains(lower, danger) {
			return true
		}
	}
	return false
}
