package bayesian

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestBinaryClassifier(t *testing.T) {
	c, err := NewBinaryClassifier(1)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.LearnPositive([]string{"spam", "spam", "spam", "spam", "spam", "spam", "ham", "apple", "cake", "taco", "app", "cat", "medicine", "medical", "dogged"})
	assert.NoError(t, err)
	err = c.LearnNegative([]string{"ham", "ham", "ham", "ham", "ham", "spam", "apple", "cake", "app", "dog", "rat", "bat", "rake", "dogged", "bothered"})
	scores, idx, _ := c.Scores([]string{"spam"})
	assert.Equal(t, Positive, idx)
	fmt.Printf("%#v\n", scores)

	scores, _, strict := c.Scores([]string{"apple"})
	assert.False(t, strict)
	fmt.Printf("%#v\n", scores)

	scores, idx, strict = c.Scores([]string{"dog"})
	assert.Equal(t, Negative, idx)
	assert.True(t, strict)
	fmt.Printf("%#v\n", scores)

	scores, _, strict = c.Scores([]string{"both"})
	assert.False(t, strict)
	fmt.Printf("%#v\n", scores)

}
