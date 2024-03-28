package ukcore

import (
	"errors"
	"fmt"
)

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

type Parser struct{ values []string }

func NewParser(values []string) *Parser { return &Parser{values: values} }

func (p *Parser) Flush() (args []string) {
	args, p.values = p.values, nil
	return
}

func (p *Parser) ParseToken() Token {
	if len(p.values) == 0 {
		return Token{Kind: KindEOF}
	}

	token := newToken(p.values[0])

	p.values = p.values[1:]
	return token
}

func (p *Parser) ParseFlags(info ParamsInfo) ([]Flag, error) {
	var flags []Flag

	for len(p.values) > 0 {
		nextToken := newToken(p.values[0])

		if nextToken.Kind == KindDelim || nextToken.Kind == KindString {
			return flags, nil
		}

		if nextToken.Kind == KindEmpty {
			// Consume the empty token
			p.values = p.values[1:]
			continue
		}

		if nextToken.Kind == KindFlag {
			// Consume the flag-name
			p.values = p.values[1:]

			flagInfo, ok := info.Flags[nextToken.Value]
			if !ok {
				return flags, errors.New("[TODO ParseFlags] got unknown flag name")
			}

			flagBool := flagInfo.IsBoolFlag()
			peekEmpty := len(p.values) == 0
			peekBool := flagBool && !peekEmpty && flagInfo.CheckBoolFlag(p.values[0])

			if !flagBool && peekEmpty {
				// Required flag-value is not available
				return flags, errors.New("[TODO ParseFlags] got non-bool flag with empty peek")
			}

			if flagBool && (peekEmpty || !peekBool) {
				// Optional flag-value is either unavailable or inappropriate
				flags = append(flags, Flag{Name: nextToken.Value, Value: "true"})
				continue
			}

			// Flag-value is available and appropriate
			flags = append(flags, Flag{Name: nextToken.Value, Value: p.values[0]})

			// Consume the flag-value
			p.values = p.values[1:]
			continue
		}

		return flags, errors.New("[TODO ParseFlags] got a bad token kind")
	}

	return flags, nil
}
