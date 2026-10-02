package names

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadWhitespaceWarnings(t *testing.T) {
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
			contents, err := json.Marshal(tc.entries)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "names.json")
			if err := os.WriteFile(path, contents, 0600); err != nil {
				t.Fatal(err)
			}
			got, warnings, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.entries) {
				t.Fatalf("dictionary values changed: got %#v, want %#v", got, tc.entries)
			}
			joined := strings.Join(warnings, "\n")
			if count := strings.Count(joined, "WARNING: names.json"); count != tc.warningCount {
				t.Fatalf("got %d warnings, want %d: %s", count, tc.warningCount, joined)
			}
			if len(warnings) != tc.warningCount {
				t.Fatalf("got %d warning lines, want %d: %#v", len(warnings), tc.warningCount, warnings)
			}
			for _, fragment := range tc.fragments {
				if !strings.Contains(joined, fragment) {
					t.Errorf("missing %q in warnings: %s", fragment, joined)
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
