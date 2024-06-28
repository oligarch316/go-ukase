package ukgen

import (
	"cmp"
	"fmt"
	"go/token"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// =============================================================================
// Data
// =============================================================================

type typeData struct{ PackageName, TypeName string }

func (td typeData) String() string {
	if td.PackageName == "" {
		return td.TypeName
	}
	return td.PackageName + "." + td.TypeName
}

type importData struct{ PackageName, PackagePath string }

func (id importData) String() string {
	return fmt.Sprintf("%s %q", id.PackageName, id.PackagePath)
}

// =============================================================================
// Load
// =============================================================================

type importStore struct {
	pathToName  map[string]string
	nameToIndex map[string]int
}

func newImportStore() importStore {
	return importStore{
		pathToName:  make(map[string]string),
		nameToIndex: make(map[string]int),
	}
}

func (g *Generator) loadImport(t reflect.Type) (typeData, error) {
	pkgPath, typeName := t.PkgPath(), t.Name()

	switch {
	case typeName == "":
		// Non-defined (anonymous) type, eg. `struct{}{}`
		return typeData{PackageName: "", TypeName: ""}, nil
	case pkgPath == "":
		// Pre-declared (native) type, eg. `string`
		return typeData{PackageName: "", TypeName: t.Name()}, nil
	}

	pkgName, err := g.loadImportName(pkgPath)
	if err != nil {
		return typeData{}, err
	}

	data := typeData{PackageName: pkgName, TypeName: t.Name()}
	return data, nil
}

func (g *Generator) loadImportName(pkgPath string) (string, error) {
	if pkgName, ok := g.imports.pathToName[pkgPath]; ok {
		return pkgName, nil
	}

	pkgBase, err := g.parsePackageBase(pkgPath)
	if err != nil {
		return "", err
	}

	pkgName, idx := pkgBase, g.imports.nameToIndex[pkgBase]
	if idx != 0 {
		pkgName = pkgName + strconv.Itoa(idx)
	}

	g.imports.nameToIndex[pkgBase] = idx + 1
	g.imports.pathToName[pkgPath] = pkgName

	return pkgName, nil
}

// =============================================================================
// Generate
// =============================================================================

func (g *Generator) generateImports() []importData {
	var list []importData

	for pkgPath, pkgName := range g.imports.pathToName {
		item := importData{PackageName: pkgName, PackagePath: pkgPath}
		list = append(list, item)
	}

	// Sort by lexicographic package path
	compare := func(a, b importData) int { return cmp.Compare(a.PackagePath, b.PackagePath) }
	slices.SortFunc(list, compare)

	return list
}

// =============================================================================
// Utility
// =============================================================================

func (Generator) parsePackageBase(pkgPath string) (string, error) {
	base := path.Base(pkgPath)

	hyphenMin, hyphenMax := 1, len(base)-2
	hyphenIdx := strings.LastIndexByte(base, '-')

	if hyphenIdx >= hyphenMin && hyphenIdx <= hyphenMax {
		base = base[hyphenIdx+1:]
	}

	if !token.IsIdentifier(base) {
		return "", fmt.Errorf("[TODO parsePackageName] invalid package base name '%s'", base)
	}

	return base, nil
}
