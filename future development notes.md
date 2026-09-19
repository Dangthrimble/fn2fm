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
  including the leading space; this has not been corrected.
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

## Paused ExifTool replacement investigation

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

- Fix the repeated key-length condition in keyword construction, which can leave
  a trailing comma and space when there is a key but no accompaniment.
- Handle the discarded `json.Unmarshal` error for malformed `names.json`.
- Preserve the supplied key text for errors: a failed map lookup currently reports
  the empty result rather than the invalid input.
- Handle errors from `os.Rename` and choose consistent failure behaviour. Filename
  and tag errors currently attempt `_rename`, ExifTool errors attempt `_error`,
  and other metadata errors skip the file. Review whether automatic renaming is
  appropriate before changing this behaviour.

Further proposals from the earlier review:

- Make name-dictionary discovery robust when the working directory differs from
  the executable directory. Consider an explicit path option and documented lookup
  rules; keep dictionaries customisable rather than permanently baking in one.
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

- Investigate removing the ExifTool dependency.
- Add a configuration file for genre or arranger handling; define the supported
  choices and metadata mapping before implementation.
- Refactor the code toward more idiomatic Go.
- Support builds for multiple platforms.
- Investigate forScore PDF-metadata parsing compatibility, particularly shared
  setlists when users have different metadata-parsing settings.

### forScore metadata parsing and shared setlists

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
