package ukhelp

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// =============================================================================
// Template Renderer
// =============================================================================

type TemplateRenderer struct {
	Name  string
	Text  string
	Funcs template.FuncMap
}

func (tr TemplateRenderer) Render(w io.Writer, data any) error {
	t, err := template.New(tr.Name).Funcs(tr.Funcs).Parse(tr.Text)
	if err != nil {
		return err
	}

	return t.Execute(w, data)
}

// =============================================================================
// Utility Functions
// =============================================================================

type renderFuncs struct{}

// ===== Has ...
func (rf renderFuncs) hasUsage(o Output) bool      { return rf.hasCommandExec(o) || rf.hasSubcommands(o) }
func (renderFuncs) hasCommandExec(o Output) bool   { return o.Command.Exec }
func (renderFuncs) hasCommandTarget(o Output) bool { return len(o.Command.Target) != 0 }
func (renderFuncs) hasArguments(o Output) bool     { return len(o.Arguments) != 0 }
func (renderFuncs) hasFlags(o Output) bool         { return len(o.Flags) != 0 }
func (renderFuncs) hasSubcommands(o Output) bool   { return len(o.Subcommands) != 0 }

// ===== Max Label ...
func (rf renderFuncs) maxLabelArguments(o Output) (max int) {
	for _, arg := range o.Arguments {
		label := rf.labelArgument(arg)
		if x := len(label); x > max {
			max = x
		}
	}
	return
}

func (rf renderFuncs) maxLabelFlags(o Output) (max int) {
	for _, flag := range o.Flags {
		label := rf.labelFlag(flag)
		if x := len(label); x > max {
			max = x
		}
	}
	return
}

func (rf renderFuncs) maxLabelSubcommands(o Output) (max int) {
	for _, subcommand := range o.Subcommands {
		label := rf.labelSubcommand(subcommand)
		if x := len(label); x > max {
			max = x
		}
	}
	return
}

// ===== Label ...
func (renderFuncs) labelArgument(o OutputArgument) string {
	switch start, end := o.Position.Start, o.Position.End; {
	case start == -1 && end == -1:
		return "..."
	case start == -1:
		return fmt.Sprintf("...%d", end)
	case end == -1:
		return fmt.Sprintf("%d...", start)
	case (end - start) < 2:
		return fmt.Sprintf("%d", start)
	default:
		return fmt.Sprintf("%d...%d", start, end)
	}
}

func (renderFuncs) labelFlag(o OutputFlag) string {
	var names []string

	for _, name := range o.Names {
		switch len(name) {
		case 0:
		case 1:
			names = append(names, "-"+name)
		default:
			names = append(names, "--"+name)
		}
	}

	return strings.Join(names, ", ")
}

func (renderFuncs) labelSubcommand(o OutputSubcommand) string {
	return o.Name
}

// ===== Describe ...
func (renderFuncs) describeCommand(o OutputCommand) string {
	switch {
	case o.Description.Long != "":
		return o.Description.Long
	case o.Description.Short != "":
		return o.Description.Short
	default:
		// TODO: Reconsider this
		return "No information available"
	}
}

func (renderFuncs) describeSubcommand(o OutputSubcommand) string {
	switch {
	case o.Description.Short != "":
		return o.Description.Short
	default:
		// TODO: Reconsider this
		return "No information available"
	}
}

func (renderFuncs) describeArgument(o OutputArgument) string {
	switch {
	case o.Description.Short != "":
		return o.Description.Short
	default:
		// TODO: Reconsider this
		return "No information available"
	}
}

func (renderFuncs) describeFlag(o OutputFlag) string {
	switch {
	case o.Description.Short != "":
		return o.Description.Short
	default:
		// TODO: Reconsider this
		return "No information available"
	}
}
