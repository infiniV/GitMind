package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/gitman/internal/adapter/ai"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/ui"
	"github.com/yourusername/gitman/internal/usecase"
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func commitCmd() *cobra.Command {
	var userPrompt string
	var useConventional bool

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Analyze changes and create an AI-powered commit",
		Long: `Analyzes your git changes using AI and helps you create meaningful commits.
The AI will suggest commit messages and determine whether to commit directly
or create a new branch based on the nature of your changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(userPrompt, useConventional)
		},
	}

	cmd.Flags().StringVarP(&userPrompt, "message", "m", "", "Additional context for the AI")
	cmd.Flags().BoolVarP(&useConventional, "conventional", "c", false, "Use conventional commit format")

	return cmd
}

func mergeCmd() *cobra.Command {
	var sourceBranch string
	var targetBranch string

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Analyze and execute an AI-powered merge",
		Long: `Analyzes a merge operation using AI and helps you merge branches intelligently.
The AI will suggest an appropriate merge strategy (squash, regular, or fast-forward)
and generate a meaningful merge commit message based on the commits being merged.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMerge(sourceBranch, targetBranch)
		},
	}

	cmd.Flags().StringVarP(&sourceBranch, "source", "s", "", "Source branch to merge from (default: current branch)")
	cmd.Flags().StringVarP(&targetBranch, "target", "t", "", "Target branch to merge into (default: parent branch)")

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

func runCommit(userPrompt string, useConventional bool) error {
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

func runMerge(sourceBranch, targetBranch string) error {
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

	// Create and launch dashboard
	model := ui.NewDashboardModel(gitOps, cwd)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("dashboard error: %w", err)
	}

	dashModel := finalModel.(ui.DashboardModel)

	if dashModel.IsCancelled() {
		return nil
	}

	// Handle dashboard actions
	action := dashModel.GetAction()
	params := dashModel.GetActionParams()

	switch action {
	case ui.ActionCommit:
		// Launch commit workflow with options from dashboard
		conventional, _ := params["conventional"].(bool)
		message, _ := params["message"].(string)
		return runCommit(message, conventional)

	case ui.ActionMerge:
		// Launch merge workflow with options from dashboard
		source, _ := params["source"].(string)
		target, _ := params["target"].(string)
		return runMerge(source, target)

	case ui.ActionSwitchBranch:
		// Switch to selected branch
		branch, _ := params["branch"].(string)
		if branch != "" {
			ui.PrintInfo(fmt.Sprintf("Switching to branch: %s", branch))
			if err := gitOps.CheckoutBranch(ctx, cwd, branch); err != nil {
				return fmt.Errorf("failed to switch branch: %w", err)
			}
			ui.PrintSuccess(fmt.Sprintf("Switched to branch: %s", branch))
		}
		return nil

	default:
		// No action, just return
		return nil
	}
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
	fmt.Printf("  Current: %s\n", cfg.AIProvider)
	fmt.Print("  Press Enter to keep current or type new provider: ")

	var provider string
	fmt.Scanln(&provider)
	if provider != "" {
		cfg.AIProvider = provider
	}

	// API Key
	fmt.Println()
	fmt.Println("Cerebras API Key:")
	fmt.Println("  Get your free API key at: https://cloud.cerebras.ai/")
	if cfg.APIKey != "" {
		fmt.Printf("  Current: %s***\n", cfg.APIKey[:min(4, len(cfg.APIKey))])
		fmt.Print("  Press Enter to keep current or paste new key: ")
	} else {
		fmt.Print("  Paste your API key: ")
	}

	var apiKey string
	fmt.Scanln(&apiKey)
	if apiKey != "" {
		cfg.APIKey = apiKey
	}

	// API Tier
	fmt.Println()
	fmt.Println("API Tier:")
	fmt.Println("  1. Free (default)")
	fmt.Println("  2. Pro")
	fmt.Printf("  Current: %s\n", cfg.APITier)
	fmt.Print("  Select (1 or 2): ")

	var tierChoice string
	fmt.Scanln(&tierChoice)
	switch tierChoice {
	case "1":
		cfg.APITier = "free"
	case "2":
		cfg.APITier = "pro"
	}

	// Conventional Commits
	fmt.Println()
	fmt.Print("Use Conventional Commits format by default? (y/N): ")
	var useConventional string
	fmt.Scanln(&useConventional)
	cfg.UseConventionalCommits = useConventional == "y" || useConventional == "Y"

	// Model Selection
	fmt.Println()
	fmt.Println("Default Model:")
	fmt.Println("  1. llama-3.3-70b (recommended, balanced)")
	fmt.Println("  2. llama3.1-8b (faster, lower quality)")
	fmt.Printf("  Current: %s\n", cfg.DefaultModel)
	fmt.Print("  Select (1 or 2): ")

	var modelChoice string
	fmt.Scanln(&modelChoice)
	switch modelChoice {
	case "1":
		cfg.DefaultModel = "llama-3.3-70b"
	case "2":
		cfg.DefaultModel = "llama3.1-8b"
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
