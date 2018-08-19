package radix

import (
	"testing"

	"math/rand"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchLeaves(t *testing.T) {
	children := []child{{Prefix: "apple"}, {Prefix: "banana"}, {Prefix: "cat"}, {Prefix: "sl"}}
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

var letterRunes = []rune("abcdefg.!*åßçêïł ")

func randString() string {
	n := rand.Intn(15) + 1
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// truncate trims a string to the given character limit.
func truncate(s string, maxChars int) string {
	runes := []rune(s)
	if len(runes) > maxChars {
		return string(runes[:maxChars])
	}
	return s
}

const iterations = 10000

func TestInsertAndFetch(t *testing.T) {

	tree, err := New(1)
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	rand.Seed(time.Now().Unix())
	words := make(map[string]struct{})
	// Make sure our insertion always works
	for i := 0; i < iterations; i++ {
		word := randString()
		if _, ok := words[word]; !ok {
			words[word] = struct{}{}
			err := tree.Insert(word, 0)
			assert.NoError(t, err)

		} else {
			_, found := tree.Find(word)
			assert.True(t, found)
			words[word] = struct{}{}
		}
	}

	// Make sure we can recover all strings in our dictionary
	for word := range words {
		_, found := tree.Find(word)
		assert.True(t, found, "Cannot find %s", word)
	}

	// test that we never find a short version of a string that is in the dictionary
	for word := range words {
		short := truncate(word, rand.Intn(len(word)))
		if _, ok := words[short]; !ok {
			_, found := tree.Find(short)
			assert.False(t, found)
		}
	}

	// Test that we never find a string not in our dictionary
	for i := 0; i < iterations; i++ {
		word := randString()
		if _, ok := words[word]; !ok {
			_, found := tree.Find(word)
			assert.False(t, found)

		}
	}
}

func BenchmarkInsert(b *testing.B) {
	b.ReportAllocs()
	tree, err := New(1)
	assert.NoError(b, err)
	assert.NotNil(b, tree)
	for i := 0; i < b.N; i++ {
		_ = tree.Insert(randString(), 0)
	}
}

func BenchmarkInsertAndFind(b *testing.B) {
	b.ReportAllocs()
	tree, err := New(1)
	assert.NoError(b, err)
	assert.NotNil(b, tree)
	for i := 0; i < b.N; i++ {
		s := randString()
		_ = tree.Insert(s, 0)
		_, _ = tree.Find(s)

	}
}

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()
	words := make(map[string]node)
	for i := 0; i < b.N; i++ {
		words[randString()] = node{Values: make([]int, 1, 1)}
	}
}
