package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func ValidateModuleCalls(ctx context.Context, modules map[cache.NSN]*types.Module) error {

	recorder := cctx.GetContextValue[diag.Recorder](ctx, types.CtxKeyRecorder)
	if recorder == nil {
		return fmt.Errorf("cannot parse without a recorder")
	}
	for nsn, m := range modules {
		// only process modules that call other modules
		if len(m.ModuleCalls.List()) > 0 {
			fmt.Printf("module: %s\n", nsn.Name)

			mcList := m.ModuleCalls.List()
			for rmNSN, mc := range mcList {
				// validate if the module call matches the input of the remote module
				for inputName := range mc.GetParams() {
					//fmt.Printf("  module calls: %s mc input: %s\n", mcNSN.Name, inputName)
					rm, ok := modules[rmNSN]
					if !ok {
						recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in modules", nsn.Name, rmNSN.Name))
						//fmt.Printf("     remote module %s from input not found\n", mcNSN.Name)
					}
					inputName := fmt.Sprintf("input.%s", inputName)
					if _, err := rm.Inputs.Get(cache.NSN{Name: inputName}); err != nil {
						recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in inputs %s", nsn.Name, rmNSN.Name, inputName))
						//fmt.Printf("    remote module %s input %s not found\n", mcNSN.Name, inputName)
					}
				}
				// validate remote call output
			}
			// validate mod dependency matches with the remote module output
			for _, modOutputDep := range m.GetModuleDependencies(ctx) {
				split := strings.Split(modOutputDep, ".")

				rmNSN := cache.NSN{Name: strings.Join([]string{split[0], split[1]}, ".")}
				if _, ok := mcList[rmNSN]; !ok {
					fmt.Printf("     remote module %s from output not found in module calls\n", rmNSN.Name)
					recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in module calls", nsn.Name, rmNSN.Name))
				}
				rm, ok := modules[rmNSN]
				if !ok {
					fmt.Printf("     remote module %s from output not found\n", rmNSN.Name)
					recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in modules", nsn.Name, rmNSN.Name))
				}

				outputName := fmt.Sprintf("output.%s", split[2])
				if _, err := rm.Outputs.Get(cache.NSN{Name: outputName}); err != nil {
					fmt.Printf("    remote module %s from output %s not found\n", rmNSN.Name, outputName)
					recorder.Record(diag.DiagErrorf("module call module from %s to %s not found in outpurs %s", nsn.Name, rmNSN.Name, outputName))
				}
			}
		}
	}
	return nil
}
