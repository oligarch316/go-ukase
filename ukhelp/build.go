package ukhelp

import (
	"context"
	"slices"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Builder
// =============================================================================

type Builder struct{ config Config }

func New(opts ...Option) Builder {
	return Builder{config: newConfig(opts)}
}

func (b Builder) Auto(subcommandName string) ukase.Option {
	return autoBuilder{builder: b, name: subcommandName}
}

func (b Builder) Build(reference ...string) ukase.Exec[struct{}] {
	return func(c context.Context, i ukase.Input) error {
		return b.exec(c, reference, i)
	}
}

func (b Builder) exec(ctx context.Context, ref []string, input ukase.Input) error {
	helpInput, err := newInput(ref, input)
	if err != nil {
		return err
	}

	helpData, err := b.config.Encode(ctx, helpInput)
	if err != nil {
		return err
	}

	return b.config.Render(b.config.Out, helpData)
}

// =============================================================================
// Auto
// =============================================================================

type autoBuilder struct {
	builder Builder
	name    string
}

func Auto(subcommandName string, opts ...Option) ukase.Option {
	return New(opts...).Auto(subcommandName)
}

func (ab autoBuilder) UkaseApply(config *ukase.Config) {
	middleware := func(s ukase.State) ukase.State {
		return &autoState{autoBuilder: ab, State: s}
	}

	config.Middleware = append(config.Middleware, middleware)
}

type autoTree map[string]autoTree

type autoState struct {
	autoBuilder
	ukase.State
	memo autoTree
}

func (as *autoState) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target []string) error {
	if err := as.State.RegisterExec(exec, spec, target); err != nil {
		return err
	}

	return as.registerHelp(target)
}

func (as *autoState) registerHelp(target []string) error {
	for _, ref := range as.sift(target) {
		helpExec := as.builder.Build(ref...)
		helpTarget := append(ref, as.name)
		helpDirective := helpExec.Bind(helpTarget...)

		as.builder.config.Log.Debug(
			"registering help exec",
			"reference", ukcore.InputTarget(helpTarget),
		)

		if err := helpDirective.UkaseRegister(as.State); err != nil {
			return err
		}
	}

	return nil
}

func (as *autoState) sift(target []string) (paths [][]string) {
	if as.memo == nil {
		paths, as.memo = [][]string{nil}, make(autoTree)
	}

	for cur, i := as.memo, 0; i < len(target); i++ {
		name := target[i]

		next, seen := cur[name]
		if !seen {
			next = make(autoTree)
			cur[name] = next
			paths = append(paths, slices.Clone(target[:i+1]))
		}

		cur = next
	}

	return
}
