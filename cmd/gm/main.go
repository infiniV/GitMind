package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/gitman/internal/adapter/ai"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/ui"
)

var (
	version = "0.1.0"
	cfgManager *config.Manager
)

func main() {
	// Initialize config manager
	var err error
	cfgManager, err = config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "gm",
		Short: "GitMind - AI-powered Git workflow automation",
		Long: `GitMind (gm) is an intelligent Git CLI manager that uses AI to generate
commit messages and help you make smart branching decisions.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// Launch dashboard when no subcommand provided
			if err := runDashboard(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(commitCmd())
	rootCmd.AddCommand(mergeCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(onboardCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func commitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Analyze changes and create an AI-powered commit",
		Long: `Analyzes your git changes using AI and helps you create meaningful commits.
The AI will suggest commit messages and determine whether to commit directly
or create a new branch based on the nature of your changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Launch dashboard which handles commit workflow
			return runDashboard()
		},
	}

	return cmd
}

func mergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Analyze and execute an AI-powered merge",
		Long: `Analyzes a merge operation using AI and helps you merge branches intelligently.
The AI will suggest an appropriate merge strategy (squash, regular, or fast-forward)
and generate a meaningful merge commit message based on the commits being merged.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Launch dashboard which handles merge workflow
			return runDashboard()
		},
	}

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure GitMind settings",
		Long:  `Interactive configuration wizard to set up API keys and preferences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig()
		},
	}

	return cmd
}

func onboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Run the GitMind setup wizard",
		Long: `Interactive onboarding wizard to set up GitMind for your workspace.
This wizard will guide you through:
  - Git repository initialization
  - GitHub integration setup
  - Branch preferences
  - Commit conventions
  - Branch naming patterns
  - AI provider configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOnboard()
		},
	}

	return cmd
}

// DEPRECATED: runCommit is no longer used. All commands now launch the unified dashboard/AppModel.
/* func runCommit(userPrompt string, useConventional bool) error {
	// Load configuration
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if API key is configured
	if cfg.APIKey == "" {
		fmt.Println("GitMind is not configured yet.")
		fmt.Println("Please run 'gm config' to set up your API key.")
		return nil
	}

	// Get API key
	apiKey, err := cfgManager.GetAPIKey(cfg)
	if err != nil {
		return fmt.Errorf("invalid API configuration: %w", err)
	}

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize dependencies
	gitOps := git.NewExecOperations()

	// Create AI provider
	providerConfig := ai.ProviderConfig{
		Model:   cfg.DefaultModel,
		Timeout: 30,
	}

	aiProvider := ai.NewCerebrasProvider(apiKey, providerConfig)

	// Create use cases
	analyzeUseCase := usecase.NewAnalyzeCommitUseCase(gitOps, aiProvider)
	executeUseCase := usecase.NewExecuteCommitUseCase(gitOps)

	// Analyze commit
	ui.PrintInfo("Analyzing changes with AI...")
	ui.PrintSubtle("Reading file contents (files will not be staged until you confirm)")

	// Use longer timeout for analysis (AI can be slow on free tier)
	analysisCtx, analysisCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer analysisCancel()

	analysisReq := usecase.AnalyzeCommitRequest{
		RepoPath:               cwd,
		UserPrompt:             userPrompt,
		UseConventionalCommits: useConventional || cfg.UseConventionalCommits,
		APIKey:                 apiKey,
		ProtectedBranches:      cfg.ProtectedBranches,
	}

	analysis, err := analyzeUseCase.Execute(analysisCtx, analysisReq)
	if err != nil {
		// Check for free tier limit error
		if freeTierErr, ok := err.(*ai.FreeTierLimitError); ok {
			ui.PrintWarning(freeTierErr.Message)
			ui.PrintInfo(fmt.Sprintf("You can upgrade your API key or wait %d seconds before trying again", freeTierErr.RetryAfter))
			return nil
		}

		// Special handling for common errors
		errMsg := err.Error()
		if errMsg == "no changes to commit" {
			ui.PrintInfo("Working directory is clean - no changes to commit")
			ui.PrintSubtle("Make some changes to your files, then run 'gm commit' again")
			return nil
		}

		return fmt.Errorf("analysis failed: %w", err)
	}

	// Show TUI for user decision
	model := ui.NewCommitViewModel(analysis)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	commitModel := finalModel.(ui.CommitViewModel)

	if commitModel.IsCancelled() {
		ui.PrintInfo("Operation cancelled by user")
		return nil
	}

	if !commitModel.IsConfirmed() {
		ui.PrintWarning("No action selected")
		return nil
	}

	// Get selected option
	option := commitModel.GetSelectedOption()
	if option == nil {
		return fmt.Errorf("no option selected")
	}

	// Handle merge action differently - run merge workflow instead
	if option.Action == domain.ActionMerge {
		fmt.Println() // Blank line
		ui.PrintInfo("Launching merge workflow...")

		// Determine target branch from branch info
		targetBranch := ""
		if analysis.BranchInfo != nil && analysis.BranchInfo.Parent() != "" {
			targetBranch = analysis.BranchInfo.Parent()
		}

		// Run merge workflow
		return runMerge("", targetBranch)
	}

	// Execute the commit
	fmt.Println() // Blank line
	ui.PrintInfo("Executing selected action...")

	// Create a fresh context for git operations (generous timeout for large repos)
	execCtx, execCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer execCancel()

	executeReq := usecase.ExecuteCommitRequest{
		RepoPath:      cwd,
		Decision:      analysis.Decision,
		Action:        option.Action,
		CommitMessage: option.Message,
		BranchName:    option.BranchName,
		StageAll:      true, // Always stage all for now
	}

	executeResp, err := executeUseCase.Execute(execCtx, executeReq)
	if err != nil {
		return fmt.Errorf("commit execution failed: %w", err)
	}

	// Show success message
	fmt.Println() // Blank line

	// Special handling for manual review
	if option.Action == domain.ActionReview {
		ui.PrintInfo(executeResp.Message)
		ui.PrintSubtle("Review your changes, then run 'gm commit' again when ready")
		return nil
	}

	ui.PrintSuccess(executeResp.Message)

	if executeResp.BranchCreated != "" {
		fmt.Printf("  %s %s\n", ui.FormatLabel("Branch:"), ui.FormatValue(executeResp.BranchCreated))
	}
	fmt.Printf("  %s %s\n", ui.FormatLabel("Message:"), ui.FormatValue(option.Message.Title()))

	return nil
}
*/

// DEPRECATED: runMerge is no longer used. All commands now launch the unified dashboard/AppModel.
/* func runMerge(sourceBranch, targetBranch string) error {
	// Load configuration
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if API key is configured
	if cfg.APIKey == "" {
		fmt.Println("GitMind is not configured yet.")
		fmt.Println("Please run 'gm config' to set up your API key.")
		return nil
	}

	// Get API key
	apiKey, err := cfgManager.GetAPIKey(cfg)
	if err != nil {
		return fmt.Errorf("invalid API configuration: %w", err)
	}

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize dependencies
	gitOps := git.NewExecOperations()

	// Create AI provider
	providerConfig := ai.ProviderConfig{
		Model:   cfg.DefaultModel,
		Timeout: 30,
	}

	aiProvider := ai.NewCerebrasProvider(apiKey, providerConfig)

	// Create use cases
	analyzeUseCase := usecase.NewAnalyzeMergeUseCase(gitOps, aiProvider)
	executeUseCase := usecase.NewExecuteMergeUseCase(gitOps)

	// Analyze merge
	ui.PrintInfo("Analyzing merge with AI...")

	analysisCtx, analysisCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer analysisCancel()

	analysisReq := usecase.AnalyzeMergeRequest{
		RepoPath:          cwd,
		SourceBranch:      sourceBranch,
		TargetBranch:      targetBranch,
		ProtectedBranches: cfg.ProtectedBranches,
		APIKey:            apiKey,
	}

	analysis, err := analyzeUseCase.Execute(analysisCtx, analysisReq)
	if err != nil {
		return fmt.Errorf("merge analysis failed: %w", err)
	}

	// Show TUI for user decision
	model := ui.NewMergeViewModel(analysis)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	mergeModel := finalModel.(ui.MergeViewModel)

	if mergeModel.IsCancelled() {
		ui.PrintInfo("Merge cancelled by user")
		return nil
	}

	if !mergeModel.IsConfirmed() {
		ui.PrintWarning("No strategy selected")
		return nil
	}

	// Get selected strategy
	strategy := mergeModel.GetSelectedStrategy()
	if strategy == "" {
		return fmt.Errorf("no strategy selected")
	}

	// Execute the merge
	fmt.Println() // Blank line
	ui.PrintInfo(fmt.Sprintf("Executing %s merge...", strategy))

	execCtx, execCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer execCancel()

	executeReq := usecase.ExecuteMergeRequest{
		RepoPath:     cwd,
		SourceBranch: analysis.SourceBranchInfo.Name(),
		TargetBranch: analysis.TargetBranch,
		Strategy:     strategy,
		MergeMessage: analysis.MergeMessage,
	}

	executeResp, err := executeUseCase.Execute(execCtx, executeReq)
	if err != nil {
		return fmt.Errorf("merge execution failed: %w", err)
	}

	// Show success message
	fmt.Println() // Blank line
	ui.PrintSuccess(executeResp.Message)

	if executeResp.MergeCommit != "" {
		fmt.Printf("  %s %s\n", ui.FormatLabel("Commit:"), ui.FormatValue(executeResp.MergeCommit))
	}
	fmt.Printf("  %s %s\n", ui.FormatLabel("Strategy:"), ui.FormatValue(executeResp.Strategy))

	return nil
}
*/

func runDashboard() error {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize git operations
	gitOps := git.NewExecOperations()

	// Check if we're in a git repo
	ctx := context.Background()
	isRepo, err := gitOps.IsGitRepo(ctx, cwd)
	if err != nil || !isRepo {
		ui.PrintWarning("Not in a git repository")
		ui.PrintInfo("Navigate to a git repository to use the dashboard")
		ui.PrintInfo("Or run 'gm config' to configure GitMind")
		return nil
	}

	// Load config
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if API key is configured
	if cfg.AI.APIKey == "" {
		ui.PrintWarning("No API key configured")
		ui.PrintInfo("Run 'gm config' or 'gm onboard' to set up your Cerebras API key")
		ui.PrintInfo("You can get a free API key at https://cloud.cerebras.ai")
		return fmt.Errorf("API key not configured")
	}

	// Create AI provider
	apiKey, err := domain.NewAPIKey(cfg.AI.APIKey, cfg.AI.Provider)
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	tier, err := domain.ParseAPITier(cfg.AI.APITier)
	if err != nil {
		tier = domain.TierUnknown
	}
	apiKey.SetTier(tier)

	providerConfig := ai.ProviderConfig{
		Model:   cfg.AI.DefaultModel,
		Timeout: 30,
	}
	aiProvider := ai.NewCerebrasProvider(apiKey, providerConfig)

	// Create and launch AppModel (unified TUI)
	model := ui.NewAppModel(gitOps, aiProvider, cfg, cfgManager, cwd)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}

func runConfig() error {
	ui.PrintInfo("GitMind Configuration Wizard")
	fmt.Println()

	// Load existing config
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// API Provider
	fmt.Println("AI Provider:")
	fmt.Printf("  Current: %s\n", cfg.AI.Provider)
	fmt.Print("  Press Enter to keep current or type new provider: ")

	var provider string
	_, _ = fmt.Scanln(&provider)
	if provider != "" {
		cfg.AI.Provider = provider
	}

	// API Key
	fmt.Println()
	fmt.Println("Cerebras API Key:")
	fmt.Println("  Get your free API key at: https://cloud.cerebras.ai/")
	if cfg.AI.APIKey != "" {
		fmt.Printf("  Current: %s***\n", cfg.AI.APIKey[:min(4, len(cfg.AI.APIKey))])
		fmt.Print("  Press Enter to keep current or paste new key: ")
	} else {
		fmt.Print("  Paste your API key: ")
	}

	var apiKey string
	_, _ = fmt.Scanln(&apiKey)
	if apiKey != "" {
		cfg.AI.APIKey = apiKey
	}

	// API Tier
	fmt.Println()
	fmt.Println("API Tier:")
	fmt.Println("  1. Free (default)")
	fmt.Println("  2. Pro")
	fmt.Printf("  Current: %s\n", cfg.AI.APITier)
	fmt.Print("  Select (1 or 2): ")

	var tierChoice string
	_, _ = fmt.Scanln(&tierChoice)
	switch tierChoice {
	case "1":
		cfg.AI.APITier = "free"
	case "2":
		cfg.AI.APITier = "pro"
	}

	// Conventional Commits
	fmt.Println()
	fmt.Print("Use Conventional Commits format by default? (y/N): ")
	var useConventional string
	_, _ = fmt.Scanln(&useConventional)
	if useConventional == "y" || useConventional == "Y" {
		cfg.Commits.Convention = "conventional"
	} else {
		cfg.Commits.Convention = "none"
	}

	// Model Selection
	fmt.Println()
	fmt.Println("Default Model:")
	fmt.Println("  1. llama-3.3-70b (recommended, balanced)")
	fmt.Println("  2. llama3.1-8b (faster, lower quality)")
	fmt.Printf("  Current: %s\n", cfg.AI.DefaultModel)
	fmt.Print("  Select (1 or 2): ")

	var modelChoice string
	_, _ = fmt.Scanln(&modelChoice)
	switch modelChoice {
	case "1":
		cfg.AI.DefaultModel = "llama-3.3-70b"
	case "2":
		cfg.AI.DefaultModel = "llama3.1-8b"
	}

	// Save configuration
	if err := cfgManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	ui.PrintSuccess(fmt.Sprintf("Configuration saved to: %s", cfgManager.ConfigPath()))
	ui.PrintInfo("You're all set! Try running 'gm commit' in a git repository")

	return nil
}

func runOnboard() error {
	ui.PrintInfo("Starting GitMind setup wizard...")
	fmt.Println()

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load existing config
	cfg, err := cfgManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create git operations
	gitOps := git.NewExecOperations()

	// Run onboarding wizard
	return ui.RunOnboarding(gitOps, cfg, cfgManager, cwd)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
