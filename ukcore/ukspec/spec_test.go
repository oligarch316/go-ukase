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
	// TODO

	type MoreStuff struct {
		Blah bool `ukflag:"blah"`
	}

	type AParams struct {
		AOne string `ukflag:"aOne"`
		ATwo string `ukflag:"aTwo"`
	}

	type BParams struct {
		BOne string    `ukflag:"bOne"`
		BTwo string    `ukflag:"bTwo"`
		More MoreStuff `ukase:"inline"`
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

	t.Logf("Type: %+v\n", params.Type)
	t.Logf("Args: %+v\n", params.Args)

	for i, flag := range params.Flags {
		t.Logf("Flag (%d): %+v\n", i, flag)
	}

	for i, inline := range params.Inlines {
		t.Logf("Inline: (%d): %+v\n", i, inline)
	}
}
