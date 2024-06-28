package ukgen

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

// =============================================================================
// Handler
// =============================================================================

type GenerateParams struct {
	Out GenerateOutput `ukflag:"o out"`
}

func (gp *GenerateParams) UkaseInit() { gp.Out.Writer = os.Stdout }

type GenerateOutput struct{ io.Writer }

func (o GenerateOutput) MarshalText() ([]byte, error) {
	switch writerT := o.Writer.(type) {
	case *os.File:
		return []byte(writerT.Name()), nil
	case fmt.Stringer:
		return []byte(writerT.String()), nil
	default:
		return []byte("unknown"), nil
	}
}

func (o *GenerateOutput) UnmarshalText(text []byte) (err error) {
	switch str := string(text); str {
	case "stdin":
		o.Writer = os.Stdin
	case "stdout":
		o.Writer = os.Stdout
	case "stderr":
		o.Writer = os.Stderr
	default:
		o.Writer, err = os.Create(str)
	}

	return
}

func (g Generator) Bind(target ...string) ukcli.Directive {
	handle := ukcli.NewHandler(g.Handle)
	return handle.Bind(target...)
}

func (g Generator) Handle(_ context.Context, params GenerateParams) error {
	return g.Generate(params.Out)
}

// =============================================================================
// Middleware
// =============================================================================

type generateState struct {
	ukcli.State
	generator *Generator
}

func (g *Generator) UkaseApply(config *ukcli.Config) {
	middleware := func(s ukcli.State) ukcli.State {
		return generateState{State: s, generator: g}
	}

	config.Middleware = append(config.Middleware, middleware)
}

func (gs generateState) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target ...string) error {
	if err := gs.State.RegisterExec(exec, spec, target...); err != nil {
		return err
	}

	return gs.generator.load(spec)
}
