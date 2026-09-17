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
	"os/exec"
)

func main() {

	var (
		err  error
		ca   map[string]string // Map of composers and arrangers with abbreviation as the key
		file string            // Song's file
		fn   string            // Song's filename without the path but including the extension
		md   string            // Song's metadata embedded in the filename
		comp string            // Song's composer(s)
		arr  string            // Song's arranger(s)
		key  string            // Song's initial key signature
		acc  string            // Whether or not the song has an accompaniment
	)

	ca, err = readComposersArrangers()
	if err != nil {
		fmt.Printf("Unable to open file of Composers and Arrangers\n")
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

		title := "-PDF:Title=" + fn
		auth := "-PDF:Author=" + comp
		subj := "-PDF:Subject=" + arr
		kw := "-PDF:Keywords="
		if len(key) > 0 {
			if len(key) > 0 {
				kw = kw + key + ", " + acc
			} else {
				kw = kw + key
			}
		} else {
			kw = kw + acc
		}
		cmd := exec.Command("exiftool", file, title, auth, subj, kw)
		err = cmd.Run()
		if err != nil {
			log.Print(err)
			os.Rename(file, file+"_error")
			continue
		}
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
	}
	json.Unmarshal([]byte(f), &ca)
	return ca, err
}
