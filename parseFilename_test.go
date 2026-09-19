package main

import (
	"testing"
)

func TestValidateFilenameAndExtension(t *testing.T) {

	t.Run("valid filename with path", func(t *testing.T) {

		got, err := validateFilenameAndExtension("! Cambrensis Choir Sheet Music/O Magnum Mysterium ~ MoLa[2#].pdf")
		want := "O Magnum Mysterium ~ MoLa[2#]"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid filename without path", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium ~ MoLa[2#].pdf")
		want := "O Magnum Mysterium ~ MoLa[2#]"

		if got != want || err != nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("wrong filename extension", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium ~ MoLa[2#].docx")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains leading or trailing spaces", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium ~ MoLa[2#] .pdf")
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

		got, err := validateFilenameAndExtension("Ö Magnum Mysterium ~ MoLa[2#].pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("filename contains 7-bit ASCII printable characters to be avoided", func(t *testing.T) {

		got, err := validateFilenameAndExtension("O Magnum Mysterium ~ MoLa<2#>.pdf")
		want := ""

		if got != want || err == nil {
			t.Errorf("got %q, want %q given", got, want)
		}
	})
}

func TestValidateMetadataTags(t *testing.T) {
	t.Run("valid metadata tags with arrangers", func(t *testing.T) {

		got, err := validateMetadataTags("A Concert Celebration ~ AnWe_MaBr[1b]+")
		want := "AnWe_MaBr[1b]+"

		if (got != want) || (err != nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("valid metadata tags without arrangers", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis) ~ KaJe[C]+")
		want := "KaJe[C]+"

		if (got != want) || (err != nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("incorrect number of tags", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis) ~ KaJe[C+")
		want := ""

		if (got != want) || (err == nil) {
			t.Errorf("got %q, want %q given", got, want)
		}
	})

	t.Run("tags in wrong order", func(t *testing.T) {

		got, err := validateMetadataTags("Sing with Joy at Christmas (Stella Natalis) ~ KaJe]C[+")
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

	for _, metadata := range []string{"CrCo[]+", "CrCo[   ]+"} {
		t.Run("empty key "+metadata, func(t *testing.T) {
			got, err := parseKey(metadata)
			if got != "" || err == nil {
				t.Fatalf("got (%q, %v), want empty-key rejection", got, err)
			}
		})
	}

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

func TestTildeOnlyFilenames(t *testing.T) {
	for _, filename := range []string{
		"Score; JoRu[C]+.pdf",
		"Score; Part Two ~ JoRu[C]+.pdf",
		"Score ~ JoRu;[C]+.pdf",
		"Score ~ Part Two ~ JoRu[C]+.pdf",
	} {
		t.Run(filename, func(t *testing.T) {
			if got, err := validateFilenameAndExtension(filename); err == nil || got != "" {
				t.Fatalf("got (%q, %v), want rejection", got, err)
			}
		})
	}
}

func TestRealFilenames(t *testing.T) {
	// Fixture values from the preserved working dictionary, including DaFo's leading space.
	names := map[string]string{
		"JoGo": "John Goss", "DWil": "David Willcocks", "HoHe": "Howard Helvey",
		"DaFo": " Dan Forrest", "MHau": "Marty Haugen", "GFHa": "George Frideric Handel",
		"KaJe": "Karl Jenkins", "IaFa": "Iain Farrington",
	}
	cases := []struct{ filename, composer, arranger, key string }{
		{"See Amid The Winter Snow ~ JoGo_DWil[G]+.pdf", "John Goss", "David Willcocks", "keysf:1, keymi:0"},
		{"Lo How a Rose E'er Blooming ~ _HoHe[F]+.pdf", "", "Howard Helvey", "keysf:-1, keymi:0"},
		{"Angels We Have Heard On High ~ _DaFo [G]+.pdf", "", " Dan Forrest", "keysf:1, keymi:0"},
		{"All Are Welcome ~ MHau[F]+.pdf", "Marty Haugen", "", "keysf:-1, keymi:0"},
		{"Let Thy Hand #2 Let Justice And Judgement ~ GFHa[Em]+.pdf", "George Frideric Handel", "", "keysf:1, keymi:1"},
		{"Celebro (Stella Natalis) ~ KaJe[4b]+.pdf", "Karl Jenkins", "", ""},
		{"Nova, Nova ~ IaFa[1#]+.pdf", "Iain Farrington", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.filename, func(t *testing.T) {
			fn, err := validateFilenameAndExtension(tc.filename)
			if err != nil {
				t.Fatal(err)
			}
			if fn != tc.filename[:len(tc.filename)-4] {
				t.Fatalf("filename altered: %q", fn)
			}
			md, err := validateMetadataTags(fn)
			if err != nil {
				t.Fatal(err)
			}
			comp, err := parseComposers(md, names)
			if err != nil {
				t.Fatal(err)
			}
			arr, err := parseArrangers(md, names)
			if err != nil {
				t.Fatal(err)
			}
			key, err := parseKey(md)
			if err != nil {
				t.Fatal(err)
			}
			acc, err := parseAccompaniment(md)
			if err != nil {
				t.Fatal(err)
			}
			if comp != tc.composer || arr != tc.arranger || key != tc.key || acc != "With Accompaniment" {
				t.Fatalf("unexpected metadata: composer=%q arranger=%q key=%q accompaniment=%q", comp, arr, key, acc)
			}
			t.Logf("composer=%q arranger=%q key=%q accompaniment=%q", comp, arr, key, acc)
		})
	}
}
