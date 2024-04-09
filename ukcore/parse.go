package ukcore

import (
	"errors"
	"fmt"

	"github.com/oligarch316/go-ukase/ukspec"
)

// =============================================================================
// Token
// =============================================================================

type TokenKind int

const (
	KindInvalid TokenKind = iota
	KindEOF
	KindEmpty
	KindDelim
	KindFlag
	KindString
)

var tokenKindToString = map[TokenKind]string{
	KindInvalid: "INVALID",
	KindEOF:     "EOF",
	KindEmpty:   "EMPTY",
	KindDelim:   "DELIM",
	KindFlag:    "FLAG",
	KindString:  "STRING",
}

func (tk TokenKind) String() string {
	if str, ok := tokenKindToString[tk]; ok {
		return str
	}
	return fmt.Sprintf("UNKNOWN(%d)", tk)
}

type Token struct {
	Kind  TokenKind
	Value string
}

func (t Token) String() string { return fmt.Sprintf("<%s>%s", t.Kind, t.Value) }

func newToken(str string) Token {
	rs := []rune(str)

	switch n := len(rs); {
	case n == 0:
		// ❬1 Empty❭ ""
		return Token{Kind: KindEmpty}
	case n == 1:
		// ❬2 Rune❭ "x" | "-"
		return Token{Kind: KindString, Value: str}
	case rs[0] != '-':
		// ❬3 String❭ "xx…"
		// • ❬1,2❭ ⇒ n > 1
		return Token{Kind: KindString, Value: str}
	case n == 2 && rs[1] == '-':
		// ❬4 Delim❭ "--"
		// • ❬3❭ ⇒ rs[0] == '-'
		return Token{Kind: KindDelim, Value: str}
	case n == 2:
		// ❬5 Short Flag❭ "-x"
		// • ❬3❭ ⇒ rs[0] == '-'
		// • ❬4❭ ⇒ rs[1] != '-'
		return Token{Kind: KindFlag, Value: string(rs[1])}
	case n > 3 && rs[1] == '-':
		// ❬6 Long Flag❭ --xx…
		// • ❬3❭ ⇒ rs[0] == '-'
		return Token{Kind: KindFlag, Value: string(rs[2:])}
	default:
		// ❬7 Invalid❭ --x | -xx…
		// • ❬1,2,5❭ ⇒ n > 2
		// • ❬3❭     ⇒ rs[0] == '-'
		// • ❬6❭     ⇒ str != "--xx…"
		return Token{Kind: KindInvalid}
	}
}

// =============================================================================
// Parser
// =============================================================================

type Parser []string

func (p Parser) Flush() []string { return p }

func (p *Parser) ConsumeToken() Token {
	if len(*p) == 0 {
		return Token{Kind: KindEOF}
	}

	val := p.consumeValue()
	return newToken(val)
}

func (p *Parser) ConsumeFlags(specs map[string]ukspec.Flag) ([]Flag, error) {
	var flags []Flag

	for len(*p) > 0 {
		nextToken := newToken((*p)[0])

		if nextToken.Kind == KindDelim || nextToken.Kind == KindString {
			return flags, nil
		}

		if nextToken.Kind == KindEmpty {
			// Consume empty token
			_ = p.consumeValue()
			continue
		}

		if nextToken.Kind == KindFlag {
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

func (p *Parser) consumeFlagValue(spec ukspec.Flag) (string, error) {
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

func (p *Parser) consumeValue() (val string) {
	val, *p = (*p)[0], (*p)[1:]
	return
}
