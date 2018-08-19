// Package bayesian implements a bayesian classifier
package bayesian

import (
	"math/big"

	"errors"

	"encoding/gob"

	"github.com/LegoRemix/bayesian/internal/radix"
)

type Classifier interface {
	Scores(doc []string) ([]*big.Float, int, bool)
	Learn(doc []string, category int) error
}

const Positive = 1
const Negative = 0

type BinaryClassifier interface {
	Scores(doc []string) ([]*big.Float, int, bool)
	LearnPositive(doc []string) error
	LearnNegative(doc []string) error
}

type classifier struct {
	Tree            radix.Tree
	SmoothingFactor float64
}

var ErrInvalidSmoothingFactor = errors.New("bayesian: invalid smoothing factor")

func init() {
	gob.Register(&classifier{})
}

func newClassifier(categories int, smoothingFactor float64) (*classifier, error) {
	tree, err := radix.New(categories)
	if err != nil {
		return nil, err
	}

	if smoothingFactor < 0 {
		return nil, ErrInvalidSmoothingFactor
	}

	return &classifier{Tree: tree, SmoothingFactor: smoothingFactor}, nil
}

// NewClassifier creates a new instance of a bayesian with n classes classifier
func NewClassifier(categories int, smoothingFactor float64) (Classifier, error) {
	return newClassifier(categories, smoothingFactor)
}

// NewBinaryClassifier creates a new bayesian classifier with two classes
func NewBinaryClassifier(smoothingFactor float64) (BinaryClassifier, error) {
	return newClassifier(2, smoothingFactor)
}

func (c *classifier) getCategoryProbs(text string) []float64 {
	if c.Tree.UniqueWords() == 0 {
		return make([]float64, c.Tree.CategoryCount(), c.Tree.CategoryCount())
	}

	uniqueWords := float64(c.Tree.UniqueWords())
	counts, seen := c.Tree.Find(text)
	if !seen {
		// if we have not seen this word, we try to smooth
		counts = make([]int, c.Tree.CategoryCount(), c.Tree.CategoryCount())
	}

	var probs []float64
	for i := range counts {
		numer := float64(counts[i]) + c.SmoothingFactor
		denom := float64(c.Tree.GetTotals()[i]) + c.SmoothingFactor*uniqueWords
		probs = append(probs, numer/denom)
	}

	return probs
}

func (c *classifier) getPriors() []float64 {
	sum := float64(0)
	var priors []float64
	for _, value := range c.Tree.GetTotals() {
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
			scores[i].Mul(scores[i], big.NewFloat(prob))
		}
	}

	sum := big.NewFloat(0.0)
	for _, score := range scores {
		sum.Add(sum, score)
	}

	for i := range scores {
		scores[i].Quo(scores[i], sum)
	}

	idx, strict := findMax(scores)
	return scores, idx, strict
}

// findMax finds the maximum of a set of scores and determines if that maximum is the only one (i.e. strict)
func findMax(scores []*big.Float) (int, bool) {
	idx := 0
	strict := true
	for i := 1; i < len(scores); i++ {
		if scores[idx].Cmp(scores[i]) < 0 {
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
		err := c.Tree.Insert(fragment, category)
		if err != nil {
			return err
		}
	}

	return nil
}

// LearnPositive learns something for the binaryClassifier as positive
func (c *classifier) LearnPositive(doc []string) error {
	return c.Learn(doc, Positive)
}

// LearnNegative learns something for the binaryClassifier as negative
func (c *classifier) LearnNegative(doc []string) error {
	return c.Learn(doc, Negative)
}
