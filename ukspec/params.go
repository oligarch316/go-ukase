package ukspec

import (
	"errors"
	"reflect"
	"strings"
)

// =============================================================================
// Params
// =============================================================================

const (
	tagKeyFlag      = "ukflag"
	tagKeyDirective = "ukase"

	tagDirectiveArgs   = "args"
	tagDirectiveInline = "inline"
)

type Params struct {
	Type    reflect.Type
	Args    *Args
	Flags   map[string]Flag
	Inlines map[reflect.Type]Inline
}

func New(t reflect.Type, opts ...Option) (Params, error) {
	if t.Kind() != reflect.Struct {
		return Params{}, errors.New("[TODO New] not a struct")
	}

	params := Params{
		Type:    t,
		Args:    nil,
		Flags:   make(map[string]Flag),
		Inlines: make(map[reflect.Type]Inline),
	}

	config := newConfig(opts)
	seed := Inline{Type: params.Type, Inlines: params.Inlines}
	state := newState(config, seed)

	for state.shift() {
		if err := params.load(state); err != nil {
			return params, err
		}
	}

	return params, nil
}

func (p *Params) load(state *state) error {
	for i := 0; i < state.Current.Type.NumField(); i++ {
		field := state.Current.Type.Field(i)
		index := append(state.Current.FieldIndex, i)

		if err := p.loadField(state, field, index); err != nil {
			return err
		}
	}

	return nil
}

func (p *Params) loadField(state *state, field reflect.StructField, index []int) error {
	if tag, ok := field.Tag.Lookup(tagKeyFlag); ok {
		return p.loadFieldFlag(state, field, tag, index)
	}

	if tag, ok := field.Tag.Lookup(tagKeyDirective); ok {
		return p.loadFieldDirective(state, field, tag, index)
	}

	return nil
}

func (p *Params) loadFieldFlag(state *state, field reflect.StructField, tag string, index []int) error {
	names := strings.Fields(tag)

	for _, name := range names {
		flag, err := newFlag(state.Config, field, name, index)
		if err != nil {
			return err
		}

		if _, exists := p.Flags[name]; exists {
			return errors.New("[TODO loadFieldFlag] flag conflict")
		}

		p.Flags[name] = flag
	}

	return nil
}

func (p *Params) loadFieldDirective(state *state, field reflect.StructField, tag string, index []int) error {
	switch directive := strings.TrimSpace(tag); directive {
	case tagDirectiveArgs:
		return p.loadFieldArgs(field, index)
	case tagDirectiveInline:
		return p.loadFieldInline(state, field, index)
	default:
		return errors.New("[TODO loadFieldDirective] invalid directive")
	}
}

func (p *Params) loadFieldArgs(field reflect.StructField, index []int) error {
	args, err := newArgs(field, index)
	if err != nil {
		return err
	}

	if p.Args == nil {
		p.Args = &args
		return nil
	}

	// ASSUMPTION:
	// BFS implies index lengths will only ever increase
	// Fail only if there's ambiguity as to the shortest
	if len(p.Args.FieldIndex) < len(args.FieldIndex) {
		return nil
	}

	return errors.New("[TODO loadFieldArgs] args conflict")
}

func (p *Params) loadFieldInline(state *state, field reflect.StructField, index []int) error {
	inline, err := newInline(field, index)
	if err != nil {
		return err
	}

	return state.push(inline)
}

// =============================================================================
// State
// =============================================================================

type state struct {
	Config  Config
	Current Inline

	queue []Inline
	seen  map[reflect.Type]struct{}
}

func newState(config Config, seed Inline) *state {
	return &state{
		Config: config,
		queue:  []Inline{seed},
		seen:   make(map[reflect.Type]struct{}),
	}
}

func (s *state) shift() bool {
	if len(s.queue) == 0 {
		return false
	}

	s.Current, s.queue = s.queue[0], s.queue[1:]
	s.seen[s.Current.Type] = struct{}{}
	return true
}

func (s *state) push(inline Inline) error {
	if _, exists := s.seen[inline.Type]; exists {
		return errors.New("[TODO push] already seen")
	}

	s.Current.Inlines[inline.Type] = inline
	s.queue = append(s.queue, inline)
	return nil
}
