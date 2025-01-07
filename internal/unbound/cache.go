package unbound

import (
	"compress/gzip"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/labstack/gommon/bytes"
)

//nolint:gochecknoglobals
var dumpMu sync.RWMutex

func DumpCache(ctx context.Context, path string) error {
	start := time.Now()

	f, err := os.CreateTemp(filepath.Dir(path), ".cache-*.txt.gz")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
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

	var size int64
	stat, err := f.Stat()
	if err == nil {
		size = stat.Size()
	}

	if err := f.Close(); err != nil {
		return err
	}

	dumpMu.Lock()
	defer dumpMu.Unlock()

	if err := os.Rename(f.Name(), path); err != nil {
		return err
	}

	slog.Info("Dumped cache", "took", time.Since(start), "size", bytes.Format(size))
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

	var size int64
	info, err := f.Stat()
	if err == nil {
		size = info.Size()
	}

	slog.Info("Loaded cache", "took", time.Since(start), "size", bytes.Format(size))
	return nil
}
