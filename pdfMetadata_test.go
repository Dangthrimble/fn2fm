package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Construct independent PDF objects rather than using the metadata writer to
// generate its own input. Metadata includes fields which fn2fm must not change.
func existingInfoPDF(signature bool) []byte {
	content := "BT /F1 12 Tf 72 700 Td (Keep this score page) Tj ET\n"
	xmp := `<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/" dc:format="application/pdf"/></rdf:RDF></x:xmpmeta>`
	form := "<< /Fields [] >>"
	if signature {
		form = "<< /Fields [9 0 R] /SigFlags 3 >>"
	}
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R /Metadata 7 0 R /AcroForm 8 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content),
		"<< /Title (Old title) /Author (Old author) /Subject (Old subject) /Keywords (Old keywords) /Producer (Original producer) /CreationDate (D:20250102123456Z) /ModDate (D:20250103123456Z) /Custom (Keep me) >>",
		fmt.Sprintf("<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream", len(xmp), xmp),
		form,
	}
	if signature {
		// Signature-bearing structure only; the bytes are not a valid certificate.
		objects = append(objects, "<< /FT /Sig /T (Signature1) /V 10 0 R >>",
			"<< /Type /Sig /Filter /Adobe.PPKLite /SubFilter /adbe.pkcs7.detached /ByteRange [0 1 2 3] /Contents <> >>")
	}
	var data bytes.Buffer
	data.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, data.Len())
		fmt.Fprintf(&data, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := data.Len()
	fmt.Fprintf(&data, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&data, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&data, "trailer\n<< /Size %d /Root 1 0 R /Info 6 0 R /ID [<0123456789abcdef> <0123456789abcdef>] >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return data.Bytes()
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func readTestPDF(t *testing.T, data []byte) *model.Context {
	t.Helper()
	ctx, err := api.ReadAndValidate(bytes.NewReader(data), pdfConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func TestWritePDFMetadata(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"existing Info", existingInfoPDF(false)},
		{"no Info classic table", readTestFile(t, "testdata/no-info-table.pdf")},
		{"no Info xref stream", readTestFile(t, "testdata/no-info-stream.pdf")},
		{"no final newline", bytes.TrimRight(existingInfoPDF(false), "\r\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Score.pdf")
			writeTestFile(t, path, tc.data)
			before := readTestPDF(t, tc.data)
			metadata := pdfMetadata{Title: "New score (test)", Author: "André J Thomas", Subject: "Dan Forrest", Keywords: "keysf:-2, keymi:0, With Accompaniment"}
			if err := writePDFMetadata(path, metadata); err != nil {
				t.Fatal(err)
			}
			updated := readTestFile(t, path)
			if !bytes.HasPrefix(updated, tc.data) {
				t.Fatal("original bytes changed")
			}
			if !bytes.Equal(readTestFile(t, path+"_original"), tc.data) {
				t.Fatal("backup differs from original")
			}
			after := readTestPDF(t, updated)
			if after.PageCount != before.PageCount || after.XRefTable.Version() != before.XRefTable.Version() || after.Read.UsingXRefStreams != before.Read.UsingXRefStreams {
				t.Fatal("page count, version or cross-reference format changed")
			}
			info, err := after.DereferenceDict(*after.Info)
			if err != nil {
				t.Fatal(err)
			}
			for key, want := range metadata.fields() {
				got, err := after.DereferenceText(info[key])
				if err != nil || got != want {
					t.Errorf("%s = %q, %v; want %q", key, got, err, want)
				}
			}
			if before.Info != nil {
				oldInfo, _ := before.DereferenceDict(*before.Info)
				for _, key := range []string{"Producer", "CreationDate", "ModDate", "Custom"} {
					if info[key] != oldInfo[key] {
						t.Errorf("unrelated Info field %s changed", key)
					}
				}
				for _, key := range []string{"Metadata", "AcroForm"} {
					if before.RootDict[key] != after.RootDict[key] {
						t.Errorf("catalog reference %s changed", key)
					}
				}
			}
			// A later edit clears absent fields but retains the first backup.
			if err := writePDFMetadata(path, pdfMetadata{Title: "Title only"}); err != nil {
				t.Fatal(err)
			}
			second := readTestPDF(t, readTestFile(t, path))
			secondInfo, _ := second.DereferenceDict(*second.Info)
			for _, key := range []string{"Author", "Subject", "Keywords"} {
				if _, exists := secondInfo[key]; exists {
					t.Errorf("empty %s was not removed", key)
				}
			}
			if !bytes.Equal(readTestFile(t, path+"_original"), tc.data) {
				t.Fatal("first backup overwritten on repeated edit")
			}
			assertNoTemporaryPDFs(t, filepath.Dir(path))
		})
	}
}

func TestUnsupportedPDFsRemainUnchanged(t *testing.T) {
	base := existingInfoPDF(false)
	cases := []struct {
		name, message string
		data          []byte
	}{
		{"malformed", "read PDF", []byte("This is not a PDF")},
		{"old PDF", "versions 1.4 through 1.7", bytes.Replace(base, []byte("%PDF-1.4"), []byte("%PDF-1.3"), 1)},
		{"PDF 2", "versions 1.4 through 1.7", bytes.Replace(base, []byte("%PDF-1.4"), []byte("%PDF-2.0"), 1)},
		{"signature field", "signature", existingInfoPDF(true)},
	}
	for _, password := range []string{"", "test-password"} {
		conf := pdfConfiguration()
		conf.UserPW = password
		conf.OwnerPW = "test-owner"
		var encrypted bytes.Buffer
		if err := api.Encrypt(bytes.NewReader(base), &encrypted, conf); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, struct {
			name, message string
			data          []byte
		}{"encrypted " + password, "encrypted/password-protected", encrypted.Bytes()})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "Score.pdf")
			writeTestFile(t, path, tc.data)
			err := writePDFMetadata(path, pdfMetadata{Title: "New title"})
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("got %v; expected %q", err, tc.message)
			}
			if !bytes.Equal(readTestFile(t, path), tc.data) {
				t.Fatal("rejected input changed")
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatal("rejected input left backup or temporary file")
			}
		})
	}
}

