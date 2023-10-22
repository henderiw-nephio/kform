package init

import (
	"context"
	"fmt"
	"reflect"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	docs "github.com/henderiw-nephio/kform/internal/docs/generated/initdocs"
	"github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax"
	"github.com/henderiw/logger/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "init DIR [flags]",
		Args:    cobra.MaximumNArgs(1),
		Short:   docs.InitShort,
		Long:    docs.InitShort + "\n" + docs.InitLong,
		Example: docs.InitExamples,
		RunE:    r.runE,
	}

	r.Command = cmd
	/*
		r.Command.Flags().StringVar(
			&r.FnConfigDir, "fn-config-dir", "", "dir where the function config files are located")
	*/
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command  *cobra.Command
	rootPath string
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	log := log.FromContext(c.Context())
	r.rootPath = args[0]
	if err := fsys.ValidateDirPath(r.rootPath); err != nil {
		return err
	}
	fs := fsys.NewDiskFS(".")
	f, err := fs.Stat(r.rootPath)
	if err != nil {
		fs.MkdirAll(r.rootPath)
	} else if !f.IsDir() {
		return fmt.Errorf("cannot initialize a pkg on a file, please provide a directory instead, file: %s", r.rootPath)
	}

	reader := pkgio.NewPkgKformInitReadWriter(r.rootPath)
	d, err := reader.Read(pkgio.NewData())
	if err != nil {
		return err
	}

	// extracts kforms from the configmaps
	kforms := map[string]*v1alpha1.K8sForm{}
	for path, data := range d.Get() {
		ko, err := fn.ParseKubeObject([]byte(data))
		if err != nil {
			continue
		}
		if ko.GetKind() == reflect.TypeOf(corev1.ConfigMap{}).Name() {
			kform, _, err := ko.NestedSubObject("data")
			if err != nil {
				continue
			}
			//fmt.Println(kform.String())
			kf := v1alpha1.K8sForm{}
			if err := yaml.Unmarshal([]byte(kform.String()), &kf); err != nil {
				continue
			}
			kforms[path] = &kf
		}
	}

	for path, kf := range kforms {
		fmt.Println("------------------")
		fmt.Println("filePath:", path)
		for _, b := range kf.Blocks {
			for blockType, b := range b.NestedBlock {
				for blockidentifier := range b.NestedBlock {
					fmt.Printf("blockType: %s, blockIdentifier: %s\n", blockType, blockidentifier)
				}
			}
		}
		//fmt.Println(kf)
	}
	ctx := c.Context()
	p := syntax.NewParser(ctx, kforms)
	execCfg, diags := p.Parse(ctx)
	if diags.Error() != nil {
		return err
	}
	fmt.Println(diags.Error())

	log.Debug("diags...")
	for _, diag := range diags {
		fmt.Printf("  %v\n", diag)
	}
	log.Debug("providers...")
	for name, provider := range execCfg.GetProviders().GetVertices() {
		fmt.Printf("  name: %v, provider: %v\n", name, provider)
	}
	log.Debug("vars...")
	for name, v := range execCfg.GetVars().GetVertices() {
		fmt.Printf("  name: %v\n", name)
		if len(v.Attributes) != 0 {
			fmt.Printf("    attributes: %v\n", v.Attributes)
		}
		if v.Provider != "" {
			fmt.Printf("    provider: %v\n", v.Provider)
			fmt.Printf("    gvk: %s\n", v.GVK.String())
		}
		if len(v.Dependencies) > 0 {
			fmt.Printf("    dep: %v\n", v.Dependencies)
		}

		down := execCfg.GetVars().GetDownVertexes(name)
		if len(down) != 0 {
			fmt.Printf("    down: %v\n", down)
		}
		up := execCfg.GetVars().GetUpVertexes(name)
		if len(up) != 0 {
			fmt.Printf("    up: %v\n", up)
		}
	}

	// init and/or restore backend
	// syntax check config -> build the dag but dont use it
	// download module
	// download providers
	// -> lock files

	return nil
}
