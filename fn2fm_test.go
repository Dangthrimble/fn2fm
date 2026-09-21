package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadComposersArrangersWhitespaceWarnings(t *testing.T) {
	// This test changes the working directory and logger; do not run it in parallel.
	previousDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDirectory); err != nil {
			t.Error(err)
		}
	})
	previousWriter, previousFlags, previousPrefix := log.Writer(), log.Flags(), log.Prefix()
	var warnings bytes.Buffer
	log.SetOutput(&warnings)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	})

	for _, tc := range []struct {
		name         string
		entries      map[string]string
		warningCount int
		fragments    []string
	}{
		{"clean and internal spaces", map[string]string{"DaFo": "Dan Forrest", "A B": "Two Names"}, 0, nil},
		{"abbreviation edges", map[string]string{" JoRu": "John Rutter", "DaFo ": "Dan Forrest"}, 2, []string{`abbreviation " JoRu"`, `abbreviation "DaFo "`}},
		{"name edges", map[string]string{"JoRu": "John Rutter ", "DaFo": " Dan Forrest"}, 2, []string{`name "John Rutter " for abbreviation "JoRu"`, `name " Dan Forrest" for abbreviation "DaFo"`}},
		{"other whitespace and both fields", map[string]string{"\tDaFo": "Dan Forrest\n", "JoRu": "\u00a0John Rutter"}, 3, []string{`abbreviation "\tDaFo"`, `name "Dan Forrest\n"`, `for abbreviation "JoRu"`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			warnings.Reset()
			contents, err := json.Marshal(tc.entries)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, "names.json")
			if err := os.WriteFile(path, contents, 0600); err != nil {
				t.Fatal(err)
			}
			got, err := readComposersArrangers()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.entries) {
				t.Fatalf("dictionary values changed: got %#v, want %#v", got, tc.entries)
			}
			if count := strings.Count(warnings.String(), "WARNING: names.json"); count != tc.warningCount {
				t.Fatalf("got %d warnings, want %d: %s", count, tc.warningCount, &warnings)
			}
			for _, fragment := range tc.fragments {
				if !strings.Contains(warnings.String(), fragment) {
					t.Errorf("missing %q in warnings: %s", fragment, &warnings)
				}
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, contents) {
				t.Fatal("dictionary file changed")
			}
		})
	}
}

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