func TestPDFFileReplacementFailures(t *testing.T) {
	for _, fail := range []string{"source changed", "backup blocked", "rename failed"} {
		t.Run(fail, func(t *testing.T) {
			dir := t.TempDir()
			path, temporary := filepath.Join(dir, "Score.pdf"), filepath.Join(dir, "temporary.pdf")
			source := existingInfoPDF(false)
			writeTestFile(t, path, source)
			writeTestFile(t, temporary, []byte("staged replacement"))
			original, _ := os.Stat(path)
			want := source
			switch fail {
			case "source changed":
				want = []byte("another editor changed this file")
				writeTestFile(t, path, want)
			case "backup blocked":
				if err := os.Mkdir(path+"_original", 0700); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			err := replacePDFFile(path, temporary, source, original, func(string, string) error {
				called = true
				return errors.New("injected rename failure")
			})
			if err == nil || called != (fail == "rename failed") {
				t.Fatalf("unexpected replacement result: %v; rename called=%t", err, called)
			}
			if !bytes.Equal(readTestFile(t, path), want) {
				t.Fatal("failed replacement altered source")
			}
		})
	}
}

func TestPDFBackupFailureCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Score.pdf")
	source := existingInfoPDF(false)
	writeTestFile(t, path, source)
	if err := os.Mkdir(path+"_original", 0700); err != nil {
		t.Fatal(err)
	}
	if err := writePDFMetadata(path, pdfMetadata{Title: "New title"}); err == nil {
		t.Fatal("expected backup failure")
	}
	if !bytes.Equal(readTestFile(t, path), source) {
		t.Fatal("backup failure changed input")
	}
	assertNoTemporaryPDFs(t, dir)
}

func assertNoTemporaryPDFs(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".fn2fm-*.pdf"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files left: %v, %v", matches, err)
	}
}

func TestMetadataKeywords(t *testing.T) {
	for _, tc := range []struct{ key, acc, want string }{
		{"keysf:0, keymi:0", "With Accompaniment", "keysf:0, keymi:0, With Accompaniment"},
		{"keysf:0, keymi:0", "", "keysf:0, keymi:0"},
		{"", "Without Accompaniment", "Without Accompaniment"},
		{"", "", ""},
	} {
		if got := metadataKeywords(tc.key, tc.acc); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
