package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kanishka-sahoo/nlch/internal/config"
	"github.com/kanishka-sahoo/nlch/internal/context"
	"github.com/kanishka-sahoo/nlch/internal/plugin"
	"github.com/kanishka-sahoo/nlch/internal/prompt"
	"github.com/kanishka-sahoo/nlch/internal/provider"
	"github.com/kanishka-sahoo/nlch/internal/shell"
	"github.com/kanishka-sahoo/nlch/internal/update"
)

// Dummy provider for demonstration
type EchoProvider struct{}

func (e *EchoProvider) Name() string { return "echo" }
func (e *EchoProvider) GenerateCommand(ctx context.Context, prompt string, opts provider.ProviderOptions) (string, error) {
	// Just echo the prompt for demonstration
	return fmt.Sprintf("echo '%s'", prompt), nil
}

func init() {
	provider.Register(&EchoProvider{})
}

func gatherContext() *context.Context {
	wd, _ := os.Getwd()
	files := []string{}
	entries, err := os.ReadDir(wd)
	if err == nil {
		for _, entry := range entries {
			files = append(files, entry.Name())
		}
	}
	ctx := &context.Context{
		WorkingDir: wd,
		Files:      files,
		GitInfo:    map[string]string{},
		Extra:      map[string]any{},
	}
	// Gather git info
	ctx.GatherGitInfo()
	// Run plugins
	for _, p := range plugin.List() {
		_ = p.Gather(ctx)
	}
	return ctx
}

const version = "0.1.0"

// This variable can be overridden at build time using -ldflags
var buildVersion = version

// cleanCommand removes markdown code blocks and extracts the actual command
func cleanCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)

	// Handle empty commands
	if cmd == "" {
		return cmd
	}

	// Remove markdown code blocks
	if strings.HasPrefix(cmd, "```") {
		lines := strings.Split(cmd, "\n")
		if len(lines) > 1 {
			// Remove first line (```bash or ```)
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
			// Remove last line (```)
			lines = lines[:len(lines)-1]
		}
		cmd = strings.Join(lines, "\n")
		cmd = strings.TrimSpace(cmd)
	}

	// Remove backticks at start/end
	cmd = strings.Trim(cmd, "`")
	cmd = strings.TrimSpace(cmd)

	// Get first non-empty line as the command
	lines := strings.Split(cmd, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}

	return cmd
}

