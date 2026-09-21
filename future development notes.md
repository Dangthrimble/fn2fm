# Future development notes

This file records project goals, confirmed filename requirements, unfinished
investigations and future work. Backlog order does not imply an agreed priority;
suggested approaches below are not settled implementation decisions.

## Goals and working approach

- Improve the existing macOS Go tool using Git and Cursor, and make executables
  available online for macOS and Windows while keeping the workflow free.
- Work in small, understandable changes. When asking the maintainer to perform
  steps, give one step at a time and wait for the result.
- Preserve existing work and the working executable while investigating the
  difference between deployed behaviour and the checked-in source.

## Confirmed filename requirements and source mismatch

The maintainer's working executable uses a tilde (`~`) between score name and
metadata. The maintainer reports its filesystem date as 1 February 2024 at 03:28.
The earliest available Git commit, `fdba9bc` (23 June 2024), already uses a
semicolon (`;`), as does the current `parseFilename.go` and README. Available
Git history does not show a tilde-to-semicolon change or establish why these
versions differ. A commit date is not the original creation date of its source,
and a filesystem timestamp alone does not establish a binary's source revision.
Do not interpret this as an intentional decision to abandon the tilde convention.

### Tilde-only reconciliation (18 September 2026)

- The maintainer explicitly chose tilde-only behaviour. The Git parser now requires
  exactly one `~` and forbids semicolons anywhere in the filename; semicolon
  backward compatibility is not supported. Earlier descriptions of the Git parser
  using a semicolon below record the pre-reconciliation state.
- Updated the existing tests and README separator examples. Added regression tests
  for semicolon rejection, repeated tildes and all seven real filenames, using a
  small fixture from the working dictionary. The module structure and earlier
  refactoring are retained. Broader named-key behaviour is unchanged. Empty keys remain rejected, as
  explicitly confirmed by the maintainer below.
- The maintainer identified the working dictionary as
  `/Users/jonathan/Documents/Choir/Cambrensis/New Songs/names.json`.
  Its 73 entries include all entries in the two source dictionaries with identical
  values and cover all seven real filenames. `DaFo` maps to `" Dan Forrest"`,
  including the leading space in the preserved snapshot. On 19 September 2026,
  the maintainer reported removing it from the live working dictionary.
- A baseline snapshot of the separate tilde source tree, working executable and
  working dictionary is preserved at `baseline-snapshots/20260918T192323Z/`.
  All 14 copied files were verified against their originals using SHA-256;
  `manifest.json` records provenance and `SHA256SUMS` records checksums.
- After applying the reconciliation, all 36 parser cases passed in the Git
  repository, including four separator-rejection cases. `go vet ./...` and
  `git diff --check` also passed. The working executable was not rebuilt.
- Before editing the Git source, disposable copies of both source variants passed
  their 25 existing cases and seven real filename cases after adapting the Git
  copy to tilde-only behaviour. These are parser checks, not proof of executable
  equivalence or PDF/forScore behaviour. No PDFs were processed.

### Executable source located (18 September 2026)

Inspection of `/Users/jonathan/go/bin/fn2fm` found Go 1.21.5 build information
for macOS amd64, module `fn2fm`, with an `internal/scoreMetadata` dependency.
The executable's debug information records its main source file as
`/Users/jonathan/go/src/github.com/Dangthrimble/fn2fm/cmd/fn2fm.go`.
That source tree still exists, and its
`internal/scoreMetadata/scoreMetadata.go` explicitly sets `mdPrefix = "~"`.
Its accompanying tests also use tilde filenames. The directory is not currently
a Git repository. No VCS revision was reported in the executable build information.

This identifies a separate source location consistent with the working executable;
it does not yet prove the files there are unchanged since that build. Compare and
preserve this tree before reconciling it with the semicolon-based repository in
`/Users/jonathan/Development/Go/fn2fm`. Do not overwrite either source tree or the
working binary as part of the investigation.

Requirements supplied by the maintainer:

- Use `~` as the separator; usual spacing is one space on each side.
- The score name may include a suite and section prefix, for example
  `Let Thy Hand #1 Let Thy Hand Be Strengthened`.
- Composer and arranger fields use abbreviations from `names.json`.
- `_` is required whenever an arranger is present, including when the composer
  is absent. Omit `_` when no arranger is present.
