// Package context defines the Context struct used for passing contextual information.
package context

import (
	"os/exec"
	"strings"
)

// Context holds information about the current environment for command generation.
type Context struct {
	WorkingDir string            // Current working directory
	GitInfo    map[string]string // Git-related info (branch, status, etc.)
	Files      []string          // List of files in the directory
	Extra      map[string]any    // Additional context from plugins
}

// GatherGitInfo populates GitInfo with branch and status if in a git repo.
func (c *Context) GatherGitInfo() {
	c.GitInfo = map[string]string{}
	// Get branch
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		c.GitInfo["branch"] = strings.TrimSpace(string(branch))
	}
	// Get status (short)
	status, err := exec.Command("git", "status", "--short").Output()
	if err == nil {
		c.GitInfo["status"] = strings.TrimSpace(string(status))
	}
}
