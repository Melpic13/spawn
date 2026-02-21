package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"spawn.dev/pkg/agent"
	"spawn.dev/pkg/localstate"
	"spawn.dev/pkg/version"
)

var builtinCapabilities = []string{"browser", "exec", "fs", "mcp", "memory", "net", "secrets", "tools"}
var builtinTools = []string{"calculator", "datetime", "json_parser"}

func newRootCmd() *cobra.Command {
	app := &cliApp{style: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)}
	var cfgFile string
	var stateFile string

	cmd := &cobra.Command{
		Use:   "spawn",
		Short: "spawn - systemd for AI agents",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			viper.SetEnvPrefix("SPAWN")
			viper.AutomaticEnv()
			if cfgFile == "" {
				return nil
			}
			viper.SetConfigFile(cfgFile)
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("read config file: %w", err)
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")
	cmd.PersistentFlags().StringVar(&stateFile, "state", "", "path to state file (default: ~/.spawn/state.json)")

	cmd.AddCommand(
		app.initCmd(),
		app.runCmd(&stateFile),
		app.startCmd(&stateFile),
		app.stopCmd(&stateFile),
		app.statusCmd(&stateFile),
		app.agentCmd(&stateFile),
		app.capabilityCmd(&stateFile),
		app.toolCmd(&stateFile),
		app.meshCmd(&stateFile),
		app.logsCmd(&stateFile),
		app.metricsCmd(&stateFile),
		app.traceCmd(&stateFile),
		app.replayCmd(&stateFile),
		app.devCmd(&stateFile),
		app.validateCmd(),
		app.lintCmd(),
		app.testCmd(),
		app.versionCmd(),
		app.doctorCmd(&cfgFile, &stateFile),
		app.upgradeCmd(),
		app.configCmd(&stateFile),
	)
	return cmd
}

type cliApp struct {
	style lipgloss.Style
}

func (a *cliApp) openStore(statePath string) (*localstate.Store, error) {
	if statePath != "" {
		return localstate.OpenAt(statePath), nil
	}
	return localstate.Open()
}

func (a *cliApp) printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func (a *cliApp) initCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize new agent project",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if !force {
				if _, err := os.Stat(name); err == nil {
					return fmt.Errorf("initialize project: %q already exists (use --force to overwrite)", name)
				}
			}
			if err := os.MkdirAll(name, 0o755); err != nil {
				return fmt.Errorf("create project dir: %w", err)
			}
			cfgPath := filepath.Join(name, "agent.yaml")
			cfg := []byte(`apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: ` + name + `
  namespace: default
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  goal: "Describe your goal"
  capabilities:
    exec:
      enabled: true
      languages: [python, nodejs, bash]
    fs:
      enabled: true
      mounts:
        - path: /workspace
          mode: rw
  sandbox:
    runtime: gvisor
    networkPolicy: restricted
    seccompProfile: strict
`)
			if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
				return fmt.Errorf("write agent config: %w", err)
			}
			fmt.Println(a.style.Render("Initialized " + cfgPath))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite if target exists")
	return cmd
}

func (a *cliApp) runCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "run [config]",
		Short: "Run agent(s) from config",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfgPath := args[0]
			cfg, err := agent.LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			if cfg.Metadata.Namespace == "" {
				cfg.Metadata.Namespace = "default"
			}

			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			agentID := uuid.NewString()
			traceID := uuid.NewString()
			now := time.Now().UTC()

			err = store.Update(func(st *localstate.State) error {
				st.Agents[cfg.Metadata.Name] = localstate.AgentRecord{
					ID:           agentID,
					Name:         cfg.Metadata.Name,
					Namespace:    cfg.Metadata.Namespace,
					ConfigPath:   cfgPath,
					State:        string(agent.StateRunning),
					Capabilities: cfg.CapabilityNames(),
					CreatedAt:    now,
					UpdatedAt:    now,
				}
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Agent: cfg.Metadata.Name, Message: "agent started"})
				st.Traces[traceID] = localstate.TraceRecord{
					ID:        traceID,
					Agent:     cfg.Metadata.Name,
					CreatedAt: now,
					UpdatedAt: now,
					Steps: []localstate.TraceStep{
						{Time: now, Message: "loaded config " + cfgPath},
						{Time: now, Message: "enabled capabilities: " + strings.Join(cfg.CapabilityNames(), ", ")},
					},
				}
				return nil
			})
			if err != nil {
				return err
			}

			fmt.Println(a.style.Render("Running agent: " + cfg.Metadata.Name))
			fmt.Println("Agent ID:", agentID)
			fmt.Println("Trace ID:", traceID)
			fmt.Println("Capabilities:", strings.Join(cfg.CapabilityNames(), ", "))
			return nil
		},
	}
}

