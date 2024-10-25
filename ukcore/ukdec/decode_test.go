package ukdec_test

import (
	"math/big"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/oligarch316/ukase/internal/itest"
	"github.com/oligarch316/ukase/ukcore"
	"github.com/oligarch316/ukase/ukcore/ukdec"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

// =============================================================================
// Utilities
// =============================================================================

func ptrTo[T any](val T) *T { return &val }

func genInput(in ...string) (input ukcore.Input) {
	input.Program = "testProgram"
	input.Target = []string{"testTarget"}

	isFlag := func(s string) bool { return strings.HasPrefix(s, "--") }
	parseFlag := func(s string) string { return strings.TrimPrefix(s, "--") }

	for ; len(in) > 1 && isFlag(in[0]); in = in[2:] {
		name, value := parseFlag(in[0]), in[1]
		input.Flags = append(input.Flags, ukcore.Flag{Name: name, Value: value})
	}

	for position, value := range in {
		input.Arguments = append(input.Arguments, ukcore.Argument{Position: position, Value: value})
	}

	return
}

// =============================================================================
// Error Tests
// =============================================================================

func TestDecodeError(t *testing.T) {
	// -------------------------------------------------------------------------
	// Error Type Shorthands
	// -------------------------------------------------------------------------

	type IPE = ukdec.InvalidParametersError
	type IFEA = ukdec.InvalidFieldError[ukcore.Argument]
	type IFEF = ukdec.InvalidFieldError[ukcore.Flag]
	type UFEA = ukdec.UnknownFieldError[ukcore.Argument]
	type UFEF = ukdec.UnknownFieldError[ukcore.Flag]

	// -------------------------------------------------------------------------
	// Subtest structure
	// -------------------------------------------------------------------------

	type subtest struct {
		name    string
		input   ukcore.Input
		compare func(error) cmp.Comparison
		params  any
	}

	var runner itest.Runner[subtest] = func(st subtest) (string, cmp.Comparison) {
		dec := ukdec.NewDecoder(st.input)
		err := dec.Decode(st.params)
		return st.name, st.compare(err)
	}

	// -------------------------------------------------------------------------
	// Invalid Parameter Tests
	// -------------------------------------------------------------------------

	t.Run("invalid parameters", func(t *testing.T) {
		subtests := []subtest{
			{
				name:    "non-pointer",
				input:   genInput(),
				compare: itest.CmpErrorAsD[IPE],
				params:  struct{}{},
			},
			{
				name:    "nil pointer",
				input:   genInput(),
				compare: itest.CmpErrorAsD[IPE],
				params:  (*struct{})(nil),
			},
			{
				name:    "non-struct",
				input:   genInput(),
				compare: itest.CmpErrorAsD[IPE],
				params:  new(string),
			},
		}

		runner.Run(t, subtests...)
	})

	// -------------------------------------------------------------------------
	// Unknown Field Tests
	// -------------------------------------------------------------------------

	t.Run("unknown field", func(t *testing.T) {
		subtests := []subtest{
			{
				name:    "unknown flag",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[UFEF],
				params:  new(struct{}),
			},
			{
				name:    "unknown argument",
				input:   genInput("lorem"),
				compare: itest.CmpErrorAsU[UFEA],
				params:  new(struct{}),
			},
		}

		runner.Run(t, subtests...)
	})

	// -------------------------------------------------------------------------
	// Invalid Field Tests
	// -------------------------------------------------------------------------

	t.Run("invalid field", func(t *testing.T) {
		subtests := []subtest{
			{
				name:    "unsupported array",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem [42]int `ukflag:"lorem"`
				}),
			},
			{
				name:    "unsupported channel",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem chan int `ukflag:"lorem"`
				}),
			},
			{
				name:    "unsupported function",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem func() `ukflag:"lorem"`
				}),
			},
			{
				name:    "unsupported map",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem map[int]int `ukflag:"lorem"`
				}),
			},
			{
				name:    "unsupported struct",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem struct{} `ukflag:"lorem"`
				}),
			},
			{
				name:    "bespoke zero value interface",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsD[IFEF],
				params: new(struct {
					Lorem interface{ bespoke() } `ukflag:"lorem"`
				}),
			},
			{
				name:    "invalid bool",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem bool `ukflag:"lorem"`
				}),
			},
			{
				name:    "NaN int",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem int `ukflag:"lorem"`
				}),
			},
			{
				name:    "NaN uint",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem uint `ukflag:"lorem"`
				}),
			},
			{
				name:    "NaN float",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem float32 `ukflag:"lorem"`
				}),
			},
			{
				name:    "NaN complex",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem complex64 `ukflag:"lorem"`
				}),
			},
			{
				name:    "invalid TextUnmarshaler",
				input:   genInput("--lorem", "ipsum"),
				compare: itest.CmpErrorAsU[IFEF],
				params: new(struct {
					Lorem *big.Int `ukflag:"lorem"`
				}),
			},
		}

		runner.Run(t, subtests...)
	})
}

