package ukgen

import (
	"bytes"
	_ "embed"
	"go/format"
	"io"
	"text/template"

	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

//go:embed generate.tmpl
var generateTemplateText string

type Generator struct {
	config  Config
	imports importStore
	params  paramsStore
}

func NewGenerator(opts ...Option) *Generator {
	return &Generator{
		config:  newConfig(opts),
		imports: newImportStore(),
		params:  newParamsStore(),
	}
}

func (g *Generator) Generate(out io.Writer) error {
	data, err := g.generate()
	if err != nil {
		return err
	}

	t, err := template.New("generate").Parse(generateTemplateText)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = out.Write(p)
	return err
}

// =============================================================================
// Data
// =============================================================================

type data struct {
	Core    coreData
	Imports []importData
	Params  []paramsData
}

// =============================================================================
// Load
// =============================================================================

func (g *Generator) load(spec ukspec.Params) error { return g.loadParams(spec) }

// =============================================================================
// Generate
// =============================================================================

func (g *Generator) generate() (data, error) {
	core, err := g.generateCore()
	if err != nil {
		return data{}, err
	}

	params, err := g.generateParams()
	if err != nil {
		return data{}, err
	}

	// NOTE:
	// Imports must be generated last
	// Prior generation routines perform import loading
	imports := g.generateImports()

	data := data{Core: core, Imports: imports, Params: params}
	return data, nil
}
