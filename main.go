package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	yesRisk := flag.Bool("yes-im-aware-of-the-risk", false, "Bypass confirmation for dangerous commands")
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

	// Generate command
	cmd, err := prov.GenerateCommand(*ctx, promptStr, opts)
	if err != nil {
		log.Fatalf("Provider error: %v", err)
	}

	// Safety check
	requireConfirm := shell.IsDangerousCommand(cmd) && !*yesRisk

	// Execute or dry-run
	exec := shell.Executor{DryRun: *dryRun}
	if err := exec.Run(cmd, requireConfirm); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
