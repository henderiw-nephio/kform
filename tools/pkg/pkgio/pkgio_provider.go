package pkgio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/data"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/grabber"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

type PkgProviderReadWriter interface {
	Reader
	Writer
	Process
	//ProcessProviderRequirements(context.Context, *Data) (*Data, error)
}

func NewPkgProviderReadWriter(rootPath string, providers cache.Cache[*address.Package]) PkgProviderReadWriter {
	path := filepath.Join(rootPath, ".kform", "providers")
	pathExists := true
	if _, err := os.Stat(path); err != nil {
		pathExists = false
	}
	fs := fsys.NewDiskFS(path)
	return &pkgProviderReadWriter{
		providers:         providers,
		validateChecksums: map[string][]*address.Package{},
		reader: &PkgReader{
			PathExists:     pathExists,
			Fsys:           fs,
			MatchFilesGlob: []string{"*"},
			// ignore rules are not required since we want to match any file
			IgnoreRules: ignore.Empty(""),
			Checksum:    true, // this flag reads the checksum
		},
		writer: &pkgProviderWriter{
			PathExists: pathExists,
			fsys:       fs,
			rootPath:   rootPath,
		},
	}
}

type pkgProviderReadWriter struct {
	providers         cache.Cache[*address.Package]
	validateChecksums map[string][]*address.Package
	reader            *PkgReader
	writer            *pkgProviderWriter
}

func (r *pkgProviderReadWriter) Read(ctx context.Context, data *data.Data) (*data.Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgProviderReadWriter) Write(ctx context.Context, data *data.Data) error {
	return r.writer.Write(ctx, data)
}

func (r *pkgProviderReadWriter) Process(ctx context.Context, data *data.Data) (*data.Data, error) {
	return r.processProviderRequirements(ctx, data)
}

// get versions from the installed providers -> right now we assume 1 provider has 1 version
// check if it is part of the candidates
// if not -> delete path/done; if yes -> check chechsum; if nok -> delete it
//

func (r *pkgProviderReadWriter) processProviderRequirements(ctx context.Context, data *data.Data) (*data.Data, error) {
	log := log.FromContext(ctx)
	// walk over the paths and delete the once that are not relevant
	// based on the provider requirements/packages
	for path, hash := range data.List() {
		for _, pkg := range r.providers.List() {
			if strings.HasPrefix(path, pkg.BasePath()) {
				installedVersion := address.GetVersionFromPath(path)
				log.Info("processProviderRequirements", "installedVersion", installedVersion)
				if pkg.HasVersion(installedVersion) {
					remoteHash, err := pkg.GetRemoteChecksum(ctx, installedVersion)
					if err != nil {
						return data, err
					}
					if string(hash) == remoteHash {
						// found and valid
						pkg.UpdateSelectedVersion(installedVersion)
						data.Delete(path)
					} else {
						// remove the files from the fsys
						if err := r.writer.fsys.RemoveAll(pkg.BasePath()); err != nil {
							return data, err
						}
						data.Delete(path)
						// update the path to install the latest
						toBeInstalledVersion := pkg.Newest()
						pkg.UpdateSelectedVersion(toBeInstalledVersion)
						data.Add(pkg.FilePath(toBeInstalledVersion), []byte(pkg.URL(toBeInstalledVersion)))
					}
				} else {
					// remove the files from the fsys
					if err := r.writer.fsys.RemoveAll(pkg.BasePath()); err != nil {
						return data, err
					}
					data.Delete(path)
					// update the path to install the latest
					toBeInstalledVersion := pkg.Newest()
					pkg.UpdateSelectedVersion(toBeInstalledVersion)
					data.Add(pkg.FilePath(toBeInstalledVersion), []byte(pkg.URL(toBeInstalledVersion)))
				}
				break
			}
		}
	}
	// it might be that no files were found so we need to add this to the data
	for _, pkg := range r.providers.List() {
		if pkg.GetSelectedVersion() == "" {
			// remove the files from the fsys
			if err := r.writer.fsys.RemoveAll(pkg.BasePath()); err != nil {
				return data, err
			}
			toBeInstalledVersion := pkg.Newest()
			pkg.UpdateSelectedVersion(toBeInstalledVersion)
			data.Add(pkg.FilePath(toBeInstalledVersion), []byte(pkg.URL(toBeInstalledVersion)))
		}
	}

	return data, nil
}

type pkgProviderWriter struct {
	PathExists bool
	fsys       fsys.FS
	rootPath   string
}

func (r *pkgProviderWriter) Write(ctx context.Context, data *data.Data) error {
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

func (r *pkgValidator) Write(ctx context.Context, data *data.Data) error {
	providers := []string{}
	for path := range data.List() {
		providers = append(providers, path)
	}
	if len(providers) > 0 {
		return fmt.Errorf("please run init, the following providers are not up to date %v", providers)
	}
	return nil
}
