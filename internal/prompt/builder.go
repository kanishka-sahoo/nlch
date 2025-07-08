// Package prompt provides utilities for building LLM prompts.
package prompt

import (
	"fmt"

	"github.com/kanishka-sahoo/nlch/internal/context"
)

// BuildPrompt constructs a structured prompt for the LLM using context and user input.
func BuildPrompt(ctx *context.Context, userInput string) string {
	// Format file list (truncate if too long)
	maxFiles := 20
	files := ctx.Files
	fileList := ""
	if len(files) > 0 {
		if len(files) > maxFiles {
			files = files[:maxFiles]
			fileList = fmt.Sprintf("%v ... (and %d more)", files, len(ctx.Files)-maxFiles)
		} else {
			fileList = fmt.Sprintf("%v", files)
		}
	} else {
		fileList = "(none)"
	}

	// Format git info
	gitInfo := ""
	if branch, ok := ctx.GitInfo["branch"]; ok && branch != "" {
		gitInfo += fmt.Sprintf("Branch: %s\n", branch)
	}
	if status, ok := ctx.GitInfo["status"]; ok && status != "" {
		gitInfo += fmt.Sprintf("Status:\n%s\n", status)
	}
	if gitInfo == "" {
		gitInfo = "No git repository detected.\n"
	}

	// Format plugin extras
	extras := ""
	if len(ctx.Extra) > 0 {
		extras += "Additional context:\n"
		for k, v := range ctx.Extra {
			extras += fmt.Sprintf("- %s: %v\n", k, v)
		}
	}

	return fmt.Sprintf(
		"You are an expert terminal assistant. Given the following project context, generate a smart, concise shell command for the user's request. Do not wrap your command in code blocks, provide it directly.\n\n"+
			"When running commands such as `ls`, make sure to pick flags to make it user-friendly. Avoid confusing the user with too much information.\n\n"+
			"If the command is potentially dangerous and destructive, write 'danger: ' before the command.\n\n"+
			"Working Directory: %s\n"+
			"Files: %s\n"+
			"Git Info:\n%s"+
			"%s"+
			"User Request: %s\n"+
			"Shell Command:",
		ctx.WorkingDir, fileList, gitInfo, extras, userInput,
	)
}
