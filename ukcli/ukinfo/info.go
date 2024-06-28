package ukinfo

import (
	"fmt"

	"github.com/oligarch316/go-ukase/ukcli"
)

type Any any

type Short string

func (s Short) Bind(target ...string) ukcli.Directive { return Bind(s, target...) }

type Description struct{ Long, Short string }

func (d Description) Bind(target ...string) ukcli.Directive { return Bind(d, target...) }

func Bind(info any, target ...string) ukcli.Directive {
	return ukcli.NewInfo(info).Bind(target...)
}

func Encode(info any) (Description, error) {
	switch infoT := info.(type) {
	case nil:
		return Description{}, nil
	case string:
		return Description{Short: infoT}, nil
	case Description:
		return infoT, nil
	default:
		return Description{}, fmt.Errorf("unknown info type '%T'", infoT)
	}
}

func Render(description Description, long bool) string {
	if long {
		return description.Long
	}
	return description.Short
}
