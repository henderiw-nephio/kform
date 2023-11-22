package parser

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

func (r *kformparser) validateModuleCalls(ctx context.Context) {
	for nsn, m := range r.modules.List() {
		// only process modules that call other modules
		if len(m.ModuleCalls.List()) > 0 {
			//fmt.Printf("module: %s\n", nsn.Name)
			mcList := m.ModuleCalls.List()
			for rmNSN, mc := range mcList {
				rm, err := r.modules.Get(rmNSN)
				if err != nil {
					r.recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in modules", nsn.Name, rmNSN.Name))
					//fmt.Printf("     remote module %s from input not found\n", mcNSN.Name)
				}
				// validate if the module call matches the input of the remote module
				for inputName := range mc.GetInputParams() {
					//fmt.Printf("  module calls: %s mc input: %s\n", mcNSN.Name, inputName)

					inputName := fmt.Sprintf("input.%s", inputName)
					if _, err := rm.Inputs.Get(cache.NSN{Name: inputName}); err != nil {
						r.recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in inputs %s", nsn.Name, rmNSN.Name, inputName))
						//fmt.Printf("    remote module %s input %s not found\n", mcNSN.Name, inputName)
					}
				}
				// validate the sourceproviders in the module call
				for targetProvider, sourceProvider := range mc.GetProviders() {
					if _, err := m.ProviderConfigs.Get(cache.NSN{Name: sourceProvider}); err != nil {
						r.recorder.Record(diag.DiagErrorf("provider module call module from %s to %s source provider %s not found", nsn.Name, rmNSN.Name, sourceProvider))
					}
					if !rm.GetProvidersFromResources(ctx).Has(cache.NSN{Name: targetProvider}) {
						r.recorder.Record(diag.DiagErrorf("provider module call module from %s to %s target provider %s not found", nsn.Name, rmNSN.Name, targetProvider))
					}
				}
			}
			// validate remote call output
			// validate mod dependency matches with the remote module output
			for modoutDep, modoutDepCtx := range m.GetModuleDependencies(ctx) {
				//fmt.Println(modoutDep, modoutDepCtx)
				split := strings.Split(modoutDep, ".")

				rmNSN := cache.NSN{Name: strings.Join([]string{split[0], split[1]}, ".")}
				if _, ok := mcList[rmNSN]; !ok {
					//fmt.Printf("     remote module %s from output not found in module calls\n", rmNSN.Name)
					r.recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in module calls fromctx: %s", nsn.Name, rmNSN.Name, modoutDepCtx))
				}
				rm, err := r.modules.Get(rmNSN)
				if err != nil {
					//fmt.Printf("     remote module %s from output not found\n", rmNSN.Name)
					r.recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in modules fromctx: %s", nsn.Name, rmNSN.Name, modoutDepCtx))
				}

				outputName := fmt.Sprintf("output.%s", split[2])
				if _, err := rm.Outputs.Get(cache.NSN{Name: outputName}); err != nil {
					//fmt.Printf("    remote module %s from output %s not found\n", rmNSN.Name, outputName)
					r.recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in outputs %s fromctx: %s", nsn.Name, rmNSN.Name, outputName, modoutDepCtx))
				}
			}
		}
	}
}

// validateProviderConfigs validates if for each provider in a child resource
// there is a provider config
func (r *kformparser) validateProviderConfigs(ctx context.Context) {
	rootModule, err := r.modules.Get(r.rootModuleName)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot validate provider configs, root module %s not found", r.rootModuleName.Name))
	}
	rootProviderConfigs := rootModule.ProviderConfigs.List()

	for cmNSN, m := range r.modules.List() {
		if m.Kind != types.ModuleKindRoot {
			for _, provider := range m.GetProvidersFromResources(ctx).UnsortedList() {
				if _, ok := rootProviderConfigs[provider]; !ok {
					r.recorder.Record(diag.DiagErrorf("no provider config in root module for child module %s, provider: %s", cmNSN.Name, provider.Name))
				}
			}
		}
	}
}

