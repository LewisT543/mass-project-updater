package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"lewist543.com/mass-project-updater/internal/config"
	"lewist543.com/mass-project-updater/internal/git"
	"lewist543.com/mass-project-updater/internal/gitlab"
	"lewist543.com/mass-project-updater/internal/model"
	"lewist543.com/mass-project-updater/internal/npm"
	"lewist543.com/mass-project-updater/internal/packagejson"
	"lewist543.com/mass-project-updater/internal/runner"
)

var (
	depsFileFlag    string
	projectPrefix   string
	maxWorkersFlag  int
	dryRunFlag      bool

	interactiveMode bool
)

func Execute() error {
	printBanner()

	rootCmd := newRootCmd()

	if len(os.Args) == 1 {
		interactiveMode = true

		// In interactive mode we control the UX via our own menu and prompts,
		// so suppress Cobra's automatic usage/error output to avoid noise.
		rootCmd.SilenceUsage = true
		rootCmd.SilenceErrors = true

		for {
			if !showInteractiveMenu(rootCmd) {
				return nil
			}

			if err := rootCmd.Execute(); err != nil {
				return err
			}

			waitForExit()
		}
	}

	interactiveMode = false
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mass-project-updater",
		Short: "CLI for running bulk operations across many GitLab projects",
		Long: `Mass Project Updater

This tool helps you run bulk operations across many GitLab projects.

Available commands:
  • update-deps  - Update dependencies across matching UI projects and open merge requests.`,
	}

	updateDepsCmd := &cobra.Command{
		Use:   "update-deps",
		Short: "Update dependencies across matching UI projects and open merge requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateDeps()
		},
	}

	updateDepsCmd.Flags().StringVarP(&depsFileFlag, "deps-file", "d", "deps.json", "Path to JSON file describing dependency updates")
	updateDepsCmd.Flags().StringVarP(&projectPrefix, "project-prefix", "p", "ui-spa", "Prefix used to filter GitLab projects by name")
	updateDepsCmd.Flags().IntVarP(&maxWorkersFlag, "max-workers", "w", 5, "Maximum number of projects to process in parallel")
	updateDepsCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Log intended operations without making any changes")

	rootCmd.AddCommand(updateDepsCmd)

	return rootCmd
}