- The initial key signature inside brackets can be a named key matching
  `[A-G][#b]?m?`, or an accidental count matching `0|[1-7][#b]`.
  Numeric signatures are valid input, not malformed named keys. The checked-in
  implementation accepts these counts but emits no forScore key metadata for
  them; do not infer major/minor from a count alone.
- Accompaniment uses `+` or `-`. The older README and current parser also allow
  it to be absent; confirm intended compatibility before changing this.
- Spaces around metadata components are allowed for readability, but leading or
  trailing spaces in the filename are not. Preserve the stated restricted ASCII
  filename policy; reconcile its exact punctuation set with tests before edits.

Representative filenames from the maintainer's successfully processed collection
(reported 18 September 2026):

```text
See Amid The Winter Snow ~ JoGo_DWil[G]+.pdf
Lo How a Rose E'er Blooming ~ _HoHe[F]+.pdf
Angels We Have Heard On High ~ _DaFo [G]+.pdf
All Are Welcome ~ MHau[F]+.pdf
Let Thy Hand #2 Let Justice And Judgement ~ GFHa[Em]+.pdf
Celebro (Stella Natalis) ~ KaJe[4b]+.pdf
Nova, Nova ~ IaFa[1#]+.pdf
```

Use representative real filenames as regression cases, with suitable name-map
fixtures, when reconciling the parser, tests and README. Locate and preserve the
working executable and its matching name dictionary; if possible, identify its
source/build provenance. Use disposable PDF copies for behaviour comparisons.
The maintainer explicitly chose to reject empty brackets `[]`, including brackets
containing only spaces. Specify the actual key where known; otherwise use an
accidental count (`0` or `1`–`7` followed by `#` or `b`) as a fallback. Counts
produce no forScore key metadata but retain distinguishing information in the
filename. They do not distinguish all possible keys (for example, relative major
and minor keys share an accidental count). The README now reflects this decision;
the parser already rejected empty keys, and regression tests cover this behaviour.
Semicolon backward compatibility was also explicitly rejected (see above).

## Dictionary requirements (19 September 2026)

- The repository dictionary is an exemplar/starter, not a copy of the maintainer's
  full working dictionary. Do not replace it with the 73-entry personal list.
- Loading `names.json` from the current working directory is intentional: it lets
  users keep and edit their dictionary alongside the PDFs they are working on.
  No executable-directory lookup is needed for this workflow.
- Added warnings on load for leading or trailing whitespace in abbreviations or
  expanded names. Processing continues, with leading and trailing whitespace
  stripped from names when producing composer/arranger metadata. Internal spaces
  are preserved. Abbreviations and dictionary files are not automatically changed.
  The baseline fixture retains its historical leading space; the regression now
  expects the trimmed metadata name while preserving the dictionary value.

## ExifTool replacement investigation

### Disposable-file comparison (19 September 2026)

- Compared the two `.pdf` files supplied in the maintainer's `New Songs` folder:
  `O Holy Night ~ AdAd_DaFo[Bb]+.pdf` (16 pages) and
  `O Nata Lux ~ DoBy[D]+.pdf` (8 pages), using its current `names.json`.
  Existing `.pdf_original` backups were not used or modified.
- Ran the preserved working executable with ExifTool 13.59 on separate copies.
  An isolated Go prototype used pdfcpu v0.15.0 and a copy of the current parser,
  updating the PDF Info dictionary and deleting Subject when no arranger exists.
  It did not invoke ExifTool to write its PDFs. The application has not adopted
  pdfcpu, and its dependencies and working executable were not changed.
- Both outputs had matching Title, Author, Subject and Keywords for both scores.
  Complete extracted XMP streams were byte-identical to the inputs in both paths.
  O Nata Lux retains its existing PDF/XMP title and author disagreement; neither
  writer resolved it. The later manual fetching result is recorded below.
- All 24 pages rendered to identical PNG bytes across inputs and both outputs
  using Poppler at 72 dpi. Representative output score pages were visually checked.
- Unlike ExifTool, pdfcpu changed PDF version 1.6 to 1.7, replaced Producer and
  CreationDate/ModDate, and removed the empty AcroForm dictionaries. Both original
  form dictionaries had empty Fields arrays. These are observed differences,
  not an approved final metadata-preservation policy.
