package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
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
