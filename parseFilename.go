package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	validExt   = ".pdf"
	mdPrefix   = "~"
	arrPrefix  = "_"
	keyPrefix  = "["
	keySuffix  = "]"
	withAcc    = "+"
	withoutAcc = "-"
)

const (
	minPrintASCII = '\u0020'        // Minimum printable ASCII value
	maxPrintASCII = '\u007E'        // Maximum printable ASCII value
	charToAvoid   = "\"*./:;<>?\\|" // Characters to avoid according to cloud storage
)

func validateFilenameAndExtension(fnExt string) (string, error) {

	var (
		ok bool
		fn string // Song's filename without the path but including the extension
	)

	// Only the filename is required, so any path prefixing the filename needs
	// to be removed.
	fnExt = filepath.Base(fnExt)

	// If the file doesn't end with validExtension, it is not a suitable file.
	fn, ok = strings.CutSuffix(fnExt, validExt)
	if !ok {
		fmt.Printf("  ERROR: %q is not a PDF file\n", fnExt)
		return "", errors.New(fmt.Sprintf("invalid file type"))
	}

	// If the filename has leading or trailing spaces, it may not be supported
	// by cloud storage.
	if fn != strings.Trim(fn, " ") {
		fmt.Printf("  ERROR: %q has leading or trailing spaces\n", fn)
		return "", errors.New(fmt.Sprintf("invalid filename"))
	}

	// If the filename doesn't have a single instance of metadataPrefix, the
	// metadata cannot be reliably parsed.
	if strings.Count(fn, mdPrefix) != 1 {
		fmt.Printf("  ERROR: %q must have a single %q in the filename to allow metadata to be parsed\n", fnExt, mdPrefix)
		return "", errors.New(fmt.Sprintf("invalid filename"))
	}

	// To avoid any complications with internationalisation, ensure the filename
	// is composed of 7-bit ASCII printable characters, excluding
	// charactersToAvoid.
	for i, w := 0, 0; i < len(fn); i += w {
		rune, width := utf8.DecodeRuneInString(fn[i:])
		if (rune < minPrintASCII) || (rune > maxPrintASCII) {
			fmt.Printf("  ERROR: %q is not a valid character in %q\n", rune, fn)
			return "", errors.New(fmt.Sprintf("invalid filename"))
		}
		if strings.Contains(charToAvoid, string(rune)) {
			fmt.Printf("  ERROR: %q is not a valid character in %q\n", rune, fn)
			return "", errors.New(fmt.Sprintf("invalid filename"))
		}
		w = width
	}

	return fn, nil
}

func validateMetadataTags(fn string) (string, error) {

	var md string // Song's metadata embedded in the filename

	// The metadata needs to be isolated to ensure it contains the correct
	// quantity of each tag.
	splitFn := strings.Split(fn, mdPrefix)
	md = strings.Trim(splitFn[1], " ")

	// The metadata cannot be parsed if it contains the incorrect quantity of
	// each tag.
	if strings.Count(md, arrPrefix) > 1 ||
		strings.Count(md, keyPrefix) != 1 ||
		strings.Count(md, keySuffix) != 1 {
		fmt.Printf("  ERROR: wrong number of metadata tags\n")
		return "", errors.New(fmt.Sprintf("invalid metadata tags"))
	}

	// The tag locations are required to determine whether any
	// arrangers are included and to ensure the tags are in the correct order.
	arrLoc := strings.Index(md, arrPrefix)
	keyLoc := strings.Index(md, keyPrefix)
	accLoc := strings.Index(md, keySuffix)

	// If arrPrefix is after keyPrefix, or keyPrefix is after keySuffix, the
	// metadata cannot be parsed.
	if (arrLoc > keyLoc) || (keyLoc > accLoc) {
		fmt.Printf("  ERROR: metadata tags in wrong order\n")
		return "", errors.New(fmt.Sprintf("invalid metadata tags"))
	}

	return string(md), nil
}

func parseComposers(md string, ca map[string]string) (string, error) {

	var splitMd []string // Metadata split into slices

	// Need to isolate the Composers' metadata to use as a key for composers.
	if strings.Contains(md, arrPrefix) {
		// The Composers' metadata is before the arrangersPrefix.
		splitMd = strings.Split(md, arrPrefix)
	} else {
		// The Composers' metadata is before the keyPrefix.
		splitMd = strings.Split(md, keyPrefix)
	}

	return findComposersOrArrangers(splitMd[0], ca)
}

