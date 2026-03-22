package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/travis-james/proxy-replay/internal/types"
)

func TestFileStorage_Save(t *testing.T) {
	dir := t.TempDir()
	fs := FileStorage{Dir: dir}

	rec := types.Recording{}

	err := fs.Save("test", rec)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	path := filepath.Join(dir, "test.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestFileStorage_Load(t *testing.T) {
	dir := t.TempDir()
	fs := FileStorage{Dir: dir}
	key := "test"
	finalPath := filepath.Join(fs.Dir, key+".json")

	rec := types.Recording{}

	data, err := json.MarshalIndent(rec, "", " ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	err = os.WriteFile(finalPath, data, 0644)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := fs.Load(key)
	if err != nil {
		t.Fatalf("load faield: %v", err)
	}

	if !reflect.DeepEqual(got, rec) {
		t.Fatalf("expected:\n%v\ngot:%v\n", rec, got)
	}
}
