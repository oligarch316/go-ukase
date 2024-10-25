package ukspec_test

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/oligarch316/ukase/internal/itest"
	"github.com/oligarch316/ukase/ukcore/ukspec"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

// =============================================================================
// Load Parameters
// =============================================================================

// -----------------------------------------------------------------------------
// Load Parameters› Error
// -----------------------------------------------------------------------------

func runParamsError[Params any, Expected error](t *testing.T) {
	_, actual := ukspec.ParametersFor[Params]()
	assert.Check(t, itest.CmpErrorAs[Expected](actual))
}

func TestLoadParametersError(t *testing.T) {
	// --- Convenient error type shorthands
	type IPE = ukspec.InvalidParametersError
	type IFE = ukspec.InvalidFieldError
	type CEA = ukspec.ConflictError[ukspec.Argument]
	type CEF = ukspec.ConflictError[ukspec.Flag]
	type CEI = ukspec.ConflictError[ukspec.Inline]

	// --- Base parameters and all inline fields must be of type `struct` or `*struct`
	{
		type Params string
		t.Run("non-struct parameters", runParamsError[Params, IPE])
	}
	{
		type Params struct {
			InlineA string `ukinline:"lorem"`
		}
		t.Run("non-struct inline", runParamsError[Params, IFE])
	}

	// --- All tagged fields must be exported
	{
		type Params struct {
			argA string `ukarg:"0"`
		}
		t.Run("unexported argument", runParamsError[Params, IFE])
	}
	{
		type Params struct {
			flagA string `ukflag:"lorem"`
		}
		t.Run("unexported flag", runParamsError[Params, IFE])
	}
	{
		type Params struct {
			inlineA struct{} `ukinline:"lorem"`
		}
		t.Run("unexported inline", runParamsError[Params, IFE])
	}

	// --- All tag values must unmarshal successfully
	{
		type Params struct {
			ArgA string `ukarg:"lorem"`
		}
		t.Run("invalid argument tag", runParamsError[Params, IFE])
	}
	{
		type Params struct {
			FlagA string `ukflag:"-lorem"`
		}
		t.Run("invalid flag tag", runParamsError[Params, IFE])
	}
	{
		type Params struct {
			InlineA struct{} `ukinline:"-lorem"`
		}
		t.Run("invalid inline tag", runParamsError[Params, IFE])
	}

	// --- Argument positions must not conflict
	{
		type Params struct {
			ArgA string `ukarg:"5"`
			ArgB string `ukarg:"5"`
		}
		t.Run("conflicting argument position indexes", runParamsError[Params, CEA])
	}
	{
		type Params struct {
			ArgA []string `ukarg:":5"`
			ArgB []string `ukarg:"4:"`
		}
		t.Run("conflicting argument position ranges", runParamsError[Params, CEA])
	}
	{
		type Params struct {
			ArgA []string `ukarg:":"`
			ArgB string   `ukarg:"5"`
		}
		t.Run("conflicting argument position range and index", runParamsError[Params, CEA])
	}

	// --- Flag names must not conflict
	{
		type Params struct {
			FlagA string `ukflag:"lorem lorem"`
		}
		t.Run("conflicting flag names within field", runParamsError[Params, CEF])
	}
	{
		type Params struct {
			FlagA string `ukflag:"lorem"`
			FlagB string `ukflag:"lorem"`
		}
		t.Run("conflicting flag names across fields", runParamsError[Params, CEF])
	}

	// --- Inline graph must not contain cycles
	// TODO:
	// Testing deeper cycles will require mutually recursive structs defined
	// at the top level.
	{
		type Params struct {
			InlineA *Params `ukinline:"lorem"`
		}
		t.Run("conflicting inline cycle", runParamsError[Params, CEI])
	}
}

// -----------------------------------------------------------------------------
// Load Parameters› Success
// -----------------------------------------------------------------------------

type fieldKey interface {
	fmt.Stringer
	lookup(ukspec.Parameters) (string, bool)
}

type argKey int

func (ak argKey) String() string { return fmt.Sprintf("argument position '%d'", ak) }
func (k argKey) lookup(p ukspec.Parameters) (string, bool) {
	v, ok := p.LookupArgument(int(k))
	return v.FieldName, ok
}

type flagKey string

func (fk flagKey) String() string { return fmt.Sprintf("flag name '%s'", string(fk)) }
func (fk flagKey) lookup(p ukspec.Parameters) (string, bool) {
	v, ok := p.LookupFlag(string(fk))
	return v.FieldName, ok
}

func cmpParamsField(key fieldKey, params ukspec.Parameters, expected string) cmp.Comparison {
	failFormat := itest.FailureFormat{
		"",
		"unexpected value for %s",
		"  actual:   %s",
		"  expected: %s",
		"",
	}

	return func() cmp.Result {
		switch actual, exists := key.lookup(params); {
		case !exists:
			return failFormat.Result(key, "❬N/A❭", expected)
		case actual != expected:
			return failFormat.Result(key, actual, expected)
		default:
			return cmp.ResultSuccess
		}
	}
}

func checkParamsFields[Params any](t *testing.T, expectedFields map[fieldKey]string) {
	params, err := ukspec.ParametersFor[Params]()
	assert.NilError(t, err)

	for key, expected := range expectedFields {
		assert.Check(t, cmpParamsField(key, params, expected))
	}
}

func runParamsFields[Params any](expectedFields map[fieldKey]string) func(*testing.T) {
	return func(t *testing.T) { checkParamsFields[Params](t, expectedFields) }
}

