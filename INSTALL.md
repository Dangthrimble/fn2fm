# Installing and updating fn2fm

These packages contain a command-line executable. You do not need Go or ExifTool
to run it. There is no graphical installer or automatic update service. Builds
are not code-signed or notarised; checksums check integrity only.

## Choose and download a package

Prefer a numbered release from the repository's
[Releases page](https://github.com/Dangthrimble/fn2fm/releases)
(for example [v0.1.1](https://github.com/Dangthrimble/fn2fm/releases/tag/v0.1.1)).
Download the Asset for your computer:

- `macos-intel`: a Mac with an Intel processor.
- `macos-apple-silicon`: a Mac with an Apple M-series chip.
- `windows-x64`: an Intel/AMD 64-bit Windows computer.
- `windows-arm64`: an ARM64 Windows computer.
- `linux-x64`: an Intel/AMD 64-bit Linux computer.
- `linux-arm64`: an ARM64 Linux computer running a 64-bit OS.

Release packages are named `fn2fm-v0.1.1-<platform>.zip` (and similarly for later
versions). Use `fn2fm --version` to confirm the stamp matches `VERSION.txt`.

For interim testing only, successful **Build and test** runs on the
[Actions page](https://github.com/Dangthrimble/fn2fm/actions) still publish
development packages named `fn2fm-dev-<run>-<commit>-<platform>.zip`. Those
downloads require signing in and are kept for 30 days. See
[GitHub's download instructions](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/download-workflow-artifacts).
Older Actions runs may only contain the three original macOS and Windows x64
packages.

Extract the downloaded ZIP (for Actions artifacts, extract the outer artifact,
then the versioned ZIP inside). Keep the extracted folder together: it contains
the executable, this guide, README, `BACKLOG.md`, `DEVELOPMENT_HISTORY.md`,
`VERSION.txt`, checksums, a starter dictionary, `LICENSE` and dependency notices.

## macOS

1. Move the extracted version folder to a convenient location, such as
   `~/Applications/fn2fm/`. Keep an older version until you have checked the new one.
2. Open Terminal. Drag the `fn2fm` executable into Terminal to insert its full path,
   add a space followed by `--version`, then press Return. For example, after
   substituting the actual folder name:

   ```sh
   "$HOME/Applications/fn2fm/<version-folder>/fn2fm" --version
   ```

3. The version should match `VERSION.txt`. If extraction removed execute permission,
   run `chmod u+x` followed by that same quoted executable path.
4. Change Terminal's current folder to the folder holding your scores and
   `names.json`, then invoke the executable by its full path with quoted PDF names.

These builds have no Apple Developer ID signature or notarisation. macOS may
require an explicit approval before first use. For a download you have verified
and trust, follow Apple's
[instructions for opening an unnotarised app](https://support.apple.com/en-gb/102445).

Example, using a temporary copy of a score and your own actual folder paths:

```sh
cd "$HOME/Documents/Score Test Copies"
"$HOME/Applications/fn2fm/<version-folder>/fn2fm" "Test Score ~ JoRu[C]+.pdf"
```

## Windows

1. Extract the package into a folder you own, such as
   `%LOCALAPPDATA%\Programs\fn2fm\<version-folder>`. Administrator rights are not
   required for this folder. Keep an older version until you have checked the new one.
2. Open PowerShell and run the executable's full path using the `&` operator:

   ```powershell
   & "$env:LOCALAPPDATA\Programs\fn2fm\<version-folder>\fn2fm.exe" --version
   ```

3. Check that the version matches `VERSION.txt`. Change PowerShell's current
   folder to the folder holding your scores and `names.json`, then process a copy:

   ```powershell
   Set-Location "$env:USERPROFILE\Documents\Score Test Copies"
   & "$env:LOCALAPPDATA\Programs\fn2fm\<version-folder>\fn2fm.exe" "Test Score ~ JoRu[C]+.pdf"
   ```

These executables have no Windows publisher signature. Windows or a managed
computer's policy (for example SmartScreen) may require approval to run them.

## Linux

1. Extract the versioned ZIP into a folder you own, for example
   `~/.local/opt/fn2fm/<version-folder>`. Keep the extracted files together.
2. Open a terminal and check the executable's version:

   ```sh
   "$HOME/.local/opt/fn2fm/<version-folder>/fn2fm" --version
   ```

3. Check that the version matches `VERSION.txt`. If extraction removed execute
   permission, run `chmod u+x` followed by the quoted executable path.
4. Change to the folder containing your score copies and `names.json`, then run:

   ```sh
   cd "$HOME/Documents/Score Test Copies"
   "$HOME/.local/opt/fn2fm/<version-folder>/fn2fm" "Test Score ~ JoRu[C]+.pdf"
   ```

Linux builds have CGO disabled and do not need Go or ExifTool installed. The CI
jobs use Ubuntu 24.04; they do not establish compatibility with every distribution.

## Keep your dictionary separate

fn2fm reads `names.json` from the folder where you run the command, not from the
executable's folder. Keep your personal dictionary alongside the scores.

The package includes `names.example.json`. Only if you do not already have a
dictionary, copy that file into your scores folder and name the copy `names.json`.
Never replace a customised `names.json` with the example during an update.

## Check the download

Each release Asset includes a `.zip.sha256` file containing the SHA-256 checksum
of its versioned ZIP. Compare it with the checksum you calculate before extracting:

```sh
shasum -a 256 fn2fm-v0.1.1-macos-intel.zip
```

On Linux, use `sha256sum` with the downloaded ZIP filename. In PowerShell, use:

```powershell
Get-FileHash -Algorithm SHA256 -LiteralPath 'fn2fm-v0.1.1-windows-x64.zip'
```

For development packages from Actions, use the `fn2fm-dev-<run>-<commit>-…`
filename instead. The ZIP also includes `SHA256SUMS.txt` for its contents.
Checksums detect changed or incomplete files; they do not provide a publisher
signature.

## Manual updates and rollback

1. Download and extract a newer release (or development build) for the same
   platform into a new version folder. The application does not check for or
   download updates itself.
2. Run the new executable with `--version`, then try it on disposable PDF copies.
3. Use the new executable's path for subsequent commands. If you previously copied
   the executable into a folder on your PATH, back up that executable before
   replacing only that file. Do not replace your dictionary or PDF collection.
4. To roll back, use the older executable again. Changing executable versions does
   not undo PDF edits; retain your PDFs and their `.pdf_original` backups.

The current application updates each PDF in place after validation and preserves
its first `.pdf_original` backup. A fixed output folder, shared dictionaries,
dedicated installers, automatic updates, signing and notarisation remain future
development options. See README for supported filenames and PDF limitations.
fn2fm uses the MIT licence in `LICENSE`; dependencies retain their own licences,
included with the package.
