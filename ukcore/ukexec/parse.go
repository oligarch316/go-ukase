package ukcore

import (
	"fmt"

	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Token
// =============================================================================

type kind int

const (
	kindInvalid kind = iota
	kindEOF
	kindEmpty
	kindDelim
	kindFlag
	kindString
)

var kindToString = map[kind]string{
	kindInvalid: "INVALID",
	kindEOF:     "EOF",
	kindEmpty:   "EMPTY",
	kindDelim:   "DELIM",
	kindFlag:    "FLAG",
	kindString:  "STRING",
}

func (k kind) String() string {
	if str, ok := kindToString[k]; ok {
		return str
	}
	return fmt.Sprintf("UNKNOWN(%d)", k)
}

type token struct {
	Kind  kind
	Value string
}

func (t token) String() string { return fmt.Sprintf("❬%s❭ %s", t.Kind, t.Value) }

func newToken(str string) token {
	rs := []rune(str)

	switch n := len(rs); {
	case n == 0:
		// ❬1 Empty❭ ""
		return token{Kind: kindEmpty}
	case n == 1:
		// ❬2 Rune❭ "x" | "-"
		return token{Kind: kindString, Value: str}
	case rs[0] != '-':
		// ❬3 String❭ "xx…"
		// • ❬1,2❭ ⇒ n > 1
		return token{Kind: kindString, Value: str}
	case n == 2 && rs[1] == '-':
		// ❬4 Delim❭ "--"
		// • ❬3❭ ⇒ rs[0] == '-'
		return token{Kind: kindDelim, Value: str}
	case n == 2:
		// ❬5 Short Flag❭ "-x"
		// • ❬3❭ ⇒ rs[0] == '-'
		// • ❬4❭ ⇒ rs[1] != '-'
		return token{Kind: kindFlag, Value: string(rs[1])}
	case n > 3 && rs[1] == '-':
		// ❬6 Long Flag❭ --xx…
		// • ❬3❭ ⇒ rs[0] == '-'
		return token{Kind: kindFlag, Value: string(rs[2:])}
	default:
		// ❬7 Invalid❭ --x | -xx…
		// • ❬1,2,5❭ ⇒ n > 2
		// • ❬3❭     ⇒ rs[0] == '-'
		// • ❬6❭     ⇒ str != "--xx…"
		return token{Kind: kindInvalid}
	}
}

// =============================================================================
// Parser
// =============================================================================

const elidePlaceholder = "true"

type parser struct {
	Position int
	Values   []string
}

func newParser(values []string) *parser { return &parser{Values: values} }

func (p *parser) consume() {
	p.Values = p.Values[1:]
	p.Position += 1
}

func (p *parser) peek() (string, bool) {
	if len(p.Values) == 0 {
		return "", false
	}
	return p.Values[0], true
}

func (p *parser) ConsumeValue() (val string, exists bool) {
	if val, exists = p.peek(); exists {
		p.consume()
	}
	return
}

func (p *parser) ConsumeToken() token {
	if val, ok := p.ConsumeValue(); ok {
		return newToken(val)
	}
	return token{Kind: kindEOF}
}

func (p *parser) ConsumeFlags(specs map[string]ukspec.Flag) ([]InputFlag, error) {
	var flags []InputFlag

	for peekVal, exists := p.peek(); exists; peekVal, exists = p.peek() {
		peekToken := newToken(peekVal)

		// ❬Delim❭ or ❬String❭ ⇒ do not consume, return flags
		if peekToken.Kind == kindDelim || peekToken.Kind == kindString {
			return flags, nil
		}

		// ❬Empty❭ ⇒ consume and continue
		if peekToken.Kind == kindEmpty {
			p.consume()
			continue
		}

		// ❬Flag❭
		if peekToken.Kind == kindFlag {
			flagName := peekToken.Value
			flagSpec, flagValid := specs[flagName]

			// Invalid flag name ⇒ do not consume, fail
			if !flagValid {
				return flags, fmt.Errorf("invalid flag '%s'", flagName)
			}

			// Consume flag name
			p.consume()

			// Consume flag value
			flagVal, err := p.consumeFlagValue(flagName, flagSpec)
			if err != nil {
				return flags, err
			}

			// Append and continue
			flags = append(flags, InputFlag{Name: flagName, Value: flagVal})
			continue
		}

		// Unexpected ⇒ do not consume, fail (internal error)
		return flags, fmt.Errorf("internal ukase parse error: unexpected token kind %s", peekToken.Kind)
	}

	// ❬EOF❭
	return flags, nil
}

func (p *parser) consumeFlagValue(name string, spec ukspec.Flag) (string, error) {
	peekVal, peekExists := p.peek()
	peekUsable := peekExists && spec.Elide.Consumable(peekVal)

	// Required value is not available
	// ⇒ do not consume, fail
	if !spec.Elide.Allow && !peekExists {
		return "", fmt.Errorf("missing value for flag '%s'", name)
	}

	// Optional value is either not available or inappropriate
	// ⇒ do not consume, return a placeholder
	if spec.Elide.Allow && !peekUsable {
		return elidePlaceholder, nil
	}

	// Value is available and appropriate
	// ⇒ consume and return value
	p.consume()
	return peekVal, nil
}
