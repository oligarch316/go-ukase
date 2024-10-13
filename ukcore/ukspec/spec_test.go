package ukspec_test

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/oligarch316/go-ukase/ukcore/ukspec"
	"github.com/oligarch316/go-ukase/ukerror"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

// =============================================================================
// Load Parameters
// =============================================================================

// -----------------------------------------------------------------------------
// Load Parameters› Error› Helpers
// -----------------------------------------------------------------------------

func compareParamsError[Expected error](actual error) cmp.Comparison {
	return func() cmp.Result {
		if expected := new(Expected); errors.As(actual, expected) {
			return cmp.ResultSuccess
		}

		message := fmt.Sprintf("unexpected error type '%T', expected '%s'", actual, reflect.TypeFor[Expected]())
		return cmp.ResultFailure(message)
	}
}

func runParamsError[Params any, Expected error](t *testing.T) {
	_, actual := ukspec.ParametersFor[Params]()
	assert.Check(t, compareParamsError[Expected](actual))
}

// -----------------------------------------------------------------------------
// Load Parameters› Error› Tests
// -----------------------------------------------------------------------------

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
// Load Parameters› Success› Helpers
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

func compareParamsField(params ukspec.Parameters, key fieldKey, expected string) cmp.Comparison {
	failureLines := []string{
		"",
		"unexpected value for %s",
		"  expected: %s",
		"  actual:   %s",
	}

	failure := func(actual string) cmp.Result {
		format := strings.Join(failureLines, "\n")
		message := fmt.Sprintf(format, key, expected, actual)
		return cmp.ResultFailure(message)
	}

	return func() cmp.Result {
		switch actual, exists := key.lookup(params); {
		case !exists:
			return failure("❬N/A❭")
		case actual != expected:
			return failure(actual)
		default:
			return cmp.ResultSuccess
		}
	}
}

func checkParamsFields[Params any](t *testing.T, expectedFields map[fieldKey]string) {
	params, err := ukspec.ParametersFor[Params]()
	assert.NilError(t, err)

	for key, expected := range expectedFields {
		assert.Check(t, compareParamsField(params, key, expected))
	}
}

func runParamsFields[Params any](expectedFields map[fieldKey]string) func(*testing.T) {
	return func(t *testing.T) { checkParamsFields[Params](t, expectedFields) }
}

// -----------------------------------------------------------------------------
// Load Parameters› Success› Tests
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// Unmarshal Tag› Helpers
// -----------------------------------------------------------------------------

type TagUnmarshaler[T any] interface {
	encoding.TextUnmarshaler
	*T
}

func loadTag[T any, TU TagUnmarshaler[T]](input string) (tag T, err error) {
	var tu TU = &tag
	return tag, tu.UnmarshalText([]byte(input))
}

// -----------------------------------------------------------------------------
// Unmarshal Tag› Tests
// -----------------------------------------------------------------------------

func TestUnmarshalTag(t *testing.T) {
	type subtest[ExpectedT any] struct {
		name     string
		input    string
		expected ExpectedT
	}

	t.Run("argument position", func(t *testing.T) {
		invalidSubtests := []subtest[error]{
			{"empty", "", ukerror.ErrDeveloper},
			{"non-digit", "lorem", ukerror.ErrDeveloper},
			{"multiple colons", "1:2:3", ukerror.ErrDeveloper},
			{"digit below minimum", "-1", ukerror.ErrDeveloper},
			{"digit above maximum", "18446744073709551616", ukerror.ErrDeveloper},
			{"unbound low equals high", ":0", ukerror.ErrDeveloper},
			{"explicit low equals high", "0:0", ukerror.ErrDeveloper},
			{"explicit low exceeds high", "1:0", ukerror.ErrDeveloper},
		}

		for _, st := range invalidSubtests {
			t.Run(st.name, func(t *testing.T) {
				_, actual := loadTag[ukspec.ArgumentPosition](st.input)
				assert.Check(t, cmp.ErrorIs(actual, st.expected), "input: %q", st.input)
			})
		}

		buildIndex := func(in any) (out *uint) {
			if in != nil {
				tmp := uint(in.(int))
				out = &tmp
			}
			return
		}

		buildPosition := func(low, high any) (out ukspec.ArgumentPosition) {
			out.Low, out.High = buildIndex(low), buildIndex(high)
			return
		}

		validSubtests := []subtest[ukspec.ArgumentPosition]{
			{"explicit digit", "0", buildPosition(0, 1)},
			{"explicit range", "0:5", buildPosition(0, 5)},
			{"unbound start", ":5", buildPosition(nil, 5)},
			{"unbound end", "0:", buildPosition(0, nil)},
			{"unbound start and end", ":", buildPosition(nil, nil)},
		}

		for _, st := range validSubtests {
			t.Run(st.name, func(t *testing.T) {
				actual, err := loadTag[ukspec.ArgumentPosition](st.input)
				assert.NilError(t, err, "input: %q", st.input)
				assert.Check(t, cmp.DeepEqual(actual, st.expected), "input: %q", st.input)
			})
		}
	})

	t.Run("flag names", func(t *testing.T) {
		invalidSubtests := []subtest[error]{
			{"empty", "", ukerror.ErrDeveloper},
			{"hyphen prefix", "-lorem", ukerror.ErrDeveloper},
		}

		for _, st := range invalidSubtests {
			t.Run(st.name, func(t *testing.T) {
				_, actual := loadTag[ukspec.FlagNames](st.input)
				assert.Check(t, cmp.ErrorIs(actual, st.expected), "input: %q", st.input)
			})
		}

		validSubtests := []subtest[ukspec.FlagNames]{
			{"single", "lorem", ukspec.FlagNames{"lorem"}},
			{"multiple", "lorem ipsum dolor", ukspec.FlagNames{"lorem", "ipsum", "dolor"}},
		}

		for _, st := range validSubtests {
			t.Run(st.name, func(t *testing.T) {
				actual, err := loadTag[ukspec.FlagNames](st.input)
				assert.NilError(t, err, "input: %q", st.input)
				assert.Check(t, cmp.DeepEqual(actual, st.expected), "input: %q", st.input)
			})
		}
	})

	t.Run("inline prefix", func(t *testing.T) {
		invalidSubtests := []subtest[error]{
			{"hyphen prefix", "-lorem", ukerror.ErrDeveloper},
			{"whitespace present", "lorem ipsum dolor", ukerror.ErrDeveloper},
		}

		for _, st := range invalidSubtests {
			t.Run(st.name, func(t *testing.T) {
				_, actual := loadTag[ukspec.InlinePrefix](st.input)
				assert.Check(t, cmp.ErrorIs(actual, st.expected), "input: %q", st.input)
			})
		}

		validSubtests := []subtest[ukspec.InlinePrefix]{
			{"empty", "", ""},
			{"basic", "lorem", "lorem"},
		}

		for _, st := range validSubtests {
			t.Run(st.name, func(t *testing.T) {
				actual, err := loadTag[ukspec.InlinePrefix](st.input)
				assert.NilError(t, err, "input: %q", st.input)
				assert.Check(t, cmp.DeepEqual(actual, st.expected), "input: %q", st.input)
			})
		}
	})
}
