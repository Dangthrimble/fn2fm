package main

import (
	"testing"
)

func TestValidateFilenameAndExtension(t *testing.T) {

	t.Run("valid filename with path", func(t *testing.T) {

		got, err := validateFilenameAndExtension("! Cambrensis Choir Sheet Music/O Magnum Mysterium; MoLa[2#].pdf")
		want := "O Magnum Mysterium; MoLa[2#]"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid filename without path", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium; MoLa[2#].pdf")
		want := "O Magnum Mysterium; MoLa[2#]"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("wrong filename extension", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium; MoLa[2#].docx")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains leading or trailing spaces", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium; MoLa[2#] .pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains wrong number of metadataPrefix", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium.pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains characters other than 7-bit ASCII printable characters", func(t *testing.T) {

		got, err := validateFilenameAndExtension("Ö Magnum Mysterium; MoLa[2#].pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains 7-bit ASCII printable characters to be avoided", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium; MoLa<2#>.pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestValidateMetadataTags(t *testing.T) {
	t.Run("valid metadata tags with arrangers", func(t *testing.T) {

		got, err := validateMetadataTags("A Concert Celebration; AnWe_MaBr[1b]+")
		want := "AnWe_MaBr[1b]+"

		if (got != want) || (err != nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid metadata tags without arrangers", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis); KaJe[C]+")
		want := "KaJe[C]+"

		if (got != want) || (err != nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("incorrect number of tags", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis); KaJe[C+")
		want := ""

		if (got != want) || (err == nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("tags in wrong order", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis); KaJe]C[+")
		want := ""

		if (got != want) || (err == nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestParseComposers(t *testing.T) {
	composersArrangers := map[string]string{
		"AnWe":   "Andrew Lloyd Webber",
		"CtEcMr": "Chris Tomlin, Ed Cash, Matt Redman",
	}

	t.Run("valid composers with arrangers", func(t *testing.T) {

		got, err := parseComposers("AnWe_MaBr[1b]+", composersArrangers)
		want := "Andrew Lloyd Webber"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid composers without arrangers", func(t *testing.T) {

		got, err := parseComposers("CtEcMr [1b]+", composersArrangers)
		want := "Chris Tomlin, Ed Cash, Matt Redman"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("no composers", func(t *testing.T) {

		got, err := parseComposers("_CtExMr [1b]+", composersArrangers)
		want := ""

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("unknown composers without arrangers", func(t *testing.T) {

		got, err := parseComposers("CtExMr [1b]+", composersArrangers)
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestParseArrangers(t *testing.T) {
	composersArrangers := map[string]string{
		"MaBr": "Mark Brymer",
	}

	t.Run("valid arrangers", func(t *testing.T) {

		got, err := parseArrangers("AnWe_ MaBr [1b]+", composersArrangers)
		want := "Mark Brymer"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("no arrangers", func(t *testing.T) {

		got, err := parseArrangers("AnWe [1b]+", composersArrangers)
		want := ""

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("unknown arrangers", func(t *testing.T) {

		got, err := parseArrangers("AnWe_ MaBx [1b]+", composersArrangers)
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestParseKey(t *testing.T) {

	t.Run("valid key signature", func(t *testing.T) {

		got, err := parseKey("CrCo[Dm]+")
		want := "keysf:-1, keymi:1"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid number of accidentals", func(t *testing.T) {

		got, err := parseKey("CtExMr [1b]+")
		want := ""

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("invalid key signature", func(t *testing.T) {

		got, err := parseKey("CtExMr [1x]+")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestParseAccompaniment(t *testing.T) {

	t.Run("with accompaniment", func(t *testing.T) {

		got, err := parseAccompaniment("CrCo[Dm]+")
		want := "With Accompaniment"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("without accompaniment", func(t *testing.T) {

		got, err := parseAccompaniment("CrCo[Dm]-")
		want := "Without Accompaniment"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("no accompaniment", func(t *testing.T) {

		got, err := parseAccompaniment("CtExMr [1b]")
		want := ""

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("invalid accompaniment", func(t *testing.T) {

		got, err := parseAccompaniment("CtExMr [1b]=")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}
