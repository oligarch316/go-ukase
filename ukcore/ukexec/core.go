package ukexec

import "github.com/oligarch316/go-ukase/ukcore/ukspec"

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
