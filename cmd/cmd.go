package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabe565/unbound-cache/internal/config"
	"github.com/gabe565/unbound-cache/internal/unbound"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "unbound-cache",
		RunE: run,
	}
	conf := config.New()
	conf.RegisterFlags(cmd)
	cmd.SetContext(config.NewContext(context.Background(), conf))
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	conf, ok := config.FromContext(cmd.Context())
	if !ok {
		panic("Config missing from command context")
	}

	if err := conf.Load(cmd); err != nil {
		return err
	}

	group, groupCtx := errgroup.WithContext(cmd.Context())

	unboundCtx, unboundCancel := context.WithCancel(groupCtx)
	defer unboundCancel()

	group.Go(func() error {
		if err := unbound.Run(unboundCtx); !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		defer unboundCancel()
		signalCtx, cancelSignal := signal.NotifyContext(unboundCtx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		defer cancelSignal()

		select {
		case <-groupCtx.Done():
			return nil
		case <-signalCtx.Done():
			ctx, cancel := context.WithTimeout(unboundCtx, time.Minute)
			defer cancel()
			return unbound.DumpCache(ctx, conf.CachePath)
		}
	})

	group.Go(func() error {
		ctx, cancel := context.WithTimeout(unboundCtx, 10*time.Minute)
		defer cancel()
		if err := unbound.AwaitHealthy(ctx); err != nil {
			return err
		}
		if err := unbound.LoadCache(ctx, conf.CachePath); !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		ticker := time.NewTicker(conf.DumpEvery)
		defer ticker.Stop()
		for {
			select {
			case <-unboundCtx.Done():
				return nil
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(unboundCtx, time.Minute)
				if err := unbound.DumpCache(ctx, conf.CachePath); err != nil {
					slog.Error("Failed to dump cache", "error", err.Error())
				}
				cancel()
			}
		}
	})

	return group.Wait()
}
