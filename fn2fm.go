/*
Fn2fm parses the filename of a music score in PDF format, deriving and writing
PDF metadata to that file that can subsequently be fetched by forScore
[https://forscore.co/].
It uses specific characters to deliniate the metadata embedded in the filename.
It requires a JSON file names.json in the folder from which it is invoked.

Usage:

	fn2fm [file ...]
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {

	var (
		err    error
		ca     map[string]string // Map of composers and arrangers with abbreviation as the key
		file   string            // Song's file
		fn     string            // Song's filename without the path but including the extension
		md     string            // Song's metadata embedded in the filename
		comp   string            // Song's composer(s)
		arr    string            // Song's arranger(s)
		key    string            // Song's initial key signature
		acc    string            // Whether or not the song has an accompaniment
		failed bool              // Whether a PDF update failed
	)

	ca, err = readComposersArrangers()
	if err != nil {
		fmt.Printf("Unable to load names.json: %v\n", err)
		os.Exit(1)
	}

	for _, file = range os.Args[1:] {
		fmt.Printf("Parsing %q...\n", file)
		fn, err = validateFilenameAndExtension(file)
		if err != nil {
			os.Rename(file, file+"_rename")
			continue
		}

		md, err = validateMetadataTags(fn)
		if err != nil {
			os.Rename(file, file+"_rename")
			continue
		}

		comp, err = parseComposers(md, ca)
		if err != nil {
			continue
		}

		arr, err = parseArrangers(md, ca)
		if err != nil {
			continue
		}

		key, err = parseKey(md)
		if err != nil {
			continue
		}

		acc, err = parseAccompaniment(md)
		if err != nil {
			continue
		}

		fmt.Printf("    Composer(s): %q\n", comp)
		fmt.Printf("    Arranger(s): %q\n", arr)
		fmt.Printf("            Key: %q\n", key)
		fmt.Printf("  Accompaniment: %q\n\n", acc)

		err = writePDFMetadata(file, pdfMetadata{
			Title: fn, Author: comp, Subject: arr, Keywords: metadataKeywords(key, acc),
		})
		if err != nil {
			log.Printf("Unable to update %q: %v", file, err)
			failed = true
			continue
		}
		fmt.Printf("  Updated PDF metadata\n")
	}
	if failed {
		os.Exit(1)
	}
}

func readComposersArrangers() (map[string]string, error) {

	var (
		err error
		f   []byte            // Contents of file of composers' and arrangers' abbreviations and names
		ca  map[string]string // Map of composers and arrangers with abbreviation as the key
	)
	f, err = os.ReadFile("names.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err := json.Unmarshal(f, &ca); err != nil {
		return nil, fmt.Errorf("invalid names.json: %w", err)
	}
	// Sort abbreviations so warnings have a consistent order.
	abbreviations := make([]string, 0, len(ca))
	for abbreviation := range ca {
		abbreviations = append(abbreviations, abbreviation)
	}
	sort.Strings(abbreviations)
	for _, abbreviation := range abbreviations {
		if abbreviation != strings.TrimSpace(abbreviation) {
			log.Printf("WARNING: names.json abbreviation %q has leading or trailing whitespace", abbreviation)
		}
		name := ca[abbreviation]
		if name != strings.TrimSpace(name) {
			log.Printf("WARNING: names.json name %q for abbreviation %q has leading or trailing whitespace", name, abbreviation)
		}
	}
	return ca, err
}
