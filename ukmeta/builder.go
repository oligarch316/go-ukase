package ukmeta

import (
	"slices"

	"github.com/oligarch316/go-ukase/ukcli"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

type Builder[Params any] func(refTarget ...string) (exec ukcli.Exec[Params], info any)

func NewBuilder[Params any](builder func(...string) (ukcli.Exec[Params], any)) Builder[Params] {
	return Builder[Params](builder)
}

func (b Builder[Params]) Auto(name string) func(ukcli.State) ukcli.State {
	return func(s ukcli.State) ukcli.State {
		return &autoState[Params]{State: s, builder: b, name: name}
	}
}

type autoTree map[string]autoTree

type autoState[Params any] struct {
	ukcli.State

	builder Builder[Params]
	name    string
	memo    autoTree
}

func (as *autoState[Params]) RegisterExec(exec ukcore.Exec, spec ukspec.Params, target ...string) error {
	if err := as.State.RegisterExec(exec, spec, target...); err != nil {
		return err
	}

	return as.registerMeta(target)
}

func (as *autoState[Params]) registerMeta(target []string) error {
	for _, path := range as.sift(target) {
		metaTarget := append(path, as.name)
		exec, info := as.builder(path...)

		if err := as.registerMetaExec(exec, metaTarget); err != nil {
			return err
		}

		if err := as.registerMetaInfo(info, metaTarget); err != nil {
			return err
		}
	}

	return nil
}

func (as *autoState[Params]) registerMetaExec(exec ukcli.Exec[Params], target []string) error {
	if exec == nil {
		return nil
	}

	metaExec := exec.Bind(target...)
	return metaExec.UkaseRegister(as.State)
}

func (as *autoState[Params]) registerMetaInfo(info any, target []string) error {
	if info == nil {
		return nil
	}

	metaInfo := ukcli.NewInfo(info).Bind(target...)
	return metaInfo.UkaseRegister(as.State)
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
