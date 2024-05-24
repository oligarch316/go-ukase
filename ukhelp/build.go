package ukhelp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Builder
// =============================================================================

type Builder struct{ config Config }

func New(opts ...Option) *Builder {
	return &Builder{config: newConfig(opts)}
}

func (b *Builder) Auto(subcommandName string) ukase.Option {
	return auto{builder: b, name: subcommandName}
}

func (b *Builder) Build(reference ...string) ukase.Exec[struct{}] {
	return func(ctx context.Context, input ukase.Input) error {
		ref := append(reference, input.Args...)

		meta, err := input.Meta(ref)
		if err != nil {
			return err
		}

		data := make(wipData)

		data.loadCommandInfo(meta)
		data.loadSubcommandInfo(meta)
		data.loadParameterInfo(input, meta)
		data.loadParameterDefaults(input, meta)

		t, err := b.config.Template(input.Target)
		if err != nil {
			return err
		}

		return t.Execute(b.config.Out, data)
	}
}

// Temporary hacky shim
type wipData map[string]any

func (wd wipData) loadCommandInfo(meta ukcore.Meta) {
	if info, ok := meta.Info(); ok {
		wd["info"] = info
		return
	}

	wd["info"] = "TODO: empty"
}

func (wd wipData) loadSubcommandInfo(meta ukcore.Meta) {
	subcommands := make(map[string]any)

	for name, subMeta := range meta.Children() {
		if info, ok := subMeta.Info(); ok {
			subcommands[name] = info
			continue
		}

		subcommands[name] = "TODO: empty"
	}

	wd["subcommands"] = subcommands
}

func (wd wipData) loadParameterInfo(input ukase.Input, meta ukcore.Meta) {
	wd["parameters"] = "TODO: not yet implemented"
}

func (wd wipData) loadParameterDefaults(input ukase.Input, meta ukcore.Meta) {
	spec, ok := meta.Spec()
	if !ok {
		wd["defaults"] = "TODO: empty"
		return
	}

	ptrVal := reflect.New(spec.Type)

	if err := input.Initialize(ptrVal.Interface()); err != nil {
		wd["defaults"] = fmt.Sprintf("TODO: initialize error: %s", err)
		return
	}

	wd["defaults"] = ptrVal.Elem().Interface()
}

// =============================================================================
// Auto
// =============================================================================

// TODO: Put this in ukopt?
func Auto(subcommandName string, opts ...Option) ukase.Option {
	return New(opts...).Auto(subcommandName)
}

type auto struct {
	builder *Builder
	name    string
}

func (a auto) UkaseApply(config *ukase.Config) {
	middleware := func(state ukase.State) ukase.State {
		return &autoState{auto: a, State: state}
	}

	config.Middleware = append(config.Middleware, middleware)
}

type autoNode map[string]autoNode

type autoState struct {
	auto
	ukase.State

	root autoNode
}

func (as *autoState) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target []string) error {
	baseErr := as.State.RegisterExec(exec, spec, target)
	helpErr := as.registerHelp(target)
	return errors.Join(baseErr, helpErr)
}

func (as *autoState) registerHelp(target []string) error {
	var errs []error

	if as.root == nil {
		as.root = make(autoNode)
		errs = append(errs, as.registerHelpTarget(nil))
	}

	node := as.root

	for i := 0; i < len(target); i++ {
		step, path := target[i], slices.Clone(target[:i+1])

		if child, ok := node[step]; ok {
			node = child
			continue
		}

		child := make(autoNode)
		node[step] = child
		errs = append(errs, as.registerHelpTarget(path))
	}

	return errors.Join(errs...)
}

func (as *autoState) registerHelpTarget(target []string) error {
	helpExec := as.builder.Build(target...)
	helpTarget := append(target, as.name)
	return helpExec.Bind(helpTarget...).UkaseRegister(as.State)
}
