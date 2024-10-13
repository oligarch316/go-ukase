package ukhelp

import (
	"io"
	"strings"
	"text/template"
)

// =============================================================================
// Template
// =============================================================================

type TemplateRenderer struct {
	Name string
	Text string
	Out  io.Writer

	DelimLeft  string
	DelimRight string
	Options    []string
	Funcs      template.FuncMap
}

func (tr TemplateRenderer) Render(data any) error {
	tmpl, err := tr.build()
	if err != nil {
		return err
	}

	return tmpl.Execute(tr.Out, data)
}

func (tr TemplateRenderer) build() (*template.Template, error) {
	return template.New(tr.Name).
		Delims(tr.DelimLeft, tr.DelimRight).
		Option(tr.Options...).
		Funcs(tr.Funcs).
		Parse(tr.Text)
}

// =============================================================================
// Template Functions
// =============================================================================

type RenderFuncs[T any] func(description T, long bool) string

func NewRenderFuncs[T any](renderDescription func(T, bool) string) RenderFuncs[T] {
	return RenderFuncs[T](renderDescription)
}

func (rf RenderFuncs[T]) Map() template.FuncMap {
	return template.FuncMap{
		"describeCommand":    rf.command,
		"describeSubcommand": rf.subcommand,
		"describeFlag":       rf.flag,
		"describeArgument":   rf.argument,

		"hasCommand":     rf.hasCommand,
		"hasSubcommands": rf.hasSubcommands,
		"hasFlags":       rf.hasFlags,
		"hasArguments":   rf.hasArguments,

		"labelCommand":    rf.labelCommand,
		"labelSubcommand": rf.labelSubcommand,
		"labelFlag":       rf.labelFlag,
		"labelArgument":   rf.labelArgument,

		"maxSubcommand": rf.maxSubcommand,
		"maxFlag":       rf.maxFlag,
		"maxArgument":   rf.maxArgument,
	}
}

// -----------------------------------------------------------------------------
// ❭ Describe
// -----------------------------------------------------------------------------

func (f RenderFuncs[T]) command(o OutputCommand[T], l bool) string       { return f(o.Description, l) }
func (f RenderFuncs[T]) subcommand(o OutputSubcommand[T], l bool) string { return f(o.Description, l) }
func (f RenderFuncs[T]) flag(o OutputFlag[T], l bool) string             { return f(o.Description, l) }
func (f RenderFuncs[T]) argument(o OutputArgument[T], l bool) string     { return f(o.Description, l) }

// -----------------------------------------------------------------------------
// ❭ Has
// -----------------------------------------------------------------------------

func (RenderFuncs[T]) hasCommand(o Output[T]) bool     { return o.Command.Exec }
func (RenderFuncs[T]) hasSubcommands(o Output[T]) bool { return len(o.Subcommands) != 0 }
func (RenderFuncs[T]) hasFlags(o Output[T]) bool       { return len(o.Flags) != 0 }
func (RenderFuncs[T]) hasArguments(o Output[T]) bool   { return len(o.Arguments) != 0 }

// -----------------------------------------------------------------------------
// ❭ Max (label)
// -----------------------------------------------------------------------------

func (r RenderFuncs[T]) maxSubcommand(o Output[T]) int { return rMax(o.Subcommands, r.labelSubcommand) }
func (r RenderFuncs[T]) maxFlag(o Output[T]) int       { return rMax(o.Flags, r.labelFlag) }
func (r RenderFuncs[T]) maxArgument(o Output[T]) int   { return rMax(o.Arguments, r.labelArgument) }

func rMax[S ~[]E, E any](list S, labelF func(E) string) (max int) {
	for _, item := range list {
		if candidate := len(labelF(item)); candidate > max {
			max = candidate
		}
	}
	return
}

// -----------------------------------------------------------------------------
// ❭ Label
// -----------------------------------------------------------------------------

func (RenderFuncs[T]) labelCommand(o OutputCommand[T]) string {
	segs := append([]string{o.Program}, o.Target...)
	return strings.Join(segs, " ")
}

func (RenderFuncs[T]) labelSubcommand(o OutputSubcommand[T]) string {
	return o.Name
}

func (RenderFuncs[T]) labelFlag(o OutputFlag[T]) string {
	var items []string

	for _, name := range o.Names {
		switch len(name) {
		case 0:
		case 1:
			items = append(items, "-"+name)
		default:
			items = append(items, "--"+name)
		}
	}

	return strings.Join(items, ", ")
}

func (RenderFuncs[T]) labelArgument(o OutputArgument[T]) string {
	// TODO: More human friendly display than half open range???
	return strings.Replace(o.Position.String(), ":", "...", 1)
}
