// Package shell provides command execution utilities for nlch.
package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

// Executor handles command execution with dry-run and confirmation support.
type Executor struct {
	DryRun bool
}

// Run executes the given shell command, optionally as a dry-run.
func (e *Executor) Run(cmd string, requireConfirm bool) error {
	fmt.Printf("> Running command `%s`...\n", cmd)
	if e.DryRun {
		fmt.Println("> This was a dry-run, thus no action was taken.")
		return nil
	}
	if requireConfirm {
		fmt.Print("> Confirm? [Y/n]: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		resp := scanner.Text()
		if resp != "" && (resp[0] == 'n' || resp[0] == 'N') {
			fmt.Println("> Aborted by user.")
			return nil
		}
	}
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}