- SHA-256 verification confirmed the two original PDFs and working dictionary
  remained unchanged. At this stage no forScore import or Windows runtime test
  had been performed; the later manual result is recorded below.
- Outputs for manual import are in `output/pdf/metadata-comparison/`, separated
  into `exiftool/` and `pdfcpu/` folders with identical filenames. The prototype,
  full metadata comparison, input hashes and page renders are temporarily in
  `/private/tmp/fn2fm-pdf-comparison/`. pdfcpu required a newer toolchain; Go 1.26.8
  was downloaded into this isolated test area, without changing project go.mod.
- The subsequent manual comparison matched for both scores (see below).
  This two-file test does not establish general PDF compatibility.

### Manual forScore result (20 September 2026)

The maintainer reported identical fetched properties for the legacy executable
and pdfcpu outputs of `O Holy Night ~ AdAd_DaFo[Bb]+.pdf`:

- Title: `O Holy Night ~ AdAd_DaFo[Bb]+`
- Composers: Adolphe Adam
- Arrangers: Dan Forrest
- Tags: With Accompaniment
- Key: B♭

The maintainer also reported identical fetched properties for both outputs of
`O Nata Lux ~ DoBy[D]+.pdf`:

- Title: `O Nata Lux ~ DoBy[D]+`
- Composers: Douglas Byler
- Arrangers: empty
- Tags: With Accompaniment
- Key: D

Both scores therefore produced matching visible properties in the maintainer's
forScore installation. For O Nata Lux, fetching yielded the intended PDF title
and composer despite the conflicting XMP values. These are maintainer-reported
results, not direct automated observations of forScore. They do not establish
general PDF compatibility, shared-setlist compatibility or Windows behaviour.

The full-rewrite prototype's changes to dates, Producer, PDF version and empty
form dictionaries prompted the separate preservation experiment below.

### Metadata-preserving incremental write (20 September 2026)

- Inspected unmodified pdfcpu v0.15.0: the full writer updates dates/Producer and
  emits a new PDF header, whereas its exported incremental writer can emit only
  selected objects. No simple preservation setting was found for the full writer.
- A separate prototype uses `api.ReadAndValidate`, updates only the existing Info
  dictionary, marks that object for incremental writing, and calls `api.WriteIncr`
  on a new disposable copy. It does not modify pdfcpu or invoke ExifTool to write.
- Both revised PDFs preserved every original input byte and appended only an Info
  update plus cross-reference/trailer data (597 bytes for O Holy Night, 534 for
  O Nata Lux). Original creation/modification dates, Producer, PDF version 1.6,
  document IDs, XMP and empty form structures were preserved in these cases.
- Automated comparison found no changes in extracted unrelated metadata. The four
  target fields matched the prior ExifTool outputs; the only visible metadata
  difference from the original inputs was removal of the leading space in
  O Holy Night's arranger. O Nata Lux already contained the intended values.
- Both outputs passed pdfcpu validation. All 24 pages rendered identically to
  the input PNGs at 72 dpi; representative output score pages were visually
  inspected. SHA-256 checks confirmed the original PDFs and dictionary unchanged.
- Revised PDFs, prototype source, pinned module files and comparison evidence are
  preserved in `output/pdf/metadata-preservation/`. They are separate from the
  first outputs in `output/pdf/metadata-comparison/`. Temporary inputs, dependencies
  and renders are in `/private/tmp/fn2fm-preservation/` and may be cleared.
- This remains an experiment for the two unencrypted PDFs with existing Info
  dictionaries. Other versions, missing Info dictionaries, signed/encrypted files
  and repeated edits require further handling/testing before integration.
  Incremental updates retain earlier metadata in prior revisions of the file.
  The app, its go.mod and working executable remain unchanged.
- Subsequent manual tests FAILED: both files appeared as single blank scores in
  forScore, with no metadata fetched. PDFgear could not open them. Acrobat opened
  them with missing/different properties and requested saving on close. These are
  maintainer-reported observations. The successful automated checks above were
  insufficient; the first incremental outputs must not be treated as compatible.

### Incremental reader-compatibility correction (20 September 2026)

