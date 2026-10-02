// Package names loads the composer/arranger abbreviation dictionary.
package names

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Load reads a JSON abbreviation map from path. Values are returned unchanged.
// Whitespace warnings are returned as message strings (no trailing newline);
// the dictionary file is never modified.
func Load(path string) (map[string]string, []string, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var ca map[string]string
	if err := json.Unmarshal(f, &ca); err != nil {
		return nil, nil, fmt.Errorf("invalid names.json: %w", err)
	}

	abbreviations := make([]string, 0, len(ca))
	for abbreviation := range ca {
		abbreviations = append(abbreviations, abbreviation)
	}
	sort.Strings(abbreviations)

	var warnings []string
	for _, abbreviation := range abbreviations {
		if abbreviation != strings.TrimSpace(abbreviation) {
			warnings = append(warnings, fmt.Sprintf("WARNING: names.json abbreviation %q has leading or trailing whitespace", abbreviation))
		}
		name := ca[abbreviation]
		if name != strings.TrimSpace(name) {
			warnings = append(warnings, fmt.Sprintf("WARNING: names.json name %q for abbreviation %q has leading or trailing whitespace", name, abbreviation))
		}
	}
	return ca, warnings, nil
}
