package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var testCommand = &cli.Command{
	Name:      "test",
	Usage:     "Run Maestro flows on a device",
	ArgsUsage: "<flow-file-or-folder>...",
	Description: `Run one or more Maestro flow files on a connected device.

Reports are generated in the output directory:
  - Default: ./reports/<timestamp>/
  - With --output: <output>/<timestamp>/
  - With --output and --flatten: <output>/ (no timestamp subfolder)

Examples:
  maestro-runner test flow.yaml
  maestro-runner test flows/
  maestro-runner test login.yaml checkout.yaml
  maestro-runner test flows/ -e USER=test -e PASS=secret
  maestro-runner test flows/ --include-tags smoke
  maestro-runner test flows/ --output ./my-reports --flatten`,
	Flags: []cli.Flag{
		// Configuration
		&cli.StringFlag{
			Name:  "config",
			Usage: "Path to workspace config.yaml",
		},

		// Environment variables
		&cli.StringSliceFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Usage:   "Environment variables (KEY=VALUE)",
		},

		// Tag filtering
		&cli.StringSliceFlag{
			Name:  "include-tags",
			Usage: "Only include flows with these tags",
		},
		&cli.StringSliceFlag{
			Name:  "exclude-tags",
			Usage: "Exclude flows with these tags",
		},

		// Output directory
		&cli.StringFlag{
			Name:  "output",
			Usage: "Output directory for reports (default: ./reports)",
		},
		&cli.BoolFlag{
			Name:  "flatten",
			Usage: "Don't create timestamp subfolder (requires --output)",
		},

		// Parallelization
		&cli.IntFlag{
			Name:  "shard-split",
			Usage: "Split tests across N devices",
		},
		&cli.IntFlag{
			Name:  "shard-all",
			Usage: "Run all tests on N devices",
		},

		// Execution modes
		&cli.BoolFlag{
			Name:    "continuous",
			Aliases: []string{"c"},
			Usage:   "Enable continuous mode for single flow",
		},

		// Web options
		&cli.BoolFlag{
			Name:  "headless",
			Usage: "Run in headless mode (web only)",
		},

		// AI options
		&cli.BoolFlag{
			Name:  "analyze",
			Usage: "Enhance output with AI insights",
		},
		&cli.StringFlag{
			Name:  "api-url",
			Usage: "API base URL",
			Value: "https://api.copilot.mobile.dev",
		},
		&cli.StringFlag{
			Name:  "api-key",
			Usage: "API key",
		},
	},
	Action: runTest,
}

// RunConfig holds the complete test run configuration.
type RunConfig struct {
	// Paths
	FlowPaths  []string
	ConfigPath string

	// Environment
	Env map[string]string

	// Filtering
	IncludeTags []string
	ExcludeTags []string

	// Output
	OutputDir string // Final resolved output directory

	// Parallelization
	ShardSplit int
	ShardAll   int

	// Execution
	Continuous bool
	Headless   bool

	// Device
	Platform string
	Device   string
	Verbose  bool
	AppFile  string
}

func runTest(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("at least one flow file or folder is required")
	}

	// Parse environment variables
	env := parseEnvVars(c.StringSlice("env"))

	// Resolve output directory
	outputDir, err := resolveOutputDir(c.String("output"), c.Bool("flatten"))
	if err != nil {
		return err
	}

	// Build run configuration
	cfg := &RunConfig{
		FlowPaths:   c.Args().Slice(),
		ConfigPath:  c.String("config"),
		Env:         env,
		IncludeTags: c.StringSlice("include-tags"),
		ExcludeTags: c.StringSlice("exclude-tags"),
		OutputDir:   outputDir,
		ShardSplit:  c.Int("shard-split"),
		ShardAll:    c.Int("shard-all"),
		Continuous:  c.Bool("continuous"),
		Headless:    c.Bool("headless"),
		Platform:    c.String("platform"),
		Device:      c.String("device"),
		Verbose:     c.Bool("verbose"),
		AppFile:     c.String("app-file"),
	}

	return executeTest(cfg)
}

// resolveOutputDir determines the output directory based on flags.
// - No --output: ./reports/<timestamp>/
// - --output given: <output>/<timestamp>/
// - --output + --flatten: <output>/ (error if --output not given)
func resolveOutputDir(output string, flatten bool) (string, error) {
	if flatten && output == "" {
		return "", fmt.Errorf("--flatten requires --output to be specified")
	}

	baseDir := output
	if baseDir == "" {
		baseDir = "./reports"
	}

	if flatten {
		return filepath.Clean(baseDir), nil
	}

	// Create timestamp-based subfolder
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(baseDir, timestamp), nil
}

func executeTest(cfg *RunConfig) error {
	// TODO: Implement test execution
	// 1. Validate all flows using validator package
	// 2. Connect to device
	// 3. Execute flows
	// 4. Generate reports in OutputDir

	fmt.Println("Test command received:")
	fmt.Printf("  Flows: %v\n", cfg.FlowPaths)
	if cfg.ConfigPath != "" {
		fmt.Printf("  Config: %s\n", cfg.ConfigPath)
	}
	if len(cfg.Env) > 0 {
		fmt.Printf("  Env: %v\n", cfg.Env)
	}
	if len(cfg.IncludeTags) > 0 {
		fmt.Printf("  Include tags: %v\n", cfg.IncludeTags)
	}
	if len(cfg.ExcludeTags) > 0 {
		fmt.Printf("  Exclude tags: %v\n", cfg.ExcludeTags)
	}
	fmt.Printf("  Output: %s\n", cfg.OutputDir)
	if cfg.Platform != "" {
		fmt.Printf("  Platform: %s\n", cfg.Platform)
	}
	if cfg.Device != "" {
		fmt.Printf("  Device: %s\n", cfg.Device)
	}

	fmt.Println("\n[Not yet implemented - will validate and run flows]")
	return nil
}

func parseEnvVars(envs []string) map[string]string {
	result := make(map[string]string)
	for _, e := range envs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}