func TestLoadParametersSuccess(t *testing.T) {
	// TODO:
	// Below is a basic stopgap
	// More thorough excercise of functionality is warranted

	{
		type Cone struct {
			FlagA string `ukflag:"a aa"`
			FlagB string `ukflag:"b bb"`
		}

		type Cube struct {
			FlagC string `ukflag:"c cc"`
			FlagD string `ukflag:"d dd"`
		}

		type Params struct {
			ArgAlpha string `ukarg:":3"`
			ArgBeta  string `ukarg:"3"`
			ArgGamma string `ukarg:"4:"`

			FlagLorem string `ukflag:"lorem"`
			FlagIpsum string `ukflag:"ipsum"`

			Cone Cone  `ukinline:"cone-"`
			Cube *Cube `ukinline:"cube-"`

			Other struct{} `ignored:""`
		}

		expected := map[fieldKey]string{
			argKey(0): "ArgAlpha",
			argKey(1): "ArgAlpha",
			argKey(2): "ArgAlpha",
			argKey(3): "ArgBeta",
			argKey(4): "ArgGamma",
			argKey(5): "ArgGamma",

			flagKey("lorem"):   "FlagLorem",
			flagKey("ipsum"):   "FlagIpsum",
			flagKey("cone-a"):  "FlagA",
			flagKey("cone-aa"): "FlagA",
			flagKey("cone-b"):  "FlagB",
			flagKey("cone-bb"): "FlagB",
			flagKey("cube-c"):  "FlagC",
			flagKey("cube-cc"): "FlagC",
			flagKey("cube-d"):  "FlagD",
			flagKey("cube-dd"): "FlagD",
		}

		t.Run("stuff", runParamsFields[Params](expected))
	}
}

// =============================================================================
// Unmarshal Tag
// =============================================================================

type TagUnmarshaler[T any] interface {
	encoding.TextUnmarshaler
	*T
}

func loadTag[T any, TU TagUnmarshaler[T]](input string) (T, error) {
	var tu TU = new(T)
	return *tu, tu.UnmarshalText([]byte(input))
}

func TestUnmarshalTagError(t *testing.T) {
	// TODO: Description Comment

	type subtest struct {
		name    string
		input   string
		compare func(error) cmp.Comparison
	}

	t.Run("argument position", func(t *testing.T) {
		subtests := []subtest{
			{"empty", "", itest.CmpErrorIsD},
			{"non-digit", "lorem", itest.CmpErrorIsD},
			{"multiple colons", "1:2:3", itest.CmpErrorIsD},
			{"digit below minimum", "-1", itest.CmpErrorIsD},
			{"digit above maximum", "18446744073709551616", itest.CmpErrorIsD},
			{"unbound low equals high", ":0", itest.CmpErrorIsD},
			{"explicit low equals high", "0:0", itest.CmpErrorIsD},
			{"explicit low exceeds high", "1:0", itest.CmpErrorIsD},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			_, err := loadTag[ukspec.ArgumentPosition](st.input)
			return st.name, st.compare(err)
		}

		itest.Run(t, runner, subtests...)
	})

	t.Run("flag names", func(t *testing.T) {
		subtests := []subtest{
			{"empty", "", itest.CmpErrorIsD},
			{"hyphen prefix", "-lorem", itest.CmpErrorIsD},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			_, err := loadTag[ukspec.FlagNames](st.input)
			return st.name, st.compare(err)
		}

		itest.Run(t, runner, subtests...)
	})

	t.Run("inline prefix", func(t *testing.T) {
		subtests := []subtest{
			{"hyphen prefix", "-lorem", itest.CmpErrorIsD},
			{"whitespace present", "lorem ipsum dolor", itest.CmpErrorIsD},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			_, err := loadTag[ukspec.InlinePrefix](st.input)
			return st.name, st.compare(err)
		}

		itest.Run(t, runner, subtests...)
	})
}

func TestUnmarshalTagSuccess(t *testing.T) {
	// TODO: Description Comment

	type subtest struct {
		name     string
		input    string
		expected any
	}

	t.Run("argument position", func(t *testing.T) {
		genPosition := func(low, high any) (out ukspec.ArgumentPosition) {
			if low != nil {
				tmp := uint(low.(int))
				out.Low = &tmp
			}

			if high != nil {
				tmp := uint(high.(int))
				out.High = &tmp
			}

			return
		}

		subtests := []subtest{
			{"explicit digit", "0", genPosition(0, 1)},
			{"explicit range", "0:5", genPosition(0, 5)},
			{"unbound start", ":5", genPosition(nil, 5)},
			{"unbound end", "0:", genPosition(0, nil)},
			{"unbound start and end", ":", genPosition(nil, nil)},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			actual, err := loadTag[ukspec.ArgumentPosition](st.input)
			return st.name, itest.CmpSequence(cmp.Nil(err), cmp.DeepEqual(actual, st.expected))
		}

		itest.Run(t, runner, subtests...)
	})

	t.Run("flag names", func(t *testing.T) {
		subtests := []subtest{
			{"single", "lorem", ukspec.FlagNames{"lorem"}},
			{"multiple", "lorem ipsum dolor", ukspec.FlagNames{"lorem", "ipsum", "dolor"}},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			actual, err := loadTag[ukspec.FlagNames](st.input)
			return st.name, itest.CmpSequence(cmp.Nil(err), cmp.DeepEqual(actual, st.expected))
		}

		itest.Run(t, runner, subtests...)
	})

	t.Run("inline prefix", func(t *testing.T) {
		subtests := []subtest{
			{"empty", "", ukspec.InlinePrefix("")},
			{"basic", "lorem", ukspec.InlinePrefix("lorem")},
		}

		runner := func(st subtest) (string, cmp.Comparison) {
			actual, err := loadTag[ukspec.InlinePrefix](st.input)
			return st.name, itest.CmpSequence(cmp.Nil(err), cmp.DeepEqual(actual, st.expected))
		}

		itest.Run(t, runner, subtests...)
	})
}
