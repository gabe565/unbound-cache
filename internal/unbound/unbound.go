package unbound

import (
	"context"
	"os"
	"os/exec"
	"time"
)

func Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "bash", "/unbound.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	return cmd.Run()
}

func AwaitHealthy(ctx context.Context) error {
	for {
		cmd := exec.CommandContext(ctx, "unbound-control", "status")
		if err := cmd.Run(); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}
		return nil
	}
}
