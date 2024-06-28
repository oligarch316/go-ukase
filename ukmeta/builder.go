package ukmeta

import (
	"slices"

	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

type Builder[Params any] func(refTarget ...string) ukcli.Exec[Params]

func NewBuilder[Params any](builder func(...string) ukcli.Exec[Params]) Builder[Params] {
	return Builder[Params](builder)
}

func (b Builder[Params]) Auto(name string) ukcli.Option {
	return autoBuilder[Params]{builder: b, name: name}
}

type autoBuilder[Params any] struct {
	builder Builder[Params]
	name    string
}

func (ab autoBuilder[Params]) UkaseApply(config *ukcli.Config) {
	middleware := func(s ukcli.State) ukcli.State {
		return &autoState[Params]{autoBuilder: ab, State: s}
	}

	config.Middleware = append(config.Middleware, middleware)
}

type autoTree map[string]autoTree

type autoState[Params any] struct {
	autoBuilder[Params]
	ukcli.State
	memo autoTree
}

func (as *autoState[Params]) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target ...string) error {
	if err := as.State.RegisterExec(exec, spec, target...); err != nil {
		return err
	}

	return as.registerHelp(target)
}

func (as *autoState[Params]) registerHelp(target []string) error {
	for _, path := range as.sift(target) {
		helpExec := as.builder(path...)
		helpTarget := append(path, as.name)
		helpDirective := helpExec.Bind(helpTarget...)

		if err := helpDirective.UkaseRegister(as.State); err != nil {
			return err
		}
	}

	return nil
}

// Ensure each sub-path of the given target is marked as visited.
// Return a list of those that have not previously been visited.
func (as *autoState[Params]) sift(target []string) [][]string {
	var paths [][]string

	if as.memo == nil {
		// Root (empty) target not yet visited

		// ⇒ Initialize memo to mark as visited
		as.memo = make(autoTree)

		// ⇒ Add empty path (nil) to the result list
		paths = [][]string{nil}
	}

	for cur, i := as.memo, 0; i < len(target); i++ {
		name := target[i]

		next, seen := cur[name]
		if !seen {
			// Path [0...i] not yet visited

			// ⇒ Initialize and add to memo to mark as visited
			next = make(autoTree)
			cur[name] = next

			// ⇒ Add path (shallow copy) to result list
			path := slices.Clone(target[:i+1])
			paths = append(paths, path)
		}

		cur = next
	}

	return paths
}
