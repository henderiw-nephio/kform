package markers

import (
	"context"
	"reflect"

	"github.com/henderiw-nephio/kform/tools/pkg/markers"
)

// Register registers all definitions for CRD generation to the given registry.
func Register(ctx context.Context) error {
	for _, def := range AllDefinitions {
		if err := def.Register(ctx); err != nil {
			return err
		}
	}

	return nil
}

// mustMakeAllWithPrefix converts each object into a marker definition using
// the object's type's with the prefix to form the marker name.
func mustMakeAllWithPrefix(prefix string, target markers.TargetType, objs ...interface{}) []*definitionWithHelp {
	defs := make([]*definitionWithHelp, len(objs))
	for i, obj := range objs {
		name := prefix + ":" + reflect.TypeOf(obj).Name()
		def, err := markers.MakeDefinition(name, target, obj)
		if err != nil {
			panic(err)
		}
		defs[i] = &definitionWithHelp{Definition: def, Help: obj.(hasHelp).Help()}
	}

	return defs
}