func parseArrangers(md string, ca map[string]string) (string, error) {

	var (
		splitMd []string // Metadata split into slices
	)
	// There can be no arrangers if the arrangersPrefix is not present.
	if !strings.Contains(md, arrPrefix) {
		return "", nil
	}

	// Need to isolate the Arrangers' metadata to use as a key for arrangers.
	// The Arrangers' metadata is after the arrangersPrefix ...
	splitMd = strings.Split(md, arrPrefix)
	// ... and before the keyPrefix
	splitMd = strings.Split(splitMd[1], keyPrefix)

	return findComposersOrArrangers(splitMd[0], ca)
}

func findComposersOrArrangers(key string, ca map[string]string) (string, error) {

	var (
		ok    bool
		names string // Name(s) of composer(s) or arranger(s)
	)

	key = strings.Trim(key, " ")
	if len(key) == 0 {
		return "", nil
	}

	names, ok = ca[key]
	if !ok {
		fmt.Printf("  ERROR: %q not found\n", key)
		return "", errors.New(fmt.Sprintf("composers or arrangers not found"))
	}

	return strings.TrimSpace(names), nil
}

func parseKey(md string) (string, error) {

	var (
		ok      bool
		splitMd []string // Metadata split into slices
		key     string   // Song's initial key signature
	)

	keys := map[string]string{
		"Cb":  "keysf:-7, keymi:0",
		"Abm": "keysf:-7, keymi:1",
		"Gb":  "keysf:-6, keymi:0",
		"Ebm": "keysf:-6, keymi:1",
		"Db":  "keysf:-5, keymi:0",
		"Bbm": "keysf:-5, keymi:1",
		"Ab":  "keysf:-4, keymi:0",
		"Fm":  "keysf:-4, keymi:1",
		"Eb":  "keysf:-3, keymi:0",
		"Cm":  "keysf:-3, keymi:1",
		"Bb":  "keysf:-2, keymi:0",
		"Gm":  "keysf:-2, keymi:1",
		"F":   "keysf:-1, keymi:0",
		"Dm":  "keysf:-1, keymi:1",
		"C":   "keysf:0, keymi:0",
		"Am":  "keysf:0, keymi:1",
		"G":   "keysf:1, keymi:0",
		"Em":  "keysf:1, keymi:1",
		"D":   "keysf:2, keymi:0",
		"Bm":  "keysf:2, keymi:1",
		"A":   "keysf:3, keymi:0",
		"F#m": "keysf:3, keymi:1",
		"E":   "keysf:4, keymi:0",
		"C#m": "keysf:4, keymi:1",
		"B":   "keysf:5, keymi:0",
		"G#m": "keysf:5, keymi:1",
		"F#":  "keysf:6, keymi:0",
		"D#m": "keysf:6, keymi:1",
		"C#":  "keysf:7, keymi:0",
		"A#m": "keysf:7, keymi:1",
		"7#":  "",
		"6#":  "",
		"5#":  "",
		"4#":  "",
		"3#":  "",
		"2#":  "",
		"1#":  "",
		"0":   "",
		"1b":  "",
		"2b":  "",
		"3b":  "",
		"4b":  "",
		"5b":  "",
		"6b":  "",
		"7b":  "",
	}

	// Need to isolate the Initial Key Signature's metadata to use as the key
	// for keys. The Initial Key Signature's metadata is after the keyPrefix ...
	splitMd = strings.Split(md, keyPrefix)
	// ... and before the keySuffix.
	splitMd = strings.Split(splitMd[1], keySuffix)
	key, ok = keys[strings.Trim(splitMd[0], " ")]
	if !ok {
		fmt.Printf("  ERROR: %q is not a valid key signature\n", key)
		return "", errors.New(fmt.Sprintf("invalid key signature"))
	}

	return key, nil
}

func parseAccompaniment(md string) (string, error) {

	var (
		splitMd []string // Metadata split into slices
		accMd   string   // Metadata indicating whether or not the song has an accompaniment
	)

	// The Accompaniment's metadata is a single character after the keySuffix.
	splitMd = strings.Split(md, keySuffix)
	accMd = strings.Trim(splitMd[1], " ")
	switch accMd {
	case "+":
		return "With Accompaniment", nil
	case "-":
		return "Without Accompaniment", nil
	case "":
		return "", nil
	default:
		fmt.Printf("  ERROR: %q is not valid for accompaniment\n", accMd)
		return "", errors.New(fmt.Sprintf("invalid accompaniment"))
	}
}
