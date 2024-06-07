package ukcore

import (
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukspec"
)

var (
	ErrInternal       = errors.New("INTERNAL ukase error")
	ErrTargetNotExist = errors.New("target does not exist")
	ErrEmptyValues    = errors.New("values are empty")
)

type ErrorExecConflict struct {
	Target           InputTarget
	Original, Update ukspec.Params
	err              error
}

type ErrorInfoConflict struct {
	Target           InputTarget
	Original, Update any
	err              error
}

type ErrorFlagConflict struct {
	Target           InputTarget
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