// validateProviderRequirements validates if the source strings of all the provider
// requirements are consistent
// first we walk through all the provider requirements referenced by all modules
// per provider we check the consistency of the source address
func (r *kformparser) validateProviderRequirements(ctx context.Context) {
	for providerNsn, nsnreqs := range r.getProviderRequirements(ctx) {
		// per provider we check the consistency of the source address
		source := ""
		for _, req := range nsnreqs {
			if source != "" && source != req.Source {
				r.recorder.Record(diag.DiagErrorf("inconsistent provider requirements for %s source1: %s, source2: %s", providerNsn.Name, source, req.Source))
			}
			source = req.Source
		}
	}
}

// validateProviderConfigs validates if for each provider in a child resource
// there is a provider config

func (r *kformparser) validateUnreferencedProviderConfigs(ctx context.Context) {
	unreferenceProviderConfigs := r.getUnReferencedProviderConfigs(ctx)
	if len(unreferenceProviderConfigs) > 0 {
		r.recorder.Record(diag.DiagWarnf("root module %s provider configs are unreferenced: %v", r.rootModuleName.Name, unreferenceProviderConfigs))
	}
}

func (r *kformparser) getUnReferencedProviderConfigs(ctx context.Context) []string {
	unreferenceProviderConfigs := []string{}

	rootModule, err := r.modules.Get(r.rootModuleName)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot validate provider config references, root module %s not found", r.rootModuleName.Name))
	}
	rootProviderConfigs := rootModule.ProviderConfigs.List()

	for _, m := range r.modules.List() {
		for _, provider := range m.GetProvidersFromResources(ctx).UnsortedList() {
			delete(rootProviderConfigs, provider)
			if len(rootProviderConfigs) == 0 {
				return unreferenceProviderConfigs
			}
		}
	}
	if len(rootProviderConfigs) > 0 {
		unreferenceProviderConfigs := make([]string, 0, len(rootProviderConfigs))
		for nsn := range rootProviderConfigs {
			unreferenceProviderConfigs = append(unreferenceProviderConfigs, nsn.Name)
		}
		sort.Strings(unreferenceProviderConfigs)
		return unreferenceProviderConfigs
	}
	return unreferenceProviderConfigs
}

func (r *kformparser) validateUnreferencedProviderRequirements(ctx context.Context) {
	rootModule, err := r.modules.Get(r.rootModuleName)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot validate provider requirements references, root module %s not found", r.rootModuleName.Name))
	}

	for cmNSN, m := range r.modules.List() {
		rootProviderReqs := m.GetProviderRequirements(ctx)
		for nsn := range rootModule.ProviderConfigs.List() {
			delete(rootProviderReqs, nsn)
			if len(rootProviderReqs) == 0 {
				continue
			}
		}
		if len(rootProviderReqs) > 0 {
			unreferenceProviderReqs := make([]string, 0, len(rootProviderReqs))
			for nsn := range rootProviderReqs {
				unreferenceProviderReqs = append(unreferenceProviderReqs, nsn.Name)
			}
			sort.Strings(unreferenceProviderReqs)
			r.recorder.Record(diag.DiagWarnf("%s module %s provider requirements are unreferenced: %v", m.Kind, cmNSN.Name, unreferenceProviderReqs))
		}
	}
}

func (r *kformparser) getProviderRequirements(ctx context.Context) map[cache.NSN][]kformpkgmetav1alpha1.Provider {
	rootModule, err := r.modules.Get(r.rootModuleName)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot validate provider requirements references, root module %s not found", r.rootModuleName.Name))
	}

	rootProviderConfigs := rootModule.ProviderConfigs.List()
	// delete the unreferenced provider configs from the provider configs
	unreferenceProviderConfigs := r.getUnReferencedProviderConfigs(ctx)
	for _, name := range unreferenceProviderConfigs {
		delete(rootProviderConfigs, cache.NSN{Name: name})
	}

	// we initialize all provider if they have aa req or not, if not the latest provider will be downloaded
	allprovreqs := map[cache.NSN][]kformpkgmetav1alpha1.Provider{}
	for nsn := range rootProviderConfigs {
		allprovreqs[nsn] = []kformpkgmetav1alpha1.Provider{}
	}

	for _, m := range r.modules.List() {
		provReqs := m.ProviderRequirements.List()
		for provNSN, provReq := range provReqs {
			if _, ok := rootProviderConfigs[provNSN]; ok {
				// since we initialized allprovreqs we dont need to check if the list is initialized
				allprovreqs[provNSN] = append(allprovreqs[provNSN], provReq)
			}
		}
	}
	return allprovreqs
}
