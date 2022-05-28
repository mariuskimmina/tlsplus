package acme

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type FileStorage struct {
    Path string
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

// Filename returns the key as a path on the file
// system prefixed by s.Path.
func (s *FileStorage) Filename(key string) string {
	return filepath.Join(s.Path, filepath.FromSlash(key))
}

