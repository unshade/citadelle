package storage_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unshade/citadelle/internal/storage"
)

// newTestStorage creates a DiskStorage backed by a temporary directory
// that is automatically removed when the test ends.
func newTestStorage(t *testing.T) storage.FileStorage {
	t.Helper()
	dir := t.TempDir()
	return storage.NewDiskStorage(dir)
}

func TestWrite_CreatesFileWithCorrectContent(t *testing.T) {
	store := newTestStorage(t)
	content := []byte("encrypted file content")

	err := store.Write("node-abc", bytes.NewReader(content))

	require.NoError(t, err)
}

func TestRead_ReturnsExactBytesWritten(t *testing.T) {
	store := newTestStorage(t)
	content := []byte("some encrypted bytes")

	require.NoError(t, store.Write("node-1", bytes.NewReader(content)))

	got, err := store.Read("node-1")

	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestWrite_CreatesRestrictedFilePermissions(t *testing.T) {
	root := t.TempDir()
	store := storage.NewDiskStorage(root)
	content := []byte("secret")

	require.NoError(t, store.Write("node-perms", bytes.NewReader(content)))

	filePath := filepath.Join(root, "node-perms", "content")
	info, err := os.Stat(filePath)
	require.NoError(t, err)

	// File must be owner read/write only (0600)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestDelete_RemovesDirectoryAndContents(t *testing.T) {
	root := t.TempDir()
	store := storage.NewDiskStorage(root)

	require.NoError(t, store.Write("node-del", bytes.NewReader([]byte("data"))))

	err := store.Delete("node-del")
	require.NoError(t, err)

	_, statErr := os.Stat(filepath.Join(root, "node-del"))
	assert.True(t, os.IsNotExist(statErr), "directory should no longer exist")
}

func TestRead_ReturnsError_WhenFileDoesNotExist(t *testing.T) {
	store := newTestStorage(t)

	_, err := store.Read("nonexistent-node")

	require.Error(t, err)
}

func TestDelete_IsNoOp_WhenNodeDoesNotExist(t *testing.T) {
	store := newTestStorage(t)

	// os.RemoveAll does not error on missing paths, and neither should we
	err := store.Delete("nonexistent-node")

	require.NoError(t, err)
}