func (a *cliApp) startCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start spawn daemon",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := store.Update(func(st *localstate.State) error {
				st.Daemon.Running = true
				st.Daemon.StartedAt = now
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Message: "daemon started"})
				return nil
			}); err != nil {
				return err
			}
			fmt.Println(a.style.Render("Daemon started"))
			return nil
		},
	}
}

func (a *cliApp) stopCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop spawn daemon",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := store.Update(func(st *localstate.State) error {
				st.Daemon.Running = false
				st.Daemon.StoppedAt = now
				for name, rec := range st.Agents {
					if rec.State == string(agent.StateRunning) {
						rec.State = string(agent.StateTerminated)
						rec.UpdatedAt = now
						st.Agents[name] = rec
					}
				}
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Message: "daemon stopped"})
				return nil
			}); err != nil {
				return err
			}
			fmt.Println(a.style.Render("Daemon stopped"))
			return nil
		},
	}
}

func (a *cliApp) statusCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon and agent status",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			runningAgents := 0
			for _, rec := range st.Agents {
				if rec.State == string(agent.StateRunning) {
					runningAgents++
				}
			}
			fmt.Println(a.style.Render("spawn status"))
			fmt.Printf("daemon.running=%t\n", st.Daemon.Running)
			if !st.Daemon.StartedAt.IsZero() {
				fmt.Printf("daemon.startedAt=%s\n", st.Daemon.StartedAt.Format(time.RFC3339))
			}
			fmt.Printf("agents.total=%d\n", len(st.Agents))
			fmt.Printf("agents.running=%d\n", runningAgents)
			fmt.Printf("logs.total=%d\n", len(st.Logs))
			return nil
		},
	}
}

func (a *cliApp) agentCmd(statePath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Agent management"}
	cmd.AddCommand(
		a.agentListCmd(statePath),
		a.agentGetCmd(statePath),
		a.agentLogsCmd(statePath),
		a.agentExecCmd(statePath),
		a.agentKillCmd(statePath),
		a.agentRestartCmd(statePath),
	)
	return cmd
}

func (a *cliApp) agentListCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all agents",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(st.Agents))
			for name := range st.Agents {
				names = append(names, name)
			}
			sort.Strings(names)
			if len(names) == 0 {
				fmt.Println("No agents found")
				return nil
			}
			fmt.Println("NAME\tSTATE\tNAMESPACE\tUPDATED")
			for _, name := range names {
				rec := st.Agents[name]
				fmt.Printf("%s\t%s\t%s\t%s\n", rec.Name, rec.State, rec.Namespace, rec.UpdatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}
}

func (a *cliApp) agentGetCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			rec, ok := st.Agents[args[0]]
			if !ok {
				return fmt.Errorf("agent %q not found", args[0])
			}
			return a.printJSON(rec)
		},
	}
}

