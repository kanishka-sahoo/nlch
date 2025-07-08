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

func main() {
	// CLI flags
	showVersion := flag.Bool("version", false, "Show version and exit")
	dryRun := flag.Bool("dry-run", false, "Show the command but do not execute it")
	model := flag.String("model", "", "Override the model to use")
	providerFlag := flag.String("provider", "", "Override the provider to use")
	yesSure := flag.Bool("yes-im-sure", false, "Bypass confirmation for all commands, including dangerous ones")
	verbose := flag.Bool("verbose", false, "Show provider and model information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nlch version %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: nlch [flags] \"Describe your command here\"")
		flag.PrintDefaults()
		os.Exit(1)
	}
	userInput := flag.Arg(0)

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

	// Safety and confirmation logic
	const DangerPrefix = "danger: "
	isDanger := shell.IsDangerousCommand(cmd)
	if isDanger && !*yesSure {
		fmt.Println("This is a dangerous command, use --yes-im-sure to bypass.")
		os.Exit(1)
	}

	// Remove danger prefix if approved by user
	if isDanger && *yesSure && strings.HasPrefix(cmd, DangerPrefix) {
		cmd = cmd[len(DangerPrefix):]
	}

	// Only confirm for non-dangerous commands
	requireConfirm := !*yesSure && !isDanger

	// Execute or dry-run
	exec := shell.Executor{DryRun: *dryRun}
	if err := exec.Run(cmd, requireConfirm); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
