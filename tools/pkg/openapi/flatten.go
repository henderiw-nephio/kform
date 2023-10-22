package openapi

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type Flattener interface {
	FlattenType(typ TypeIdent) *apiext.JSONSchemaProps
}

func newFlattener(p *parser) Flattener {
	return &flattener{
		Parser:          p,
		LookupReference: identFromRef,
		flattenedTypes:  map[TypeIdent]apiext.JSONSchemaProps{},
	}
}

// Flattener knows how to take a root type, and flatten all references in it
// into a single, flat type.  Flattened types are cached, so it's relatively
// cheap to make repeated calls with the same type.
type flattener struct {
	// Parser is used to lookup package and type details, and parse in new packages.
	Parser *parser

	LookupReference func(ref string, contextPkg *loader.Package) (TypeIdent, error)

	// flattenedTypes hold the flattened version of each seen type for later reuse.
	flattenedTypes map[TypeIdent]apiext.JSONSchemaProps
}

// cacheType saves the flattened version of the given type for later reuse
func (r *flattener) cacheType(typ TypeIdent, schema apiext.JSONSchemaProps) {
	r.flattenedTypes[typ] = schema
}

// loadUnflattenedSchema fetches a fresh, unflattened schema from the parser.
func (r *flattener) loadUnflattenedSchema(typ TypeIdent) (*apiext.JSONSchemaProps, error) {
	r.Parser.NeedSchemaFor(typ)

	baseSchema, found := r.Parser.Schemata[typ]
	if !found {
		return nil, fmt.Errorf("unable to locate schema for type %s", typ)
	}
	return &baseSchema, nil
}

// FlattenType flattens the given pre-loaded type, removing any references from it.
// It deep-copies the schema first, so it won't affect the parser's version of the schema.
func (r *flattener) FlattenType(typ TypeIdent) *apiext.JSONSchemaProps {
	if cachedSchema, isCached := r.flattenedTypes[typ]; isCached {
		return &cachedSchema
	}
	baseSchema, err := r.loadUnflattenedSchema(typ)
	if err != nil {
		typ.Package.AddError(err)
		return nil
	}
	resSchema := r.FlattenSchema(*baseSchema, typ.Package)
	r.cacheType(typ, *resSchema)
	return resSchema
}

// FlattenSchema flattens the given schema, removing any references.
// It deep-copies the schema first, so the input schema won't be affected.
func (r *flattener) FlattenSchema(baseSchema apiext.JSONSchemaProps, currentPackage *loader.Package) *apiext.JSONSchemaProps {
	resSchema := baseSchema.DeepCopy()
	EditSchema(resSchema, &flattenVisitor{
		flattener:      r,
		currentPackage: currentPackage,
	})

	return resSchema
}

// ErrorRecorder knows how to record errors.  It wraps the part of
// pkg/loader.Package that we need to record errors in places were it might not
// make sense to have a loader.Package
type ErrorRecorder interface {
	// AddError records that the given error occurred.
	// See the documentation on loader.Package.AddError for more information.
	AddError(error)
}

// isOrNil checks if val is nil if val is of a nillable type, otherwise,
// it compares val to valInt (which should probably be the zero value).
func isOrNil(val reflect.Value, valInt interface{}, zeroInt interface{}) bool {
	switch valKind := val.Kind(); valKind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return valInt == zeroInt
	}
}

