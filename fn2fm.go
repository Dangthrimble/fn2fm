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
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"fn2fm/internal/filename"
	"fn2fm/internal/names"
	"fn2fm/internal/pdfmeta"
)

// Set by the build workflow; source builds retain the development label.
var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("fn2fm %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

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

	var warnings []string
	ca, warnings, err = names.Load("names.json")
	if err != nil {
		// Match prior behaviour: log the underlying error for missing/unreadable files.
		if !strings.Contains(err.Error(), "invalid names.json:") {
			log.Println(err)
		}
		fmt.Printf("Unable to load names.json: %v\n", err)
		os.Exit(1)
	}
	for _, warning := range warnings {
		log.Println(warning)
	}

	for _, file = range os.Args[1:] {
		fmt.Printf("Parsing %q...\n", file)
		fn, err = filename.ValidateAndExtension(file)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			os.Rename(file, file+"_rename")
			continue
		}

		md, err = filename.ValidateMetadataTags(fn)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			os.Rename(file, file+"_rename")
			continue
		}

		comp, err = filename.ParseComposers(md, ca)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		arr, err = filename.ParseArrangers(md, ca)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		key, err = filename.ParseKey(md)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		acc, err = filename.ParseAccompaniment(md)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("    Composer(s): %q\n", comp)
		fmt.Printf("    Arranger(s): %q\n", arr)
		fmt.Printf("            Key: %q\n", key)
		fmt.Printf("  Accompaniment: %q\n\n", acc)

		err = pdfmeta.Write(file, pdfmeta.Metadata{
			Title: fn, Author: comp, Subject: arr, Keywords: pdfmeta.Keywords(key, acc),
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
