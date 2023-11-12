package pkgio

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/xlab/treeprint"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
)

type PkgTreeReadWriter interface {
	Reader
	Writer
}

func NewPkgTreeReadWriter(path string) PkgTreeReadWriter {

	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(IgnoreFileMatch[0])
	f, err := fs.Open(IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}

	return &pkgTreeReadWriter{
		reader: &PkgReader{
			PathExists:     true,
			Fsys:           fsys.NewDiskFS(path),
			MatchFilesGlob: YAMLMatch,
			IgnoreRules:    ignoreRules,
		},
		writer: &pkgTreeWriter{
			fsys:     fs,
			writer:   os.Stdout,
			rootPath: path,
		},
	}
}

type pkgTreeReadWriter struct {
	reader *PkgReader
	writer *pkgTreeWriter
}

func (r *pkgTreeReadWriter) Read(ctx context.Context, data *Data) (*Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgTreeReadWriter) Write(ctx context.Context, data *Data) error {
	return r.writer.Write(ctx, data)
}

const (
	// TreeStructurePackage configures TreeWriter to generate the tree structure off of the
	// packages.
	//TreeStructurePackage TreeStructure = "directory"
	// %q holds the package name
	PkgNameFormat = "Package %q"
)

type pkgTreeWriter struct {
	writer   io.Writer
	fsys     fsys.FS
	rootPath string
}

func (r *pkgTreeWriter) Write(ctx context.Context, data *Data) error {
	indexByPkgDir := r.index(data)

	// create the new tree
	tree := treeprint.New()
	// add each package to the tree
	treeIndex := map[string]treeprint.Tree{}
	keys := r.sort(indexByPkgDir)
	for _, pkg := range keys {
		// create a branch for this package -- search for the parent package and create
		// the branch under it -- requires that the keys are sorted
		branch := tree
		for parent, subTree := range treeIndex {
			if strings.HasPrefix(pkg, parent) {
				// found a package whose path is a prefix to our own, use this
				// package if a closer one isn't found
				branch = subTree
				// don't break, continue searching for more closely related ancestors
			}
		}
		// create a new branch for the package
		createOk := pkg != "." // special edge case logic for tree on current working dir
		if createOk {
			branch = branch.AddBranch(branchName(r.fsys, pkg))
		}

		// cache the branch for this package
		treeIndex[pkg] = branch

		// print each resource in the package
		for i := range indexByPkgDir[pkg] {
			var err error
			if _, err = r.doResource(indexByPkgDir[pkg][i], "", branch); err != nil {
				return err
			}
		}
	}
	_, err := r.fsys.Stat(kformpkgmetav1alpha1.KformFileName)
	if !os.IsNotExist(err) {
		// if Kptfile exists in the root directory, it is a kpt package
		// print only package name and not entire path
		tree.SetValue(fmt.Sprintf(PkgNameFormat, filepath.Base(r.rootPath)))
	} else {
		// else it is just a directory, so print only directory name
		tree.SetValue(filepath.Base(r.rootPath))
	}

	out := tree.String()
	_, err = io.WriteString(r.writer, out)
	return err
}

// branchName takes the root directory and relative path to the directory
// and returns the branch name
func branchName(fs fsys.FS, dirRelPath string) string {
	name := filepath.Base(dirRelPath)
	_, err := fs.Stat(filepath.Join(dirRelPath, kformpkgmetav1alpha1.KformFileName))
	if !os.IsNotExist(err) {
		// add Package prefix indicating that it is a separate package as it has
		// KFormFile
		return fmt.Sprintf(PkgNameFormat, name)
	}
	return name
}

// index indexes the Resources by their package
func (p pkgTreeWriter) index(data *Data) map[string][]*fn.KubeObject {
	indexByPkgDir := map[string][]*fn.KubeObject{}
	for path, data := range data.List() {
		ko, err := fn.ParseKubeObject([]byte(data))
		if err != nil {
			continue
		}
		ko.SetAnnotation(kioutil.PathAnnotation, path)
		// TODO check uniqueness
		indexByPkgDir[filepath.Dir(path)] = append(indexByPkgDir[filepath.Dir(path)], ko)
	}
	return indexByPkgDir
}

// sort sorts the Resources in the index in display order and returns the ordered
// keys for the index
//
// Packages are sorted by sub dir name
// Resources within a package are sorted by: [namespace, name, kind, apiVersion]
func (r pkgTreeWriter) sort(indexByPackage map[string][]*fn.KubeObject) []string {
	var keys []string
	for k := range indexByPackage {
		kubeobjects := indexByPackage[k]
		// sort the nodes for each sub dir
		sort.Slice(kubeobjects, func(i, j int) bool { return compareNodes(kubeobjects[i], kubeobjects[j]) })
		keys = append(keys, k)
	}

	// return the package names sorted lexicographically
	sort.Strings(keys)
	return keys
}

func (r pkgTreeWriter) doResource(ko *fn.KubeObject, metaString string, branch treeprint.Tree) (treeprint.Tree, error) {

	value := fmt.Sprintf("%s %s", ko.GetKind(), ko.GetName())
	if len(ko.GetNamespace()) > 0 {
		value = fmt.Sprintf("%s %s/%s", ko.GetKind(), ko.GetNamespace(), ko.GetName())
	}

	if metaString == "" {
		path := ko.GetAnnotation(kioutil.PathAnnotation)
		path = filepath.Base(path)
		metaString = path
	}
	n := branch.AddMetaBranch(metaString, value)
	return n, nil
}

func compareNodes(i, j *fn.KubeObject) bool {
	// compare namespace
	if i.GetNamespace() != j.GetNamespace() {
		return i.GetNamespace() < j.GetNamespace()
	}

	// compare name
	if i.GetName() != j.GetName() {
		return i.GetName() < j.GetName()
	}

	// compare kind
	if i.GetKind() != j.GetKind() {
		return i.GetKind() < j.GetKind()
	}

	// compare apiVersion
	if i.GetAPIVersion() != j.GetAPIVersion() {
		return i.GetAPIVersion() < j.GetAPIVersion()
	}
	return true
}
