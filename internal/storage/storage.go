package storage

import (
	"io"
	"os"
	"path/filepath"
)

type FileStorage interface {
	Write(id string, r io.Reader) error
	Read(id string) ([]byte, error)
	Delete(id string) error
}

type DiskStorage struct {
	root string
}

func NewDiskStorage(root string) FileStorage {
	return &DiskStorage{root: root}
}

func (s *DiskStorage) Write(id string, r io.Reader) error {
	dir := filepath.Join(s.root, id)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "content"), data, 0600)
}

func (s *DiskStorage) Read(id string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.root, id, "content"))
}

func (s *DiskStorage) Delete(id string) error {
	return os.RemoveAll(filepath.Join(s.root, id))
}
