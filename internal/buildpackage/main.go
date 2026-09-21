// buildpackage creates a portable development archive from a tested executable.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

type entry struct {
	name string
	data []byte
	mode fs.FileMode
}

func main() {
	binary := flag.String("binary", "", "path of the tested executable")
	version := flag.String("version", "", "version embedded in the executable")
	out := flag.String("out", "dist/packages", "directory for ZIP and checksum")
	flag.Parse()
	if err := packageBinary(*binary, *version, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func packageBinary(binary, version, out string) error {
	if binary == "" || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`).MatchString(version) {
		return errors.New("supply -binary and a version containing letters, digits, dots, underscores, pluses or hyphens")
	}
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return err
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	osName, arch := settings["GOOS"], settings["GOARCH"]
	platform := map[string]string{"darwin/amd64": "macos-intel", "darwin/arm64": "macos-apple-silicon", "windows/amd64": "windows-x64"}[osName+"/"+arch]
	if info.Path != "fn2fm" || platform == "" || settings["CGO_ENABLED"] != "0" {
		return errors.New("expected a standalone fn2fm binary for a supported platform")
	}
	absoluteBinary, err := filepath.Abs(binary)
	if err != nil {
		return err
	}
	// Trimmed builds do not reliably retain linker flags in build information.
	// Packaging runs on the target platform, so ask the actual executable.
	versionOutput, err := exec.Command(absoluteBinary, "--version").Output()
	if err != nil || string(versionOutput) != fmt.Sprintf("fn2fm %s (%s/%s)\n", version, osName, arch) {
		return fmt.Errorf("executable does not report the requested version/target: %q (%v)", versionOutput, err)
	}
	stem := "fn2fm-" + version + "-" + platform
	executable := "fn2fm"
	if osName == "windows" {
		executable += ".exe"
	}
	var files []entry
	add := func(name, path string, mode fs.FileMode) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, entry{name, data, mode})
		return nil
	}
	for _, file := range []struct {
		name, path string
		mode       fs.FileMode
	}{
		{executable, binary, 0755},
		{"INSTALL.md", "INSTALL.md", 0644},
		{"README.md", "README.md", 0644},
		{"names.example.json", "names.json", 0644},
	} {
		if err := add(file.name, file.path, file.mode); err != nil {
			return err
		}
	}
	versionText := fmt.Sprintf("fn2fm %s\nPlatform: %s (%s/%s)\nCommit: %s\nSource modified: %s\nGo: %s\nDevelopment build; updates are manual.\n", version, platform, osName, arch, settings["vcs.revision"], settings["vcs.modified"], info.GoVersion)
	files = append(files, entry{"VERSION.txt", []byte(versionText), 0644})
	licenses, err := dependencyLicenses(info.Deps)
	if err != nil {
		return err
	}
	files = append(files, licenses...)
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	archivePath := filepath.Join(out, stem+".zip")
	if err := writeArchive(archivePath, stem, files); err != nil {
		return err
	}
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return err
	}
	checksum := fmt.Sprintf("%x  %s\n", sha256.Sum256(data), filepath.Base(archivePath))
	if err := os.WriteFile(archivePath+".sha256", []byte(checksum), 0644); err != nil {
		return err
	}
	fmt.Println(archivePath)
	return nil
}

func dependencyLicenses(deps []*debug.Module) ([]entry, error) {
	var files []entry
	var index strings.Builder
	index.WriteString("Third-party software included in this build\n\nThese notices apply to dependencies, not to fn2fm's own code.\nThe project licence for fn2fm has not yet been selected.\n\n")
	goRoot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(goRoot)), "LICENSE"))
	if err != nil {
		return nil, err
	}
	files = append(files, entry{"THIRD_PARTY_LICENSES/Go/LICENSE", data, 0644})
	index.WriteString("Go standard library/runtime: THIRD_PARTY_LICENSES/Go/LICENSE\n")
	for _, dep := range deps {
		if dep.Replace != nil {
			return nil, fmt.Errorf("cannot package an unreviewed module replacement: %s", dep.Path)
		}
		output, err := exec.Command("go", "list", "-m", "-json", dep.Path).Output()
		if err != nil {
			return nil, fmt.Errorf("locate dependency %s: %w", dep.Path, err)
		}
		var module struct{ Dir, Version string }
		if err := json.Unmarshal(output, &module); err != nil {
			return nil, err
		}
		if module.Dir == "" || module.Version != dep.Version {
			return nil, fmt.Errorf("dependency cache does not match binary: %s", dep.Path)
		}
		found := false
		err = filepath.WalkDir(module.Dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.Type().IsRegular() {
				return nil
			}
			upper := strings.ToUpper(d.Name())
			if !(upper == "PATENTS" || strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "NOTICE") || strings.HasPrefix(upper, "COPYING")) {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(module.Dir, path)
			if err != nil {
				return err
			}
			name := "THIRD_PARTY_LICENSES/" + dep.Path + "@" + dep.Version + "/" + filepath.ToSlash(rel)
			files = append(files, entry{name, data, 0644})
			found = true
			return nil
		})
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("no licence found for %s", dep.Path)
		}
		fmt.Fprintf(&index, "%s %s: THIRD_PARTY_LICENSES/%s@%s/\n", dep.Path, dep.Version, dep.Path, dep.Version)
	}
	files = append(files, entry{"THIRD_PARTY_NOTICES.txt", []byte(index.String()), 0644})
	return files, nil
}

func writeArchive(path, stem string, files []entry) (err error) {
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	var manifest strings.Builder
	for _, file := range files {
		fmt.Fprintf(&manifest, "%x  %s\n", sha256.Sum256(file.data), file.name)
	}
	files = append(files, entry{"SHA256SUMS.txt", []byte(manifest.String()), 0644})
	output, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		output.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	zw := zip.NewWriter(output)
	for _, file := range files {
		header := &zip.FileHeader{Name: stem + "/" + file.name, Method: zip.Deflate}
		header.SetMode(file.mode)
		header.SetModTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := writer.Write(file.data); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	// Reopen the ZIP and compare every byte and mode to the intended package.
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	if len(archive.File) != len(files) {
		return errors.New("package verification failed: wrong file count")
	}
	for i, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		want := files[i]
		if err != nil || file.Name != stem+"/"+want.name || file.Mode().Perm() != want.mode || !bytes.Equal(data, want.data) {
			return fmt.Errorf("package verification failed: %s", file.Name)
		}
	}
	return nil
}