- Confirmed the failed outputs still match their recorded hashes. qpdf --check
  accepted both, so adding it alone would not have caught this failure.
- Reproduced the problem using Apple's PDFKit from a local command-line check:
  neither failed file opened. CoreGraphics reported failure to find the start of
  the cross-reference section. No direct control of forScore was involved.
- The prototype forced `WriteXRefStream = false` despite these inputs using
  cross-reference streams. Changing only this setting to `true` produced files
  PDFKit could open correctly, with 16 and 8 pages and all document attributes
  matching the ExifTool outputs. This isolates the cross-reference output choice
  as the cause for these files; it is not a general compatibility guarantee.
- All 24 revised pages rendered identically to the ExifTool copies using Apple
  CoreGraphics at 72 dpi, and identically to the original inputs using Poppler.
  The preservation checks still passed: target metadata matched, XMP and dates/
  Producer/PDF version were preserved, and original input bytes were unchanged.
  The new appended updates are 684 and 619 bytes respectively.
- Revised files, prototype and reader-check evidence are preserved separately in
  `output/pdf/metadata-preservation-v2/`. The failed files remain in
  `output/pdf/metadata-preservation/`, whose README now marks the failure.
  Application source, dependencies, working executable and originals are unchanged.
- The maintainer subsequently confirmed all five fetched fields match the legacy
  results for both v2 files in forScore: O Holy Night has Adolphe Adam, Dan Forrest,
  With Accompaniment and B♭; O Nata Lux has Douglas Byler, an empty arranger,
  With Accompaniment and D. Both titles include the full filename metadata as
  expected. These are maintainer-reported forScore results. Page display/counts
  were not explicitly reported in this follow-up.
- The maintainer subsequently confirmed both v2 files open and display all pages
  in PDFgear and Adobe Acrobat. PDFgear properties were reported correct.
  Acrobat screenshots show a remaining metadata discrepancy: O Holy Night has
  the expected Title and Subject but blank Author and Keywords; O Nata Lux shows
  Title `BP133_O Nata Lux 1.6DF`, Author `Dan Forrest`, and blank Subject/Keywords.
  Its empty Subject is expected because there is no arranger.
- Read-only checks confirmed both v2 files still match the generated hashes.
  The four target PDF Info fields match the legacy ExifTool output, and the
  complete XMP bytes are identical between legacy and v2 for each score.
  O Nata Lux's Acrobat title/author exactly match its retained XMP title/creator.
  O Holy Night's XMP contains none of the corresponding title/creator/description/
  subject/Keywords entries. Both PDFs still contain the intended Info fields.
- The maintainer then supplied Acrobat screenshots of their legacy outputs, which
  show the same displayed properties as v2 for both scores, and reported correct
  PDFgear properties for the legacy files. The legacy screenshots show files in
  the maintainer's Box library; their exact bytes were not independently verified.
  These observations establish that the displayed Acrobat discrepancies also
  occur in the legacy workflow for these scores, rather than being newly introduced
  by v2. They do not establish universal reader compatibility or prove Acrobat's
  complete Info/XMP reconciliation algorithm.
- The maintainer also supplied Acrobat screenshots from before any fn2fm use.
  O Holy Night displayed blank Title, Author, Subject and Keywords; after either
  legacy processing or v2, Title and Subject are populated while Author/Keywords
  remain blank. O Nata Lux already displayed `BP133_O Nata Lux 1.6DF` and
  `Dan Forrest`, with blank Subject/Keywords; these displayed values remain the
  same after either writer. Both original screenshots show dates, creator and
  producer, so blank descriptive fields must not be equated with an absent Info
  dictionary. These are screenshot observations, not new binary-level checks of
  the original Box-library files.
- Keep Acrobat/Info/XMP consistency as a separate development issue. Synchronising
  the intended fields in both metadata stores would be a behaviour change from
  the legacy writer; it has not been implemented or approved as the replacement's
  initial behaviour.
- The subsequent missing-Info experiment is recorded below.

### Missing Info dictionary support (20 September 2026)

- Extended only the isolated prototype to create an Info dictionary when absent,
  allocating a new object rather than reusing an earlier revision's free object.
  It now selects the input's cross-reference format for the incremental update.
