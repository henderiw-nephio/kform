package fns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw-nephio/kform/kform-plugin/plugin"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewProviderFn(cfg *Config) fn.BlockInstanceRunner {
	return &provider{
		rootModuleName:    cfg.RootModuleName,
		vars:              cfg.Vars,
		recorder:          cfg.Recorder,
		providerInventory: cfg.ProviderInventory,
		providerInstances: cfg.ProviderInstances,
	}
}

type provider struct {
	// initialized from the vertexContext
	rootModuleName string
	// dynamic injection required
	vars              cache.Cache[vars.Variable]
	recorder          recorder.Recorder[record.Record]
	providerInventory cache.Cache[types.Provider]
	providerInstances cache.Cache[plugin.Provider]
}

func (r *provider) Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(r.rootModuleName, vCtx))
	log.Info("run provider")

	// render the config
	renderer := &Renderer{Vars: cache.New[vars.Variable]()}
	d, err := renderer.RenderConfig(ctx, vCtx.BlockName, vCtx.BlockContext.Config, localVars)
	if err != nil {
		return err
	}
	if vCtx.BlockContext.Attributes != nil && vCtx.BlockContext.Attributes.Schema == nil {
		return fmt.Errorf("cannot add type meta without a schema for %s", vctx.GetContext(r.rootModuleName, vCtx))
	}
	d, err = AddTypeMeta(ctx, *vCtx.BlockContext.Attributes.Schema, d)
	if err != nil {
		return fmt.Errorf("cannot add type meta for %s, err: %s", vctx.GetContext(r.rootModuleName, vCtx), err.Error())
	}
	providerConfigByte, err := json.Marshal(d)
	if err != nil {
		log.Error("cannot json marshal config", "error", err.Error())
		return err
	}
	log.Info("providerConfig", "config", string(providerConfigByte))

	// initialize the provider
	p, err := r.providerInventory.Get(cache.NSN{Name: vCtx.BlockName})
	if err != nil {
		log.Error("provider not found in inventory", "err", err)
		return fmt.Errorf("provider %s not found in inventory err: %s", vctx.GetContext(r.rootModuleName, vCtx), err.Error())
	}
	provider, err := p.Initializer()
	if err != nil {
		return err
	}
	// add the provide client to the cache - so we know what to delete
	r.providerInstances.Upsert(ctx, cache.NSN{Name: vCtx.BlockName}, provider)

	// configure the provider
	cfgresp, err := provider.Configure(ctx, &kfplugin1.Configure_Request{
		Config: providerConfigByte,
	})
	if err != nil {
		log.Error("failed to configure provider", "error", err.Error())
		return fmt.Errorf("provider %s not found in inventory err: %s", vctx.GetContext(r.rootModuleName, vCtx), err.Error())
	}
	if diag.Diagnostics(cfgresp.Diagnostics).HasError() {
		log.Error("failed to configure provider", "error", err.Error())
		return fmt.Errorf("provider %s not found in inventory err: %s", vctx.GetContext(r.rootModuleName, vCtx), err.Error())
	}
	log.Info("configure response", "diag", cfgresp.Diagnostics)

	return nil
}