func (a *cliApp) agentLogsCmd(statePath *string) *cobra.Command {
	var tail int
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "Stream agent logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			filtered := make([]localstate.LogEntry, 0)
			for _, entry := range st.Logs {
				if entry.Agent == args[0] {
					filtered = append(filtered, entry)
				}
			}
			if tail > 0 && len(filtered) > tail {
				filtered = filtered[len(filtered)-tail:]
			}
			for _, entry := range filtered {
				fmt.Printf("%s [%s] %s\n", entry.Time.Format(time.RFC3339), strings.ToUpper(entry.Level), entry.Message)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&tail, "tail", 50, "number of log lines")
	return cmd
}

func (a *cliApp) agentExecCmd(statePath *string) *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "exec <name> <cmd>",
		Short: "Execute command in agent",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			agentName := args[0]
			command := args[1]
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			rec, ok := st.Agents[agentName]
			if !ok {
				return fmt.Errorf("agent %q not found", agentName)
			}
			if rec.State != string(agent.StateRunning) {
				return fmt.Errorf("agent %q is not running", agentName)
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			execCmd := exec.CommandContext(ctx, "sh", "-lc", command)
			out, err := execCmd.CombinedOutput()
			now := time.Now().UTC()
			updateErr := store.Update(func(st *localstate.State) error {
				level := "info"
				msg := fmt.Sprintf("exec `%s` succeeded", command)
				if err != nil {
					level = "error"
					msg = fmt.Sprintf("exec `%s` failed: %v", command, err)
				}
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: level, Agent: agentName, Message: msg})
				traceID := uuid.NewString()
				st.Traces[traceID] = localstate.TraceRecord{
					ID:        traceID,
					Agent:     agentName,
					CreatedAt: now,
					UpdatedAt: now,
					Steps: []localstate.TraceStep{
						{Time: now, Message: "executed command: " + command},
						{Time: now, Message: "output bytes: " + strconv.Itoa(len(out))},
					},
				}
				return nil
			})
			if updateErr != nil {
				return updateErr
			}
			fmt.Print(string(out))
			if err != nil {
				return fmt.Errorf("execute command: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "execution timeout")
	return cmd
}

func (a *cliApp) agentKillCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "kill <name>",
		Short: "Terminate agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			return store.Update(func(st *localstate.State) error {
				rec, ok := st.Agents[args[0]]
				if !ok {
					return fmt.Errorf("agent %q not found", args[0])
				}
				rec.State = string(agent.StateTerminated)
				rec.UpdatedAt = now
				st.Agents[args[0]] = rec
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "warn", Agent: args[0], Message: "agent terminated"})
				fmt.Println(a.style.Render("Terminated " + args[0]))
				return nil
			})
		},
	}
}

func (a *cliApp) agentRestartCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			return store.Update(func(st *localstate.State) error {
				rec, ok := st.Agents[args[0]]
				if !ok {
					return fmt.Errorf("agent %q not found", args[0])
				}
				rec.State = string(agent.StateRunning)
				rec.UpdatedAt = now
				st.Agents[args[0]] = rec
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Agent: args[0], Message: "agent restarted"})
				fmt.Println(a.style.Render("Restarted " + args[0]))
				return nil
			})
		},
	}
}

func (a *cliApp) capabilityCmd(statePath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "capability", Short: "Capability management"}
	cmd.AddCommand(
		a.capabilityListCmd(statePath),
		a.capabilityInstallCmd(statePath),
		a.capabilityConfigCmd(statePath),
	)
	return cmd
}

func (a *cliApp) capabilityListCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available capabilities",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			all := append([]string{}, builtinCapabilities...)
			all = append(all, st.InstalledCapabilities...)
			all = dedupeSorted(all)
			for _, name := range all {
				fmt.Println(name)
			}
			return nil
		},
	}
}

func (a *cliApp) capabilityInstallCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "install <name>",
		Short: "Install capability plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			name := args[0]
			now := time.Now().UTC()
			if err := store.Update(func(st *localstate.State) error {
				st.InstalledCapabilities = append(st.InstalledCapabilities, name)
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Message: "capability installed: " + name})
				return nil
			}); err != nil {
				return err
			}
			fmt.Println(a.style.Render("Installed capability: " + name))
			return nil
		},
	}
}