// =============================================================================
// Success Tests
// =============================================================================

func TestDecodeDirect(t *testing.T) {
	// Decode into direct types
	// • Scope› Direct types = { bool, numeric, string }
	// • Scope› No recursion

	// -------------------------------------------------------------------------
	// Arguments
	// -------------------------------------------------------------------------

	type ParamsArgs struct {
		ArgBool    bool       `ukarg:"0"`
		ArgInt     int        `ukarg:"1"`
		ArgUint    uint       `ukarg:"2"`
		ArgFloat   float64    `ukarg:"3"`
		ArgComplex complex128 `ukarg:"4"`
		ArgString  string     `ukarg:"5"`
	}

	inputArgs := []string{
		"true",   // (0) ArgBool
		"-42",    // (1) ArgInt
		"42",     // (2) ArgUint
		"42.42",  // (3) ArgFloat
		"42+42i", // (4) ArgComplex
		"lorem",  // (5) ArgString
	}

	expectedArgs := ParamsArgs{
		ArgBool:    true,
		ArgInt:     -42,
		ArgUint:    42,
		ArgFloat:   42.42,
		ArgComplex: complex(42, 42),
		ArgString:  "lorem",
	}

	// -------------------------------------------------------------------------
	// Flags
	// -------------------------------------------------------------------------

	type ParamsFlags struct {
		FlagBool    bool       `ukflag:"flagBool"`
		FlagInt     int        `ukflag:"flagInt"`
		FlagUint    uint       `ukflag:"flagUint"`
		FlagFloat   float64    `ukflag:"flagFloat"`
		FlagComplex complex128 `ukflag:"flagComplex"`
		FlagString  string     `ukflag:"flagString"`
	}

	inputFlags := []string{
		"--flagBool", "true",
		"--flagInt", "-42",
		"--flagUint", "42",
		"--flagFloat", "42.42",
		"--flagComplex", "42+42i",
		"--flagString", "lorem",
	}

	expectedFlags := ParamsFlags{
		FlagBool:    true,
		FlagInt:     -42,
		FlagUint:    42,
		FlagFloat:   42.42,
		FlagComplex: complex(42, 42),
		FlagString:  "lorem",
	}

	// -------------------------------------------------------------------------
	// Test
	// -------------------------------------------------------------------------

	type Params struct {
		Args  ParamsArgs  `ukinline:""`
		Flags ParamsFlags `ukinline:""`
	}

	input := genInput(append(inputFlags, inputArgs...)...)
	expected := Params{Args: expectedArgs, Flags: expectedFlags}
	actual, err := ukdec.DecodeFor[Params](input)

	assert.NilError(t, err)
	assert.DeepEqual(t, actual, expected)
}

