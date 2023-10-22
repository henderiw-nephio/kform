package loader

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/types"
	"os"
	"sync"

	"golang.org/x/tools/go/packages"
)

// Package is a single, unique Go package that can be
// lazily parsed and type-checked.  Packages should not
// be constructed directly -- instead, use LoadRoots.
type Package struct {
	*packages.Package

	imports map[string]*Package

	loader *loader
	sync.Mutex
}

// Imports returns the imports for the given package, indexed by
// package path (*not* name in any particular file).
func (p *Package) Imports() map[string]*Package {
	if p.imports == nil {
		p.imports = p.loader.packagesFor(p.Package.Imports)
	}

	return p.imports
}

// NeedSyntax indicates that a parsed AST is needed for this package.
// Actual ASTs can be accessed via the Syntax field.
func (p *Package) NeedSyntax() {
	if p.Syntax != nil {
		return
	}
	out := make([]*ast.File, len(p.CompiledGoFiles))
	var wg sync.WaitGroup
	wg.Add(len(p.CompiledGoFiles))
	for i, filename := range p.CompiledGoFiles {
		go func(i int, filename string) {
			defer wg.Done()
			src, err := os.ReadFile(filename)
			if err != nil {
				p.AddError(err)
				return
			}
			out[i], err = p.loader.parseFile(filename, src)
			if err != nil {
				p.AddError(err)
				return
			}
		}(i, filename)
	}
	wg.Wait()
	for _, file := range out {
		if file == nil {
			return
		}
	}
	p.Syntax = out
}

// NeedTypesInfo indicates that type-checking information is needed for this package.
// Actual type-checking information can be accessed via the Types and TypesInfo fields.
func (p *Package) NeedTypesInfo() {
	if p.TypesInfo != nil {
		return
	}
	p.NeedSyntax()
	p.typeCheck()
}

// typeCheck type-checks the given package.
func (p *Package) typeCheck() {
	// don't conflict with typeCheckFromExportData

	p.TypesInfo = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	p.Fset = p.loader.cfg.Fset
	p.Types = types.NewPackage(p.PkgPath, p.Name)

	importer := importerFunc(func(path string) (*types.Package, error) {
		if path == "unsafe" {
			return types.Unsafe, nil
		}

		// The imports map is keyed by import path.
		importedPkg := p.Imports()[path]
		if importedPkg == nil {
			return nil, fmt.Errorf("package %q possibly creates an import loop", path)
		}

		// it's possible to have a call to check in parallel to a call to this
		// if one package in the package graph gets its dependency filtered out,
		// but another doesn't (so one wants a "placeholder" package here, and another
		// wants the full check).
		//
		// Thus, we need to lock here (at least for the time being) to avoid
		// races between the above write to `pkg.Types` and this checking of
		// importedPkg.Types.
		importedPkg.Lock()
		defer importedPkg.Unlock()

		if importedPkg.Types != nil && importedPkg.Types.Complete() {
			return importedPkg.Types, nil
		}

		// if we haven't already loaded typecheck data, we don't care about this package's types
		return types.NewPackage(importedPkg.PkgPath, importedPkg.Name), nil
	})

	var errs []error

	// type-check
	checkConfig := &types.Config{
		Importer: importer,

		IgnoreFuncBodies: true, // we only need decl-level info

		Error: func(err error) {
			errs = append(errs, err)
		},

		Sizes: p.TypesSizes,
	}
	if err := types.NewChecker(checkConfig, p.loader.cfg.Fset, p.Types, p.TypesInfo).Files(p.Syntax); err != nil {
		errs = append(errs, err)
	}

	// make sure that if a given sub-import is ill-typed, we mark this package as ill-typed as well.
	illTyped := len(errs) > 0
	if !illTyped {
		for _, importedPkg := range p.Imports() {
			if importedPkg.IllTyped {
				illTyped = true
				break
			}
		}
	}
	p.IllTyped = illTyped

	// publish errors to the package error list.
	for _, err := range errs {
		p.AddError(err)
	}
}

// AddError adds an error to the errors associated with the given package.
func (p *Package) AddError(err error) {
	switch typedErr := err.(type) {
	case *os.PathError:
		// file-reading errors
		p.Errors = append(p.Errors, packages.Error{
			Pos:  typedErr.Path + ":1",
			Msg:  typedErr.Err.Error(),
			Kind: packages.ParseError,
		})
	case scanner.ErrorList:
		// parsing/scanning errors
		for _, subErr := range typedErr {
			p.Errors = append(p.Errors, packages.Error{
				Pos:  subErr.Pos.String(),
				Msg:  subErr.Msg,
				Kind: packages.ParseError,
			})
		}
	case types.Error:
		// type-checking errors
		p.Errors = append(p.Errors, packages.Error{
			Pos:  typedErr.Fset.Position(typedErr.Pos).String(),
			Msg:  typedErr.Msg,
			Kind: packages.TypeError,
		})
	case ErrList:
		for _, subErr := range typedErr {
			p.AddError(subErr)
		}
	case PositionedError:
		p.Errors = append(p.Errors, packages.Error{
			Pos:  p.loader.cfg.Fset.Position(typedErr.Pos).String(),
			Msg:  typedErr.Error(),
			Kind: packages.UnknownError,
		})
	default:
		// should only happen for external errors, like ref checking
		p.Errors = append(p.Errors, packages.Error{
			Pos:  p.ID + ":-",
			Msg:  err.Error(),
			Kind: packages.UnknownError,
		})
	}
}

// EachType calls the given callback for each (gendecl, typespec) combo in the
// given package.  Generally, using markers.EachType is better when working
// with marker data, and has a more convinient representation.
func (p *Package) EachType(cb TypeCallback) {
	visitor := &typeVisitor{
		callback: cb,
	}
	p.NeedSyntax()
	for _, file := range p.Syntax {
		visitor.file = file
		ast.Walk(visitor, file)
	}
}

// importFunc is an implementation of the single-method
// types.Importer interface based on a function value.
type importerFunc func(path string) (*types.Package, error)

func (f importerFunc) Import(path string) (*types.Package, error) { return f(path) }
