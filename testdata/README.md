# PDF fixtures

`no-info-table.pdf` and `no-info-stream.pdf` are generated two-page test documents,
not music scores. Neither contains a PDF Info dictionary. Their pages contain
visible test text so reader/rendering regressions can be checked.

The classic-table fixture is PDF 1.4. The stream fixture was converted to PDF 1.5
with qpdf's `--object-streams=generate --force-version=1.5` options. The generation
script and original verification evidence are preserved in the local development
artifacts at `output/pdf/metadata-no-info/`. The tests read these included fixture
bytes directly; qpdf, ExifTool and macOS frameworks are not required by `go test`.
The repository marks PDFs as binary in `.gitattributes` so Windows checkout does
not alter line endings and invalidate the fixtures' byte offsets.
