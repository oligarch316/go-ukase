package ukcore_test

import (
	"testing"

	"github.com/oligarch316/go-ukase/ukcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsInfo(t *testing.T) {
	t.Run("valid kind", func(t *testing.T) {
		type StructParams struct{}

		_, err := ukcore.ParamsInfoOf(StructParams{})
		assert.NoError(t, err)
	})

	// Params must be of kind struct
	t.Run("invalid kind", func(t *testing.T) {
		type StringParams string

		targetIs := ukcore.ErrParams
		targetAs := new(ukcore.ErrorParamsKind)

		_, err := ukcore.ParamsInfoOf(StringParams(""))
		require.ErrorIs(t, err, targetIs)
		assert.ErrorAs(t, err, targetAs)
	})
}

func TestParamsInfoArgs(t *testing.T) {
	t.Run("valid field", func(t *testing.T) {
		t.Skip("TODO")
	})

	// At most one field may be tagged as args
	t.Run("conflicting args", func(t *testing.T) {
		type Params struct {
			ParamOne []string `ukargs:""`
			ParamTwo []string `ukargs:""`
		}

		targetIs := ukcore.ErrParams
		targetAs := new(ukcore.ErrorParamsArgsConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, targetIs)
		assert.ErrorAs(t, err, targetAs)
	})

	t.Run("conflicting args embedded", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("conflicting args inlined", func(t *testing.T) {
		t.Skip("TODO")
	})
}

func TestParamsInfoFlag(t *testing.T) {
	t.Run("valid stuff...", func(t *testing.T) {
		t.Skip("TODO")
	})

	// Flag names must be unique across fields
	t.Run("conflicting flag explicit", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName"`
			ParamTwo string `ukflag:"flagName"`
		}

		targetIs := ukcore.ErrParams
		targetAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, targetIs)
		assert.ErrorAs(t, err, targetAs)
	})

	// Flag names must be unique across fields
	t.Run("conflicting flag implicit", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName"`
			FlagName string
		}

		targetIs := ukcore.ErrParams
		targetAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, targetIs)
		assert.ErrorAs(t, err, targetAs)
	})

	t.Run("conflicting flag embedded", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("conflicting flag inlined", func(t *testing.T) {
		t.Skip("TODO")
	})

	// Flag names must be unique within a field
	t.Run("conflicting flag internal", func(t *testing.T) {
		type Params struct {
			ParamOne string `ukflag:"flagName flagName"`
		}

		targetIs := ukcore.ErrParams
		targetAs := new(ukcore.ErrorParamsFlagConflict)

		_, err := ukcore.ParamsInfoOf(Params{})
		require.ErrorIs(t, err, targetIs)
		assert.ErrorAs(t, err, targetAs)
	})
}
