package ukenc_test

import (
	"testing"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukenc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genInput(t *testing.T, flagPairs ...string) ukcore.Input {
	nPairs := len(flagPairs)
	if nPairs%2 != 0 {
		require.Fail(t, "genInput() called with uneven number of flag pairs")
	}

	var flags []ukcore.Flag
	for i := 0; i < len(flagPairs); i += 2 {
		name, value := flagPairs[i], flagPairs[i+1]
		flags = append(flags, ukcore.Flag{Name: name, Value: value})
	}

	return ukcore.Input{Target: []string{"testTarget"}, Flags: flags}
}

func TestDecodeError(t *testing.T) {
	// TODO: ???
	// - field kind == reflect.Invalid
	// - field kind == reflect.Uintptr
	// - field kind == reflect.UnsafePointer

	type subtest struct {
		name     string
		expected error
		input    ukcore.Input
		params   any
	}

	subtests := []subtest{
		// ----- Invalid parameters
		{
			name:     "params non-pointer",
			expected: new(ukenc.ErrorDecodeParams),
			params:   struct{}{},
		},
		{
			name:     "params nil pointer",
			expected: new(ukenc.ErrorDecodeParams),
			params:   (*struct{})(nil),
		},
		{
			name:     "params non-struct",
			expected: new(ukenc.ErrorDecodeParams),
			params:   new(string),
		},

		// ----- Invalid parameters field
		{
			name:     "field array",
			expected: new(ukenc.ErrorDecodeField),
			input:    genInput(t, "flagA", "valA"),
			params: new(struct {
				A [42]int `ukflag:"flagA"`
			}),
		},
		{
			name:     "field channel",
			expected: new(ukenc.ErrorDecodeField),
			input:    genInput(t, "flagA", "valA"),
			params: new(struct {
				A chan int `ukflag:"flagA"`
			}),
		},
		{
			name:     "field function",
			expected: new(ukenc.ErrorDecodeField),
			input:    genInput(t, "flagA", "valA"),
			params: new(struct {
				A func() `ukflag:"flagA"`
			}),
		},
		{
			name:     "field map",
			expected: new(ukenc.ErrorDecodeField),
			input:    genInput(t, "flagA", "valA"),
			params: new(struct {
				A map[int]int `ukflag:"flagA"`
			}),
		},
		{
			name:     "field struct",
			expected: new(ukenc.ErrorDecodeField),
			input:    genInput(t, "flagA", "valA"),
			params: new(struct {
				A struct{} `ukflag:"flagA"`
			}),
		},

		// ----- Invalid flag
		{
			name:     "flag name missing",
			expected: new(ukenc.ErrorDecodeFlagName),
			input:    genInput(t, "flagA", "valA"),
			params:   new(struct{}),
		},
		{
			name:     "flag value invalid bool",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A bool `ukflag:"flagA"`
			}),
		},
		{
			name:     "flag value NaN int",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A int `ukflag:"flagA"`
			}),
		},
		{
			name:     "flag value NaN uint",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A uint `ukflag:"flagA"`
			}),
		},
		{
			name:     "flag value NaN float",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A float32 `ukflag:"flagA"`
			}),
		},
		{
			name:     "flag value NaN complex",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A complex64 `ukflag:"flagA"`
			}),
		},
	}

	for _, subtest := range subtests {
		st := subtest

		t.Run(st.name, func(t *testing.T) {
			err := ukenc.NewDecoder(st.input).Decode(st.params)
			require.ErrorIs(t, err, ukenc.ErrDecode)
			require.ErrorAs(t, err, st.expected)
		})
	}
}

func TestDecodeBasic(t *testing.T) {
	type Params struct {
		ParamBool    bool       `ukflag:"flagBool"`
		ParamInt     int        `ukflag:"flagInt"`
		ParamUint    uint       `ukflag:"flagUint"`
		ParamFloat   float64    `ukflag:"flagFloat"`
		ParamComplex complex128 `ukflag:"flagComplex"`
		ParamString  string     `ukflag:"flagString"`
	}

	input := genInput(t,
		"flagBool", "true",
		"flagInt", "-42",
		"flagUint", "42",
		"flagFloat", "42.42",
		"flagComplex", "42+42i",
		"flagString", "forty-two",
	)

	params := new(Params)

	err := ukenc.NewDecoder(input).Decode(params)
	require.NoError(t, err, "check Decoder error")
	assert.Equal(t, bool(true), params.ParamBool, "check ParamBool")
	assert.Equal(t, int(-42), params.ParamInt, "check ParamInt")
	assert.Equal(t, uint(42), params.ParamUint, "check ParamUint")
	assert.Equal(t, float64(42.42), params.ParamFloat, "check ParamFloat")
	assert.Equal(t, complex(42, 42), params.ParamComplex, "check ParamComplex")
	assert.Equal(t, string("forty-two"), params.ParamString, "check ParamString")
}