func (a *cliApp) capabilityConfigCmd(statePath *string) *cobra.Command {
	var setKV string
	cmd := &cobra.Command{
		Use:   "config <name>",
		Short: "Configure capability",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			if setKV == "" {
				st, err := store.Load()
				if err != nil {
					return err
				}
				cfg := st.CapabilityConfig[name]
				if len(cfg) == 0 {
					fmt.Printf("No capability config for %s\n", name)
					return nil
				}
				return a.printJSON(cfg)
			}
			k, v, err := parseKV(setKV)
			if err != nil {
				return err
			}
			return store.Update(func(st *localstate.State) error {
				if st.CapabilityConfig[name] == nil {
					st.CapabilityConfig[name] = map[string]string{}
				}
				st.CapabilityConfig[name][k] = v
				fmt.Printf("Set capability %s config %s=%s\n", name, k, v)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&setKV, "set", "", "set key=value")
	return cmd
}

func (a *cliApp) toolCmd(statePath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "tool", Short: "Tool management"}
	cmd.AddCommand(
		a.toolListCmd(statePath),
		a.toolRegisterCmd(statePath),
		a.toolInvokeCmd(statePath),
	)
	return cmd
}

func (a *cliApp) toolListCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered tools",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			for _, name := range builtinTools {
				fmt.Printf("%s\t(builtin)\n", name)
			}
			names := make([]string, 0, len(st.Tools))
			for name := range st.Tools {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				rec := st.Tools[name]
				fmt.Printf("%s\t%s\n", rec.Name, rec.SchemaPath)
			}
			return nil
		},
	}
}

func (a *cliApp) toolRegisterCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "register <schema>",
		Short: "Register new tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := args[0]
			b, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read tool schema: %w", err)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(b, &payload); err != nil {
				if err := yaml.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("decode tool schema: %w", err)
				}
			}
			name, _ := payload["name"].(string)
			if name == "" {
				name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := store.Update(func(st *localstate.State) error {
				st.Tools[name] = localstate.ToolRecord{Name: name, SchemaPath: path, RegisteredAt: now}
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Message: "tool registered: " + name})
				return nil
			}); err != nil {
				return err
			}
			fmt.Printf("Registered tool %s from %s\n", name, path)
			return nil
		},
	}
}

func (a *cliApp) toolInvokeCmd(statePath *string) *cobra.Command {
	var input string
	cmd := &cobra.Command{
		Use:   "invoke <name>",
		Short: "Manually invoke tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if contains(builtinTools, name) {
				out, err := invokeBuiltinTool(name, input)
				if err != nil {
					return err
				}
				fmt.Println(out)
				return nil
			}
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			rec, ok := st.Tools[name]
			if !ok {
				return fmt.Errorf("tool %q not found", name)
			}
			fmt.Printf("Tool %s is registered at %s\n", rec.Name, rec.SchemaPath)
			fmt.Println("Custom tool execution requires runtime plugin integration")
			return nil
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "input payload for tool invocation")
	return cmd
}

func (a *cliApp) meshCmd(statePath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "mesh", Short: "Mesh commands"}
	cmd.AddCommand(
		a.meshStatusCmd(statePath),
		a.meshChannelsCmd(statePath),
		a.meshSendCmd(statePath),
	)
	return cmd
}

func (a *cliApp) meshStatusCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show mesh topology",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			messageCount := 0
			for _, msgs := range st.MeshChannels {
				messageCount += len(msgs)
			}
			fmt.Printf("channels=%d\n", len(st.MeshChannels))
			fmt.Printf("messages=%d\n", messageCount)
			return nil
		},
	}
}

func (a *cliApp) meshChannelsCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "channels",
		Short: "List communication channels",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(st.MeshChannels))
			for name := range st.MeshChannels {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Printf("%s\t%d messages\n", name, len(st.MeshChannels[name]))
			}
			return nil
		},
	}
}

