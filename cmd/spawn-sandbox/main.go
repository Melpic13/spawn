package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"spawn.dev/pkg/sandbox"
)

func main() {
	cmd := &cobra.Command{
		Use:   "spawn-sandbox",
		Short: "sandbox helper",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("command required")
			}
			runtime := sandbox.NewNativeRuntime()
			sb, err := runtime.Create(context.Background(), sandbox.DefaultConfig())
			if err != nil {
				return err
			}
			if err := sb.Start(context.Background()); err != nil {
				return err
			}
			res, err := sb.Exec(context.Background(), &sandbox.Command{Path: args[0], Args: args[1:], Timeout: 60 * time.Second})
			if err != nil {
				return err
			}
			fmt.Print(res.Stdout)
			if res.ExitCode != 0 {
				return fmt.Errorf("sandbox command exited with code %d", res.ExitCode)
			}
			return nil
		},
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
