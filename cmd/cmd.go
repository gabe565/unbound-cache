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

	group, ctx := errgroup.WithContext(cmd.Context())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Runs Unbound
	group.Go(func() error {
		if err := unbound.Run(ctx); !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	// Loads previous cacheduring startup
	var loaded bool
	group.Go(func() error {
		subCtx, subCancel := context.WithTimeout(ctx, 10*time.Minute)
		defer subCancel()
		if err := unbound.AwaitHealthy(subCtx); err != nil {
			return err
		}
		if err := unbound.LoadCache(subCtx, conf.CachePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		loaded = true
		return nil
	})

	// Dumps cache before exit
	group.Go(func() error {
		defer cancel()
		signalCtx, cancelSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		defer cancelSignal()

		select {
		case <-ctx.Done():
			return nil
		case <-signalCtx.Done():
			if loaded {
				subCtx, subCancel := context.WithTimeout(ctx, time.Minute)
				defer subCancel()
				return unbound.DumpCache(subCtx, conf.CachePath)
			}
			return nil
		}
	})

	// Regularly dumps cache
	group.Go(func() error {
		ticker := time.NewTicker(conf.DumpEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if loaded {
					subCtx, subCancel := context.WithTimeout(ctx, time.Minute)
					if err := unbound.DumpCache(subCtx, conf.CachePath); err != nil {
						slog.Error("Failed to dump cache", "error", err.Error())
					}
					subCancel()
				}
			}
		}
	})

	return group.Wait()
}
