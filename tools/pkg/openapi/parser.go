package openapi

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
	"github.com/henderiw-nephio/kform/tools/pkg/markers"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

type Parser interface {
	// NeedPackage indicates that types and type-checking information
	// is needed for the given package.
	NeedPackage(pkg *loader.Package)
	NeedOpenAPIFor(typeIdent TypeIdent)
	FindProviderAPIs() []TypeIdent
	PrintSchemata()
}

func (r *parser) PrintSchemata() {
	for typeident, schema := range r.Schemata {
		b, _ := yaml.Marshal(schema)
		fmt.Println(typeident)
		fmt.Println(string(b))
	}
}

func NewParser(ctx context.Context) Parser {
	allowDangerousTypes := cctx.GetContextValue[bool](ctx, "allowDangerousTypes")
	ignoreUnexportedFields := cctx.GetContextValue[bool](ctx, "ignoreUnexportedFields")
	generateEmbeddedObjectMeta := cctx.GetContextValue[bool](ctx, "generateEmbeddedObjectMeta")
	collector := cctx.GetContextValue[markers.Collector](ctx, "collector")
	if collector == nil {
		collector = markers.NewCollector(ctx)
	}
	checker := cctx.GetContextValue[loader.TypeChecker](ctx, "checker")
	if checker == nil {
		checker = loader.NewTypeChecker()
	}

	p := &parser{
		Collector:                  collector,
		Checker:                    checker,
		Types:                      map[TypeIdent]*markers.TypeInfo{},
		Schemata:                   map[TypeIdent]apiext.JSONSchemaProps{},
		FlattenedSchemata:          map[TypeIdent]apiext.JSONSchemaProps{},
		PackageOverrides:           map[string]PackageOverride{},
		packages:                   map[*loader.Package]struct{}{},
		AllowDangerousTypes:        allowDangerousTypes,
		IgnoreUnexportedFields:     ignoreUnexportedFields,
		GenerateEmbeddedObjectMeta: generateEmbeddedObjectMeta,
	}
	p.flattener = newFlattener(p)
	return p
}

type parser struct {
	Collector markers.Collector
	// checker stores persistent partial type-checking/reference-traversal information.
	Checker loader.TypeChecker
	// flattens the schema
	flattener Flattener

	// Types contains the known TypeInfo for this parser.
	Types map[TypeIdent]*markers.TypeInfo
	// Schemata contains the known OpenAPI JSONSchemata for this parser.
	Schemata map[TypeIdent]apiext.JSONSchemaProps
	// FlattenedSchemata contains fully flattened schemata for validation.
	// Each schema has been flattened by the flattener,
	// and then embedded fields have been flattened with FlattenEmbedded.
	FlattenedSchemata map[TypeIdent]apiext.JSONSchemaProps

	// PackageOverrides indicates that the loading of any package with
	// the given path should be handled by the given overrider.
	PackageOverrides map[string]PackageOverride

	// packages marks packages as loaded, to avoid re-loading them.
	packages map[*loader.Package]struct{}

	// AllowDangerousTypes controls the handling of non-recommended types such as float. If
	// false (the default), these types are not supported.
	// There is a continuum here:
	//    1. Types that are always supported.
	//    2. Types that are allowed by default, but not recommended (warning emitted when they are encountered as per PR #443).
	//       Possibly they are allowed by default for historical reasons and may even be "on their way out" at some point in the future.
	//    3. Types that are not allowed by default, not recommended, but there are some legitimate reasons to need them in certain corner cases.
	//       Possibly these types should also emit a warning as per PR #443 even when they are "switched on" (an integration point between
	//       this feature and #443 if desired). This is the category that this flag deals with.
	//    4. Types that are not allowed and will not be allowed, possibly because it just "doesn't make sense" or possibly
	//       because the implementation is too difficult/clunky to promote them to category 3.
	// TODO: Should we have a more formal mechanism for putting "type patterns" in each of the above categories?
	AllowDangerousTypes bool

	// IgnoreUnexportedFields specifies if unexported fields on the struct should be skipped
	IgnoreUnexportedFields bool

	// GenerateEmbeddedObjectMeta specifies if any embedded ObjectMeta should be generated
	GenerateEmbeddedObjectMeta bool
}

