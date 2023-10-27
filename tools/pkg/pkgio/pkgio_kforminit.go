package pkgio

import (
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
)

type PkgKformInitReader interface {
	Reader
}

func NewPkgKformInitReadWriter(path string) PkgKformInitReader {
	// TBD do we add validation here
	fs := fsys.NewDiskFS(path)
	// Ignore file processing should be done here
	ignoreRules := ignore.Empty(IgnoreFileMatch[0])
	f, err := fs.Open(IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}
	defer f.Close()
	return &pkgKformInitReader{
		reader: &PkgReader{
			Fsys:           fs,
			MatchFilesGlob: YAMLMatch,
			IgnoreRules:    ignoreRules,
			SkipDir:        true,
		},
	}
}

type pkgKformInitReader struct {
	reader *PkgReader
}

func (r *pkgKformInitReader) Read(data *Data) (*Data, error) {
	return r.reader.Read(data)
}
