package index

import (
	"io"
	"os"
	"path/filepath"
)

// writeFileAtomic writes path through a synced temp file in the same directory and renames
// it into place, so readers never see a partially written or truncated chunk. The temp name
// is hidden and does not parse as a block range, so index folder walkers ignore leftovers.
func writeFileAtomic(path string, write func(io.Writer) error) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	if err := write(tmp); err != nil {
		return err
	}
	// CreateTemp is owner-only; chunks are published world-readable.
	if err := tmp.Chmod(0644); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
