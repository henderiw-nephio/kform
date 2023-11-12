package parser

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
	koe "github.com/nephio-project/nephio/krm-functions/lib/kubeobject"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

// TODO need to enhance this to differentiate running incluster versus out of cluster
func (r *moduleparser) getKforms(ctx context.Context) (*kformpkgmetav1alpha1.KformFile, map[string]*types.Kform, error) {
	log := log.FromContext(ctx)
	var kfile *kformpkgmetav1alpha1.KformFile
	kforms := map[string]*types.Kform{}
	// read the directory
	/*
		if err := fsys.ValidateDirPath(r.path); err != nil {
			return kfile, kforms, fmt.Errorf("cannot parse module dir path invalid, err: %s", err.Error())
		}
	*/
	// check the path from the current directory
	if _, err := r.fsys.Stat("."); err != nil {
		return kfile, kforms, fmt.Errorf("cannot parse module dir does not exist, err: %s", err.Error())
	}
	ignoreRules := ignore.Empty(pkgio.IgnoreFileMatch[0])
	f, err := r.fsys.Open(pkgio.IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}
	//defer f.Close()
	reader := pkgio.PkgReader{
		PathExists:     true,
		Fsys:           r.fsys,
		MatchFilesGlob: pkgio.YAMLMatch,
		IgnoreRules:    ignoreRules,
		SkipDir:        true,
	}
	d, err := reader.Read(ctx, pkgio.NewData())
	if err != nil {
		return kfile, kforms, err
	}
	// extracts kforms from the configmaps
	for path, data := range d.List() {
		ko, err := fn.ParseKubeObject([]byte(data))
		if err != nil {
			log.Error("kubeObject parsing failed", "path", filepath.Join(r.path, path), "err", err.Error())
			continue
		}
		if ko.GetKind() == reflect.TypeOf(corev1.ConfigMap{}).Name() {
			kform, _, err := ko.NestedSubObject("data")
			if err != nil {
				return kfile, kforms, fmt.Errorf("data not present in configmap file: %s", path)
			}
			kf := types.Kform{}
			if err := yaml.Unmarshal([]byte(kform.String()), &kf); err != nil {
				return kfile, kforms, fmt.Errorf("unmarshal error kform in file: %s", path)
			}
			kforms[path] = &kf
			log.Debug("kform", "path", path, "kform", kf.Blocks)
		}
		if ko.GetKind() == reflect.TypeOf(kformpkgmetav1alpha1.KformFile{}).Name() {
			if kfile != nil {
				return kfile, kforms, fmt.Errorf("cannot have 2 kform file resource in the package")
			}
			kfKOE, err := koe.NewFromKubeObject[kformpkgmetav1alpha1.KformFile](ko)
			if err != nil {
				return kfile, kforms, err
			}
			kfile, err = kfKOE.GetGoStruct()
			if err != nil {
				return kfile, kforms, err
			}
		}
	}
	return kfile, kforms, nil
}
