package kformpkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	ko "github.com/nephio-project/nephio/krm-functions/lib/kubeobject"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func New(ctx context.Context, fsys fs.FS, pkgKind kformpkgmetav1alpha1.PkgKind) (*Pkg, error) {
	pkgPath := cctx.GetContextValue[string](ctx, CtxkeyPath)
	if pkgPath == "" {
		return nil, ErrNoPkgPath
	}
	r := &Pkg{
		fsys:          fsys,
		kformFile:     nil,
		pkgPath:       pkgPath,
		parentPkgPath: filepath.Dir(pkgPath),
		pkgName:       filepath.Base(pkgPath),
		pkgKind:       pkgKind,
	}
	return r, nil
}

type Pkg struct {
	fsys          fs.FS
	kformFile     *kformpkgmetav1alpha1.KformFile
	pkgPath       string
	parentPkgPath string
	pkgName       string
	pkgKind       kformpkgmetav1alpha1.PkgKind
}

func (r *Pkg) GetPkgName() string {
	return r.pkgName
}

// Kformfile returns the Kformfile meta resource by lazy loading it from the filesytem.
// A nil value represents an implicit package.
func (r *Pkg) Kformfile() (*kformpkgmetav1alpha1.KformFile, error) {
	if r.kformFile == nil {
		kf, err := r.ReadKformfile()
		if err != nil {
			return nil, err
		}
		r.kformFile = kf
	}
	return r.kformFile, nil
}

func (r *Pkg) ReadKformfile() (*kformpkgmetav1alpha1.KformFile, error) {
	f, err := r.fsys.Open(kformpkgmetav1alpha1.KformFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	fnko, err := fn.ParseKubeObject(b)
	if err != nil {
		return nil, err
	}
	kfFileKOE, err := ko.NewFromKubeObject[kformpkgmetav1alpha1.KformFile](fnko)
	if err != nil {
		return nil, err
	}
	return kfFileKOE.GetGoStruct()
}

func (r *Pkg) WriteKformfile() error {
	if _, err := os.Stat(filepath.Join(r.pkgPath, kformpkgmetav1alpha1.KformFileName)); err != nil {
		// file does not exist
		kf := kformpkgmetav1alpha1.BuildKptFile(
			metav1.ObjectMeta{Name: r.pkgName},
			kformpkgmetav1alpha1.KformFileSpec{
				Kind: r.pkgKind,
			},
		)
		koe, err := ko.NewFromGoStruct(kf)
		if err != nil {
			return err
		}
		return WriteFile(r.pkgPath, kformpkgmetav1alpha1.KformFileName, []byte(koe.String()))

	}
	return nil
}

func (r *Pkg) WriteReadmeFile() error {
	if _, err := os.Stat(filepath.Join(r.pkgPath, ReadmeFile)); err != nil {
		buff := &bytes.Buffer{}
		t, err := template.New("readme").Parse(readmeTemplate)
		if err != nil {
			return err
		}
		readmeTemplateData := map[string]string{
			"Name":        r.pkgName,
			"Description": r.pkgName,
			"PkgPath":     r.pkgPath,
		}
		err = t.Execute(buff, readmeTemplateData)
		if err != nil {
			return err
		}

		// Replace single quotes with backticks.
		b := strings.ReplaceAll(buff.String(), "'", "`")

		return WriteFile(r.pkgPath, ReadmeFile, []byte(b))
	}
	return nil
}

func (r *Pkg) WriteIgnoreFile() error {
	if _, err := os.Stat(filepath.Join(r.pkgPath, IgnoreFile)); err != nil {
		return WriteFile(r.pkgPath, IgnoreFile, nil)
	}
	return nil
}

func WriteFile(filePath, fileName string, b []byte) error {
	fmt.Println("creating", fileName)
	return func() error {
		f, err := os.Create(filepath.Join(filePath, fileName))
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write([]byte(b)); err != nil {
			return err
		}
		return nil
	}()
}

// readmeTemplate is the content for the automatically generated README.md file.
// It uses ' instead of ` since golang doesn't allow using ` in a raw string
// literal. We do a replace on the content before printing.
var readmeTemplate = `# {{.Name}}

## Description
{{.Description}}

## Usage

### View package content
'kform pkg tree {{.PkgPath}}'
Details: https://kform.dev/reference/cli/pkg/tree/

`