// flattenAllOfInto copies properties from src to dst, then copies the properties
// of each item in src's allOf to dst's properties as well.
func flattenAllOfInto(dst *apiext.JSONSchemaProps, src apiext.JSONSchemaProps, errRec ErrorRecorder) {
	if len(src.AllOf) > 0 {
		for _, embedded := range src.AllOf {
			flattenAllOfInto(dst, embedded, errRec)
		}
	}

	dstVal := reflect.Indirect(reflect.ValueOf(dst))
	srcVal := reflect.ValueOf(src)
	typ := dstVal.Type()

	srcRemainder := apiext.JSONSchemaProps{}
	srcRemVal := reflect.Indirect(reflect.ValueOf(&srcRemainder))
	dstRemainder := apiext.JSONSchemaProps{}
	dstRemVal := reflect.Indirect(reflect.ValueOf(&dstRemainder))
	hoisted := false

	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := typ.Field(i).Name
		switch fieldName {
		case "AllOf":
			// don't merge because we deal with it above
			continue
		case "Title", "Description", "Example", "ExternalDocs":
			// don't merge because we pre-merge to properly preserve field docs
			continue
		}
		srcField := srcVal.Field(i)
		fldTyp := srcField.Type()
		zeroVal := reflect.Zero(fldTyp)
		zeroInt := zeroVal.Interface()
		srcInt := srcField.Interface()

		if isOrNil(srcField, srcInt, zeroInt) {
			// nothing to copy from src, continue
			continue
		}

		dstField := dstVal.Field(i)
		dstInt := dstField.Interface()
		if isOrNil(dstField, dstInt, zeroInt) {
			// dst is empty, continue
			dstField.Set(srcField)
			continue
		}

		if fldTyp.Comparable() && srcInt == dstInt {
			// same value, continue
			continue
		}

		// resolve conflict
		switch fieldName {
		case "Properties":
			// merge if possible, use all of otherwise
			srcMap := srcInt.(map[string]apiext.JSONSchemaProps)
			dstMap := dstInt.(map[string]apiext.JSONSchemaProps)

			for k, v := range srcMap {
				dstProp, exists := dstMap[k]
				if !exists {
					dstMap[k] = v
					continue
				}
				flattenAllOfInto(&dstProp, v, errRec)
				dstMap[k] = dstProp
			}
		case "Required":
			// merge
			dstField.Set(reflect.AppendSlice(dstField, srcField))
		case "Type":
			if srcInt != dstInt {
				// TODO(directxman12): figure out how to attach this back to a useful point in the Go source or in the schema
				errRec.AddError(fmt.Errorf("conflicting types in allOf branches in schema: %s vs %s", dstInt, srcInt))
			}
			// keep the destination value, for now
		// TODO(directxman12): Default -- use field?
		// TODO(directxman12):
		// - Dependencies: if field x is present, then either schema validates or all props are present
		// - AdditionalItems: like AdditionalProperties
		// - Definitions: common named validation sets that can be references (merge, bail if duplicate)
		case "AdditionalProperties":
			// as of the time of writing, `allows: false` is not allowed, so we don't have to handle it
			srcProps := srcInt.(*apiext.JSONSchemaPropsOrBool)
			if srcProps.Schema == nil {
				// nothing to merge
				continue
			}
			dstProps := dstInt.(*apiext.JSONSchemaPropsOrBool)
			if dstProps.Schema == nil {
				dstProps.Schema = &apiext.JSONSchemaProps{}
			}
			flattenAllOfInto(dstProps.Schema, *srcProps.Schema, errRec)
		case "XPreserveUnknownFields":
			dstField.Set(srcField)
		case "XMapType":
			dstField.Set(srcField)
		// NB(directxman12): no need to explicitly handle nullable -- false is considered to be the zero value
		// TODO(directxman12): src isn't necessarily the field value -- it's just the most recent allOf entry
		default:
			// hoist into allOf...
			hoisted = true

			srcRemVal.Field(i).Set(srcField)
			dstRemVal.Field(i).Set(dstField)
			// ...and clear the original
			dstField.Set(zeroVal)
		}
	}

	if hoisted {
		dst.AllOf = append(dst.AllOf, dstRemainder, srcRemainder)
	}

	// dedup required
	if len(dst.Required) > 0 {
		reqUniq := make(map[string]struct{})
		for _, req := range dst.Required {
			reqUniq[req] = struct{}{}
		}
		dst.Required = make([]string, 0, len(reqUniq))
		for req := range reqUniq {
			dst.Required = append(dst.Required, req)
		}
		// be deterministic
		sort.Strings(dst.Required)
	}
}

// allOfVisitor recursively visits allOf fields in the schema,
// merging nested allOf properties into the root schema.
type allOfVisitor struct {
	// errRec is used to record errors while flattening (like two conflicting
	// field values used in an allOf)
	errRec ErrorRecorder
}

func (v *allOfVisitor) Visit(schema *apiext.JSONSchemaProps) SchemaVisitor {
	if schema == nil {
		return v
	}

	// clear this now so that we can safely preserve edits made my flattenAllOfInto
	origAllOf := schema.AllOf
	schema.AllOf = nil

	for _, embedded := range origAllOf {
		flattenAllOfInto(schema, embedded, v.errRec)
	}
	return v
}

// NB(directxman12): FlattenEmbedded is separate from Flattener because
// some tooling wants to flatten out embedded fields, but only actually
// flatten a few specific types first.

// FlattenEmbedded flattens embedded fields (represented via AllOf) which have
// already had their references resolved into simple properties in the containing
// schema.
func FlattenEmbedded(schema *apiext.JSONSchemaProps, errRec ErrorRecorder) *apiext.JSONSchemaProps {
	outSchema := schema.DeepCopy()
	EditSchema(outSchema, &allOfVisitor{errRec: errRec})
	return outSchema
}

