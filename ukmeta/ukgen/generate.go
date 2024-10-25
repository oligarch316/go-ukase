package ukgen

import (
	_ "embed"

	"bytes"
	"go/format"
	"io"
	"text/template"

	"github.com/oligarch316/ukase/ukcore/ukspec"
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
	g.config.Log.Info("generating data")
	data, err := g.generate()
	if err != nil {
		return err
	}

	g.config.Log.Info("parsing template")
	t, err := template.New("generate").Parse(generateTemplateText)
	if err != nil {
		return err
	}

	g.config.Log.Info("rendering template")
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	g.config.Log.Info("formatting result")
	p, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	g.config.Log.Info("writing result")
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

func (g *Generator) load(spec ukspec.Parameters) error { return g.loadParams(spec) }

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
