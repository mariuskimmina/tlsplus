package acme

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

type FileStorage struct {
	Path string
}

func NewFileStorage(path string) *FileStorage {
    return &FileStorage{
        Path: path,
    }
}

// Exists returns true if key exists in s.
func (s *FileStorage) Exists(_ context.Context, key string) bool {
	_, err := os.Stat(s.Filename(key))
	return !errors.Is(err, fs.ErrNotExist)
}

// Store saves value at key.
func (s *FileStorage) Store(_ context.Context, key string, value []byte) error {
	filename := s.Filename(key)
	err := os.MkdirAll(filepath.Dir(filename), 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, value, 0600)
}

// Load retrieves the value at key.
func (s *FileStorage) Load(_ context.Context, key string) ([]byte, error) {
	return os.ReadFile(s.Filename(key))
}

// Delete deletes the value at key.
func (s *FileStorage) Delete(_ context.Context, key string) error {
	return os.Remove(s.Filename(key))
}

// List returns all keys that match prefix.
func (s *FileStorage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	var keys []string
	walkPrefix := s.Filename(prefix)

	err := filepath.Walk(walkPrefix, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return fmt.Errorf("%s: file info is nil", fpath)
		}
		if fpath == walkPrefix {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		suffix, err := filepath.Rel(walkPrefix, fpath)
		if err != nil {
			return fmt.Errorf("%s: could not make path relative: %v", fpath, err)
		}
		keys = append(keys, path.Join(prefix, suffix))

		if !recursive && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	return keys, err
}

// Stat returns information about key.
func (s *FileStorage) Stat(_ context.Context, key string) (error) {
    return nil
}

// Filename returns the key as a path on the file
// system prefixed by s.Path.
func (s *FileStorage) Filename(key string) string {
	return filepath.Join(s.Path, filepath.FromSlash(key))
}

// Lock obtains a lock named by the given key. It blocks
// until the lock can be obtained or an error is returned.
func (s *FileStorage) Lock(ctx context.Context, key string) error {
    return nil
}

// Unlock releases the lock for name.
func (s *FileStorage) Unlock(_ context.Context, key string) error {
    return nil
}

func (s *FileStorage) String() string {
	return "FileStorage:" + s.Path
}


// atomicallyCreateFile atomically creates the file
// identified by filename if it doesn't already exist.
func atomicallyCreateFile(filename string, writeLockInfo bool) error {
	// no need to check this error, we only really care about the file creation error
	_ = os.MkdirAll(filepath.Dir(filename), 0700)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if writeLockInfo {
		now := time.Now()
		meta := lockMeta{
			Created: now,
			Updated: now,
		}
		if err := json.NewEncoder(f).Encode(meta); err != nil {
			return err
		}
		// see https://github.com/caddyserver/caddy/issues/3954
		if err := f.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// homeDir returns the best guess of the current user's home
// directory from environment variables. If unknown, "." (the
// current directory) is returned instead.
func homeDir() string {
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		drive := os.Getenv("HOMEDRIVE")
		path := os.Getenv("HOMEPATH")
		home = drive + path
		if drive == "" || path == "" {
			home = os.Getenv("USERPROFILE")
		}
	}
	if home == "" {
		home = "."
	}
	return home
}

func dataDir() string {
	baseDir := filepath.Join(homeDir(), ".local", "share")
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		baseDir = xdgData
	}
	return filepath.Join(baseDir, "certmagic")
}

// lockMeta is written into a lock file.
type lockMeta struct {
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

// lockFreshnessInterval is how often to update
// a lock's timestamp. Locks with a timestamp
// more than this duration in the past (plus a
// grace period for latency) can be considered
// stale.
const lockFreshnessInterval = 5 * time.Second

// fileLockPollInterval is how frequently
// to check the existence of a lock file
const fileLockPollInterval = 1 * time.Second

// Interface guard
var _ Storage = (*FileStorage)(nil)

