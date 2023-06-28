//go:build darwin && arm64

package virtualizationframework

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"golang.org/x/sys/unix"
)

func (hv *VirtualizationFramework) cloneVM(ctx context.Context, id, name string) (cfg *VirtualMachineConfig, err error) {
	defer func() {
		if err != nil {
			os.RemoveAll(filepath.Join(hv.cfg.WorkingDirectory, id))
		}
	}()

	imageDir := filepath.Join(hv.cfg.ImageDirectory, name)
	workingDir := filepath.Join(hv.cfg.WorkingDirectory, id)

	rawVmCfg, err := os.ReadFile(filepath.Join(imageDir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("reading vm config: %w", err)
	}

	cfg = &VirtualMachineConfig{}
	if err := json.Unmarshal(rawVmCfg, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling vm config: %w", err)
	}

	if err := os.MkdirAll(workingDir, 0o777); err != nil {
		return nil, fmt.Errorf("creating image directory: %w", err)
	}

	f, err := os.Open(filepath.Join(imageDir, "archive.tar.zst"))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, extractFromDisk(imageDir, workingDir)
	}
	if err != nil {
		return nil, fmt.Errorf("opening compressed archive: %w", err)
	}
	defer f.Close()

	return cfg, extractFromArchive(f, workingDir)
}

func extractFromDisk(imageDir, workingDir string) error {
	for _, pathname := range []string{"disk.img", "nvram.bin"} {
		srcpath := filepath.Join(imageDir, pathname)
		dstpath := filepath.Join(workingDir, pathname)

		// use the more efficient clonefile is possible: this will error if cross-device.
		if err := unix.Clonefile(srcpath, dstpath, unix.CLONE_NOFOLLOW); err == nil {
			continue
		}

		src, err := os.Open(srcpath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", srcpath, err)
		}
		defer src.Close()

		dst, err := os.Create(dstpath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", dstpath, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copying %s -> %s: %w", srcpath, dstpath, err)
		}

		src.Close()
		if err := dst.Close(); err != nil {
			return fmt.Errorf("flushing %s: %w", dstpath, err)
		}
	}

	return nil
}

func extractFromArchive(r io.Reader, workingDir string) error {
	zr, err := zstd.NewReader(r)
	if err != nil {
		return fmt.Errorf("creating zstd reader: %w", err)
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar entry: %w", err)
		}

		dstpath := filepath.Join(workingDir, hdr.Name)

		dst, err := os.Create(dstpath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", dstpath, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, tr); err != nil {
			return fmt.Errorf("copying %s -> %s: %w", hdr.Name, dstpath, err)
		}
		if err := dst.Close(); err != nil {
			return fmt.Errorf("flushing %s: %w", dstpath, err)
		}
	}

	return nil
}
