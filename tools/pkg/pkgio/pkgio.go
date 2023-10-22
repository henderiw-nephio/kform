package pkgio

import (
	"fmt"
	"io/fs"
	"sync"
)

const kformOciPkgExt = "kformpkg"

var ReadmeFileMatch = []string{"README.md"}
var IgnoreFileMatch = []string{".kformignore"}
var PkgFileMatch = []string{"KformFile.yaml"}
var MarkdownMatch = []string{"*.md"}
var YAMLMatch = []string{"*.yaml", "*.yml"}
var JSONMatch = []string{"*.json"}
var PkgMatch = []string{fmt.Sprintf("*.%s", kformOciPkgExt)}

type Reader interface {
	Read(*result) (*result, error)
}

type Writer interface {
	Write(*result) error
}

type FilterFn func(path string, d fs.DirEntry) (bool, error)

type Pipeline struct {
	Inputs  []Reader `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs []Writer `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

func (r Pipeline) Execute() error {
	result := newResult()

	// read from the inputs
	for _, i := range r.Inputs {
		var err error
		result, err = i.Read(result)
		if err != nil {
			return err
		}
	}
	// TODO filter/processor
	// write to the outputs
	for _, o := range r.Outputs {
		if err := o.Write(result); err != nil {
			return err
		}
	}
	return nil
}

type result struct {
	d map[string][]byte
	m sync.RWMutex
}

func newResult() *result {
	return &result{
		d: map[string][]byte{},
	}
}

func (r *result) add(name string, b []byte) {
	r.m.Lock()
	defer r.m.Unlock()
	r.d[name] = b
}

func (r *result) get() map[string]string {
	r.m.RLock()
	defer r.m.RUnlock()
	f := make(map[string]string, len(r.d))
	for n, b := range r.d {
		f[n] = string(b)
	}
	return f
}

/*
func (r *result) print() {
	r.m.RLock()
	defer r.m.RUnlock()
	for path, data := range r.d {
		fmt.Printf("## file: %s ##\n", path)
		fmt.Println(string(data))
	}
}
*/
