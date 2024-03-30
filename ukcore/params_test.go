package ukcore_test

import (
	"testing"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO:
// - Success cases for embedded `ukargs`
// - Success cases for inlined `ukargs`
// - Success cases for embedded `ukflag`
// - success cases for inline `ukflag`

type customBoolFalse struct{}

func (customBoolFalse) IsBoolFlag() bool { return false }

type customBoolTrue struct{}

func (customBoolTrue) IsBoolFlag() bool            { return true }
func (customBoolTrue) CheckBoolFlag(s string) bool { return s == "custom" }

func TestParamsInfo(t *testing.T) {
	t.Run("valid kind", func(t *testing.T) {
		type Params struct{}

		info, err := ukcore.ParamsInfoOf(Params{})
		require.NoError(t, err, "check error")
		assert.Equal(t, "Params", info.TypeName, "check TypeName")
	})

	t.Run("invalid kind", func(t *testing.T) {
		type Params string

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsKind)

		_, err := ukcore.ParamsInfoOf(Params(""))
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})
}

func TestParamsInfoArgs(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		type Params struct {
			Args []string `ukargs:"args doc"`
		}

		info, err := ukcore.ParamsInfoOf(Params{})
		require.NoError(t, err, "check error")
		require.NotNil(t, info.Args, "check Args")
		assert.Equal(t, "args doc", info.Args.Doc, "check Args.Doc")
		assert.Equal(t, "Args", info.Args.FieldName, "check Args.FieldName")
	})

	t.Run("conflicting args", func(t *testing.T) {
		type Params struct {
			ParamOne []string `ukargs:"First occurence"`
			ParamTwo []string `ukargs:"Second occurence"`
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsArgsConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})

	t.Run("conflicting args embedded", func(t *testing.T) {
		type Embedded struct {
			EmbedOne []string `ukargs:"First occurence"`
		}

		type Params struct {
			Embedded
			ParamOne []string `ukargs:"Second occurence"`
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsArgsConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})

	t.Run("conflicting args inlined", func(t *testing.T) {
		t.Skip("TODO")
	})
}

func TestParamsInfoFlag(t *testing.T) {
	t.Run("valid flags", func(t *testing.T) {
		type Params struct {
			ParamOne   string
			ParamTwo   string          `ukflag:"flagTwo   - flagTwo doc"`
			ParamThree bool            `ukflag:"flagThree - flagThree doc"`
			ParamFour  customBoolFalse `ukflag:"flagFour  - flagFour doc"`
			ParamFive  customBoolTrue  `ukflag:"flagFive  - flagFive doc"`
			Skipped    string          `ukflag:"-"`
		}

		info, err := ukcore.ParamsInfoOf(Params{})
		require.NoError(t, err, "check error")
		require.Equal(t, 5, len(info.Flags), "check len(Flags)")

		lookupFlag := func(flagName string) (ukcore.ParamsInfoFlag, bool) {
			flagInfo, ok := info.Flags[flagName]
			return flagInfo, assert.Truef(t, ok, "check flag %s exists", flagName)
		}

		// Implicitly named
		if flagOne, ok := lookupFlag("paramOne"); ok {
			assert.Empty(t, flagOne.Doc, "check paramOne Doc")
			assert.Equal(t, "ParamOne", flagOne.FieldName, "check paramOne FieldName")
			assert.False(t, flagOne.IsBoolFlag(), "check paramOne IsBoolFlag()")
		}

		// Explictly named
		if flagTwo, ok := lookupFlag("flagTwo"); ok {
			assert.Equal(t, "flagTwo doc", flagTwo.Doc, "check flagTwo Doc")
			assert.Equal(t, "ParamTwo", flagTwo.FieldName, "check flagTwo FieldName")
			assert.False(t, flagTwo.IsBoolFlag(), "check flagTwo IsBoolFlag()")
		}

		// Implicitly `IsBoolFlag() == true`
		if flagThree, ok := lookupFlag("flagThree"); ok {
			assert.Equal(t, "flagThree doc", flagThree.Doc, "check flagThree Doc")
			assert.Equal(t, "ParamThree", flagThree.FieldName, "check flagThree FieldName")
			assert.True(t, flagThree.IsBoolFlag(), "check flagThree IsBoolFlag()")

			checkTrue := flagThree.CheckBoolFlag("true")
			checkUnknown := flagThree.CheckBoolFlag("unknown")
			assert.True(t, checkTrue, `check flagThree CheckBoolFlag("true")`)
			assert.False(t, checkUnknown, `check flagThree CheckBoolFlag("unknown")`)
		}

		// Custom `IsBoolFlag() == false`
		if flagFour, ok := lookupFlag("flagFour"); ok {
			assert.Equal(t, "flagFour doc", flagFour.Doc, "check flagFour Doc")
			assert.Equal(t, "ParamFour", flagFour.FieldName, "check flagFour FieldName")
			assert.False(t, flagFour.IsBoolFlag(), "check flagFour IsBoolFlag()")
		}

		// Custom `IsBoolFlag() == true` + `CheckBoolFlag(...) bool { ... }`
		if flagFive, ok := lookupFlag("flagFive"); ok {
			assert.Equal(t, "flagFive doc", flagFive.Doc, "check flagFiv Doc")
			assert.Equal(t, "ParamFive", flagFive.FieldName, "check flagFive FieldName")
			assert.True(t, flagFive.IsBoolFlag(), "check flagFive IsBoolFlag()")

			checkTrue := flagFive.CheckBoolFlag("true")
			checkCustom := flagFive.CheckBoolFlag("custom")
			assert.False(t, checkTrue, `check flagFive CheckBoolFlag("true")`)
			assert.True(t, checkCustom, `check flagFive CheckBoolFlag("custom")`)
		}
	})

	t.Run("conflicting flag explicit", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName - First occurence"`
			ParamTwo string `ukflag:"flagName - Second occurence"`
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})

	t.Run("conflicting flag implicit", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName"`
			FlagName string
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})

	t.Run("conflicting flag embedded", func(t *testing.T) {
		type Embedded struct {
			EmbedOne string `ukflag:"flagName - First occurence"`
		}

		type Params struct {
			Embedded
			ParamOne string `ukflag:"flagName - Second occurence"`
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})

	t.Run("conflicting flag inlined", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("conflicting flag internal", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName flagName - Double occurence"`
		}

		expectedIs := ukcore.ErrParams
		expectedAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, expectedIs)
		assert.ErrorAs(t, err, expectedAs)
	})
}
