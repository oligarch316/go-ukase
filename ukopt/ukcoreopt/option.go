package ukcoreopt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oligarch316/go-ukase"
	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukspec"
)

var (
	errExecExists = errors.New("exec already exists")
	errInfoExists = errors.New("info already exists")
)

var (
	// TODO: Document
	ExecUnspecifiedFail ExecUnspecified = execUnspecifiedFail

	// TODO: Document
	ExecConflictFail ExecConflict = conflictStatic[ukspec.Params](false, errExecExists)

	// TODO: Document
	ExecConflictFirstWins ExecConflict = conflictStatic[ukspec.Params](false, nil)

	// TODO: Document
	ExecConflictLastWins ExecConflict = conflictStatic[ukspec.Params](true, nil)

	// TODO: Document
	InfoConflictFail InfoConflict = conflictStatic[any](false, errInfoExists)

	// TODO: Document
	InfoConflictFirstWins InfoConflict = conflictStatic[any](false, nil)

	// TODO: Document
	InfoConflictLastWins InfoConflict = conflictStatic[any](true, nil)

	// TODO: Document
	FlagConflictLoose FlagConflict = flagConflictLoose

	// TODO: Document
	FlagConflictStrict FlagConflict = flagConflictStrict
)

// TODO: Better name
func conflictStatic[T any](overwrite bool, err error) func(_, _ T) (bool, error) {
	return func(_, _ T) (bool, error) { return overwrite, err }
}

// TODO: Document
type ExecUnspecified ukcore.Exec

func (o ExecUnspecified) UkaseApplyCore(c *ukcore.Config) { c.ExecUnspecified = ukcore.Exec(o) }
func (o ExecUnspecified) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

func execUnspecifiedFail(_ context.Context, i ukcore.Input) error {
	return fmt.Errorf("unspecified target '%s'", strings.Join(i.Target, " "))
}

// TODO: Document
type ExecConflict func(original, update ukspec.Params) (overwrite bool, err error)

func (o ExecConflict) UkaseApplyCore(c *ukcore.Config) { c.ExecConflict = o }
func (o ExecConflict) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

// TODO: Document
type InfoConflict func(original, update any) (overwrite bool, err error)

func (o InfoConflict) UkaseApplyCore(c *ukcore.Config) { c.InfoConflict = o }
func (o InfoConflict) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

// TODO: Document
type FlagConflict func(original, update ukspec.Flag) error

func (o FlagConflict) UkaseApplyCore(c *ukcore.Config) { c.FlagConflict = o }
func (o FlagConflict) UkaseApply(c *ukase.Config)      { c.Core = append(c.Core, o) }

func flagConflictLoose(o, u ukspec.Flag) error {
	if o.Elide.Allow != u.Elide.Allow {
		return fmt.Errorf("incompatible elide behavior '%t' and '%t'", o.Elide.Allow, u.Elide.Allow)
	}

	return nil
}

func flagConflictStrict(o, u ukspec.Flag) error {
	if o.Type != u.Type {
		return fmt.Errorf("incompatible types '%T' and '%T'", o.Type, u.Type)
	}

	return nil
}
