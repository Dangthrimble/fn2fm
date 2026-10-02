package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchivePreservesExecutableAndDictionary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.zip")
	binary := []byte("test executable bytes")
	dictionary := []byte(`{"JoRu":"John Rutter"}`)
	if err := writeArchive(path, "fn2fm-dev-test", []entry{
		{"fn2fm", binary, 0755},
		{"names.example.json", dictionary, 0644},
	}); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	contents := map[string][]byte{}
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimPrefix(file.Name, "fn2fm-dev-test/")
		contents[name] = data
		if name == "fn2fm" && file.Mode().Perm() != 0755 {
			t.Fatal("executable permissions lost")
		}
	}
	if !bytes.Equal(contents["fn2fm"], binary) || !bytes.Equal(contents["names.example.json"], dictionary) {
		t.Fatal("packaged bytes differ")
	}
	if _, exists := contents["names.json"]; exists {
		t.Fatal("package could overwrite a personal dictionary")
	}
	manifest := string(contents["SHA256SUMS.txt"])
	for name, data := range map[string][]byte{"fn2fm": binary, "names.example.json": dictionary} {
		if !strings.Contains(manifest, fmt.Sprintf("%x  %s\n", sha256.Sum256(data), name)) {
			t.Fatalf("missing or incorrect checksum for %s", name)
		}
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeArchive(path, "replacement", nil); err == nil {
		t.Fatal("existing package was overwritten")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("existing package changed after refusal")
	}
}
