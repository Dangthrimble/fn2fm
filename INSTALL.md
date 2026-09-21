# Installing and updating fn2fm development builds

These packages contain a command-line executable. You do not need Go or ExifTool
to run it. There is no graphical installer or automatic update service.

## Choose and download a package

On the repository's GitHub **Actions** page, open a successful **Build and test**
run and download the package for your computer from **Artifacts**:

- `macos-intel`: a Mac with an Intel processor.
- `macos-apple-silicon`: a Mac with an Apple M-series chip.
- `windows-x64`: an Intel/AMD 64-bit Windows computer.

Each package is named `fn2fm-dev-<run>-<commit>-<platform>.zip`. The version identifies
the build run and source commit. These are development builds, not numbered stable
releases. GitHub artifact downloads require signing in; these artifacts are kept
for 30 days. A later successful run provides a new set of packages. See
[GitHub's download instructions](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/download-workflow-artifacts).

Extract the downloaded artifact, then the versioned ZIP inside it. Keep the
extracted folder together: it contains the executable, this guide, README,
`VERSION.txt`, checksums, a starter dictionary and dependency licence notices.

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

These development builds have no Apple Developer ID signature or notarisation.
macOS may require an explicit approval before first use. For a download you have
verified and trust, follow Apple's
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

These development executables have no Windows publisher signature. Windows or a
managed computer's policy may require approval to run them.

## Keep your dictionary separate

fn2fm reads `names.json` from the folder where you run the command, not from the
executable's folder. Keep your personal dictionary alongside the scores.

The package includes `names.example.json`. Only if you do not already have a
dictionary, copy that file into your scores folder and name the copy `names.json`.
Never replace a customised `names.json` with the example during an update.

## Check the download

Each artifact includes a `.zip.sha256` file containing the SHA-256 checksum of its
versioned ZIP. Compare it with the checksum you calculate before extracting:

```sh
shasum -a 256 fn2fm-dev-<run>-<commit>-macos-intel.zip
```

In PowerShell, use:

```powershell
Get-FileHash -Algorithm SHA256 -LiteralPath 'fn2fm-dev-<run>-<commit>-windows-x64.zip'
```

The ZIP also includes `SHA256SUMS.txt` for its contents. Checksums detect changed
or incomplete files; they do not provide a publisher signature.

## Manual updates and rollback

1. Download and extract a newer successful build for the same platform into a new
   version folder. The application does not check for or download updates itself.
2. Run the new executable with `--version`, then try it on disposable PDF copies.
3. Use the new executable's path for subsequent commands. If you previously copied
   the executable into a folder on your PATH, back up that executable before
   replacing only that file. Do not replace your dictionary or PDF collection.
4. To roll back, use the older executable again. Changing executable versions does
   not undo PDF edits; retain your PDFs and their `.pdf_original` backups.

The current application updates each PDF in place after validation and preserves
its first `.pdf_original` backup. A fixed output folder, shared dictionaries,
dedicated installers and automatic updates remain future development options.
See README for supported filenames and PDF limitations. A project licence for
fn2fm has not yet been selected; included dependency notices cover those components.
