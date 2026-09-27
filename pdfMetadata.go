package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type pdfMetadata struct {
	Title, Author, Subject, Keywords string
}

func (m pdfMetadata) fields() map[string]string {
	return map[string]string{"Title": m.Title, "Author": m.Author, "Subject": m.Subject, "Keywords": m.Keywords}
}

func metadataKeywords(key, accompaniment string) string {
	if key != "" && accompaniment != "" {
		return key + ", " + accompaniment
	}
	return key + accompaniment
}

var disablePDFConfig sync.Once

func pdfConfiguration() *model.Configuration {
	// The standalone writer needs neither a user configuration directory nor fonts.
	disablePDFConfig.Do(api.DisableConfigDir)
	conf := model.NewDefaultConfiguration()
	conf.Offline = true
	conf.ValidationMode = model.ValidationRelaxed
	conf.WriteObjectStream = false
	return conf
}

// writePDFMetadata stages and validates an incremental update before replacing
// the source. The first source is retained as path + "_original", as with ExifTool.
func writePDFMetadata(path string, metadata pdfMetadata) error {
	original, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !original.Mode().IsRegular() {
		return errors.New("only regular PDF files are supported; symbolic links are not followed")
	}
	if original.Mode().Perm()&0222 == 0 {
		return errors.New("PDF is read-only")
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	conf := pdfConfiguration()
	ctx, err := api.ReadContext(bytes.NewReader(source), conf)
	if err != nil {
		if errors.Is(err, pdfcpu.ErrWrongPassword) || errors.Is(err, pdfcpu.ErrOwnerPasswordRequired) {
			return errors.New("encrypted/password-protected PDFs are not supported")
		}
		return fmt.Errorf("read PDF: %w", err)
	}
	if ctx.Encrypt != nil {
		return errors.New("encrypted/password-protected PDFs are not supported")
	}
	if ctx.XRefTable.Version() < model.V14 || ctx.XRefTable.Version() > model.V17 {
		return errors.New("only PDF versions 1.4 through 1.7 are currently supported")
	}
	if ctx.Read.Hybrid || ctx.Read.RepairOffset != 0 {
		return errors.New("hybrid or repaired cross-reference data is not supported")
	}
	// Check objects before validation too: a signature need not advertise SigFlags.
	for _, entry := range ctx.Table {
		if entry == nil || entry.Free {
			continue
		}
		if d, ok := entry.Object.(types.Dict); ok {
			if d["Type"] == types.Name("Sig") || d["Type"] == types.Name("DocTimeStamp") || d["FT"] == types.Name("Sig") {
				return errors.New("signed PDFs and signature fields are not supported")
			}
		}
	}
	if err := api.ValidateContext(ctx); err != nil {
		return fmt.Errorf("validate input PDF: %w", err)
	}
	if ctx.SignatureExist || ctx.AppendOnly || len(ctx.Signatures) > 0 || ctx.RootDict["Perms"] != nil {
		return errors.New("signed PDFs and signature permissions are not supported")
	}
	// Switching cross-reference formats broke Apple readers in the prototype.
	ctx.WriteXRefStream = ctx.Read.UsingXRefStreams
	if ctx.Info == nil {
		entry := model.NewXRefTableEntryGen0(types.NewDict())
		number := ctx.InsertNew(*entry)
		ctx.Info = types.NewIndirectRef(number, 0)
	}
	info, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil {
		return fmt.Errorf("read PDF Info dictionary: %w", err)
	}
	if info == nil {
		return errors.New("PDF Info reference does not resolve to a dictionary")
	}
	for key, value := range metadata.fields() {
		if value == "" {
			delete(info, key)
			continue
		}
		encoded, err := types.EscapedUTF16String(value)
		if err != nil {
			return fmt.Errorf("encode %s: %w", key, err)
		}
		info[key] = types.StringLiteral(*encoded)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".fn2fm-*.pdf")
	if err != nil {
		return fmt.Errorf("create temporary PDF: %w", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if _, err := temp.Write(source); err != nil {
		return fmt.Errorf("copy PDF: %w", err)
	}
	offset := int64(len(source))
	// A final %%EOF comment without a newline must not swallow the new object.
	if len(source) > 0 && source[len(source)-1] != '\n' && source[len(source)-1] != '\r' {
		if _, err := temp.WriteString("\n"); err != nil {
			return err
		}
		offset++
	}
	ctx.Write.Increment = true
	ctx.Write.Offset = offset
	ctx.Write.IncrementWithObjNr(ctx.Info.ObjectNumber.Value())
	if err := api.WriteIncr(ctx, temp, conf); err != nil {
		return fmt.Errorf("write PDF metadata: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("flush temporary PDF: %w", err)
	}
	// Close every handle before replacement, including on Windows.
	if err := temp.Close(); err != nil {
		return err
	}
	if err := validatePDFUpdate(temp.Name(), source, ctx.PageCount, metadata); err != nil {
		return err
	}
	if err := os.Chmod(temp.Name(), original.Mode().Perm()); err != nil {
		return fmt.Errorf("preserve file permissions: %w", err)
	}
	return replacePDFFile(path, temp.Name(), source, original, os.Rename)
}

func validatePDFUpdate(path string, source []byte, pages int, metadata pdfMetadata) error {
	updated, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(updated, source) {
		return errors.New("PDF verification failed: original bytes changed")
	}
	ctx, err := api.ReadAndValidate(bytes.NewReader(updated), pdfConfiguration())
	if err != nil {
		return fmt.Errorf("validate updated PDF: %w", err)
	}
	if ctx.PageCount != pages || ctx.Info == nil {
		return errors.New("PDF verification failed: page count or metadata dictionary changed unexpectedly")
	}
	info, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil {
		return err
	}
	for key, want := range metadata.fields() {
		if want == "" && info[key] == nil {
			continue
		}
		got, err := ctx.DereferenceText(info[key])
		if err != nil || got != want || (want == "" && info[key] != nil) {
			return fmt.Errorf("PDF verification failed for %s", key)
		}
	}
	return nil
}

func unchangedPDF(path string, source []byte, original os.FileInfo) error {
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !current.Mode().IsRegular() || !os.SameFile(current, original) || current.Mode() != original.Mode() {
		return errors.New("source PDF changed during processing; retry after closing other editors")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, source) {
		return errors.New("source PDF changed during processing; retry after closing other editors")
	}
	return nil
}

func replacePDFFile(path, temporary string, source []byte, original os.FileInfo, rename func(string, string) error) error {
	if err := unchangedPDF(path, source, original); err != nil {
		return err
	}
	if err := preservePDFBackup(path+"_original", source, original.Mode().Perm()); err != nil {
		return fmt.Errorf("preserve PDF backup: %w", err)
	}
	if err := unchangedPDF(path, source, original); err != nil {
		return err
	}
	// Never remove the destination first. A failed replacement leaves it in place.
	if err := rename(temporary, path); err != nil {
		return fmt.Errorf("replace PDF (original retained): %w", err)
	}
	return nil
}

func preservePDFBackup(path string, source []byte, mode os.FileMode) (err error) {
	backup, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return statErr
		}
		if !info.Mode().IsRegular() {
			return errors.New("existing backup is not a regular file")
		}
		return nil // Preserve the first backup, even after subsequent updates.
	}
	if err != nil {
		return err
	}
	defer func() {
		backup.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	if _, err = backup.Write(source); err != nil {
		return err
	}
	if err = backup.Chmod(mode); err != nil {
		return err
	}
	if err = backup.Sync(); err != nil {
		return err
	}
	return backup.Close()
}
