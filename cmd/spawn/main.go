package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"spawn.dev/pkg/agent"
	"spawn.dev/pkg/version"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var cfgFile string
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	cmd := &cobra.Command{
		Use:   "spawn",
		Short: "spawn - systemd for AI agents",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			viper.SetEnvPrefix("SPAWN")
			viper.AutomaticEnv()
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
				if err := viper.ReadInConfig(); err != nil {
					return fmt.Errorf("read config file: %w", err)
				}
			}
			_ = cmd
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")

	cmd.AddCommand(
		initCmd(style),
		runCmd(style),
		startCmd(style),
		stopCmd(style),
		statusCmd(style),
		agentCmd(style),
		capabilityCmd(style),
		toolCmd(style),
		meshCmd(style),
		stub(style, "logs", "Stream all logs"),
		stub(style, "metrics", "Show metrics"),
		stubWithArg(style, "trace", "id", "Get trace details"),
		stubWithArg(style, "replay", "id", "Replay agent decision"),
		stub(style, "dev", "Start development mode"),
		validateCmd(style),
		stub(style, "lint", "Lint agent configs"),
		stub(style, "test", "Run agent tests"),
		versionCmd(style),
		stub(style, "doctor", "Diagnose installation"),
		stub(style, "upgrade", "Upgrade spawn"),
		stub(style, "config", "Manage global config"),
	)

	return cmd
}

func initCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize new agent project",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if err := os.MkdirAll(name, 0o755); err != nil {
				return fmt.Errorf("create project dir: %w", err)
			}
			cfg := []byte(`apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: ` + name + `
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  goal: "Describe your goal"
  capabilities:
    exec:
      enabled: true
      languages: [python, nodejs, bash]
  sandbox:
    runtime: gvisor
    networkPolicy: restricted
    seccompProfile: strict
`)
			if err := os.WriteFile(name+"/agent.yaml", cfg, 0o644); err != nil {
				return fmt.Errorf("write agent config: %w", err)
			}
			fmt.Println(style.Render("Initialized " + name + "/agent.yaml"))
			return nil
		},
	}
}

func runCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{
		Use:   "run [config]",
		Short: "Run agent(s) from config",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := agent.LoadConfig(args[0])
			if err != nil {
				return err
			}
			fmt.Println(style.Render("Running agent: " + cfg.Metadata.Name))
			fmt.Println("Capabilities:", strings.Join(cfg.CapabilityNames(), ", "))
			return nil
		},
	}
}

func startCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{Use: "start", Short: "Start spawn daemon", Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(style.Render("Daemon start requested"))
	}}
}

func stopCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{Use: "stop", Short: "Stop spawn daemon", Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(style.Render("Daemon stop requested"))
	}}
}

func statusCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Show daemon and agent status", Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(style.Render("spawn status: healthy"))
	}}
}

func agentCmd(style lipgloss.Style) *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Agent management"}
	cmd.AddCommand(
		stub(style, "list", "List all agents"),
		stubWithArg(style, "get", "name", "Get agent details"),
		stubWithArg(style, "logs", "name", "Stream agent logs"),
		stubWithArgs(style, "exec", []string{"name", "cmd"}, "Execute command in agent"),
		stubWithArg(style, "kill", "name", "Terminate agent"),
		stubWithArg(style, "restart", "name", "Restart agent"),
	)
	return cmd
}

func capabilityCmd(style lipgloss.Style) *cobra.Command {
	cmd := &cobra.Command{Use: "capability", Short: "Capability management"}
	cmd.AddCommand(
		stub(style, "list", "List available capabilities"),
		stubWithArg(style, "install", "name", "Install capability plugin"),
		stubWithArg(style, "config", "name", "Configure capability"),
	)
	return cmd
}

func toolCmd(style lipgloss.Style) *cobra.Command {
	cmd := &cobra.Command{Use: "tool", Short: "Tool management"}
	cmd.AddCommand(
		stub(style, "list", "List registered tools"),
		stubWithArg(style, "register", "schema", "Register new tool"),
		stubWithArg(style, "invoke", "name", "Manually invoke tool"),
	)
	return cmd
}

func meshCmd(style lipgloss.Style) *cobra.Command {
	cmd := &cobra.Command{Use: "mesh", Short: "Mesh commands"}
	cmd.AddCommand(
		stub(style, "status", "Show mesh topology"),
		stub(style, "channels", "List communication channels"),
		stubWithArg(style, "send", "channel", "Send message to channel"),
	)
	return cmd
}

func validateCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [config]",
		Short: "Validate configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			_, err := agent.LoadConfig(args[0])
			if err != nil {
				return err
			}
			fmt.Println(style.Render("Configuration is valid"))
			return nil
		},
	}
}

func versionCmd(style lipgloss.Style) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(style.Render(version.Info()))
		},
	}
}

func stub(style lipgloss.Style, use, short string) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(style.Render(short + " (stub)"))
	}}
}

func stubWithArg(style lipgloss.Style, use, arg, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <" + arg + ">",
		Short: short,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			fmt.Println(style.Render(short + ": " + args[0]))
		},
	}
}

func stubWithArgs(style lipgloss.Style, use string, args []string, short string) *cobra.Command {
	argList := ""
	for _, arg := range args {
		argList += " <" + arg + ">"
	}
	return &cobra.Command{
		Use:   use + argList,
		Short: short,
		Args:  cobra.ExactArgs(len(args)),
		Run: func(_ *cobra.Command, vals []string) {
			fmt.Println(style.Render(short + ": " + strings.Join(vals, " ")))
		},
	}
}

func init() {
	cobra.OnInitialize(func() {
		_ = context.Background()
		_ = time.Now()
	})
}
