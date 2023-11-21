package pkgio

import (
	"context"
)

const kformOciPkgExt = "kformpkg"

var ReadmeFileMatch = []string{"README.md"}
var IgnoreFileMatch = []string{".kformignore"}
var PkgFileMatch = []string{"KformFile.yaml"}
var MarkdownMatch = []string{"*.md"}
var YAMLMatch = []string{"*.yaml", "*.yml"}
var JSONMatch = []string{"*.json"}
var MatchAll = []string{"*"}

//var PkgMatch = []string{fmt.Sprintf("*.%s", kformOciPkgExt)}

type Reader interface {
	Read(context.Context, *Data) (*Data, error)
}

type Writer interface {
	Write(context.Context, *Data) error
}

type Process interface {
	Process(context.Context, *Data) (*Data, error)
}

type Pipeline struct {
	Inputs     []Reader  `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Processors []Process `json:"processors,omitempty" yaml:"processors,omitempty"`
	Outputs    []Writer  `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

func (r Pipeline) Execute(ctx context.Context) error {
	data := NewData()
	var err error
	// read from the inputs
	for _, i := range r.Inputs {
		data, err = i.Read(ctx, data)
		if err != nil {
			return err
		}
	}
	//data.Print()
	for _, p := range r.Processors {
		data, err = p.Process(ctx, data)
		if err != nil {
			return err
		}
	}
	//data.Print()
	// write to the outputs
	for _, o := range r.Outputs {
		if err := o.Write(ctx, data); err != nil {
			return err
		}
	}
	return nil
}