func (a *cliApp) meshSendCmd(statePath *string) *cobra.Command {
	var message string
	var from string
	cmd := &cobra.Command{
		Use:   "send <channel>",
		Short: "Send message to channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("mesh send: --message is required")
			}
			channel := args[0]
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := store.Update(func(st *localstate.State) error {
				st.MeshChannels[channel] = append(st.MeshChannels[channel], localstate.MeshMessage{Channel: channel, From: from, Payload: message, SentAt: now})
				st.Logs = append(st.Logs, localstate.LogEntry{Time: now, Level: "info", Message: fmt.Sprintf("mesh send channel=%s message=%q", channel, message)})
				return nil
			}); err != nil {
				return err
			}
			fmt.Printf("Message sent to %s\n", channel)
			return nil
		},
	}
	cmd.Flags().StringVar(&message, "message", "", "message payload")
	cmd.Flags().StringVar(&from, "from", "", "sender id")
	return cmd
}

func (a *cliApp) logsCmd(statePath *string) *cobra.Command {
	var tail int
	var agentFilter string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream all logs",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			entries := st.Logs
			if agentFilter != "" {
				filtered := entries[:0]
				for _, entry := range entries {
					if entry.Agent == agentFilter {
						filtered = append(filtered, entry)
					}
				}
				entries = filtered
			}
			if tail > 0 && len(entries) > tail {
				entries = entries[len(entries)-tail:]
			}
			for _, entry := range entries {
				agentName := entry.Agent
				if agentName == "" {
					agentName = "system"
				}
				fmt.Printf("%s [%s] (%s) %s\n", entry.Time.Format(time.RFC3339), strings.ToUpper(entry.Level), agentName, entry.Message)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&tail, "tail", 100, "number of log lines")
	cmd.Flags().StringVar(&agentFilter, "agent", "", "filter by agent name")
	return cmd
}

func (a *cliApp) metricsCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Show metrics",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			running := 0
			for _, rec := range st.Agents {
				if rec.State == string(agent.StateRunning) {
					running++
				}
			}
			messageCount := 0
			for _, msgs := range st.MeshChannels {
				messageCount += len(msgs)
			}
			fmt.Printf("daemon_running %t\n", st.Daemon.Running)
			fmt.Printf("agents_total %d\n", len(st.Agents))
			fmt.Printf("agents_running %d\n", running)
			fmt.Printf("logs_total %d\n", len(st.Logs))
			fmt.Printf("traces_total %d\n", len(st.Traces))
			fmt.Printf("mesh_channels_total %d\n", len(st.MeshChannels))
			fmt.Printf("mesh_messages_total %d\n", messageCount)
			return nil
		},
	}
}

func (a *cliApp) traceCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "trace <id>",
		Short: "Get trace details",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			trace, ok := st.Traces[args[0]]
			if !ok {
				return fmt.Errorf("trace %q not found", args[0])
			}
			return a.printJSON(trace)
		},
	}
}

func (a *cliApp) replayCmd(statePath *string) *cobra.Command {
	var stepDelay time.Duration
	cmd := &cobra.Command{
		Use:   "replay <id>",
		Short: "Replay agent decision",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			trace, ok := st.Traces[args[0]]
			if !ok {
				return fmt.Errorf("trace %q not found", args[0])
			}
			for i, step := range trace.Steps {
				fmt.Printf("%d. %s %s\n", i+1, step.Time.Format(time.RFC3339), step.Message)
				if stepDelay > 0 {
					time.Sleep(stepDelay)
				}
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&stepDelay, "step-delay", 0, "delay between replay steps (e.g. 250ms)")
	return cmd
}

func (a *cliApp) devCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Start development mode",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			if err := store.Update(func(st *localstate.State) error {
				st.Config["mode"] = "dev"
				st.Logs = append(st.Logs, localstate.LogEntry{Time: time.Now().UTC(), Level: "info", Message: "development mode enabled"})
				return nil
			}); err != nil {
				return err
			}
			fmt.Println(a.style.Render("Development mode enabled"))
			return nil
		},
	}
}