// PackageOverride overrides the loading of some package
// (potentially setting custom schemata, etc).  It must
// call AddPackage if it wants to continue with the default
// loading behavior.
type PackageOverride func(p *parser, pkg *loader.Package)

// AddPackage indicates that types and type-checking information is needed
// for the the given package, *ignoring* overrides.
// Generally, consumers should call NeedPackage, while PackageOverrides should
// call AddPackage to continue with the normal loading procedure.
func (r *parser) AddPackage(pkg *loader.Package) {
	if _, checked := r.packages[pkg]; checked {
		return
	}
	r.indexTypes(pkg)
	r.Checker.Check(pkg)
	r.packages[pkg] = struct{}{}
}

// NeedPackage indicates that types and type-checking information
// is needed for the given package.
func (r *parser) NeedPackage(pkg *loader.Package) {
	if _, checked := r.packages[pkg]; checked {
		return
	}
	// overrides are going to be written without vendor.  This is why we index by the actual
	// object when we can.
	if override, overridden := r.PackageOverrides[loader.NonVendorPath(pkg.PkgPath)]; overridden {
		override(r, pkg)
		r.packages[pkg] = struct{}{}
		return
	}
	r.AddPackage(pkg)

}

// indexTypes loads all types in the package into Types.
func (r *parser) indexTypes(pkg *loader.Package) {
	// autodetect
	pkgMarkers, err := markers.PackageMarkers(r.Collector, pkg)
	if err != nil {
		pkg.AddError(err)
	} else {
		if skipPkg := pkgMarkers.Get("kubebuilder:skip"); skipPkg != nil {
			return
		}
	}

	fmt.Println(pkgMarkers)

	if err := markers.EachType(r.Collector, pkg, func(info *markers.TypeInfo) {
		ident := TypeIdent{
			Package: pkg,
			Name:    info.Name,
		}

		r.Types[ident] = info
	}); err != nil {
		pkg.AddError(err)
	}
}

func (r *parser) FindProviderAPIs() []TypeIdent {
	// keeps track of unique typeident
	tis := map[TypeIdent]struct{}{}
	for typeIdent := range r.Types {
		if typeIdent.Name == "ProviderAPI" {
			tis[typeIdent] = struct{}{}
		}
	}

	typeIdents := make([]TypeIdent, 0, len(tis))
	for ti := range tis {
		typeIdents = append(typeIdents, ti)
	}
	return typeIdents
}

func (r *parser) NeedOpenAPIFor(typeIdent TypeIdent) {
	fmt.Println(typeIdent.Name, typeIdent.Package.ID)

	r.NeedFlattenedSchemaFor(typeIdent)
}

func (r *parser) NeedFlattenedSchemaFor(typ TypeIdent) {
	if _, knownSchema := r.FlattenedSchemata[typ]; knownSchema {
		return
	}

	r.NeedSchemaFor(typ)
	partialFlattened := r.flattener.FlattenType(typ)
	fullyFlattened := FlattenEmbedded(partialFlattened, typ.Package)

	r.FlattenedSchemata[typ] = *fullyFlattened
}

// NeedSchemaFor indicates that a schema should be generated for the given type.
func (r *parser) NeedSchemaFor(typ TypeIdent) {
	r.NeedPackage(typ.Package)
	if _, knownSchema := r.Schemata[typ]; knownSchema {
		return
	}

	info, knownInfo := r.Types[typ]
	if !knownInfo {
		typ.Package.AddError(fmt.Errorf("unknown type %s", typ))
		return
	}

	// avoid tripping recursive schemata, like ManagedFields, by adding an empty WIP schema
	r.Schemata[typ] = apiext.JSONSchemaProps{}

	schemaCtx := newSchemaContext(typ.Package, r, r.AllowDangerousTypes, r.IgnoreUnexportedFields)
	ctxForInfo := schemaCtx.ForInfo(info)

	pkgMarkers, err := markers.PackageMarkers(r.Collector, typ.Package)
	if err != nil {
		typ.Package.AddError(err)
	}
	ctxForInfo.PackageMarkers = pkgMarkers

	schema := infoToSchema(ctxForInfo)

	r.Schemata[typ] = *schema
}