func main() {
	// Set the build version for the update package
	update.BuildVersion = buildVersion

	// CLI flags
	showVersion := flag.Bool("version", false, "Show version and exit")
	dryRun := flag.Bool("dry-run", false, "Show the command but do not execute it")
	model := flag.String("model", "", "Override the model to use")
	providerFlag := flag.String("provider", "", "Override the provider to use")
	yesSure := flag.Bool("yes-im-sure", false, "Bypass confirmation for all commands, including dangerous ones")
	verbose := flag.Bool("verbose", false, "Show provider and model information")
	updateFlag := flag.Bool("update", false, "Check for and install updates")
	checkUpdate := flag.Bool("check-update", false, "Check for updates without installing")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nlch version %s\n", buildVersion)
		os.Exit(0)
	}

	if *updateFlag {
		if err := update.AutoUpdate(false); err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		os.Exit(0)
	}

	if *checkUpdate {
		release, hasUpdate, err := update.CheckForUpdates()
		if err != nil {
			log.Fatalf("Update check failed: %v", err)
		}
		if hasUpdate {
			fmt.Printf("New version available: %s (current: v%s)\n", release.TagName, update.GetCurrentVersion())
			fmt.Println("Run 'nlch --update' to install the update.")
		} else {
			fmt.Println("nlch is up to date.")
		}
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: nlch [flags] \"Describe your command here\"")
		flag.PrintDefaults()
		os.Exit(1)
	}
	userInput := flag.Arg(0)

	// Check for updates in the background (non-blocking)
	update.NotifyUpdateAvailable()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Gather context
	ctx := gatherContext()

	// Build prompt
	promptStr := prompt.BuildPrompt(ctx, userInput)

	// Select provider
	providerName := cfg.DefaultProvider
	if *providerFlag != "" {
		providerName = *providerFlag
	}
	prov, ok := provider.Get(providerName)
	if !ok {
		log.Fatalf("Provider '%s' not found. Available: %v", providerName, provider.List())
	}

	// Provider options
	opts := provider.ProviderOptions{
		Model:    *model,
		Provider: providerName,
	}

	if *verbose {
		fmt.Printf("Provider: %s\n", providerName)
		modelUsed := opts.Model
		// Try to get model from provider instance if not overridden
		if modelUsed == "" {
			// Try to get model from known provider types
			switch p := prov.(type) {
			case interface{ Model() string }:
				modelUsed = p.Model()
			case interface{ GetModel() string }:
				modelUsed = p.GetModel()
			case *provider.OpenRouterProvider:
				modelUsed = p.Model
			default:
				// Fallback to config
				if provCfg, ok := cfg.Providers[providerName]; ok {
					modelUsed = provCfg.DefaultModel
				}
			}
		}
		fmt.Printf("Model: %s\n", modelUsed)
	}

	// Generate command
	cmd, err := prov.GenerateCommand(*ctx, promptStr, opts)
	if err != nil {
		log.Fatalf("Provider error: %v", err)
	}

	// Clean up the command (remove markdown code blocks, etc.)
	cmd = cleanCommand(cmd)

	// Safety and confirmation logic - let LLM decide what's dangerous
	const DangerPrefix = "danger: "
	isDanger := strings.HasPrefix(cmd, DangerPrefix)
	if isDanger && !*yesSure {
		fmt.Println("This is a dangerous command, use --yes-im-sure to bypass.")
		os.Exit(1)
	}

	// Remove danger prefix if approved by user
	if *yesSure && isDanger {
		cmd = cmd[len(DangerPrefix):]
	}

	// Only confirm for non-dangerous commands
	requireConfirm := !*yesSure && !isDanger

	// Execute or dry-run with retry logic
	exec := shell.Executor{DryRun: *dryRun}
	stdout, stderr, err := exec.Run(cmd, requireConfirm)

	// If command failed and not in dry-run mode, ask LLM to fix it
	if err != nil && !*dryRun {
		fmt.Println("\n> Command failed. Asking LLM to provide a corrected version...")

		// Build a prompt with the error information
		errorPrompt := fmt.Sprintf(
			"The previous command failed:\n"+
				"Command: %s\n"+
				"Error: %s\n"+
				"Stderr: %s\n"+
				"Stdout: %s\n\n"+
				"Please provide a corrected command for the original request: %s\n"+
				"Return ONLY the shell command, nothing else. Do not use markdown code blocks.",
			cmd, err.Error(), stderr, stdout, userInput)

		// Get corrected command from LLM
		correctedCmd, corrErr := prov.GenerateCommand(*ctx, errorPrompt, opts)
		if corrErr != nil {
			log.Fatalf("Failed to get corrected command: %v", corrErr)
		}

		// Clean up the corrected command (remove markdown code blocks, etc.)
		correctedCmd = cleanCommand(correctedCmd)

		// Check if we got a valid corrected command
		if strings.TrimSpace(correctedCmd) == "" {
			log.Fatalf("LLM did not provide a valid corrected command")
		}

		// Check if corrected command is dangerous
		isCorrectedDanger := strings.HasPrefix(correctedCmd, DangerPrefix)
		if isCorrectedDanger && !*yesSure {
			fmt.Println("The corrected command is dangerous, use --yes-im-sure to bypass.")
			os.Exit(1)
		}

		// Remove danger prefix if approved
		if *yesSure && isCorrectedDanger {
			correctedCmd = correctedCmd[len(DangerPrefix):]
		}

		// Execute corrected command (with confirmation if not bypassed)
		requireCorrectedConfirm := !*yesSure && !isCorrectedDanger
		fmt.Printf("\n> Trying corrected command: %s\n", correctedCmd)
		_, _, corrErr = exec.Run(correctedCmd, requireCorrectedConfirm)
		if corrErr != nil {
			log.Fatalf("Corrected command also failed: %v", corrErr)
		}
	} else if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
