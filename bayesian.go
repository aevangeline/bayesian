// Package bayesian implements a bayesian classifier
package bayesian

import (
	"math/big"

	"github.com/LegoRemix/bayesian/internal/radix"
)

type Classifier interface {
	Scores(doc []string) ([]*big.Float, int, bool)
	Learn(doc []string, category int) error
}

type classifier struct {
	tree            radix.Tree
	smoothingFactor float64
}

// New creates a new instance of a boolean classifier
func New(categories int, smoothingFactor float64) (Classifier, error) {
	tree, err := radix.New(categories)
	if err != nil {
		return nil, err
	}

	return &classifier{tree: tree, smoothingFactor: smoothingFactor}, nil
}

func (c *classifier) getCategoryProbs(text string) []float64 {
	if c.tree.UniqueWords() == 0 {
		return make([]float64, c.tree.CategoryCount(), c.tree.CategoryCount())
	}

	uniqueWords := float64(c.tree.UniqueWords())
	counts, seen := c.tree.Find(text)
	if !seen {
		// if we have not seen this word, we try to smooth
		counts = make([]int, c.tree.CategoryCount(), c.tree.CategoryCount())
	}

	var probs []float64
	for i := range counts {
		numer := float64(counts[i]) + c.smoothingFactor
		denom := float64(c.tree.GetTotals()[i]) + c.smoothingFactor*uniqueWords
		probs = append(probs, numer/denom)
	}

	return probs
}

func (c *classifier) getPriors() []float64 {
	sum := float64(0)
	var priors []float64
	for _, value := range c.tree.GetTotals() {
		total := float64(value)
		sum += total
		priors = append(priors, total)
	}

	if sum != 0 {
		for i := range priors {
			priors[i] /= sum
		}
	}
	return priors
}

// Scores computes the probability that a given document belongs to each of the categories we are tracking
func (c *classifier) Scores(doc []string) ([]*big.Float, int, bool) {
	var scores []*big.Float
	priors := c.getPriors()
	for _, prior := range priors {
		scores = append(scores, big.NewFloat(prior))
	}

	// calculate the scores for each category
	for _, word := range doc {
		wordProbs := c.getCategoryProbs(word)
		for i, prob := range wordProbs {
			scores[i] = scores[i].Mul(scores[i], big.NewFloat(prob))
		}
	}

	sum := big.NewFloat(0.0)
	for _, score := range scores {
		sum = sum.Add(sum, score)
	}

	for i := range scores {
		scores[i] = scores[i].Quo(scores[i], sum)
	}

	idx, strict := findMax(scores)
	return scores, idx, strict
}

// findMax finds the maximum of a set of scores and determines if that maximum is the only one (i.e. strict)
func findMax(scores []*big.Float) (int, bool) {
	idx := 0
	strict := true
	for i := 1; i < len(scores); i++ {
		if scores[idx].Cmp(scores[i]) > 0 {
			idx = i
			strict = true
		} else if scores[idx].Cmp(scores[i]) == 0 {
			strict = false
		}
	}
	return idx, strict
}

// Learn learns all of the words in a given document as members of a given category
func (c *classifier) Learn(doc []string, category int) error {
	for _, fragment := range doc {
		err := c.tree.Insert(fragment, category)
		if err != nil {
			return err
		}
	}

	return nil
}