// RefParts splits a reference produced by the schema generator into its component
// type name and package name (if it's a cross-package reference).  Note that
// referenced packages *must* be looked up relative to the current package.
func RefParts(ref string) (typ string, pkgName string, err error) {
	if !strings.HasPrefix(ref, defPrefix) {
		return "", "", fmt.Errorf("non-standard reference link %q", ref)
	}
	ref = ref[len(defPrefix):]
	// decode the json pointer encodings
	ref = strings.Replace(ref, "~1", "/", -1)
	ref = strings.Replace(ref, "~0", "~", -1)
	nameParts := strings.SplitN(ref, "~", 2)

	if len(nameParts) == 1 {
		// local reference
		return nameParts[0], "", nil
	}
	// cross-package reference
	return nameParts[1], nameParts[0], nil
}

// identFromRef converts the given schema ref from the given package back
// into the TypeIdent that it represents.
func identFromRef(ref string, contextPkg *loader.Package) (TypeIdent, error) {
	typ, pkgName, err := RefParts(ref)
	if err != nil {
		return TypeIdent{}, err
	}

	if pkgName == "" {
		// a local reference
		return TypeIdent{
			Name:    typ,
			Package: contextPkg,
		}, nil
	}

	// an external reference
	return TypeIdent{
		Name:    typ,
		Package: contextPkg.Imports()[pkgName],
	}, nil
}

// preserveFields copies documentation fields from src into dst, preserving
// field-level documentation when flattening, and preserving field-level validation
// as allOf entries.
func preserveFields(dst *apiext.JSONSchemaProps, src apiext.JSONSchemaProps) {
	srcDesc := src.Description
	srcTitle := src.Title
	srcExDoc := src.ExternalDocs
	srcEx := src.Example

	src.Description, src.Title, src.ExternalDocs, src.Example = "", "", nil, nil

	src.Ref = nil
	*dst = apiext.JSONSchemaProps{
		AllOf: []apiext.JSONSchemaProps{*dst, src},

		// keep these, in case the source field doesn't specify anything useful
		Description:  dst.Description,
		Title:        dst.Title,
		ExternalDocs: dst.ExternalDocs,
		Example:      dst.Example,
	}

	if srcDesc != "" {
		dst.Description = srcDesc
	}
	if srcTitle != "" {
		dst.Title = srcTitle
	}
	if srcExDoc != nil {
		dst.ExternalDocs = srcExDoc
	}
	if srcEx != nil {
		dst.Example = srcEx
	}
}

// flattenVisitor visits each node in the schema, recursively flattening references.
type flattenVisitor struct {
	*flattener

	currentPackage *loader.Package
	currentType    *TypeIdent
	currentSchema  *apiext.JSONSchemaProps
	originalField  apiext.JSONSchemaProps
}

func (r *flattenVisitor) Visit(baseSchema *apiext.JSONSchemaProps) SchemaVisitor {
	if baseSchema == nil {
		// end-of-node marker, cache the results
		if r.currentType != nil {
			r.cacheType(*r.currentType, *r.currentSchema)
			// preserve field information *after* caching so that we don't
			// accidentally cache field-level information onto the schema for
			// the type in general.
			preserveFields(r.currentSchema, r.originalField)
		}
		return r
	}

	// if we get a type that's just a ref, resolve it
	if baseSchema.Ref != nil && len(*baseSchema.Ref) > 0 {
		// resolve this ref
		refIdent, err := r.LookupReference(*baseSchema.Ref, r.currentPackage)
		if err != nil {
			r.currentPackage.AddError(err)
			return nil
		}

		// load and potentially flatten the schema

		// check the cache first...
		if refSchemaCached, isCached := r.flattenedTypes[refIdent]; isCached {
			// shallow copy is fine, it's just to avoid overwriting the doc fields
			preserveFields(&refSchemaCached, *baseSchema)
			*baseSchema = refSchemaCached
			return nil // don't recurse, we're done
		}

		// ...otherwise, we need to flatten
		refSchema, err := r.loadUnflattenedSchema(refIdent)
		if err != nil {
			r.currentPackage.AddError(err)
			return nil
		}
		refSchema = refSchema.DeepCopy()

		// keep field around to preserve field-level validation, docs, etc
		origField := *baseSchema
		*baseSchema = *refSchema

		// avoid loops (which shouldn't exist, but just in case)
		// by marking a nil cached pointer before we start recursing
		r.cacheType(refIdent, apiext.JSONSchemaProps{})

		return &flattenVisitor{
			flattener: r.flattener,

			currentPackage: refIdent.Package,
			currentType:    &refIdent,
			currentSchema:  baseSchema,
			originalField:  origField,
		}
	}

	// otherwise, continue recursing...
	if r.currentType != nil {
		// ...but don't accidentally end this node early (for caching purposes)
		return &flattenVisitor{
			flattener:      r.flattener,
			currentPackage: r.currentPackage,
		}
	}

	return r
}
