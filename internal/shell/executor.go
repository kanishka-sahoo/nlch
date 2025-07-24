// Package shell provides command execution utilities for nlch.
package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Executor handles command execution with dry-run and confirmation support.
type Executor struct {
	DryRun bool
}

// Run executes the given shell command, optionally as a dry-run.
// Returns the command output and error for potential retry logic.
func (e *Executor) Run(cmd string, requireConfirm bool) (stdout, stderr string, err error) {
	fmt.Printf("> Running command `%s`...\n", cmd)
	if e.DryRun {
		fmt.Println("> This was a dry-run, thus no action was taken.")
		return "", "", nil
	}
	if requireConfirm {
		fmt.Print("> Confirm? [Y/n]: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		resp := scanner.Text()
		if resp != "" && (resp[0] == 'n' || resp[0] == 'N') {
			fmt.Println("> Aborted by user.")
			return "", "", nil
		}
	}

	command := exec.Command("bash", "-c", cmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = &stdoutBuf
	command.Stderr = &stderrBuf
	command.Stdin = os.Stdin

	err = command.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	// Still print to console for user visibility
	if stdout != "" {
		fmt.Print(stdout)
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	return stdout, stderr, err
}
