package ukspec

import (
	"errors"
	"fmt"
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

var Empty, _ = For[struct{}]()

type Params struct {
	Type    reflect.Type
	Args    *Args
	Flags   []Flag
	Inlines []Inline

	FlagIndex map[string]Flag
}

func For[T any](opts ...Option) (Params, error) {
	t := reflect.TypeFor[T]()
	return New(t, opts...)
}

func Of(v any, opts ...Option) (Params, error) {
	t := reflect.TypeOf(v)
	return New(t, opts...)
}

func New(t reflect.Type, opts ...Option) (Params, error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return Params{}, errors.New("[TODO New] not a struct or struct pointer")
	}

	params := Params{
		Type:      t,
		Args:      nil,
		Flags:     nil,
		Inlines:   nil,
		FlagIndex: make(map[string]Flag),
	}

	state := newState(t, opts)

	for state.shift() {
		if err := params.load(state); err != nil {
			return params, err
		}
	}

	return params, nil
}

func (p *Params) load(state *state) error {
	for i := 0; i < state.Head.Type.NumField(); i++ {
		field := state.Head.Type.Field(i)
		index := append(state.Head.Index, i)

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

	flag, err := newFlag(state.Config, field, names, index)
	if err != nil {
		return err
	}

	for _, name := range names {
		if _, exists := p.FlagIndex[name]; exists {
			return fmt.Errorf("[TODO loadFieldFlag] flag conflict on name '%s'", name)
		}

		p.FlagIndex[name] = flag
	}

	p.Flags = append(p.Flags, flag)
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

	p.Inlines = append(p.Inlines, inline)
	return state.push(inline)
}

// =============================================================================
// State
// =============================================================================

type queueItem struct {
	Tier  int
	Type  reflect.Type
	Index []int
}

type state struct {
	Config Config
	Seen   map[reflect.Type]int

	Head queueItem
	Tail []queueItem
}

func newState(t reflect.Type, opts []Option) *state {
	item := queueItem{Tier: 0, Type: t}

	return &state{
		Config: newConfig(opts),
		Seen:   make(map[reflect.Type]int),
		Tail:   []queueItem{item},
	}
}

func (s *state) shift() bool {
	if len(s.Tail) == 0 {
		return false
	}

	s.Seen[s.Head.Type] = s.Head.Tier
	s.Head, s.Tail = s.Tail[0], s.Tail[1:]
	return true
}

func (s *state) push(inline Inline) error {
	item := queueItem{
		Tier:  s.Head.Tier + 1,
		Type:  inline.Type,
		Index: inline.FieldIndex,
	}

	if seenTier, ok := s.Seen[item.Type]; ok && seenTier < item.Tier {
		return fmt.Errorf("[TODO push] inline type '%s' already seen (cycle)", item.Type)
	}

	s.Tail = append(s.Tail, item)
	return nil
}