func (a *cliApp) validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [config]",
		Short: "Validate configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			_, err := agent.LoadConfig(args[0])
			if err != nil {
				return err
			}
			fmt.Println(a.style.Render("Configuration is valid"))
			return nil
		},
	}
}

func (a *cliApp) lintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint [path]",
		Short: "Lint agent configs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			root := "configs/agents"
			if len(args) == 1 {
				root = args[0]
			}
			var files []string
			err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("walk configs: %w", err)
			}
			if len(files) == 0 {
				return fmt.Errorf("no yaml files found in %s", root)
			}
			var lintErrs []string
			for _, file := range files {
				if _, err := agent.LoadConfig(file); err != nil {
					lintErrs = append(lintErrs, fmt.Sprintf("%s: %v", file, err))
				}
			}
			if len(lintErrs) > 0 {
				for _, msg := range lintErrs {
					fmt.Println(msg)
				}
				return fmt.Errorf("lint failed for %d file(s)", len(lintErrs))
			}
			fmt.Printf("Lint passed for %d file(s)\n", len(files))
			return nil
		},
	}
}

func (a *cliApp) testCmd() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run agent tests",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			testCmd := exec.CommandContext(ctx, "go", "test", "./...")
			testCmd.Stdout = os.Stdout
			testCmd.Stderr = os.Stderr
			if err := testCmd.Run(); err != nil {
				return fmt.Errorf("run go test: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "test timeout")
	return cmd
}

func (a *cliApp) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(a.style.Render(version.Info()))
		},
	}
}

func (a *cliApp) doctorCmd(cfgFile, statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose installation",
		RunE: func(_ *cobra.Command, _ []string) error {
			checks := []string{"go", "docker", "runsc", "firecracker"}
			for _, bin := range checks {
				path, err := exec.LookPath(bin)
				if err != nil {
					fmt.Printf("[WARN] %s not found\n", bin)
					continue
				}
				fmt.Printf("[OK] %s: %s\n", bin, path)
			}

			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			if _, err := store.Load(); err != nil {
				return fmt.Errorf("state health: %w", err)
			}
			fmt.Printf("[OK] state file: %s\n", store.Path())

			if *cfgFile != "" {
				if _, err := os.Stat(*cfgFile); err != nil {
					fmt.Printf("[WARN] config file: %v\n", err)
				} else {
					fmt.Printf("[OK] config file: %s\n", *cfgFile)
				}
			}
			return nil
		},
	}
}

func (a *cliApp) upgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade spawn",
		RunE: func(_ *cobra.Command, _ []string) error {
			latest, err := latestRelease(context.Background())
			if err != nil {
				return err
			}
			fmt.Printf("Current: %s\n", version.Info())
			fmt.Printf("Latest:  %s\n", latest.Tag)
			fmt.Printf("Release: %s\n", latest.URL)
			fmt.Println("Upgrade using your installation method (brew/go-install/binary release).")
			return nil
		},
	}
}

func (a *cliApp) configCmd(statePath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage global config"}
	cmd.AddCommand(
		a.configListCmd(statePath),
		a.configGetCmd(statePath),
		a.configSetCmd(statePath),
	)
	return cmd
}

func (a *cliApp) configListCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List global config",
		RunE: func(_ *cobra.Command, _ []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(st.Config))
			for key := range st.Config {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("%s=%s\n", key, st.Config[key])
			}
			return nil
		},
	}
}

func (a *cliApp) configGetCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			st, err := store.Load()
			if err != nil {
				return err
			}
			val, ok := st.Config[args[0]]
			if !ok {
				return fmt.Errorf("config key %q not found", args[0])
			}
			fmt.Println(val)
			return nil
		},
	}
}

