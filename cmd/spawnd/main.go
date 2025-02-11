package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"spawn.dev/pkg/config"
	"spawn.dev/pkg/gateway"
)

func main() {
	cmd := &cobra.Command{
		Use:   "spawnd",
		Short: "spawn daemon",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			gw := gateway.New(gateway.Config{
				GRPCAddr: fmt.Sprintf(":%d", cfg.Server.Ports.GRPC),
				RESTAddr: fmt.Sprintf(":%d", cfg.Server.Ports.REST),
				WSAddr:   fmt.Sprintf(":%d", cfg.Server.Ports.REST+1),
			})
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := gw.Start(ctx); err != nil {
				return err
			}
			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return gw.Stop(shutdownCtx)
		},
	}
	cmd.Flags().String("config", "configs/spawn.yaml", "daemon config path")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