- Created three disposable two-page fixtures and verified they had no `/Info`
  reference: a PDF 1.4 classic-table file with all four target fields, a PDF 1.5
  cross-reference-stream file with no composer and a numeric key, and a PDF 1.4
  title-only case. Empty target fields remain absent. This tests genuinely missing
  dictionaries, rather than assuming blank reader properties mean no dictionary.
- Also processed copies of the two earlier ExifTool outputs as regression cases
  for existing dictionaries. All five passed pdfcpu validation, qpdf checks and
  Apple PDFKit opening, page-count, text-length and target-metadata checks.
  All 30 pages rendered identically before/after with CoreGraphics at 72 dpi.
  Both pages of the classic-table fixture were visually checked using Poppler.
- Every original input byte was preserved, unrelated extracted PDF/XMP metadata
  was unchanged, and input/dictionary hashes remained unchanged. No dates or
  Producer were added to the new dictionaries. Prototype `go vet ./...` passed.
- Inputs, outputs, prototype, fixture generator and verification evidence are
  preserved at `output/pdf/metadata-no-info/`. Original music files, application
  source, project dependencies and the working executable remain unchanged.
- These are automated local macOS results, not new manual forScore, PDFgear,
  Acrobat or Windows results. Signed/encrypted files, PDFs older than 1.4 and
  production file-replacement handling remain outside this experiment.
- The maintainer approved integration; its implementation is recorded below.

### Standalone writer integration (20 September 2026)

- Replaced the development app's ExifTool invocation with the incremental writer
  in `pdfMetadata.go`, pinned to pdfcpu v0.15.0. The minimum Go version is now
  1.25. The built executable needs no ExifTool or Go runtime installation;
  `names.json` remains an editable file loaded from the working directory.
- Writes use a temporary file in the source folder. Before replacement, the app
  checks the original-byte prefix, PDF validation, page count and all four target
  fields, closes the temporary file and preserves basic permissions. It keeps
  the first `.pdf_original` backup without overwriting existing regular backups.
  It checks for source changes before replacement and never deletes the source
  as a preliminary step. Temporary files are removed on ordinary failures.
- Failed PDF writes now report an error, retain the source filename/content and
  produce a nonzero command exit status, instead of renaming to `_error`.
  Existing filename/tag `_rename` behaviour remains unchanged. No claim is made
  of protection against concurrent editors or power loss at every possible point;
  other editors should be closed. Filesystem timestamps, ACLs and extended
  attributes are not copied; PDF-internal dates and unrelated metadata are kept.
- The initial writer rejects encrypted/password-protected PDFs, signature fields
  and signatures, PDF versions outside 1.4–1.7, hybrid/repaired cross-reference
  structures, symbolic links and read-only PDFs. These limitations are explicit
  guardrails, not newly implemented support for those formats. XMP remains
  unchanged; synchronising Acrobat's properties is still separate future work.
- Corrected keyword joining so a key without accompaniment has no trailing comma.
  Malformed dictionary JSON now stops the command before PDF updates, preventing
  processing with a partially decoded dictionary. Other parser rules are retained.
- Automated tests and vet passed. New tests cover missing/existing dictionaries,
  both cross-reference formats, Unicode metadata, empty-field removal, repeated
  edits, preserved first backups, missing final newlines, malformed/version/
  encryption/signature-structure rejection, backup/replacement failures and source
  changes. The signature fixture is structural, not a real cryptographic signature.
  Command tests run with no external executable available and cover dictionary
  errors and unchanged filenames after write failures.
- Ran the newly built macOS Intel app on five disposable copies with an empty
  executable search path. All five backups matched their inputs. qpdf, metadata
  comparison and PDFKit checks passed; all 30 pages rendered identically with
  CoreGraphics at 72 dpi. First-page Poppler renders of both scores were visually
  checked. PDF Info/XMP preservation checks passed. The two score outputs differ
  from the preceding prototype only in the allocated cross-reference-stream object
  number; identical output bytes are not asserted.
- Cross-compilation with CGO disabled passed for macOS Intel, macOS Apple Silicon
  and Windows x64. Only the Intel macOS build was executed here. Evidence and
  disposable PDFs are preserved at `output/pdf/standalone-integration/`; development
  binaries are in ignored `dist/standalone-integration/`. Hashes confirmed the music
  originals/backups, working dictionary and installed legacy executable unchanged.
  No commit, push or deployment was performed.
