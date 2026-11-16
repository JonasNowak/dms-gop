package gops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetaModules(t *testing.T) {
	g := NewGopsUtilWithProviders(nil, nil, nil, nil, nil, nil, nil, nil)
	mods, err := g.GetModules()
	require.NoError(t, err)
	assert.Subset(t, mods.Available, []string{"sensors", "process-io", "process-net", "process-members"})

	for _, bad := range [][]string{{"nope"}, {"process-io", "nope"}} {
		_, err := g.GetMeta(context.Background(), bad, MetaParams{})
		assert.Error(t, err, "%v", bad)
	}
}

func TestMetaParamsProcessOptions(t *testing.T) {
	opts, modifier := MetaParams{}.processOptions([]string{"process-net"})
	assert.True(t, opts.Net && modifier)
	assert.False(t, opts.IO)
	opts, _ = MetaParams{}.processOptions([]string{"processes"})
	assert.False(t, opts.Net)

	// Omitted ids stay nil and an empty list stays non-nil: neither reads a GPU.
	opts, _ = MetaParams{}.processOptions(nil)
	assert.False(t, opts.IO || opts.Net)
}
