package ukenc_test

import (
	"testing"
	"time"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/oligarch316/go-ukase/ukreflect/ukenc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO:
// Replace the use of time.Date as custom field with something simpler
// It's too tedious to read/write all these time.Date(...) calls

// TODO: Test decoding multiple of the same flag into basic (int)
// TODO: Test decoding multiple of the same flag into collection (slice)
// TODO: Test args decoding
// TODO: Test embedded struct fields
// TODO: Test inlined struct fields

func pointerTo[T any](val T) *T { return &val }

func genInput(t *testing.T, flagPairs ...string) ukcore.Input {
	nPairs := len(flagPairs)
	if nPairs%2 != 0 {
		require.Fail(t, "genInput() called with uneven number of flag pairs")
	}

	var flags []ukcore.InputFlag
	for i := 0; i < len(flagPairs); i += 2 {
		name, value := flagPairs[i], flagPairs[i+1]
		flags = append(flags, ukcore.InputFlag{Name: name, Value: value})
	}

	return ukcore.Input{Target: []string{"testTarget"}, Flags: flags}
}

func TestDecodeError(t *testing.T) {
	type subtest struct {
		name     string
		expected error
		input    ukcore.Input
		params   any
	}

	subtests := []subtest{
		// ===== Invalid parameters
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

		// ===== Invalid parameters field
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
		// TODO:
		// {
		// 	name:     "field bespoke zero value interface",
		// 	expected: new(ukenc.ErrorDecodeField),
		// 	input:    genInput(t, "flagA", "valA"),
		// 	params: new(struct {
		// 		A interface{ Bespoke() } `ukflag:"flagA"`
		// 	}),
		// },

		// ===== Invalid flag name
		{
			name:     "flag name missing",
			expected: new(ukenc.ErrorDecodeFlagName),
			input:    genInput(t, "flagA", "valA"),
			params:   new(struct{}),
		},

		// ===== Invalid flag value
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
		{
			name:     "flag value invalid TextUnmarshaler",
			expected: new(ukenc.ErrorDecodeFlagValue),
			input:    genInput(t, "flagA", "invalid"),
			params: new(struct {
				A time.Time `ukflag:"flagA"`
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

func TestDecodeDirect(t *testing.T) {
	// Decode into simple native go types

	type Params struct {
		ParamBool    bool       `ukflag:"flagBool"`
		ParamInt     int        `ukflag:"flagInt"`
		ParamUint    uint       `ukflag:"flagUint"`
		ParamFloat   float64    `ukflag:"flagFloat"`
		ParamComplex complex128 `ukflag:"flagComplex"`
		ParamString  string     `ukflag:"flagString"`
		ParamSlice   []int      `ukflag:"flagSlice"`
	}

	input := genInput(t,
		"flagBool", "true",
		"flagInt", "-42",
		"flagUint", "42",
		"flagFloat", "42.42",
		"flagComplex", "42+42i",
		"flagString", "forty-two",
		"flagSlice", "-42",
	)

	expected := Params{
		ParamBool:    true,
		ParamInt:     -42,
		ParamUint:    42,
		ParamFloat:   42.42,
		ParamComplex: complex(42, 42),
		ParamString:  "forty-two",
		ParamSlice:   []int{-42},
	}

	var actual Params

	err := ukenc.NewDecoder(input).Decode(&actual)
	require.NoError(t, err, "check Decode error")
	assert.Equal(t, expected, actual, "check result")
}

func TestDecodeCustom(t *testing.T) {
	// Decode into types that implement custom unmarshaling logic

	type Params struct {
		ParamTime time.Time `ukflag:"flagTime"`
	}

	input := genInput(t, "flagTime", "1985-04-12T23:20:50Z")

	expected := Params{
		ParamTime: time.Date(1985, 4, 12, 23, 20, 50, 0, time.UTC),
	}

	var actual Params

	err := ukenc.NewDecoder(input).Decode(&actual)
	require.NoError(t, err, "check Decode error")
	assert.Equal(t, expected, actual, "check result")
}

func TestDecodeIndirect(t *testing.T) {
	// Decode into zero value indirect types, themselves containing direct types
	// • Interfaces should be loaded with string values
	// • Pointers should have the correct element type created and loaded

	type Params struct {
		ParamAny     any         `ukflag:"flagAny"`
		ParamBool    *bool       `ukflag:"flagBool"`
		ParamInt     *int        `ukflag:"flagInt"`
		ParamUint    *uint       `ukflag:"flagUint"`
		ParamFloat   *float64    `ukflag:"flagFloat"`
		ParamComplex *complex128 `ukflag:"flagComplex"`
		ParamString  *string     `ukflag:"flagString"`
	}

	input := genInput(t,
		"flagAny", "old-thing",
		"flagBool", "true",
		"flagInt", "-42",
		"flagUint", "42",
		"flagFloat", "42.42",
		"flagComplex", "42+42i",
		"flagString", "forty-two",
	)

	expected := Params{
		ParamAny:     "old-thing",
		ParamBool:    pointerTo(true),
		ParamInt:     pointerTo(-42),
		ParamUint:    pointerTo[uint](42),
		ParamFloat:   pointerTo(42.42),
		ParamComplex: pointerTo(complex(42, 42)),
		ParamString:  pointerTo("forty-two"),
	}

	var actual Params

	err := ukenc.NewDecoder(input).Decode(&actual)
	require.NoError(t, err, "check Decode error")
	assert.Equal(t, expected, actual, "check result")
}

func TestDecodeBaroque(t *testing.T) {
	// Decode into esoteric combinations of types

	t.Run("interface->direct", func(t *testing.T) {
		type Params struct {
			ParamAny any `ukflag:"flagAny"`
		}

		var (
			input    = genInput(t, "flagAny", "true")
			expected = Params{ParamAny: true}
			actual   = Params{ParamAny: false}
		)

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("interface->custom", func(t *testing.T) {
		type Params struct {
			ParamAny any `ukflag:"flagAny"`
		}

		input := genInput(t, "flagAny", "1985-04-12T23:20:50Z")

		expected := Params{
			ParamAny: time.Date(1985, 4, 12, 23, 20, 50, 0, time.UTC),
		}

		actual := Params{
			ParamAny: time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		}

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("interface->pointer", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("interface->pointer->custom", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("pointer->custom", func(t *testing.T) {
		type Params struct {
			ParamTime *time.Time `ukflag:"flagTime"`
		}

		input := genInput(t, "flagTime", "1985-04-12T23:20:50Z")

		expected := Params{
			ParamTime: pointerTo(time.Date(1985, 4, 12, 23, 20, 50, 0, time.UTC)),
		}

		var actual Params

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("pointer->interface", func(t *testing.T) {
		type Params struct {
			ParamAny *any `ukflag:"flagAny"`
		}

		input := genInput(t, "flagAny", "old-thing")
		expected := Params{ParamAny: pointerTo[any]("old-thing")}
		actual := Params{ParamAny: pointerTo[any](nil)}

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("pointer->interface->direct", func(t *testing.T) {
		type Params struct {
			ParamAny *any `ukflag:"flagAny"`
		}

		input := genInput(t, "flagAny", "true")

		expected := Params{
			ParamAny: pointerTo[any](true),
		}

		actual := Params{
			ParamAny: pointerTo[any](false),
		}

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("pointer->interface->custom", func(t *testing.T) {
		t.Skip("TODO")
	})
}

func TestDecodeEmbedded(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type Embedded struct {
			ParamEmbedded string `ukflag:"flagEmbedded"`
		}

		type Params struct {
			Embedded      `ukase:"inline"`
			ParamStandard string `ukflag:"flagStandard"`
		}

		input := genInput(t,
			"flagEmbedded", "valEmbedded",
			"flagStandard", "valStandard",
		)

		var actual, expected Params

		expected.ParamEmbedded = "valEmbedded"
		expected.ParamStandard = "valStandard"

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("pointer zero", func(t *testing.T) {
		type Embedded struct {
			ParamEmbedded string `ukflag:"flagEmbedded"`
		}

		type Params struct {
			*Embedded     `ukase:"inline"`
			ParamStandard string `ukflag:"flagStandard"`
		}

		input := genInput(t,
			"flagEmbedded", "valEmbedded",
			"flagStandard", "valStandard",
		)

		expected := Params{
			Embedded:      &Embedded{ParamEmbedded: "valEmbedded"},
			ParamStandard: "valStandard",
		}

		actual := Params{}

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})

	t.Run("pointer informed", func(t *testing.T) {
		type Embedded struct {
			ParamEmbedded string `ukflag:"flagEmbedded"`
		}

		type Params struct {
			*Embedded     `ukase:"inline"`
			ParamStandard string `ukflag:"flagStandard"`
		}

		input := genInput(t,
			"flagEmbedded", "valEmbedded",
			"flagStandard", "valStandard",
		)

		expected := Params{
			Embedded:      &Embedded{ParamEmbedded: "valEmbedded"},
			ParamStandard: "valStandard",
		}

		actual := Params{
			Embedded:      &Embedded{ParamEmbedded: "defaultEmbedded"},
			ParamStandard: "defaultStandard",
		}

		err := ukenc.NewDecoder(input).Decode(&actual)
		require.NoError(t, err, "check Decode error")
		assert.Equal(t, expected, actual, "check result")
	})
}
