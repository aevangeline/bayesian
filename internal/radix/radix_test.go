package radix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchLeaves(t *testing.T) {
	children := []radixChild{{prefix: "apple"}, {prefix: "banana"}, {prefix: "cat"}, {prefix: "sl"}}
	numChildren := len(children)
	idx, match, lcp := searchChildren(children, "x")
	assert.Equal(t, numChildren, idx)
	assert.Equal(t, nomatch, match)
	assert.Equal(t, "", lcp)
	idx, match, lcp = searchChildren(children, "app")
	assert.Equal(t, 0, idx)
	assert.Equal(t, super, match)
	assert.Equal(t, "app", lcp)
	idx, match, lcp = searchChildren(children, "slow")
	assert.Equal(t, 3, idx)
	assert.Equal(t, substring, match)
	assert.Equal(t, "sl", lcp)
	idx, match, lcp = searchChildren(children, "cab")
	assert.Equal(t, 2, idx)
	assert.Equal(t, shared_prefix, match)
	assert.Equal(t, "ca", lcp)
	idx, match, lcp = searchChildren(children, "banana")
	assert.Equal(t, 1, idx)
	assert.Equal(t, exact, match)
	assert.Equal(t, "banana", lcp)

}
