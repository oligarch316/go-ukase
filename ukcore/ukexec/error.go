package ukexec

import (
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukcore/ukspec"
)

var (
	ErrTargetNotExist = errors.New("target does not exist")
	ErrMissingProgram = errors.New("missing program name")
)

type ErrorExecConflict struct {
	Target           []string
	Original, Update ukspec.Params
	err              error
}

type ErrorInfoConflict struct {
	Target           []string
	Original, Update any
	err              error
}

type ErrorFlagConflict struct {
	Target           []string
	Name             string
	Original, Update ukspec.Flag
	err              error
}

func (eec ErrorExecConflict) Unwrap() error { return eec.err }
func (eic ErrorInfoConflict) Unwrap() error { return eic.err }
func (efc ErrorFlagConflict) Unwrap() error { return efc.err }

func (efc ErrorFlagConflict) Error() string {
	return fmt.Sprintf(
		"conflicting flag specifications for name '%s': %s",
		efc.Name, efc.err,
	)
}

func (eec ErrorExecConflict) Error() string {
	return fmt.Sprintf(
		"conflicting exec specifications for target '%s': %s",
		eec.Target, eec.err,
	)
}

func (eic ErrorInfoConflict) Error() string {
	return fmt.Sprintf(
		"conflicting info specifications for target '%s': %s",
		eic.Target, eic.err,
	)
}

type ErrorParse struct {
	Target   []string
	Position int
	err      error
}

func (ep ErrorParse) Unwrap() error { return ep.err }
func (ep ErrorParse) Error() string { return ep.err.Error() }
