package pkgio

import (
	"fmt"
	"io/fs"
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
	Read(*Data) (*Data, error)
}

type Writer interface {
	Write(*Data) error
}

type FilterFn func(path string, d fs.DirEntry) (bool, error)

type Pipeline struct {
	Inputs  []Reader `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs []Writer `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

func (r Pipeline) Execute() error {
	data := NewData()

	// read from the inputs
	for _, i := range r.Inputs {
		var err error
		data, err = i.Read(data)
		if err != nil {
			return err
		}
	}
	// TODO filter/processor
	// write to the outputs
	for _, o := range r.Outputs {
		if err := o.Write(data); err != nil {
			return err
		}
	}
	return nil
}
