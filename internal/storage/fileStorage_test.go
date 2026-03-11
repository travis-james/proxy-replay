package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileStorage_Save(t *testing.T) {
	dir := t.TempDir()
	fs := FileStorage{Dir: dir}

	rec := Recording{}

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

	rec := Recording{}

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

func TestFileStorage_List(t *testing.T) {
	dir := t.TempDir()
	fs := FileStorage{Dir: dir}
	key := "test"
	finalPath := filepath.Join(fs.Dir, key+".json")

	rec := Recording{}

	data, err := json.MarshalIndent(rec, "", " ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	err = os.WriteFile(finalPath, data, 0644)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := fs.List()
	if err != nil {
		t.Fatalf("load faield: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 recording, got %d", len(got))
	}

	if got[0].Key != key {
		t.Fatalf("expected key: %s, got: %s", key, got[0].Key)
	}

	if got[0].SizeBytes <= 0 {
		t.Fatalf("expected positive file size")
	}

	if got[0].Timestamp.IsZero() {
		t.Fatalf("expected timestamp to be set")
	}
}
