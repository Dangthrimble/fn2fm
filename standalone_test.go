package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// CI supplies the absolute path of a separately built app. Ordinary unit tests
// remain runnable without installing or overwriting any fn2fm executable.
func TestBuiltExecutable(t *testing.T) {
	binary := os.Getenv("FN2FM_TEST_BINARY")
	if binary == "" {
		t.Skip("set FN2FM_TEST_BINARY to the absolute path of the built app")
	}
	if !filepath.IsAbs(binary) {
		t.Fatal("FN2FM_TEST_BINARY must be an absolute path")
	}
	if info, err := os.Stat(binary); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("cannot use built app %q: %v", binary, err)
	}

	for _, tc := range []struct {
		fixture, filename string
		want              pdfMetadata
	}{
		{"no-info-table.pdf", "Table Score ~ JoRu_DaFo[C]+.pdf", pdfMetadata{Author: "John Rutter", Subject: "Dan Forrest", Keywords: "keysf:0, keymi:0, With Accompaniment"}},
		{"no-info-stream.pdf", "Stream Score ~ _DaFo[4b]-.pdf", pdfMetadata{Subject: "Dan Forrest", Keywords: "Without Accompaniment"}},
		{"no-info-table.pdf", "Title Score ~ [0].pdf", pdfMetadata{}},
	} {
		t.Run(tc.filename, func(t *testing.T) {
			dir := builtCommandDirectory(t)
			path := filepath.Join(dir, tc.filename)
			source := readTestFile(t, filepath.Join("testdata", tc.fixture))
			writeTestFile(t, path, source)
			tc.want.Title = strings.TrimSuffix(tc.filename, ".pdf")

			output, err := runBuiltCommand(t, binary, dir, path)
			if err != nil {
				t.Fatalf("built app failed: %v\n%s", err, output)
			}
			assertBuiltMetadata(t, path, source, tc.want)

			// Change the filename to remove arranger/key/accompaniment and rerun.
			// Keep the first backup alongside the renamed PDF to exercise its reuse.
			nextPath := filepath.Join(dir, "Updated Score ~ JoRu[0].pdf")
			if err := os.Rename(path, nextPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(path+"_original", nextPath+"_original"); err != nil {
				t.Fatal(err)
			}
			first := readTestFile(t, nextPath)
			output, err = runBuiltCommand(t, binary, dir, nextPath)
			if err != nil {
				t.Fatalf("second update failed: %v\n%s", err, output)
			}
			assertBuiltMetadata(t, nextPath, source, pdfMetadata{Title: "Updated Score ~ JoRu[0]", Author: "John Rutter"})
			if !bytes.HasPrefix(readTestFile(t, nextPath), first) {
				t.Fatal("second update did not retain the first PDF revision")
			}
		})
	}

	t.Run("invalid PDF retains filename and bytes", func(t *testing.T) {
		dir := builtCommandDirectory(t)
		path := filepath.Join(dir, "Invalid Score ~ JoRu[C]+.pdf")
		source := []byte("This is not a PDF")
		writeTestFile(t, path, source)
		output, err := runBuiltCommand(t, binary, dir, path)
		if err == nil || !strings.Contains(string(output), "Unable to update") {
			t.Fatalf("expected a reported write failure: %v\n%s", err, output)
		}
		if !bytes.Equal(readTestFile(t, path), source) {
			t.Fatal("failed command altered the source")
		}
		if _, err := os.Lstat(path + "_original"); !os.IsNotExist(err) {
			t.Fatal("rejected input created a backup")
		}
		assertNoTemporaryPDFs(t, dir)
	})

	t.Run("Windows open file prevents replacement", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows file sharing semantics require a Windows runner")
		}
		dir := builtCommandDirectory(t)
		path := filepath.Join(dir, "Open Score ~ JoRu[C]+.pdf")
		source := readTestFile(t, "testdata/no-info-table.pdf")
		writeTestFile(t, path, source)
		// Go opens this handle with read/write sharing, but without delete sharing.
		// The child can read the PDF yet cannot replace it until this handle closes.
		held, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer held.Close()
		output, err := runBuiltCommand(t, binary, dir, path)
		if err == nil || !strings.Contains(string(output), "replace PDF (original retained)") {
			t.Fatalf("expected a replacement failure: %v\n%s", err, output)
		}
		if !bytes.Equal(readTestFile(t, path), source) || !bytes.Equal(readTestFile(t, path+"_original"), source) {
			t.Fatal("replacement failure changed original or backup")
		}
		assertNoTemporaryPDFs(t, dir)
		if err := held.Close(); err != nil {
			t.Fatal(err)
		}
		output, err = runBuiltCommand(t, binary, dir, path)
		if err != nil {
			t.Fatalf("retry after closing PDF failed: %v\n%s", err, output)
		}
		assertBuiltMetadata(t, path, source, pdfMetadata{Title: "Open Score ~ JoRu[C]+", Author: "John Rutter", Keywords: "keysf:0, keymi:0, With Accompaniment"})
	})
}

func builtCommandDirectory(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "Scores with spaces")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "names.json"), []byte(`{"JoRu":" John Rutter ","DaFo":" Dan Forrest "}`))
	return dir
}

func runBuiltCommand(t *testing.T, binary, dir, path string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, path)
	cmd.Dir = dir
	// Keep Windows system variables, but make external programs unavailable.
	cmd.Env = append(os.Environ(), "PATH="+dir)
	return cmd.CombinedOutput()
}

func assertBuiltMetadata(t *testing.T, path string, original []byte, want pdfMetadata) {
	t.Helper()
	updated := readTestFile(t, path)
	if !bytes.HasPrefix(updated, original) {
		t.Fatal("source bytes were not preserved")
	}
	if !bytes.Equal(readTestFile(t, path+"_original"), original) {
		t.Fatal("first backup was not preserved")
	}
	ctx := readTestPDF(t, updated)
	if ctx.PageCount != 2 || ctx.Info == nil {
		t.Fatal("incorrect page count or missing metadata dictionary")
	}
	info, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range want.fields() {
		if value == "" {
			if _, exists := info[key]; exists {
				t.Errorf("empty %s was not removed", key)
			}
			continue
		}
		got, err := ctx.DereferenceText(info[key])
		if err != nil || got != value {
			t.Errorf("%s = %q, %v; want %q", key, got, err, value)
		}
	}
	assertNoTemporaryPDFs(t, filepath.Dir(path))
}
