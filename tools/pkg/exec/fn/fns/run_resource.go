package fns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw-nephio/kform/kform-plugin/plugin"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewResourceFn(cfg *Config) fn.BlockInstanceRunner {
	return &resource{
		rootModuleName:    cfg.RootModuleName,
		vars:              cfg.Vars,
		providerInstances: cfg.ProviderInstances,
	}
}

type resource struct {
	rootModuleName    string
	vars              cache.Cache[vars.Variable]
	providerInstances cache.Cache[plugin.Provider]
}

func (r *resource) Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error {
	// NOTE: forEach or count expected and its respective values will be represented in localVars
	// ForEach: each.key/value
	// Count: count.index

	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(r.rootModuleName, vCtx))
	log.Info("run block instance start...")

	// 1. render the config of the resource with variable subtitution
	if vCtx.BlockContext.Config == nil {
		// Pressence of the config should be checked in the syntax validation
		return fmt.Errorf("cannot run without config for %s", vctx.GetContext(r.rootModuleName, vCtx))
	}
	if vCtx.BlockContext.Attributes != nil && vCtx.BlockContext.Attributes.Schema == nil {
		return fmt.Errorf("cannot run without a schema for %s", vctx.GetContext(r.rootModuleName, vCtx))
	}
	// adds the metaType to the config
	renderer := &Renderer{
		Vars:   r.vars,
		Schema: *vCtx.BlockContext.Attributes.Schema,
	}
	d, err := renderer.RenderConfigOrValue(ctx, vCtx.BlockName, vCtx.BlockContext.Config, localVars)
	if err != nil {
		return fmt.Errorf("cannot render config for %s", vctx.GetContext(r.rootModuleName, vCtx))
	}
	/*
		d, err = AddTypeMeta(ctx, *vCtx.BlockContext.Attributes.Schema, d)
		if err != nil {
			return fmt.Errorf("cannot add type meta for %s, err: %s", vctx.GetContext(r.rootModuleName, vCtx), err.Error())
		}
	*/
	log.Info("data raw", "req", d)

	b, err := json.Marshal(d)
	if err != nil {
		log.Error("cannot json marshal list", "error", err.Error())
		return err
	}
	log.Info("data json", "req", string(b))

	// 2. run provider
	// lookup the provider in the provider instances
	// based on the blockType run either data or resource
	// add the data in the variable
	fmt.Println("provider", vCtx.Provider)

	provider, err := r.providerInstances.Get(cache.NSN{Name: vCtx.Provider})
	if err != nil {
		log.Info("cannot get provider", "error", err.Error())
		return err
	}

	switch vCtx.BlockType {
	case types.BlockTypeData:
		resp, err := provider.ReadDataSource(ctx, &kfplugin1.ReadDataSource_Request{
			Name: strings.Split(vCtx.BlockName, ".")[0],
			Obj: b,
		})
		if err != nil {
			log.Error("cannot read resource", "error", err.Error())
			return err
		}
		if diag.Diagnostics(resp.Diagnostics).HasError() {
			log.Error("request failed", "error", diag.Diagnostics(resp.Diagnostics).Error())
			return err
		}
		b = resp.Obj
	case types.BlockTypeResource:
		resp, err := provider.CreateResource(ctx, &kfplugin1.CreateResource_Request{
			Name: strings.Split(vCtx.BlockName, ".")[0],
			Obj: b,
		})
		if err != nil {
			log.Error("cannot read resource", "error", err.Error())
			return err
		}
		if diag.Diagnostics(resp.Diagnostics).HasError() {
			log.Error("request failed", "error", diag.Diagnostics(resp.Diagnostics).Error())
			return err
		}
		b = resp.Obj
	case types.BlockTypeList:
		// TBD how do we deal with a list
		resp, err := provider.ListDataSource(ctx, &kfplugin1.ListDataSource_Request{
			Name: strings.Split(vCtx.BlockName, ".")[0],
			Obj: b,
		})
		if err != nil {
			log.Error("cannot read resource", "error", err.Error())
			return err
		}
		if diag.Diagnostics(resp.Diagnostics).HasError() {
			log.Error("request failed", "error", diag.Diagnostics(resp.Diagnostics).Error())
			return err
		}
		b = resp.Obj
	default:
		return fmt.Errorf("unexpected blockType, expected %v, got %s", types.ResourceBlockTypes, vCtx.BlockType)
	}

	if err := json.Unmarshal(b, &d); err != nil {
		log.Error("cannot unmarshal resp", "error", err.Error())
		return err
	}
	log.Info("data response", "resp", string(b))

	if err := renderer.updateVars(ctx, vCtx.BlockName, d, localVars); err != nil {
		return err
	}

	log.Info("run block instance finished...")
	return nil
}