func TestDecodeIndirect(t *testing.T) {
	// Decode into indirect types
	// • Scope› Indirect types = { interface, pointer }
	// • Scope› Indirect<Direct> types, 1 level of recursion
	// • Scope› Input fields are uninformed (zero-value)
	//
	// • Expect› Interface fields are loaded with string values
	// • Expect› Pointer fields have correct element type created and loaded

	type Params struct {
		FlagAny     any         `ukflag:"flagAny"`
		FlagBool    *bool       `ukflag:"flagBool"`
		FlagInt     *int        `ukflag:"flagInt"`
		FlagUint    *uint       `ukflag:"flagUint"`
		FlagFloat   *float64    `ukflag:"flagFloat"`
		FlagComplex *complex128 `ukflag:"flagComplex"`
		FlagString  *string     `ukflag:"flagString"`
	}

	input := genInput(
		"--flagAny", "lorem",
		"--flagBool", "true",
		"--flagInt", "-42",
		"--flagUint", "42",
		"--flagFloat", "42.42",
		"--flagComplex", "42+42i",
		"--flagString", "lorem",
	)

	expected := Params{
		FlagAny:     "lorem",
		FlagBool:    ptrTo(true),
		FlagInt:     ptrTo(-42),
		FlagUint:    ptrTo[uint](42),
		FlagFloat:   ptrTo(42.42),
		FlagComplex: ptrTo(42 + 42i),
		FlagString:  ptrTo("lorem"),
	}

	actual, err := ukdec.DecodeFor[Params](input)

	assert.NilError(t, err)
	assert.DeepEqual(t, actual, expected)
}

func TestDecodeContainer(t *testing.T) {
	// Decode into container types
	// • Scope› Container types = { slice }
	// • Scope› Container<Direct> types, 1 level of recursion
	// • Scope› Input fields are uninformed (zero-value)
	//
	// • Expect› Slice fields have correct element types created and loaded

	type Params struct {
		Lorem []int `ukflag:"lorem"`
	}

	input := genInput("--lorem", "-42")
	expected := Params{Lorem: []int{-42}}
	actual, err := ukdec.DecodeFor[Params](input)

	assert.NilError(t, err)
	assert.DeepEqual(t, actual, expected)
}

func TestDecodeCustom(t *testing.T) {
	// Decode into custom types
	// • Scope› Custom types = { encoding.TextUnmarshaler }

	type Params struct {
		Lorem *big.Int `ukflag:"lorem"`
	}

	input := genInput("--lorem", "42")
	expected := Params{Lorem: big.NewInt(42)}
	actual, err := ukdec.DecodeFor[Params](input)

	equalBigInt := func(a, b *big.Int) bool { return a.Cmp(b) == 0 }
	equalOpt := gocmp.Comparer(equalBigInt)

	assert.NilError(t, err)
	assert.DeepEqual(t, actual, expected, equalOpt)
}

