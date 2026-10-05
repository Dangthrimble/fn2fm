package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Invoke the real command entry point in a subprocess so os.Exit and working
// directory behaviour are covered without needing a separately installed app.
func TestFN2FMCLIHelper(t *testing.T) {
	if os.Getenv("FN2FM_TEST_MAIN") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"fn2fm"}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing helper arguments")
}

func TestFN2FMStandaloneCommand(t *testing.T) {
	for _, tc := range []struct {
		name, dictionary string
		input            []byte
		wantError        bool
	}{
		{"no external programs", `{"JoRu":" John Rutter "}`, existingInfoPDF(false), false},
		{"unsupported input", `{"JoRu":"John Rutter"}`, []byte("not a PDF"), true},
		{"malformed dictionary", `{"JoRu":`, existingInfoPDF(false), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			name := "Score ~ JoRu[C].pdf"
			path := filepath.Join(dir, name)
			writeTestFile(t, path, tc.input)
			writeTestFile(t, filepath.Join(dir, "names.json"), []byte(tc.dictionary))
			cmd := exec.Command(os.Args[0], "-test.run=^TestFN2FMCLIHelper$", "--", name)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "FN2FM_TEST_MAIN=1", "PATH="+dir)
			output, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected exit: %v\n%s", err, output)
			}
			if tc.name == "malformed dictionary" && !strings.Contains(string(output), "Unable to load names.json: invalid names.json:") {
				t.Fatalf("missing malformed dictionary diagnostic: %s", output)
			}
			if tc.wantError {
				if !bytes.Equal(readTestFile(t, path), tc.input) {
					t.Fatal("failed command changed or renamed original PDF")
				}
				if _, err := os.Lstat(path + "_original"); !os.IsNotExist(err) {
					t.Fatal("failed command created a backup")
				}
				return
			}
			ctx := readTestPDF(t, readTestFile(t, path))
			if ctx.Title != "Score ~ JoRu[C]" || ctx.Author != "John Rutter" || ctx.Subject != "" || ctx.Keywords != "keysf:0, keymi:0" {
				t.Fatalf("incorrect CLI metadata: %q / %q / %q / %q", ctx.Title, ctx.Author, ctx.Subject, ctx.Keywords)
			}
			if !bytes.Equal(readTestFile(t, path+"_original"), tc.input) {
				t.Fatal("command did not preserve original backup")
			}
		})
	}
}

func TestFN2FMRenameFailure(t *testing.T) {
	dir := t.TempDir()
	name := "Bad Score.pdf"
	path := filepath.Join(dir, name)
	source := existingInfoPDF(false)
	writeTestFile(t, path, source)
	writeTestFile(t, filepath.Join(dir, "names.json"), []byte(`{"JoRu":"John Rutter"}`))
	if err := os.Mkdir(path+"_rename", 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestFN2FMCLIHelper$", "--", name)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "FN2FM_TEST_MAIN=1", "PATH="+dir)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected nonzero exit when rename fails\n%s", output)
	}
	if !strings.Contains(string(output), "unable to rename") {
		t.Fatalf("missing rename failure diagnostic: %s", output)
	}
	if !bytes.Equal(readTestFile(t, path), source) {
		t.Fatal("failed rename changed or removed the original PDF")
	}
}

func TestFN2FMEmptyKeyCommand(t *testing.T) {
	for _, tc := range []struct{ suffix, keywords string }{
		{"[]", ""},
		{"[]+", "With Accompaniment"},
		{"[   ]-", "Without Accompaniment"},
	} {
		t.Run(tc.suffix, func(t *testing.T) {
			dir := t.TempDir()
			name := "Atonal Score ~ JoRu" + tc.suffix + ".pdf"
			path := filepath.Join(dir, name)
			source := existingInfoPDF(false)
			writeTestFile(t, path, source)
			writeTestFile(t, filepath.Join(dir, "names.json"), []byte(`{"JoRu":"John Rutter"}`))
			cmd := exec.Command(os.Args[0], "-test.run=^TestFN2FMCLIHelper$", "--", name)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "FN2FM_TEST_MAIN=1", "PATH="+dir)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("command failed: %v\n%s", err, output)
			}
			ctx := readTestPDF(t, readTestFile(t, path))
			if ctx.Title != strings.TrimSuffix(name, ".pdf") || ctx.Author != "John Rutter" || ctx.Keywords != tc.keywords {
				t.Fatalf("unexpected metadata: title=%q author=%q keywords=%q", ctx.Title, ctx.Author, ctx.Keywords)
			}
			if !bytes.Equal(readTestFile(t, path+"_original"), source) {
				t.Fatal("command did not preserve original backup")
			}
		})
	}
}