- Next proposed step: run the Windows build against disposable PDFs on Windows
  to verify replacement, backups and metadata there. Cross-compilation alone does
  not establish Windows runtime compatibility.

### Windows workflow prepared locally (20 September 2026)

- The maintainer has no Windows environment. They approved adding a GitHub Actions
  workflow locally so a hosted Windows runner can execute the tests later.
- Added `.github/workflows/windows.yml`: Windows Server 2025 x64, Go from go.mod,
  CGO disabled, existing tests and vet, a standalone executable build and a separate
  run of executable-level checks. It triggers on pushes, pull requests and manual
  dispatch, with read-only repository permissions and a 15-minute timeout.
  Checkout v7.0.1 and setup-go v7.0.0 are pinned to their verified release commits.
- Added `standalone_test.go`. `FN2FM_TEST_BINARY` selects the actual built app;
  missing configuration skips this optional test, but an invalid supplied path
  fails it. The workflow supplies the newly built executable's absolute path.
  The tests use generated fixtures and temporary dictionaries/folders, with no
  external programs on the app's search path. They check paths with spaces,
  metadata, repeated edits, removal of absent fields, first-backup preservation,
  rejected input and cleanup. A Windows-only case holds the PDF open, expects
  replacement to fail without changing it, then closes it and verifies retry.
- Local verification passed: all tests including the built macOS executable
  checks, vet, Windows x64 test-binary cross-compilation, workflow YAML/configuration
  checks and `git diff --check`. The Windows-only sharing check was skipped on
  macOS. Local execution used Go 1.26.8; the workflow is configured to use the
  minimum version declared in go.mod. Neither the hosted runner nor its PowerShell
  steps have executed yet, so no Windows runtime success is claimed.
- Application code, personal PDFs/dictionary and installed executable were not
  changed by this step. No commit, push, release or hosted workflow run was made.
- Next proposed step: commit and push the development app, tests, generated
  fixtures and workflow to trigger the Windows job, then inspect its result.
  Keep baseline snapshots, personal score outputs and local binaries out of the
  commit; those are investigation artifacts, not required workflow inputs.

### First successful hosted Windows run (21 September 2026)

- With the maintainer's approval, committed and pushed the standalone app, tests,
  generated fixtures, workflow and notes as `07d4700` on `improve-cross-platform`.
  This includes the newly requested persistent output-location configuration
  (including Box/Dropbox) and shared `names.json` options in the backlog below;
  those options remain future work, not implemented behaviour.
- Added `.gitattributes` to treat PDFs as binary. This prevents Windows checkout
  from converting fixture line endings and invalidating PDF byte offsets.
  Personal PDFs, working dictionary, installed executable, baseline snapshots and
  local investigation/build artifacts were excluded from the commit and push.