func printBanner() {
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║         Mass Project Updater CLI             ║")
	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Println("║  Run bulk operations across many GitLab UI   ║")
	fmt.Println("║  projects with a single command.             ║")
	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Println("║  Available options:                          ║")
	fmt.Println("║    • update-deps  - Update dependencies      ║")
	fmt.Println("║                     across matching UI       ║")
	fmt.Println("║                     projects and open MRs.   ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()
}

func runUpdateDeps() error {
	cfg := config.Load()

	// Allow CLI flags to override config defaults where appropriate.
	if maxWorkersFlag > 0 {
		cfg.MaxWorkers = maxWorkersFlag
	}

	// Ensure required config values, prompting the user in interactive mode when missing.
	var err error
	cfg, err = ensureConfig(cfg)
	if err != nil {
		return err
	}

	// Decide dry-run behavior:
	// - In interactive mode, prompt the user each time.
	// - In non-interactive mode, respect the --dry-run flag for scripting/CI.
	if interactiveMode {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Run in dry-run mode? [y/N]: ")
		input, _ := reader.ReadString('\n')
		choice := strings.ToLower(strings.TrimSpace(input))

		cfg.DryRun = choice == "y" || choice == "yes"
	} else if dryRunFlag {
		cfg.DryRun = true
	}

	if err := prepareFilesystem(cfg); err != nil {
		return err
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	updates, err := packagejson.LoadUpdates(depsFileFlag)
	if err != nil {
		return err
	}

	projects, err := gitlab.FetchUIProjects(
		cfg.GitlabBase,
		cfg.GitlabToken,
		cfg.GroupID,
		projectPrefix,
	)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("No projects found matching the given criteria.")
		return nil
	}

	if interactiveMode {
		fmt.Println("\nThe following projects will be processed:")
		for i, p := range projects {
			fmt.Printf("  %d) %s (%s)\n", i+1, p.Name, p.SSHURL)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nDo you want to continue with these projects? [y/N]: ")
		input, _ := reader.ReadString('\n')
		choice := strings.ToLower(strings.TrimSpace(input))

		if choice != "y" && choice != "yes" {
			fmt.Println("Aborting update-deps run and returning to main menu.")
			return nil
		}
	}

	gitRunner := git.New()
	npmRunner := npm.New()
	mrClient := gitlab.NewMRClient(cfg.GitlabBase, cfg.GitlabToken)

	results := runner.RunAllProjects(projects, cfg.MaxWorkers, func(p model.Project) (string, error) {
		mrURL, err := runner.RunProject(cfg, p, updates, gitRunner, npmRunner, mrClient, "", logger)
		if err != nil {
			log.Printf("[ERROR] Project %s: %v", p.Name, err)
		} else {
			log.Printf("[OK] Project %s processed, MR: %s", p.Name, mrURL)
		}
		return mrURL, err
	})

	// Report summary
	var failed int
	for _, r := range results {
		if r.Err != nil {
			failed++
			log.Printf("[FAIL] %s: %v", r.Project.Name, r.Err)
		} else {
			log.Printf("[SUCCESS] %s MR: %s", r.Project.Name, r.MRURL)
		}
	}

	fmt.Printf("Completed %d projects: %d failed, %d succeeded\n", len(projects), failed, len(projects)-failed)
	return nil
}

func ensureConfig(cfg config.Config) (config.Config, error) {
	// Non-interactive mode: validate and error out if required values are missing.
	if !interactiveMode {
		missing := []string{}
		if cfg.GitlabBase == "" {
			missing = append(missing, "GITLAB_BASE_URL")
		}
		if cfg.GitlabToken == "" {
			missing = append(missing, "GITLAB_TOKEN")
		}
		if cfg.GroupID == "" {
			missing = append(missing, "GITLAB_GROUP_ID")
		}

		if len(missing) > 0 {
			return cfg, fmt.Errorf("missing required configuration: %s (set environment variables or run interactively)", strings.Join(missing, ", "))
		}

		return cfg, nil
	}

	// Interactive mode: prompt the user for any missing values.
	reader := bufio.NewReader(os.Stdin)

	if cfg.GitlabBase == "" {
		fmt.Println("GITLAB_BASE_URL env var is not set. Please enter it below.")
		fmt.Print("Enter GitLab base URL (e.g. https://gitlab.example.com): ")
		input, _ := reader.ReadString('\n')
		cfg.GitlabBase = strings.TrimSpace(input)
	}

	if cfg.GitlabToken == "" {
		fmt.Println("GITLAB_TOKEN env var is not set. Please enter it below.")
		fmt.Print("Enter GitLab access token: ")
		input, _ := reader.ReadString('\n')
		cfg.GitlabToken = strings.TrimSpace(input)
	}

	if cfg.GroupID == "" {
		fmt.Println("GITLAB_GROUP_ID env var is not set. Please enter it below.")	
		fmt.Print("Enter GitLab group ID: ")
		input, _ := reader.ReadString('\n')
		cfg.GroupID = strings.TrimSpace(input)
	}

	// BranchName already has a sensible default, but allow overriding interactively if empty.
	if cfg.BranchName == "" {
		fmt.Print("Enter branch name to use for updates [chore/dependency-updates]: ")
		input, _ := reader.ReadString('\n')
		value := strings.TrimSpace(input)
		if value != "" {
			cfg.BranchName = value
		}
	}

	return cfg, nil
}

func showInteractiveMenu(rootCmd *cobra.Command) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select an option:")
	fmt.Println("  1) Update dependencies across matching UI projects and open merge requests")
	fmt.Println("  0) Exit")
	fmt.Print("\nEnter choice: ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	switch choice {
	case "1":
		// Route into the existing cobra command as if user had typed `update-deps`.
		rootCmd.SetArgs([]string{"update-deps"})
		return true
	case "0", "":
		fmt.Println("Exiting.")
		return false
	default:
		fmt.Println("Unrecognized choice, exiting.")
		return false
	}
}

func waitForExit() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nPress Enter to return to the main menu...")
	_, _ = reader.ReadString('\n')
}

func prepareFilesystem(cfg config.Config) error {
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		return err
	}

	if err := os.RemoveAll(cfg.ErrorLogDir); err != nil {
		return err
	}

	return os.MkdirAll(cfg.ErrorLogDir, 0755)
}

