//go:build darwin && arm64

package virtualizationframework

import (
	"archive/tar"
	"bytes"
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
	var buf []byte
	for _, pathname := range []string{"disk.img", "nvram.bin"} {
		srcpath := filepath.Join(imageDir, pathname)
		dstpath := filepath.Join(workingDir, pathname)

		// use the more efficient clonefile is possible: this will error if cross-device.
		if err := unix.Clonefile(srcpath, dstpath, unix.CLONE_NOFOLLOW); err == nil {
			continue
		}

		if len(buf) == 0 { // lazily initialize buf size when needed
			buf = make([]byte, len(sparseBlock))
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

		if err := sparseCopyBuffer(dst, src, buf); err != nil {
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
	buf := make([]byte, len(sparseBlock))
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

		if err := sparseCopyBuffer(dst, tr, buf); err != nil {
			return fmt.Errorf("copying %s -> %s: %w", hdr.Name, dstpath, err)
		}
		if err := dst.Close(); err != nil {
			return fmt.Errorf("flushing %s: %w", dstpath, err)
		}
	}

	return nil
}

var sparseBlock = make([]byte, 64*1024)

func sparseCopyBuffer(dst io.WriterAt, src io.Reader, buf []byte) error {
	if buf == nil || len(buf) < len(sparseBlock) {
		panic("sparse copy buffer cannot be smaller than the sparse block size")
	}

	var offset int64

	for {
		n, err := src.Read(buf)
		if !bytes.Equal(buf[:n], sparseBlock[:n]) {
			if _, err := dst.WriteAt(buf[:n], offset); err != nil {
				return err
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		offset += int64(n)
	}
}
