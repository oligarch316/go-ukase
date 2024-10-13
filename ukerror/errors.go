package ukerror

import "github.com/oligarch316/go-ukase/internal/ierror"

// TODO: Document
var ErrAny = ierror.ErrAny

// TODO: Document
var (
	ErrInternal  = ierror.ErrInternal
	ErrDeveloper = ierror.ErrDeveloper
	ErrUser      = ierror.ErrUser
)

// TODO: Document
var (
	ErrDec  = ierror.ErrDec
	ErrExec = ierror.ErrExec
	ErrInit = ierror.ErrInit
	ErrSpec = ierror.ErrSpec
)
