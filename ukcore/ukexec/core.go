package ukcore

import (
	"context"
	"fmt"
	"strings"

	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Exec
// =============================================================================

type Exec func(context.Context, Input) error

// =============================================================================
// Input
// =============================================================================

type Input struct {
	Program string
	Target  InputTarget
	Args    []string
	Flags   []InputFlag
}

type InputTarget []string

func (i InputTarget) String() string { return strings.Join(i, "›") }

type InputFlag struct{ Name, Value string }

func (i InputFlag) String() string {
	// TODO: Move this logic into ukspec?

	switch len(i.Name) {
	case 0:
		return "<invalid>"
	case 1:
		return fmt.Sprintf("-%s %s", i.Name, i.Value)
	default:
		return fmt.Sprintf("--%s %s", i.Name, i.Value)
	}
}

// =============================================================================
// Meta
// =============================================================================

type Meta struct {
	Exec bool
	Info any
	Spec ukspec.Params

	children map[string]*muxNode
}

func newMeta(node *muxNode) Meta {
	meta := Meta{
		Exec:     node.exec != nil,
		Info:     nil,
		Spec:     ukspec.Empty,
		children: node.children,
	}

	if node.info != nil {
		meta.Info = node.info
	}

	if node.spec != nil {
		meta.Spec = *node.spec
	}

	return meta
}

func (m Meta) Children() map[string]Meta {
	res := make(map[string]Meta)

	for childName, child := range m.children {
		res[childName] = newMeta(child)
	}

	return res
}