- GitHub Actions [run 35550647408](https://github.com/Dangthrimble/fn2fm/actions/runs/35550647408)
  completed successfully on Windows Server 2025 x64 using Go 1.25.0. All tests,
  vet, standalone executable build and executable-level checks passed. The log
  confirms the Windows open-file/replacement/retry case ran and passed, rather
  than being skipped. Both cross-reference formats, title-only metadata, repeated
  updates, first backups, paths with spaces and rejected input passed as well.
- This is actual automated Windows execution, not cross-compilation alone. It
  does not establish PDF-reader display behaviour, other Windows versions or
  universal PDF compatibility. The local music files, backups, dictionary and
  installed legacy executable were rechecked by hash and remain unchanged.
- Next proposed step: prepare downloadable development builds for macOS Intel,
  macOS Apple Silicon and Windows x64.

### Versioned development packages (21 September 2026)

- The maintainer approved ready-to-run versioned packages and manual installation/
  update instructions. Dedicated installers and automatic updates are separate
  future work; no stable GitHub Release or installation over the working binary
  is part of this step.
- Added `--version`, which works without a dictionary or PDF arguments. The build
  workflow embeds `dev-<run number>-<12-character commit>` in each executable.
- Replaced the Windows-only workflow with `build.yml`: native macOS Intel,
  macOS Apple Silicon and Windows x64 runners test and vet the source, build each
  versioned executable and exercise it before creating a ZIP. Successful push
  and manual runs upload development artifacts retained for 30 days; pull requests
  test packaging without uploading downloads.
- `internal/buildpackage` includes the executable, `INSTALL.md`, README, version/
  commit/platform details, content/archive SHA-256 checksums and dependency licence
  notices. It verifies the actual executable's version and every ZIP entry's bytes
  and permissions. The starter dictionary is packaged as `names.example.json`,
  preventing extraction from replacing a personal `names.json`.
- `INSTALL.md` covers choosing a platform, first use from Terminal/PowerShell,
  dictionary placement, manual updates, rollback and checksum checks. These builds
  have no Developer ID/notarisation or Windows publisher certificate. The project's
  own licence remains undecided; packaging dependency notices does not select one.
- Local macOS tests and vet passed, including the version command without a
  dictionary and ZIP preservation/checksum tests. An extracted Intel package ran
  with the expected version; independent checks verified its ZIP/content hashes,
  executable permission and absence of a packaged working `names.json`.
- Committed and pushed the packaging implementation as `866c1d7`. GitHub Actions
  [run 35565795539](https://github.com/Dangthrimble/fn2fm/actions/runs/35565795539)
  completed successfully for all three native platforms using Go 1.25.0. Tests,
  vet, executable-level checks, packaging and artifact uploads passed. The
  published development version is `dev-1-866c1d767085`.
- Downloaded all three packages to ignored `dist/development/35565795539/` and
  independently verified their archive checksums and all 22 entries in each ZIP.
  Each records the expected source commit and an unmodified source checkout;
  executable permissions and the example-only dictionary were verified. The
  downloaded Intel executable also returned the expected version on this Mac.
  Verification evidence is in `output/builds/35565795539/download-verification.json`.
- Hash checks confirmed the installed legacy executable, personal dictionary,
  original score files and their existing backups remain unchanged. These checks
  establish package integrity and automated runtime results, not new manual
  PDF-reader checks or signed-installer compatibility.
- Development artifact downloads require GitHub sign-in and expire on
  21 October 2026. A stable release, dedicated installers and automatic updates
  remain future work. Next proposed step: choose the project's licence before
  preparing a stable public release.

### Earlier investigation context

The setup conversation proposed comparing ExifTool with pdfcpu before committing
to removing ExifTool. pdfcpu is a candidate, not an adopted dependency. Compare
Title, Author, Subject and Keywords, relevant PDF/XMP metadata, and actual forScore
import behaviour; matching field names alone does not prove equivalent results.

Two identical disposable copies were prepared in `/tmp/fn2fm-exiftool` and
`/tmp/fn2fm-pdfcpu`, both named `Test Score ~ JoRu[C]+.pdf`. The source PDF already
contained metadata; it was not blank. The conversation ended before any result
from processing the ExifTool copy or performing the pdfcpu comparison was shown.
Temporary files may need recreating. The previously suggested `go run .` command
would use the semicolon-based checked-in parser, so resolve the baseline mismatch
first; it may reject and rename the tilde-named test copy.

ExifTool 13.59 at `/opt/local/bin/exiftool` was reported during setup. If replacement
does not meet compatibility requirements, retain ExifTool and investigate portable
discovery or bundling. Bundling/extracting Windows ExifTool was suggested as an
option; packaging, licensing and update handling must be checked before adoption.

## Additional issues identified during setup

These findings were also checked against the local source on 18 September 2026:

- The repeated key-length condition and discarded `json.Unmarshal` error were
  fixed during standalone writer integration above.
- Preserve the supplied key text for errors: a failed map lookup currently reports
  the empty result rather than the invalid input.
- Handle errors from filename/tag `os.Rename` calls and review their `_rename`
  behaviour. Other metadata parsing errors still skip the file; the standalone
  writer now retains filenames on PDF update errors (see integration above).

Further proposals from the earlier review:

- An explicit dictionary-path option was an earlier suggestion, not an agreed
  change. Current-working-directory lookup is intentional (see requirements above).
- Reconcile README name-code examples with the supplied dictionary; consider an
  example dictionary or overridable defaults.
- Add command-line help/options and integration tests. Extend existing tests for
  changed behaviour rather than replacing the current suite wholesale.
- Make wildcard handling consistent across macOS and Windows shells.
- Consider named Go string constants for reusable regex fragments if regex-based
  parsing improves clarity. The regex conversation explored this technique; it
  did not decide to replace the existing parser or define the full grammar.

## Distribution proposals and completed setup

- Proposed initial builds: macOS Intel, macOS Apple Silicon and Windows x64.
  Windows ARM64 was a possible later addition, not an agreed initial target.
- Consider GitHub Actions for checks and builds, then tagged GitHub Releases with
  downloadable executables and checksums. Verify current service limits when
  implementing the no-cost workflow.
- Choose a project licence before presenting the project as open source. Signing,
  notarisation and an optional download website were deferred considerations.
- GitHub repository: https://github.com/Dangthrimble/fn2fm. Development branch at
  this handover: `improve-cross-platform`.
- Existing variable/return refactoring is preserved in commit `f2a8734`.
  README renaming, ignore-rule cleanup, removal of stray gofmt documentation and
  obsolete local files were completed during setup; do not repeat them as backlog.
- The setup conversation recorded passing tests and vet checks and reported
  cross-compilation for the three initial targets. These are historical results,
  not verification of future changes or runtime behaviour on each platform.
- GitHub plugin troubleshooting added no product requirements. Later setup
  recorded successful GitHub authentication and a push, so the earlier connection
  failure should not be treated as a continuing blocker.

## Existing review items

- Review use of `:=` to avoid shadowing.
- Review use of named return parameters and blank returns.
- Support use of glob to pass in wildcards.

## Additional backlog

- Support multiple versions of a song.
- Support suites of music (e.g. oratorios, symphonies, operas and musicals) and
  their individual movements.
- Let the user select a file and supply the details from which to create its
  filename, rather than requiring the user to rename it first.
- Support renaming files with password-protected metadata.
- Check filename uniqueness against existing files.
- Prepare a stable release after selecting the project licence. Versioned
  development packages and native automated checks for all three initial
  platforms are complete (see above).
- Consider dedicated installers and automatic updates; current packages use
  manual installation and manual updates as documented in `INSTALL.md`.
- Add a configuration file for genre or arranger handling; define the supported
  choices and metadata mapping before implementation.
- Optionally retain configuration between application invocations, including a
  predefined output location so generated files can always be written to the same
  destination. Allow that destination to be on Box or Dropbox. Decide whether to
  use a locally synchronised folder or direct cloud access during design; neither
  approach is yet selected or implemented.
- Allow multiple users of the application to share `names.json`. Define how the
  shared dictionary is located and how updates and concurrent edits are handled;
  the current per-working-directory dictionary behaviour remains unchanged.
- Refactor the code toward more idiomatic Go.
- Consider additional platforms beyond the three supported development builds.
- Investigate forScore PDF-metadata parsing compatibility, particularly shared
  setlists when users have different metadata-parsing settings.

### forScore metadata parsing and shared setlists

forScore calls this operation **Fetching PDF metadata**. Its
[official instructions](https://forscore.co/kb/fetching-pdf-metadata/) describe
manual fetching and automatic fetching for newly added files.

The maintainer supplied these steps for their installed version:
- Automatic: Settings > Advanced options > Metadata > Automatic fetching for new
  files. This is currently turned off, as reported by the maintainer.
- Per file: open the file's properties, tap the ellipsis, select **Fetch…**, then
  tap the tick to save.

Use per-file fetching for the comparison; enabling automatic fetching globally
is unnecessary. These instructions do not establish a successful comparison:
the actual fetched values still need to be checked in forScore.

forScore can optionally parse embedded PDF metadata and use it as visible metadata
within the app. The reported corner case is that doing so sometimes changes how a
score is identified within forScore. A setlist containing that score may then fail
to share successfully with someone who has the same PDFs but does not parse their
metadata.

Investigate which embedded metadata fields affect score identification and setlist
references, and whether fn2fm can generate metadata that preserves compatibility
with parsing either enabled or disabled. This is a forScore interoperability
investigation, rather than a general filename-normalisation task; the cause and
solution have not yet been established.
