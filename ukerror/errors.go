package ukerror

import "github.com/oligarch316/go-ukase/internal/ierrors"

// TODO: Document
var ErrAny = ierrors.ErrAny

// TODO: Document
var (
	ErrInternal  = ierrors.ErrInternal
	ErrDeveloper = ierrors.ErrDeveloper
	ErrUser      = ierrors.ErrUser
)

// TODO: Document
var (
	ErrDec  = ierrors.ErrDec
	ErrExec = ierrors.ErrExec
	ErrInit = ierrors.ErrInit
	ErrSpec = ierrors.ErrSpec
)
