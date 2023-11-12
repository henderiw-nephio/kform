package pkgio

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/grabber"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
)

type PkgProviderReadWriter interface {
	Reader
	Writer
	Process
	//ProcessProviderRequirements(context.Context, *Data) (*Data, error)
}

func NewPkgProviderReadWriter(rootPath string, pkgs []*address.Package) PkgProviderReadWriter {
	path := filepath.Join(rootPath, ".kform", "providers")
	pathExists := true
	if _, err := os.Stat(path); err != nil {
		pathExists = false
	}
	fs := fsys.NewDiskFS(path)
	return &pkgProviderReadWriter{
		pkgs:              pkgs,
		validateChecksums: map[string][]*address.Package{},
		reader: &PkgReader{
			PathExists:     pathExists,
			Fsys:           fs,
			MatchFilesGlob: []string{"*"},
			// ignore rules are not required since we want to match any file
			IgnoreRules: ignore.Empty(""),
			Checksum:    true,
		},

		writer: &pkgProviderWriter{
			PathExists: pathExists,
			fsys:       fs,
			rootPath:   rootPath,
		},
	}
}

type pkgProviderReadWriter struct {
	pkgs              []*address.Package
	validateChecksums map[string][]*address.Package
	reader            *PkgReader
	writer            *pkgProviderWriter
}

func (r *pkgProviderReadWriter) Read(ctx context.Context, data *Data) (*Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgProviderReadWriter) Write(ctx context.Context, data *Data) error {
	return r.writer.Write(ctx, data)
}

func (r *pkgProviderReadWriter) Process(ctx context.Context, data *Data) (*Data, error) {
	return r.processProviderRequirements(ctx, data)
}

func (r *pkgProviderReadWriter) processProviderRequirements(ctx context.Context, data *Data) (*Data, error) {
	// delete the paths/files that should not be in the directory
	// based on the requirements
	for path := range data.List() {
		found := false
		for _, pkg := range r.pkgs {
			if path == pkg.Path() {
				found = true
				break
			}
		}
		if !found {
			// we could add diagnostics to this process, to indicate with a warning
			// these files exists but should not be there
			data.Delete(path)
		}
	}

	for _, pkg := range r.pkgs {
		if !data.Exists(pkg.Path()) {
			data.Add(pkg.Path(), nil)
		} else {
			if !pkg.IsLocal() {
				if _, ok := r.validateChecksums[pkg.ChecksumURL()]; !ok {
					r.validateChecksums[pkg.ChecksumURL()] = []*address.Package{}
				}
				r.validateChecksums[pkg.ChecksumURL()] = append(r.validateChecksums[pkg.ChecksumURL()], pkg)
			}
		}
	}
	data, err := r.processChecksums(ctx, data)
	if err != nil {
		return data, err
	}
	return r.prepareWriter(ctx, data)
}

// the remainingfiles in the data
func (r *pkgProviderReadWriter) processChecksums(ctx context.Context, data *Data) (*Data, error) {
	remoteCheckSums := map[string]string{}
	for checksumURL, pkgs := range r.validateChecksums {
		resp, err := http.Get(checksumURL)
		if err != nil {
			return data, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return data, fmt.Errorf("failed to download checksum file, status code: %d", resp.StatusCode)
		}

		s := bufio.NewScanner(resp.Body)
		for s.Scan() {
			line := s.Text()

			for _, pkg := range pkgs {
				if strings.HasSuffix(line, pkg.Filename()) {
					if _, ok := remoteCheckSums[pkg.Path()]; ok {
						return data, fmt.Errorf("duplicate checksum for: %s", pkg.Path())
					}
					remoteCheckSums[pkg.Path()] = strings.TrimSpace(strings.TrimSuffix(line, pkg.Filename()))
				}
			}
		}

		if err := s.Err(); err != nil {
			return data, err
		}

		for _, pkg := range pkgs {
			if _, ok := remoteCheckSums[pkg.Path()]; !ok {
				return data, fmt.Errorf("checksum for %s not found in the url %s", pkg.Path(), checksumURL)
			}
		}
	}
	for path, hash := range remoteCheckSums {
		b, err := data.Get(path)
		if err != nil {
			// this should never happen, since we added this path to the data before
			return data, fmt.Errorf("path %s not found", path)
		}
		if string(b) == hash {
			data.Delete(path)
		}
	}
	return data, nil
}

// prepareWriter updates the data with the URL such that the writer
// knows the file and
func (r *pkgProviderReadWriter) prepareWriter(ctx context.Context, data *Data) (*Data, error) {
	for _, pkg := range r.pkgs {
		if data.Exists(pkg.Path()) {
			data.Update(pkg.Path(), []byte(pkg.URL()))
		}
	}
	return data, nil
}

type pkgProviderWriter struct {
	PathExists bool
	fsys       fsys.FS
	rootPath   string
}

func (r *pkgProviderWriter) Write(ctx context.Context, data *Data) error {
	providerPath := filepath.Join(r.rootPath, ".kform", "providers")
	if !r.PathExists {
		os.MkdirAll(providerPath, 0755|os.ModeDir)
		r.fsys = fsys.NewDiskFS(providerPath)
	}

	fileLocs := map[string][]string{}
	for path, url := range data.List() {
		// create a dir for all the paths
		r.fsys.MkdirAll(filepath.Dir(path))
		fileLocs[filepath.Join(providerPath, path)] = []string{url}
	}

	respch, err := grabber.GetBatch(ctx, 3, fileLocs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	// start a ticker to update progress every 200ms
	t := time.NewTicker(200 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grabber.Response, 0)
	for completed < grabber.GetTotalURLs(fileLocs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					// print final result
					if resp.Err() != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Err())
					} else {
						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesComplete(), resp.Size, int(100*resp.Progress()))
					}
					// mark completed
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesComplete(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}
	t.Stop()
	return nil
}

type PkgValidator interface {
	Writer
}

func NewPkgValidator() PkgValidator {
	return &pkgValidator{}
}

type pkgValidator struct{}

func (r *pkgValidator) Write(ctx context.Context, data *Data) error {
	providers := []string{}
	for path := range data.List() {
		providers = append(providers, path)
	}
	if len(providers) > 0 {
		return fmt.Errorf("please run init, the following providers are not up to date %v", providers)
	}
	return nil
}
