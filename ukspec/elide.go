package ukspec

import "reflect"

type elideDecideSet map[string]struct{}

func NewDecideSet(allowed ...string) func(string) bool {
	set := make(elideDecideSet)
	for _, item := range allowed {
		set[item] = struct{}{}
	}
	return set.Decide
}

func (eds elideDecideSet) Decide(text string) bool {
	_, valid := eds[text]
	return !valid
}

type ElideConfig struct {
	AllowBoolType   bool
	AllowIsBoolFlag bool
	DecideDefault   func(string) bool
}

type Elide struct {
	Allow  bool
	Decide func(string) bool
}

func newElide(config ElideConfig, field reflect.StructField) Elide {
	type decider interface{ UkaseElide(string) bool }
	type allower interface{ UkaseElide() bool }
	type isBooler interface{ IsBoolFlag() bool }

	elide := Elide{Allow: false, Decide: config.DecideDefault}
	zero := reflect.New(field.Type).Interface()

	if x, ok := zero.(decider); ok {
		elide.Allow, elide.Decide = true, x.UkaseElide
		return elide
	}

	if x, ok := zero.(allower); ok {
		elide.Allow = x.UkaseElide()
		return elide
	}

	if config.AllowIsBoolFlag {
		if x, ok := zero.(isBooler); ok {
			elide.Allow = x.IsBoolFlag()
			return elide
		}
	}

	if config.AllowBoolType {
		switch zero.(type) {
		case *bool, **bool:
			elide.Allow = true
			return elide
		}
	}

	return elide
}
