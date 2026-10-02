package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"fn2fm/internal/pdfmeta"
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
	ctx, err := api.ReadAndValidate(bytes.NewReader(data), pdfmeta.Configuration())
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func assertNoTemporaryPDFs(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".fn2fm-*.pdf"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files left: %v, %v", matches, err)
	}
}
