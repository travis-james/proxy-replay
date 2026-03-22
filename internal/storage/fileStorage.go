package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/travis-james/proxy-replay/internal/types"
)

type FileStorage struct {
	Dir string // Directory to write to.
}

// Save a recorded Request/Response pair to key (file name).
func (fs FileStorage) Save(key string, rec types.Recording) (err error) {
	if err := os.MkdirAll(fs.Dir, 0755); err != nil {
		return err
	}
	// Since os.WriteFile requires multiple system calls to complete,
	// a failure mid-operation can leave the file in a partially written
	// state.
	// SO INSTEAD:
	// 1. write contents to a temp file in the same directory.
	// 2. Typically the OS buffers data in memory. 'fsync' forces OS to
	// flush the buffer to disk.
	// 3. Use rename so temp file becomes the target file. Reason being
	// this is atomic, either we get an update, or the old file stays
	// untouched.
	finalPath := filepath.Join(fs.Dir, key+".json")
	data, err := json.MarshalIndent(rec, "", " ")
	if err != nil {
		return err
	}

	// create temp file in same directory.
	tmpFile, err := os.CreateTemp(fs.Dir, "temp-*")
	if err != nil {
		return err
	}
	// write data.
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return err
	}
	// sync/force data out of buffer to disk.
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return err
	}
	// close temp file.
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return err
	}

	return os.Rename(tmpFile.Name(), finalPath)
}

// Load a saved req/resp.
func (fs FileStorage) Load(key string) (rec types.Recording, err error) {
	finalPath := filepath.Join(fs.Dir, key+".json")
	data, err := os.ReadFile(finalPath)
	if err != nil {
		return rec, err
	}

	err = json.Unmarshal(data, &rec)
	if err != nil {
		return rec, err
	}
	return rec, nil
}
