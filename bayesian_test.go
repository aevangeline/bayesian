package bayesian

import (
	"testing"

	"fmt"

	"bytes"
	"encoding/gob"

	"github.com/stretchr/testify/assert"
)

func TestBinaryClassifier(t *testing.T) {
	c, err := NewBinaryClassifier(1)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	_ = c.LearnPositive([]string{"spam", "spam", "spam", "spam",
		"spam", "spam", "ham", "apple", "cake",
		"taco", "app", "cat", "medicine",
		"medical", "dogged"})

	assert.NoError(t, err)
	_ = c.LearnNegative([]string{"ham", "ham", "ham", "ham",
		"ham", "spam", "apple", "cake", "app",
		"dog", "rat", "bat", "rake", "dogged", "bothered"})

	_, idx, _ := c.Scores([]string{"spam"})
	assert.Equal(t, Positive, idx)

	_, _, strict := c.Scores([]string{"apple"})
	assert.False(t, strict)

	_, idx, strict = c.Scores([]string{"dog"})
	assert.Equal(t, Negative, idx)
	assert.True(t, strict)

	_, _, strict = c.Scores([]string{"both"})
	assert.False(t, strict)

}

func TestEncodeDecode(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)

	c, err := NewBinaryClassifier(1)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.LearnPositive([]string{"spam", "spam",
		"spam", "spam", "spam",
		"spam", "ham", "apple", "cake", "taco",
		"app", "cat", "medicine", "medical", "dogged"})

	assert.NoError(t, err)
	err = c.LearnNegative([]string{"ham", "ham", "ham",
		"ham", "ham", "spam", "apple",
		"cake", "app", "dog", "rat", "bat",
		"rake", "dogged", "bothered"})

	assert.NoError(t, err)

	err = enc.Encode(&c)
	assert.NoError(t, err)

	var c2 BinaryClassifier
	err = dec.Decode(&c2)
	assert.NoError(t, err)

	scores, idx, _ := c2.Scores([]string{"spam"})
	assert.Equal(t, Positive, idx)
	fmt.Printf("%#v\n", scores)

}
