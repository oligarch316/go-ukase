package ukspec

import (
	"fmt"
	"reflect"
	"strings"
)

// =============================================================================
// State
// =============================================================================

type scope struct {
	Inline
	Trail []Inline
}

type state struct {
	Config Config
	Scope  scope

	argumentList []Argument
	flagList     []Flag
	inlineList   []Inline

	flagMap    map[string]Flag
	scopeQueue []scope
}

func newState(config Config, t reflect.Type) *state {
	seed := scope{Inline: Inline{FieldType: t, FieldName: "<Root>"}}

	return &state{
		Config:     config,
		flagMap:    make(map[string]Flag),
		scopeQueue: []scope{seed},
	}
}

func (s *state) Shift() bool {
	if len(s.scopeQueue) == 0 {
		return false
	}

	s.Scope, s.scopeQueue = s.scopeQueue[0], s.scopeQueue[1:]
	return true
}

// =============================================================================
// State› Insert
// =============================================================================

func (s *state) InsertArgument(update Argument) error {
	idx := 0
	uLow, uHigh := update.Position.Low, update.Position.High

	for ; idx < len(s.argumentList); idx++ {
		original := s.argumentList[idx]
		oLow, oHigh := original.Position.Low, original.Position.High

		if uLow != nil && oHigh != nil && *oHigh <= *uLow {
			// Original precedes update (disjoint)
			continue
		}

		if oLow != nil && uHigh != nil && *uHigh <= *oLow {
			// Update precedes original (disjoint)
			break
		}

		// Original and update intersect
		err := fmt.Errorf("intersecting argument positions '%s' and '%s'", update.Position, original.Position)
		return ConflictError[Argument]{Trail: s.Scope.Trail, Original: original, Update: update, err: err}
	}

	s.argumentList = append(s.argumentList, Argument{})
	copy(s.argumentList[idx+1:], s.argumentList[idx:])
	s.argumentList[idx] = update

	return nil
}

func (s *state) InsertFlag(update Flag) error {
	for _, name := range update.Names {
		if original, exists := s.flagMap[name]; exists {
			err := fmt.Errorf("duplicated flag name '%s'", name)
			return ConflictError[Flag]{Trail: s.Scope.Trail, Original: original, Update: update, err: err}
		}

		s.flagMap[name] = update
	}

	s.flagList = append(s.flagList, update)
	return nil
}

func (s *state) InsertInline(update Inline) error {
	updateTrail := append(s.Scope.Trail, s.Scope.Inline)
	updateScope := scope{Inline: update, Trail: updateTrail}

	for i, ancestor := range updateTrail {
		if ancestor.FieldType == update.FieldType {
			var cycleNames []string
			for _, item := range append(updateTrail[i:], update) {
				cycleNames = append(cycleNames, item.FieldName)
			}

			err := fmt.Errorf("inline cycle '%s'", strings.Join(cycleNames, " → "))
			return ConflictError[Inline]{Trail: s.Scope.Trail, Original: ancestor, Update: update, err: err}
		}
	}

	s.inlineList = append(s.inlineList, update)
	s.scopeQueue = append(s.scopeQueue, updateScope)
	return nil
}
