package migrations

import (
	"io/fs"
	"testing"
)

func TestFilesEmbedded(t *testing.T) {
	entries, err := fs.ReadDir(Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no embedded migration files")
	}
	for _, e := range entries {
		t.Log("embedded:", e.Name())
	}
}
