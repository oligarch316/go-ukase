package ukcore

import (
	"errors"
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

func (t token) String() string { return fmt.Sprintf("<%s>%s", t.Kind, t.Value) }

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

type parser []string

func (p *parser) ConsumeToken() token {
	if len(*p) == 0 {
		return token{Kind: kindEOF}
	}

	val := p.consumeValue()
	return newToken(val)
}

func (p *parser) ConsumeFlags(specs map[string]ukspec.Flag) ([]Flag, error) {
	var flags []Flag

	for len(*p) > 0 {
		nextToken := newToken((*p)[0])

		if nextToken.Kind == kindDelim || nextToken.Kind == kindString {
			return flags, nil
		}

		if nextToken.Kind == kindEmpty {
			// Consume empty token
			_ = p.consumeValue()
			continue
		}

		if nextToken.Kind == kindFlag {
			// Consume the flag-name
			flagName, _ := nextToken.Value, p.consumeValue()

			flagSpec, ok := specs[flagName]
			if !ok {
				return flags, errors.New("[TODO ConsumeFlags] got an unknown flag name")
			}

			// Consume the flag-value
			flagVal, err := p.consumeFlagValue(flagSpec)
			if err != nil {
				return flags, err
			}

			flags = append(flags, Flag{Name: flagName, Value: flagVal})
			continue
		}

		return flags, errors.New("[TODO ConsumeFlags] got a bad token kind")
	}

	return flags, nil
}

func (p *parser) consumeFlagValue(spec ukspec.Flag) (string, error) {
	peekEmpty := len(*p) == 0
	peekValid := !peekEmpty && spec.Elide.Consumable((*p)[0])

	if !spec.Elide.Allow && peekEmpty {
		// Required flag-value is not available
		// ⇒ Fail
		return "", errors.New("[TODO consumeFlagValue] got a non-eliable flag with empty peek")
	}

	if spec.Elide.Allow && (peekEmpty || !peekValid) {
		// Optional flag-value is either unavailable or inappropriate
		// ⇒ Return placeholder as flag-value
		return "true", nil
	}

	// Flag value is available and appropriate
	// ⇒ Consume and return actual value as flag-value
	return p.consumeValue(), nil
}

func (p *parser) consumeValue() (val string) {
	val, *p = (*p)[0], (*p)[1:]
	return
}
