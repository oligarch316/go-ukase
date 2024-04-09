package ukspec_test

import (
	"testing"

	"github.com/oligarch316/go-ukase/ukspec"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	t.Skip("TODO")
}

func TestBasic(t *testing.T) {
	type AParams struct {
		AOne string `ukflag:"aOne"`
		ATwo string `ukflag:"aTwo"`
	}

	type BParams struct {
		BOne string `ukflag:"bOne"`
		BTwo string `ukflag:"bTwo"`
	}

	type Params struct {
		A AParams `ukase:"inline"`
		B BParams `ukase:"inline"`

		One string `ukflag:"one"`
		Two string `ukflag:"two"`

		Args []string `ukase:"args"`
	}

	params, err := ukspec.For[Params]()
	require.NoError(t, err, "check Create error")

	t.Logf("TODO: %+v\n", params)
}
