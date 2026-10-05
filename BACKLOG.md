# Backlog

Outstanding work and unresolved questions for fn2fm. Order does not imply an
agreed priority, and suggested approaches are not settled decisions. Completed
work and its rationale are in [DEVELOPMENT_HISTORY.md](DEVELOPMENT_HISTORY.md);
current usage is in [README.md](README.md).

## Reliability and command-line behaviour

- Handle errors from filename/tag `os.Rename` calls and review the automatic
  `_rename` behaviour. Other metadata parsing errors still skip the file; PDF
  update failures already retain the filename and contents.
- Add command-line help and consider a dry-run option and a success/failure
  summary. Keep these as proposals until their behaviour is defined.
- Make wildcard/glob handling consistent across macOS and Windows shells.
- Reconcile README name-code examples with the starter dictionary. Preserve the
  distinction between the repository example and personal working dictionaries.
- Extend the existing regression and integration tests as behaviour changes.
  Standalone command and built-executable tests already exist; do not replace
  the suite wholesale or treat integration testing as entirely absent.

## Scores and filename creation

- Support multiple versions of a song.
- Support suites of music (for example oratorios, symphonies, operas and musicals)
  and their individual movements. Existing title prefixes can be retained today;
  richer suite/movement support remains undefined.
- Let the user select a file and supply the details needed to create its filename,
  rather than requiring the user to rename it first.
- Support renaming files with password-protected metadata. Define the intended
  operation separately from writing encrypted PDFs, which is currently rejected.
- Check filename uniqueness against existing files.
- Confirm the intended compatibility of an absent accompaniment marker before
  changing it; the current parser and README permit omission.
- Keep the restricted ASCII filename policy. Reconcile the exact punctuation
  set with tests before any future changes. Preserve the distinction between
  forScore's title restrictions, documented in its 6 March 2024 support reply,
  and fn2fm's broader cloud-storage/ASCII restrictions (see the README).
- Resolve the named-key grammar if broadening support: the earlier requirement
  used `[A-G][#b]?m?`, while the implementation accepts a fixed list of keys.
  Accidental counts remain valid and do not imply major or minor. Empty keys
  are now supported and are no longer an outstanding item.

## Configuration and dictionaries

- Add configuration for genre or arranger handling; define the supported choices
  and metadata mapping before implementation.
- Optionally retain configuration between invocations, including a predefined
  output location. Allow a destination on Box or Dropbox; choose between a locally
  synchronised folder and direct cloud access during design. Neither is implemented.
- Allow multiple users to share `names.json`. Define dictionary discovery,
  updates and concurrent-edit handling.
- An explicit dictionary-path option was suggested but not agreed. Loading from
  the current working directory is intentional, so users can keep their dictionary
  with their PDFs. Executable-directory lookup is not needed for that workflow.

## PDF metadata and forScore compatibility

- Investigate consistency between PDF Info and XMP as a separate behaviour change.
  The current writer preserves XMP, matching the legacy workflow. Acrobat can
  therefore show older or blank properties despite the intended Info values.
  Synchronising both stores has not been approved as current behaviour.
- Broaden compatibility evidence beyond the tested fixtures and two music scores.
  Current restrictions include encrypted/signed PDFs, PDF versions outside
  1.4–1.7 and hybrid/repaired cross-reference structures. Supporting them requires
  a separate design and tests; no commitment to remove these restrictions is made.
- Investigate shared-setlist compatibility when users have different forScore
  metadata-fetching settings, as described below.

### forScore metadata fetching and shared setlists

The earlier investigation recorded forScore's term **Fetching PDF metadata** and
its [official instructions](https://forscore.co/kb/fetching-pdf-metadata/) for manual
fetching and automatic fetching for newly added files. Recheck the instructions
against the installed version when investigating.

The maintainer supplied these steps for their installed version:

- Automatic: Settings > Advanced options > Metadata > Automatic fetching for new
  files. This was turned off when reported.
- Per file: open the file's properties, tap the ellipsis, select **Fetch…**, then
  tap the tick to save.

Use per-file fetching for comparisons; enabling automatic fetching globally is
unnecessary. Check the actual fetched values. The earlier successful two-score
comparison is recorded in the history and does not resolve shared-setlist behaviour.

The reported corner case is that fetching embedded metadata sometimes changes
how a score is identified within forScore. A setlist may then fail to share with
someone who has the same PDFs but does not fetch their metadata. The cause and
solution have not been established.

Investigate which fields affect score identification and setlist references, and
whether fn2fm can preserve compatibility with fetching enabled or disabled. This
is a forScore interoperability issue, not a general filename-normalisation task.

## Distribution

- Prepare a numbered stable release from the tested MIT-licensed source.
  Versioned development packages, checksums and native automated checks for the
  six configured platforms are complete. Use tagged GitHub Releases for durable
  downloads; current development artifacts require sign-in and expire.
- Consider dedicated installers and automatic updates. Current packages use
  manual installation and updates as described in [INSTALL.md](INSTALL.md).
- Consider signing, macOS notarisation and an optional download website.
- Consider platforms beyond the six configured targets: macOS Intel/Apple Silicon,
  Windows x64/ARM64 and Linux x64/ARM64. All six passed hosted runtime and packaging
  verification on 27 September 2026; see the development history for the run.
- Verify service limits when extending the no-cost distribution workflow.
- Manual Windows PDF-reader display checks remain distinct from the successful
  hosted Windows runtime tests.

## Code maintenance

- Review use of `:=` to avoid shadowing.
- Review named return parameters and blank returns; move toward idiomatic Go
  where it improves clarity without undoing preserved refactoring.
- Consider named string constants for reusable regex fragments only if regex-based
  parsing improves clarity. Replacing the parser was discussed, not decided.
- Ignore the binary created by `go build ./cmd/fn2fm` when no `-o` path is given.
  `.gitignore` currently only ignores `/fn2fm` at the repository root, so that
  binary can sit in `cmd/fn2fm/` and be committed by mistake.
- Rename `filename.ValidateAndExtension` to something clearer. The old name was
  `validateFilenameAndExtension`; the exported name was shortened during the
  layout move and is harder to read.
- Let dictionary errors and whitespace warnings use the path that was actually
  loaded. `names.Load` takes a path but still says `names.json` in its messages.
  That matches today's CLI, which always loads `names.json` from the current
  folder.
- Tidy how a missing or invalid dictionary is reported. The CLI still logs the
  underlying error, then prints `Unable to load names.json`, and skips the extra
  log for malformed JSON by looking at the error text. Behaviour should stay the
  same unless a change is agreed.
- Share the PDF test helpers instead of keeping two copies: one in
  `cmd/fn2fm/pdf_test_helpers_test.go` and the same helpers in
  `internal/pdfmeta/write_test.go`.
- Remove or use the unused `withAcc` and `withoutAcc` constants in
  `internal/filename`. They were already unused before the layout move. The
  accompaniment parser still checks `+` and `-` as literals.

## Remaining provenance question

The separate tilde source tree and installed legacy executable were located and
preserved. The maintainer recalls that the executable probably came from a
modification to the Git source after the 2024 forScore email, changing semicolons
to tildes. File timestamps do not establish a contrary sequence. The exact
source/build revision and whether the preserved source is unchanged since that
build remain unverified.
The forScore reason for excluding semicolons is now documented from the 2024
support correspondence; it is no longer an unresolved policy question.
Investigate further only if exact historical build provenance is needed;
this does not block the reconciled tilde-only implementation.