func (a *cliApp) configSetCmd(statePath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			store, err := a.openStore(*statePath)
			if err != nil {
				return err
			}
			key, value := args[0], args[1]
			if err := store.Update(func(st *localstate.State) error {
				st.Config[key] = value
				return nil
			}); err != nil {
				return err
			}
			fmt.Printf("Set %s=%s\n", key, value)
			return nil
		},
	}
}

type releaseInfo struct {
	Tag string `json:"tag_name"`
	URL string `json:"html_url"`
}

func latestRelease(ctx context.Context) (*releaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/spawndev/spawn/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("build release request: %w", err)
	}
	req.Header.Set("User-Agent", "spawn-cli")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch latest release: unexpected status %d", resp.StatusCode)
	}
	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode release response: %w", err)
	}
	if info.Tag == "" {
		return nil, errors.New("latest release response missing tag_name")
	}
	return &info, nil
}

func invokeBuiltinTool(name, input string) (string, error) {
	switch name {
	case "calculator":
		if strings.TrimSpace(input) == "" {
			return "", fmt.Errorf("calculator requires --input expression")
		}
		value, err := evalExpression(strings.TrimSpace(input))
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case "datetime":
		return time.Now().UTC().Format(time.RFC3339), nil
	case "json_parser":
		if strings.TrimSpace(input) == "" {
			return "", fmt.Errorf("json_parser requires --input JSON")
		}
		var payload interface{}
		if err := json.Unmarshal([]byte(input), &payload); err != nil {
			return "", fmt.Errorf("parse json input: %w", err)
		}
		b, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return "", fmt.Errorf("format json output: %w", err)
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("unknown builtin tool %q", name)
	}
}

func evalExpression(expr string) (float64, error) {
	// Tiny expression evaluator for calculator tool with +,-,*,/ and parentheses.
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	parser := &exprParser{tokens: tokens}
	value, err := parser.parseExpr()
	if err != nil {
		return 0, err
	}
	if parser.pos != len(parser.tokens) {
		return 0, fmt.Errorf("unexpected token %q", parser.tokens[parser.pos])
	}
	return value, nil
}

type exprParser struct {
	tokens []string
	pos    int
}

func (p *exprParser) parseExpr() (float64, error) {
	lhs, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "+" && op != "-" {
			break
		}
		p.pos++
		rhs, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == "+" {
			lhs += rhs
		} else {
			lhs -= rhs
		}
	}
	return lhs, nil
}

func (p *exprParser) parseTerm() (float64, error) {
	lhs, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "*" && op != "/" {
			break
		}
		p.pos++
		rhs, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if op == "*" {
			lhs *= rhs
		} else {
			if rhs == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			lhs /= rhs
		}
	}
	return lhs, nil
}

func (p *exprParser) parseFactor() (float64, error) {
	if p.pos >= len(p.tokens) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	tok := p.tokens[p.pos]
	if tok == "(" {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos] != ")" {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	}
	if tok == "+" || tok == "-" {
		p.pos++
		v, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if tok == "-" {
			return -v, nil
		}
		return v, nil
	}
	value, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", tok)
	}
	p.pos++
	return value, nil
}

func tokenize(expr string) ([]string, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, fmt.Errorf("empty expression")
	}
	tokens := make([]string, 0, len(trimmed))
	buf := strings.Builder{}
	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}
	for _, r := range trimmed {
		switch {
		case r == ' ' || r == '\t' || r == '\n':
			flush()
		case strings.ContainsRune("+-*/()", r):
			flush()
			tokens = append(tokens, string(r))
		case (r >= '0' && r <= '9') || r == '.':
			buf.WriteRune(r)
		default:
			return nil, fmt.Errorf("invalid character %q", r)
		}
	}
	flush()
	return tokens, nil
}

func parseKV(kv string) (string, string, error) {
	parts := strings.SplitN(kv, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key=value: %q", kv)
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" {
		return "", "", fmt.Errorf("invalid key=value: empty key")
	}
	return key, value, nil
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func dedupeSorted(items []string) []string {
	set := map[string]struct{}{}
	for _, item := range items {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
