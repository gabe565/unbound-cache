package unbound

import (
	"compress/gzip"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

//nolint:gochecknoglobals
var dumpMu sync.RWMutex

func DumpCache(ctx context.Context, path string) error {
	dumpMu.Lock()
	defer dumpMu.Unlock()

	start := time.Now()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	gzw := gzip.NewWriter(f)
	defer func() {
		_ = gzw.Close()
	}()

	cmd := exec.CommandContext(ctx, "unbound-control", "dump_cache")
	cmd.Stderr = os.Stderr
	cmd.Stdout = gzw
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := gzw.Close(); err != nil {
		return err
	}

	var size uint64
	stat, err := f.Stat()
	if err == nil {
		size = uint64(stat.Size())
	}

	if err := f.Close(); err != nil {
		return err
	}

	slog.Info("Dumped cache", "took", time.Since(start), "size", humanize.IBytes(size))
	return nil
}

func LoadCache(ctx context.Context, path string) error {
	dumpMu.RLock()
	defer dumpMu.RUnlock()

	start := time.Now()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() {
		_ = gzr.Close()
	}()

	cmd := exec.CommandContext(ctx, "unbound-control", "load_cache")
	cmd.Stdin = gzr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := gzr.Close(); err != nil {
		return err
	}

	var size uint64
	info, err := f.Stat()
	if err == nil {
		size = uint64(info.Size())
	}

	slog.Info("Loaded cache", "took", time.Since(start), "size", humanize.IBytes(size))
	return nil
}