func TestDecodeBaroque(t *testing.T) {
	// Decode into complex/esoteric types
	// • Scope› Any level of recursion

	t.Run("interface to direct", func(t *testing.T) {
		type Params struct {
			Lorem any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "true")
		expected := Params{Lorem: true}
		actual := Params{Lorem: false}

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)

		// TODO: Also test...
		// input := <ditto>
		// actual := Params{Lorem: 42}
		// check CmpErrorAsU[ukdoc.InvalidFieldError]
	})

	t.Run("interface to pointer", func(t *testing.T) {
		type Params struct {
			Lorem any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "true")
		expected := Params{Lorem: ptrTo(true)}
		actual := Params{Lorem: ptrTo(false)}

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)

		// TODO: Also test...
		// input := <ditto>
		// actual := Params{Lorem: new(int)}
		// check CmpErrorAsU[ukdoc.InvalidFieldError]
	})

	t.Run("interface to custom", func(t *testing.T) {
		type Params struct {
			Lorem any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "42")
		expected := Params{Lorem: big.NewInt(42)}
		actual := Params{Lorem: big.NewInt(0)}

		equalBigInt := func(a, b *big.Int) bool { return a.Cmp(b) == 0 }
		equalOpt := gocmp.Comparer(equalBigInt)

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected, equalOpt)
	})

	t.Run("pointer to interface", func(t *testing.T) {
		type Params struct {
			Lorem *any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "ipsum")
		expected := Params{Lorem: ptrTo[any]("ipsum")}
		actual := Params{Lorem: ptrTo[any](nil)}

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("pointer to interface to direct", func(t *testing.T) {
		type Params struct {
			Lorem *any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "true")
		expected := Params{Lorem: ptrTo[any](true)}
		actual := Params{Lorem: ptrTo[any](false)}

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("pointer to interface to custom", func(t *testing.T) {
		type Params struct {
			Lorem *any `ukflag:"lorem"`
		}

		input := genInput("--lorem", "42")
		expected := Params{Lorem: ptrTo[any](big.NewInt(42))}
		actual := Params{Lorem: ptrTo[any](big.NewInt(0))}

		equalBigInt := func(a, b *big.Int) bool { return a.Cmp(b) == 0 }
		equalOpt := gocmp.Comparer(equalBigInt)

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected, equalOpt)
	})
}

func TestDecodeEmbedded(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type Embedded struct {
			Lorem string `ukflag:"lorem"`
		}

		type Params struct {
			Embedded `ukinline:""`
			Ipsum    string `ukflag:"ipsum"`
		}

		input := genInput(
			"--lorem", "lorem-val",
			"--ipsum", "ipsum-val",
		)

		var expected, actual Params

		expected.Lorem = "lorem-val"
		expected.Ipsum = "ipsum-val"

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("pointer with zero value", func(t *testing.T) {
		type Embedded struct {
			Lorem string `ukflag:"lorem"`
		}

		type Params struct {
			*Embedded `ukinline:""`
			Ipsum     string `ukflag:"ipsum"`
		}

		input := genInput(
			"--lorem", "lorem-val",
			"--ipsum", "ipsum-val",
		)

		var expected, actual Params

		expected.Embedded = new(Embedded)
		expected.Lorem = "lorem-val"
		expected.Ipsum = "ipsum-val"

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("pointer with non-zero value", func(t *testing.T) {
		type Embedded struct {
			Lorem string `ukflag:"lorem"`
		}

		type Params struct {
			*Embedded `ukinline:""`
			Ipsum     string `ukflag:"ipsum"`
		}

		input := genInput(
			"--lorem", "lorem-val",
			"--ipsum", "ipsum-val",
		)

		var expected, actual Params

		expected.Embedded = new(Embedded)
		expected.Lorem = "lorem-val"
		expected.Ipsum = "ipsum-val"

		actual.Embedded = new(Embedded)
		actual.Lorem = "existing-val"
		actual.Ipsum = "existing-val"

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("prefixed", func(t *testing.T) {
		type Embedded struct {
			InnerLorem string `ukflag:"lorem"`
			InnerIpsum string `ukflag:"ipsum"`
		}

		type Params struct {
			Embedded `ukinline:"inner-"`
			Lorem    string `ukflag:"lorem"`
			Ipsum    string `ukflag:"ipsum"`
		}

		input := genInput(
			"--inner-lorem", "inner-lorem-val",
			"--inner-ipsum", "inner-ipsum-val",
			"--lorem", "lorem-val",
			"--ipsum", "ipsum-val",
		)

		var expected, actual Params

		expected.InnerLorem = "inner-lorem-val"
		expected.InnerIpsum = "inner-ipsum-val"
		expected.Lorem = "lorem-val"
		expected.Ipsum = "ipsum-val"

		assert.NilError(t, ukdec.Decode(input, &actual))
		assert.DeepEqual(t, actual, expected)
	})
}
