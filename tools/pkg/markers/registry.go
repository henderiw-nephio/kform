package markers

import (
	"context"
	"fmt"
	"sync"
)

// Registry keeps track of registered definitions, and allows for easy lookup.
type Registry interface {
	Register(def *Definition) error
	Lookup(name string, target TargetType) *Definition
	AddHelp(def *Definition, help *DefinitionHelp)
	Print()
}

func NewRegistry(ctx context.Context) Registry {
	return &registry{
		forPkg:   map[string]*Definition{},
		forType:  map[string]*Definition{},
		forField: map[string]*Definition{},
		helpFor:  map[*Definition]*DefinitionHelp{},
	}
}

// Registry keeps track of registered definitions, and allows for easy lookup.
// It's thread-safe.
type registry struct {
	forPkg   map[string]*Definition
	forType  map[string]*Definition
	forField map[string]*Definition
	helpFor  map[*Definition]*DefinitionHelp

	m sync.RWMutex
	//initOnce sync.Once
}

func (r *registry) Print() {
	fmt.Println("pkgs...")
	for name, d := range r.forPkg {
		fmt.Printf("  pkg: %s, def: %v\n", name, d)
	}
	fmt.Println("types...")
	for name, d := range r.forType {
		fmt.Printf("  typ: %s, def: %v\n", name, d)
	}
	fmt.Println("fields...")
	for name, d := range r.forField {
		fmt.Printf("  fld: %s, def: %v\n", name, d)
	}
	fmt.Println("helps...")
	for d, help := range r.helpFor {
		fmt.Printf("  hlp: %s, def: %v\n", d.Name, help)
	}
}

// Register registers the given marker definition with this registry for later lookup.
func (r *registry) Register(def *Definition) error {
	r.m.Lock()
	defer r.m.Unlock()

	switch def.Target {
	case DescribesPackage:
		r.forPkg[def.Name] = def
	case DescribesType:
		r.forType[def.Name] = def
	case DescribesField:
		r.forField[def.Name] = def
	default:
		return fmt.Errorf("unknown target type %v", def.Target)
	}
	return nil
}

// AddHelp stores the given help in the registry, marking it as associated with
// the given definition.
func (r *registry) AddHelp(def *Definition, help *DefinitionHelp) {
	r.m.Lock()
	defer r.m.Unlock()

	r.helpFor[def] = help
}

// Lookup fetches the definition corresponding to the given name and target type.
func (r *registry) Lookup(name string, target TargetType) *Definition {
	r.m.RLock()
	defer r.m.RUnlock()

	switch target {
	case DescribesPackage:
		return tryAnonLookup(name, r.forPkg)
	case DescribesType:
		return tryAnonLookup(name, r.forType)
	case DescribesField:
		return tryAnonLookup(name, r.forField)
	default:
		return nil
	}
}

// tryAnonLookup tries looking up the given marker as both an struct-based
// marker and an anonymous marker, returning whichever format matches first,
// preferring the longer (anonymous) name in case of conflicts.
func tryAnonLookup(name string, defs map[string]*Definition) *Definition {
	// NB(directxman12): we look up anonymous names first to work with
	// legacy style marker definitions that have a namespaced approach
	// (e.g. deepcopy-gen, which uses `+k8s:deepcopy-gen=foo,bar` *and*
	// `+k8s.io:deepcopy-gen:interfaces=foo`).
	name, anonName, _ := splitMarker(name)
	//fmt.Println("name: ", name)
	//fmt.Println("anonName: ", anonName)
	//fmt.Println("fields: ", fields)
	if def, exists := defs[anonName]; exists {
		return def
	}

	return defs[name]
}

// Must panics on errors creating definitions.
func Must(def *Definition, err error) *Definition {
	if err != nil {
		panic(err)
	}
	return def
}
