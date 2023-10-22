package markers

import (
	"context"
	"fmt"
	"go/ast"
	"sync"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func (r *collector) Print() {
	r.m.RLock()
	defer r.m.RUnlock()
	for pkg, astNodes := range r.byPackage {
		fmt.Println("pkg: ", pkg)
		for _, markerValues := range astNodes {
			fmt.Printf(" astNode: , markerValues: %v\n", markerValues)
		}
	}
}

// Collector collects and parses marker comments defined in the registry
// from package source code.
type Collector interface {
	MarkersInPackage(pkg *loader.Package) (map[ast.Node]MarkerValues, error)
	Print()
}

func NewCollector(ctx context.Context) Collector {
	registry := cctx.GetContextValue[Registry](ctx, "registry")
	if registry == nil {
		registry = NewRegistry(ctx)
	}

	return &collector{
		Registry:  registry,
		byPackage: map[*loader.Package]map[ast.Node]MarkerValues{},
	}
}

type collector struct {
	Registry

	byPackage map[*loader.Package]map[ast.Node]MarkerValues
	m         sync.RWMutex
}

// MarkersInPackage computes the marker values by node for the given package.  Results
// are cached by package ID, so this is safe to call repeatedly from different functions.
// Each file in the package is treated as a distinct node.
//
// We consider a marker to be associated with a given AST node if either of the following are true:
//
// - it's in the Godoc for that AST node
//
//   - it's in the closest non-godoc comment group above that node,
//     *and* that node is a type or field node, *and* [it's either
//     registered as type-level *or* it's not registered as being
//     package-level]
//
//   - it's not in the Godoc of a node, doesn't meet the above criteria, and
//     isn't in a struct definition (in which case it's package-level)
func (r *collector) MarkersInPackage(pkg *loader.Package) (map[ast.Node]MarkerValues, error) {
	markers := r.getPkgMarkers(pkg)
	if markers != nil {
		return markers, nil
	}

	pkg.NeedSyntax()
	nodeMarkersRaw := r.associatePkgMarkers(pkg)
	markers, err := r.parseMarkersInPackage(nodeMarkersRaw)
	if err != nil {
		return nil, err
	}

	r.addPkgMarkers(pkg, markers)
	return markers, nil
}

func (r *collector) getPkgMarkers(pkg *loader.Package) map[ast.Node]MarkerValues {
	r.m.Lock()
	defer r.m.Unlock()
	if markers, exist := r.byPackage[pkg]; exist {
		return markers
	}
	return nil
}

func (r *collector) addPkgMarkers(pkg *loader.Package, markers map[ast.Node]MarkerValues) {
	r.m.Lock()
	defer r.m.Unlock()

	r.byPackage[pkg] = markers
}

// associatePkgMarkers associates markers with AST nodes in the given package.
func (r *collector) associatePkgMarkers(pkg *loader.Package) map[ast.Node][]markerComment {
	nodeMarkers := make(map[ast.Node][]markerComment)
	for _, file := range pkg.Syntax {
		fileNodeMarkers := r.associateFileMarkers(file)
		for node, markers := range fileNodeMarkers {
			nodeMarkers[node] = append(nodeMarkers[node], markers...)
		}
	}

	return nodeMarkers
}

// associateFileMarkers associates markers with AST nodes in the given file.
func (r *collector) associateFileMarkers(file *ast.File) map[ast.Node][]markerComment {
	// grab all the raw marker comments by node
	visitor := markerSubVisitor{
		collectPackageLevel: true,
		markerVisitor: &markerVisitor{
			nodeMarkers: make(map[ast.Node][]markerComment),
			allComments: file.Comments,
		},
	}
	ast.Walk(visitor, file)

	// grab the last package-level comments at the end of the file (if any)
	lastFileMarkers := visitor.markersBetween(false, visitor.commentInd, len(visitor.allComments))
	visitor.pkgMarkers = append(visitor.pkgMarkers, lastFileMarkers...)

	// figure out if any type-level markers are actually package-level markers
	for node, markers := range visitor.nodeMarkers {
		_, isType := node.(*ast.TypeSpec)
		if !isType {
			continue
		}
		endOfMarkers := 0
		for _, marker := range markers {
			if marker.fromGodoc {
				// markers from godoc are never package level
				markers[endOfMarkers] = marker
				endOfMarkers++
				continue
			}
			markerText := marker.Text()
			typeDef := r.Registry.Lookup(markerText, DescribesType)
			if typeDef != nil {
				// prefer assuming type-level markers
				markers[endOfMarkers] = marker
				endOfMarkers++
				continue
			}
			def := r.Registry.Lookup(markerText, DescribesPackage)
			if def == nil {
				// assume type-level unless proven otherwise
				markers[endOfMarkers] = marker
				endOfMarkers++
				continue
			}
			// it's package-level, since a package-level definition exists
			visitor.pkgMarkers = append(visitor.pkgMarkers, marker)
		}
		visitor.nodeMarkers[node] = markers[:endOfMarkers] // re-set after trimming the package markers
	}
	visitor.nodeMarkers[file] = visitor.pkgMarkers

	return visitor.nodeMarkers
}

// parseMarkersInPackage parses the given raw marker comments into output values using the registry.
func (r *collector) parseMarkersInPackage(nodeMarkersRaw map[ast.Node][]markerComment) (map[ast.Node]MarkerValues, error) {
	var errors []error
	nodeMarkerValues := make(map[ast.Node]MarkerValues)
	for node, markersRaw := range nodeMarkersRaw {
		var target TargetType
		switch node.(type) {
		case *ast.File:
			target = DescribesPackage
		case *ast.Field:
			target = DescribesField
		default:
			target = DescribesType
		}
		markerVals := make(map[string][]interface{})
		for _, markerRaw := range markersRaw {
			markerText := markerRaw.Text()
			def := r.Registry.Lookup(markerText, target)
			if def == nil {
				continue
			}
			val, err := def.Parse(markerText)
			if err != nil {
				errors = append(errors, loader.ErrFromNode(err, markerRaw))
				continue
			}
			markerVals[def.Name] = append(markerVals[def.Name], val)
		}
		nodeMarkerValues[node] = markerVals
	}

	return nodeMarkerValues, loader.MaybeErrList(errors)
}
